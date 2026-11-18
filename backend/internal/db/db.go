// Package db owns PostgreSQL connectivity and schema migrations.
// Legacy SQLite connectivity is retained for off-machine transfer fixtures.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func OpenPostgres(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	connectionURL, err := url.Parse(dsn)
	if err != nil || (connectionURL.Scheme != "postgres" && connectionURL.Scheme != "postgresql") {
		return nil, fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	options := connectionURL.Query()
	switch options.Get("sslmode") {
	case "", "require", "verify-full":
		options.Set("sslmode", "verify-full")
	case "disable":
	default:
		return nil, fmt.Errorf("DATABASE_URL sslmode must be verify-full, require, or explicitly disable")
	}
	connectionURL.RawQuery = options.Encode()
	config, err := pgx.ParseConfig(connectionURL.String())
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL")
	}
	database := stdlib.OpenDB(*config)
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return database, nil
}

func MigratePostgres(database *sql.DB) error {
	migrations, err := loadMigrationsFrom("postgres_migrations")
	if err != nil {
		return err
	}
	transaction, err := database.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()
	if _, err := transaction.Exec(`SELECT pg_advisory_xact_lock(743002019)`); err != nil {
		return err
	}
	if _, err := transaction.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return err
	}
	var current int
	if err := transaction.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current); err != nil {
		return err
	}
	if len(migrations) == 0 || current > migrations[len(migrations)-1].Version {
		return fmt.Errorf("unsupported PostgreSQL schema version %d", current)
	}
	for _, migration := range migrations {
		if migration.Version <= current {
			continue
		}
		if _, err := transaction.Exec(migration.SQL); err != nil {
			return fmt.Errorf("apply PostgreSQL migration %d: %w", migration.Version, err)
		}
		if _, err := transaction.Exec(
			`INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)`,
			migration.Version, migration.Name, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return err
		}
	}
	return transaction.Commit()
}

// CurrentVersion reports the highest applied migration version.
func CurrentVersion(d *sql.DB) (int, error) {
	var v int
	err := d.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}
	return v, nil
}
