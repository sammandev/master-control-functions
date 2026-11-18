//go:build sqlite_migration

package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"

	"webprogfndb/internal/db"
)

type table struct {
	name string
	key  string
}

var tables = []table{
	{"functions", "id"},
	{"parameter_definitions", "id"},
	{"criteria_definitions", "id"},
	{"parameter_links", "id"},
	{"criteria_links", "id"},
	{"examples", "id"},
	{"import_runs", "id"},
	{"import_normalizations", "id"},
	{"sequences", "id"},
	{"sequence_steps", "id"},
	{"devices", "name"},
}

type report struct {
	Count  int64
	SHA256 string
}

func Migrate(ctx context.Context, source, target *sql.DB, verifyOnly bool) error {
	snapshot, err := source.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer snapshot.Rollback()
	if err := validateSource(ctx, snapshot); err != nil {
		return err
	}
	if !verifyOnly {
		if err := db.MigratePostgres(target); err != nil {
			return err
		}
	}
	var encoding string
	if err := target.QueryRowContext(ctx, "SHOW server_encoding").Scan(&encoding); err != nil {
		return err
	}
	if encoding != "UTF8" {
		return fmt.Errorf("target must use UTF8 encoding for byte-identical text verification")
	}
	version, err := db.CurrentVersion(target)
	if err != nil {
		return err
	}
	if version != 6 {
		return fmt.Errorf("target schema must be version 6, got %d", version)
	}
	options := &sql.TxOptions{Isolation: sql.LevelSerializable}
	if verifyOnly {
		options = &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}
	}
	transaction, err := target.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if !verifyOnly {
		identifiers := make([]string, len(tables))
		for index, item := range tables {
			identifiers[index] = pgx.Identifier{item.name}.Sanitize()
		}
		if _, err := transaction.ExecContext(ctx, "LOCK TABLE "+strings.Join(identifiers, ", ")+" IN ACCESS EXCLUSIVE MODE"); err != nil {
			return err
		}
		for _, item := range tables {
			var count int64
			if err := transaction.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+pgx.Identifier{item.name}.Sanitize()).Scan(&count); err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("target table %s is not empty; use --verify-only after a successful transfer", item.name)
			}
		}
	}
	for _, item := range tables {
		if err := transferTable(ctx, snapshot, transaction, item, verifyOnly); err != nil {
			return fmt.Errorf("table %s: %w", item.name, err)
		}
	}
	if !verifyOnly {
		for _, item := range tables {
			if item.key != "id" {
				continue
			}
			identifier := pgx.Identifier{item.name}.Sanitize()
			var maximum int64
			if err := transaction.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM "+identifier).Scan(&maximum); err != nil {
				return err
			}
			if maximum == math.MaxInt64 {
				return fmt.Errorf("table %s has exhausted its identity range", item.name)
			}
			next := max(maximum+1, int64(1))
			if _, err := transaction.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN id RESTART WITH %d", identifier, next)); err != nil {
				return err
			}
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit status may be uncertain; rerun --verify-only before retrying: %w", err)
	}
	if verifyOnly {
		fmt.Println("Verification passed: all 11 application tables match the SQLite snapshot.")
	} else {
		fmt.Println("Migration committed: all 11 application tables verified; SQLite source unchanged.")
	}
	return nil
}

func OpenSource(path string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source must be an existing regular SQLite file")
	}
	connectionURL := &url.URL{Scheme: "file", Path: filepath.ToSlash(absolute)}
	options := url.Values{"mode": {"ro"}, "_pragma": {"query_only(ON)"}}
	connectionURL.RawQuery = options.Encode()
	source, err := sql.Open("sqlite", connectionURL.String())
	if err != nil {
		return nil, err
	}
	source.SetMaxOpenConns(1)
	return source, nil
}

func validateSource(ctx context.Context, source *sql.Tx) error {
	var encoding string
	if err := source.QueryRowContext(ctx, "PRAGMA encoding").Scan(&encoding); err != nil {
		return err
	}
	if encoding != "UTF-8" {
		return fmt.Errorf("source must use UTF-8 encoding for byte-identical text verification")
	}
	var integrity string
	if err := source.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&integrity); err != nil {
		return err
	}
	if integrity != "ok" {
		return fmt.Errorf("SQLite integrity check failed")
	}
	var version int
	if err := source.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version); err != nil {
		return err
	}
	if version != 6 {
		return fmt.Errorf("source schema must be version 6; got %d (upgrade with the old server before backing up)", version)
	}
	foreignKeys, err := source.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer foreignKeys.Close()
	if foreignKeys.Next() {
		return fmt.Errorf("source has a broken sequence/import foreign key; orphan-capable links are not foreign keys")
	}
	if err := foreignKeys.Err(); err != nil {
		return err
	}
	if err := foreignKeys.Close(); err != nil {
		return err
	}
	known := map[string]bool{"schema_migrations": true}
	for _, item := range tables {
		known[item.name] = true
	}
	rows, err := source.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		if !known[name] {
			return fmt.Errorf("unrecognized source table %s; refusing to silently omit it", name)
		}
	}
	return rows.Err()
}

func transferTable(ctx context.Context, source, target *sql.Tx, item table, verifyOnly bool) error {
	identifier := pgx.Identifier{item.name}.Sanitize()
	sourceOrder := pgx.Identifier{item.key}.Sanitize()
	targetOrder := sourceOrder
	if item.key == "name" {
		sourceOrder += " COLLATE BINARY"
		targetOrder += ` COLLATE "C"`
	}
	rows, err := source.QueryContext(ctx, "SELECT * FROM "+identifier+" ORDER BY "+sourceOrder)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	quoted := make([]string, len(columns))
	parameters := make([]string, len(columns))
	for index, column := range columns {
		quoted[index] = pgx.Identifier{column}.Sanitize()
		parameters[index] = fmt.Sprintf("$%d", index+1)
	}
	columnList := strings.Join(quoted, ", ")
	var insert *sql.Stmt
	if !verifyOnly {
		insert, err = target.PrepareContext(ctx, "INSERT INTO "+identifier+" ("+columnList+") VALUES ("+strings.Join(parameters, ", ")+")")
		if err != nil {
			return err
		}
		defer insert.Close()
	}
	sourceReport, err := fingerprint(rows, func(values []any) error {
		if insert != nil {
			_, err := insert.ExecContext(ctx, values...)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	targetRows, err := target.QueryContext(ctx, "SELECT "+columnList+" FROM "+identifier+" ORDER BY "+targetOrder)
	if err != nil {
		return err
	}
	defer targetRows.Close()
	targetReport, err := fingerprint(targetRows, nil)
	if err != nil {
		return err
	}
	if sourceReport != targetReport {
		return fmt.Errorf("row count or SHA-256 content mismatch; transfer will not commit")
	}
	fmt.Printf("verified %-24s rows=%d sha256=%s\n", item.name, sourceReport.Count, sourceReport.SHA256)
	return nil
}

func fingerprint(rows *sql.Rows, consume func([]any) error) (report, error) {
	columns, err := rows.Columns()
	if err != nil {
		return report{}, err
	}
	digest := sha256.New()
	encoder := json.NewEncoder(digest)
	if err := encoder.Encode(columns); err != nil {
		return report{}, err
	}
	var count int64
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return report{}, err
		}
		for index, value := range values {
			if bytes, ok := value.([]byte); ok {
				values[index] = string(bytes)
			}
		}
		if consume != nil {
			if err := consume(values); err != nil {
				return report{}, err
			}
		}
		if err := encoder.Encode(values); err != nil {
			return report{}, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return report{}, err
	}
	return report{Count: count, SHA256: hex.EncodeToString(digest.Sum(nil))}, nil
}
