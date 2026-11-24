package db_test

import (
	"database/sql"
	"strings"
	"testing"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

func TestPostgresRejectsPlaintextFallbackWithoutEchoingCredentials(t *testing.T) {
	for _, mode := range []string{"allow", "prefer", "unexpected"} {
		connection, err := db.OpenPostgres("postgresql://user:do-not-echo@127.0.0.1:1/catalog?sslmode=" + mode)
		if err == nil || connection != nil {
			t.Fatalf("TLS mode %s must be rejected before connecting", mode)
		}
		if strings.Contains(err.Error(), "do-not-echo") {
			t.Fatal("configuration error exposed credentials")
		}
	}
}

func openTemp(t *testing.T) *sql.DB {
	t.Helper()
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestMigrateCreatesSchemaVersionTable(t *testing.T) {
	d := openTemp(t)

	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var name string
	err := d.QueryRow(
		`SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'schema_migrations'`,
	).Scan(&name)
	if err != nil {
		t.Fatalf("schema_migrations table not found: %v", err)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	d := openTemp(t)

	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var count int
	if err := d.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected at least one applied migration")
	}
}

func TestCurrentVersionOnFreshDatabase(t *testing.T) {
	d := openTemp(t)

	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	v, err := db.CurrentVersion(d)
	if err != nil {
		t.Fatalf("CurrentVersion: %v", err)
	}
	if v != 11 {
		t.Fatalf("expected version 11 on fresh database (migrations through 0011), got %d", v)
	}
}

func TestFunctionsTableExistsWithNaturalKey(t *testing.T) {
	d := openTemp(t)

	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// The natural key (source, interface, name) must be enforced by the
	// schema: a duplicate insert is the workbook's most common defect.
	if _, err := d.Exec(`INSERT INTO functions
		(source, interface, name, created_at, updated_at)
		VALUES ('DLL', 'WIFI_TEST', 'CLOSE', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if _, err := d.Exec(`INSERT INTO functions
		(source, interface, name, created_at, updated_at)
		VALUES ('DLL', 'WIFI_TEST', 'CLOSE', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err == nil {
		t.Fatalf("duplicate (source, interface, name) must be rejected")
	}

	// Distinct interface with same name is a different Function.
	if _, err := d.Exec(`INSERT INTO functions
		(source, interface, name, created_at, updated_at)
		VALUES ('BUILT-IN', 'RS232', 'CLOSE', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("distinct-key insert rejected: %v", err)
	}
}
