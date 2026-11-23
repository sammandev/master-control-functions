//go:build sqlite_migration

package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

const LegacyTransferVersion = 6

func MigratePostgresTransferBaseline(database *sql.DB) error {
	return migratePostgres(database, LegacyTransferVersion)
}

func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(ON)"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	return database, nil
}

func MigrateSQLiteTransferBaseline(database *sql.DB) error {
	if _, err := database.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return err
	}
	var current int
	if err := database.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current); err != nil {
		return err
	}
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if migration.Version > LegacyTransferVersion {
			break
		}
		if migration.Version <= current {
			continue
		}
		transaction, err := database.Begin()
		if err != nil {
			return err
		}
		if _, err := transaction.Exec(migration.SQL); err != nil {
			transaction.Rollback()
			return err
		}
		if _, err := transaction.Exec(
			`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
			migration.Version, migration.Name, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			transaction.Rollback()
			return err
		}
		if err := transaction.Commit(); err != nil {
			return err
		}
	}
	return nil
}
