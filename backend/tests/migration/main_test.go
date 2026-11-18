//go:build sqlite_migration

package migration_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"webprogfndb/internal/db"
	"webprogfndb/internal/migration"
	"webprogfndb/tests/support/testdb"
)

func sourceFixture(t *testing.T) (string, *sql.DB) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.db")
	writable, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(writable); err != nil {
		writable.Close()
		t.Fatal(err)
	}
	_, err = writable.Exec(`
		INSERT INTO functions (id, source, interface, name, convention, created_at, updated_at)
		VALUES (41, 'DLL', 'WIFI_TEST', 'EXISTING', 2, 't', 't');
		INSERT INTO parameter_definitions (id, name, device_name, required, created_at, updated_at)
		VALUES (51, 'P', 'WIFI_TEST', 1, 't', 't');
		INSERT INTO criteria_definitions (id, name, default_value, created_at, updated_at)
		VALUES (52, 'C', '', 't', 't');
		INSERT INTO parameter_links (id, function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES (61, 'DLL', 'ORPHAN', 'WIFI_TEST', 'P', 1, 't', 't');
		INSERT INTO criteria_links (id, function_source, function_name, name, position, created_at, updated_at)
		VALUES (62, 'DLL', 'ORPHAN', 'C', 2, 't', 't');
		INSERT INTO examples (id, function_source, function_name, script, created_at, updated_at)
		VALUES (71, 'DLL', 'ORPHAN', 'raw|example', 't', 't');
		INSERT INTO import_runs (id, file_name, reset, started_at, finished_at)
		VALUES (81, 'fixture.xlsx', 1, 't', 't');
		INSERT INTO import_normalizations (id, run_id, table_name, row_key, column_name, source_value, normalized_value, rule)
		VALUES (82, 81, 'functions', 'key', 'name', 'a', 'A', 'upper');
		INSERT INTO sequences (id, name, created_at, updated_at) VALUES (91, 'Sequence', 't', 't');
		INSERT INTO sequence_steps (id, sequence_id, position, function_source, function_interface, function_name)
		VALUES (92, 91, 1, 'DLL', 'WIFI_TEST', 'ORPHAN');
		INSERT INTO devices (name, created_at, updated_at) VALUES ('DEVICE', 't', 't');
	`)
	if err == nil {
		_, err = writable.Exec("UPDATE examples SET meaning = ? WHERE id = 71", "caf\u00e9")
	}
	if err == nil {
		for _, name := range []string{"Z", "caf\u00e9", "\u4e2d"} {
			if _, err = writable.Exec("INSERT INTO devices (name, created_at, updated_at) VALUES (?, 't', 't')", name); err != nil {
				break
			}
		}
	}
	closeErr := writable.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	readonly, err := migration.OpenSource(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { readonly.Close() })
	return path, readonly
}

func TestTransferPreservesAllRowsAndResetsIdentities(t *testing.T) {
	target, err := testdb.Open(t)
	if err != nil {
		t.Fatal(err)
	}
	path, source := sourceFixture(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := migration.Migrate(ctx, source, target, false); err != nil {
		t.Fatal(err)
	}
	if err := migration.Migrate(ctx, source, target, true); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("read-only transfer modified the source file")
	}
	var next int64
	if err := target.QueryRow(`INSERT INTO functions (source, interface, name, created_at, updated_at)
		VALUES ('DLL', 'WIFI_TEST', 'NEXT', 't', 't') RETURNING id`).Scan(&next); err != nil {
		t.Fatal(err)
	}
	if next != 42 {
		t.Fatalf("next identity = %d, want 42", next)
	}
	var orphan string
	var position sql.NullInt64
	if err := target.QueryRow("SELECT function_name, position FROM parameter_links WHERE id = 61").Scan(&orphan, &position); err != nil {
		t.Fatal(err)
	}
	if orphan != "ORPHAN" || position.Valid {
		t.Fatal("orphan link or nullable position changed")
	}
}

func TestTransferFailureRollsBackAndCanBeRetried(t *testing.T) {
	target, err := testdb.Open(t)
	if err != nil {
		t.Fatal(err)
	}
	path, source := sourceFixture(t)
	writable, err := db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Exec("UPDATE parameter_definitions SET description = ?", "invalid\x00text"); err != nil {
		writable.Close()
		t.Fatal(err)
	}
	writable.Close()
	ctx := context.Background()
	if err := migration.Migrate(ctx, source, target, false); err == nil {
		t.Fatal("PostgreSQL must reject NUL text after the earlier table was copied")
	}
	var count int64
	if err := target.QueryRow("SELECT COUNT(*) FROM functions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partially copied rows survived rollback")
	}
	writable, err = db.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Exec("UPDATE parameter_definitions SET description = 'fixed'"); err != nil {
		writable.Close()
		t.Fatal(err)
	}
	writable.Close()
	if err := migration.Migrate(ctx, source, target, false); err != nil {
		t.Fatalf("retry after rollback: %v", err)
	}
}

func TestTransferRefusesNonemptyTarget(t *testing.T) {
	target, err := testdb.Open(t)
	if err != nil {
		t.Fatal(err)
	}
	_, source := sourceFixture(t)
	if err := db.MigratePostgres(target); err != nil {
		t.Fatal(err)
	}
	if _, err := target.Exec(`INSERT INTO functions (id, source, interface, name, created_at, updated_at)
		VALUES (900, 'DLL', 'WIFI_TEST', 'KEEP', 't', 't')`); err != nil {
		t.Fatal(err)
	}
	if err := migration.Migrate(context.Background(), source, target, false); err == nil {
		t.Fatal("nonempty target must be rejected")
	}
	var name string
	if err := target.QueryRow("SELECT name FROM functions WHERE id = 900").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "KEEP" {
		t.Fatal("existing target row was changed")
	}
}
