package db

import (
	"strings"
	"testing"
)

func TestLoadTroubleshootingMigrations(t *testing.T) {
	pgMigrations, err := loadMigrationsFrom("postgres_migrations")
	if err != nil {
		t.Fatalf("loadPostgresMigrations: %v", err)
	}
	var foundPG bool
	for _, m := range pgMigrations {
		if m.Version == 9 && m.Name == "0009_troubleshooting.sql" {
			foundPG = true
			for _, snippet := range []string{
				"CREATE TABLE troubleshooting_guides",
				"CREATE TABLE troubleshooting_steps",
				"CREATE VIEW mcp_v1_troubleshooting_guides",
				"CREATE VIEW mcp_v1_troubleshooting_steps",
				"ERRCODE:104",
				"ERRCODE:201",
			} {
				if !strings.Contains(m.SQL, snippet) {
					t.Fatalf("0009_troubleshooting.sql missing %q", snippet)
				}
			}
		}
	}
	if !foundPG {
		t.Fatalf("0009_troubleshooting.sql not found in postgres migrations")
	}

	sqliteMigrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	var foundSQLite bool
	for _, m := range sqliteMigrations {
		if m.Version == 7 && m.Name == "0007_troubleshooting.sql" {
			foundSQLite = true
			if !strings.Contains(m.SQL, "CREATE TABLE troubleshooting_guides") {
				t.Fatalf("0007_troubleshooting.sql missing CREATE TABLE troubleshooting_guides")
			}
		}
	}
	if !foundSQLite {
		t.Fatalf("0007_troubleshooting.sql not found in sqlite migrations")
	}
}
