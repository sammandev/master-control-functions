package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
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

func TestTroubleshootingRegistryMigrationDefinition(t *testing.T) {
	migrations, err := loadMigrationsFrom("postgres_migrations")
	if err != nil {
		t.Fatalf("loadPostgresMigrations: %v", err)
	}
	migration := findMigration(t, migrations, 10, "0010_troubleshooting_registry.sql")

	for _, snippet := range []string{
		"CREATE TABLE troubleshooting_categories",
		"CREATE TABLE troubleshooting_severities",
		"DROP CONSTRAINT troubleshooting_guides_category_check",
		"DROP CONSTRAINT troubleshooting_guides_severity_check",
		"REFERENCES troubleshooting_categories (key)",
		"REFERENCES troubleshooting_severities (key)",
		"('connection', 'DUT Connection & Setup',",
		"('logging-reporting', 'Logging & Reporting',",
		"('critical', 'Critical (Blocks Flow)',",
		"('info', 'Info (Best Practice)',",
		"'destructive'",
		"'warn'",
		"'outline'",
		"'default'",
	} {
		if !strings.Contains(migration.SQL, snippet) {
			t.Errorf("0010_troubleshooting_registry.sql missing %q", snippet)
		}
	}
}

func TestTroubleshootingRegistryMigrationAppliesToExistingGuides(t *testing.T) {
	migrations, err := loadMigrationsFrom("postgres_migrations")
	if err != nil {
		t.Fatalf("loadPostgresMigrations: %v", err)
	}
	migration := findMigration(t, migrations, 10, "0010_troubleshooting_registry.sql")
	database := openIsolatedPostgres(t)

	transaction, err := database.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer transaction.Rollback()
	for _, candidate := range migrations {
		if candidate.Version >= 10 {
			continue
		}
		if _, err := transaction.Exec(candidate.SQL); err != nil {
			t.Fatalf("apply migration %d (%s): %v", candidate.Version, candidate.Name, err)
		}
	}
	if _, err := transaction.Exec(`INSERT INTO troubleshooting_guides
		(id, title, category, severity, created_at, updated_at)
		VALUES (99, 'Preserved custom guide', 'production', 'info', 't', 't')`); err != nil {
		t.Fatalf("insert pre-taxonomy guide: %v", err)
	}
	wantGuideTaxonomies := snapshotGuideTaxonomies(t, transaction)
	if _, err := transaction.Exec(migration.SQL); err != nil {
		t.Fatalf("apply registry migration: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	assertGuideKeysPreserved(t, database, wantGuideTaxonomies)
	assertRegistryRows(t, database)
	assertRegistryConstraints(t, database)
}

func findMigration(t *testing.T, migrations []migration, version int, name string) migration {
	t.Helper()
	for _, candidate := range migrations {
		if candidate.Version == version && candidate.Name == name {
			return candidate
		}
	}
	t.Fatalf("%s not found in postgres migrations", name)
	return migration{}
}

func openIsolatedPostgres(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration test: set TEST_DATABASE_URL on the authorized off-machine test host")
	}
	admin, err := OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("open admin database: %v", err)
	}
	token := make([]byte, 12)
	if _, err := rand.Read(token); err != nil {
		admin.Close()
		t.Fatalf("generate schema name: %v", err)
	}
	schema := "wpf_migration_test_" + hex.EncodeToString(token)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec("CREATE SCHEMA " + identifier); err != nil {
		admin.Close()
		t.Fatalf("create schema: %v", err)
	}
	connectionURL, err := url.Parse(dsn)
	if err != nil {
		admin.Exec("DROP SCHEMA " + identifier + " CASCADE")
		admin.Close()
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	options := connectionURL.Query()
	options.Set("search_path", schema)
	connectionURL.RawQuery = options.Encode()
	database, err := OpenPostgres(connectionURL.String())
	if err != nil {
		admin.Exec("DROP SCHEMA " + identifier + " CASCADE")
		admin.Close()
		t.Fatalf("open isolated database: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
		if _, err := admin.Exec("DROP SCHEMA " + identifier + " CASCADE"); err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
		admin.Close()
	})
	return database
}

type guideTaxonomy struct {
	category string
	severity string
}

type guideQueryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

func snapshotGuideTaxonomies(t *testing.T, queryer guideQueryer) map[int64]guideTaxonomy {
	t.Helper()
	rows, err := queryer.Query(`SELECT id, category, severity FROM troubleshooting_guides ORDER BY id`)
	if err != nil {
		t.Fatalf("snapshot guide taxonomies: %v", err)
	}
	defer rows.Close()

	snapshot := make(map[int64]guideTaxonomy)
	for rows.Next() {
		var id int64
		var taxonomy guideTaxonomy
		if err := rows.Scan(&id, &taxonomy.category, &taxonomy.severity); err != nil {
			t.Fatalf("scan guide taxonomy snapshot: %v", err)
		}
		snapshot[id] = taxonomy
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate guide taxonomy snapshot: %v", err)
	}
	return snapshot
}

func assertGuideKeysPreserved(t *testing.T, database *sql.DB, want map[int64]guideTaxonomy) {
	t.Helper()
	got := snapshotGuideTaxonomies(t, database)
	if len(got) != len(want) {
		t.Fatalf("preserved guide count = %d, want %d", len(got), len(want))
	}
	for id, wantTaxonomy := range want {
		gotTaxonomy, ok := got[id]
		if !ok {
			t.Errorf("guide %d disappeared during registry migration", id)
			continue
		}
		if gotTaxonomy != wantTaxonomy {
			t.Errorf("guide %d taxonomy = (%q, %q), want (%q, %q)", id,
				gotTaxonomy.category, gotTaxonomy.severity,
				wantTaxonomy.category, wantTaxonomy.severity)
		}
	}
	var customTitle string
	if err := database.QueryRow(`SELECT title FROM troubleshooting_guides WHERE id = 99`).Scan(&customTitle); err != nil {
		t.Fatalf("query custom guide: %v", err)
	}
	if customTitle != "Preserved custom guide" {
		t.Fatalf("custom guide changed: %q", customTitle)
	}
}

func assertRegistryRows(t *testing.T, database *sql.DB) {
	t.Helper()
	var categoryCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM troubleshooting_categories`).Scan(&categoryCount); err != nil {
		t.Fatalf("count troubleshooting categories: %v", err)
	}
	if categoryCount != 11 {
		t.Fatalf("troubleshooting category count = %d, want 11", categoryCount)
	}
	var severityCount int
	if err := database.QueryRow(`SELECT COUNT(*) FROM troubleshooting_severities`).Scan(&severityCount); err != nil {
		t.Fatalf("count troubleshooting severities: %v", err)
	}
	if severityCount != 5 {
		t.Fatalf("troubleshooting severity count = %d, want 5", severityCount)
	}

	categories := []struct {
		key, label string
		order      int
	}{
		{"connection", "DUT Connection & Setup", 1},
		{"execution", "Script & Flow Execution", 2},
		{"measurement", "Measurement & Criteria", 3},
		{"production", "Production Run & Handoff", 4},
		{"fixture-hardware", "Fixture & Hardware", 5},
		{"instrument-calibration", "Instrument & Calibration", 6},
		{"dut-firmware-image", "DUT Firmware & Image", 7},
		{"network-shopfloor-mes", "Network / Shopfloor / MES", 8},
		{"ini-spec-configuration", "INI & Spec Configuration", 9},
		{"station-environment", "Station Environment", 10},
		{"logging-reporting", "Logging & Reporting", 11},
	}
	for _, want := range categories {
		var label string
		var order int
		var builtIn bool
		if err := database.QueryRow(`SELECT label, sort_order, built_in
			FROM troubleshooting_categories WHERE key = $1`, want.key).Scan(&label, &order, &builtIn); err != nil {
			t.Fatalf("category %q: %v", want.key, err)
		}
		if label != want.label || order != want.order || !builtIn {
			t.Errorf("category %q = (%q, %d, %t), want (%q, %d, true)", want.key, label, order, builtIn, want.label, want.order)
		}
	}

	severities := []struct {
		key, label, tone string
		order            int
	}{
		{"critical", "Critical (Blocks Flow)", "destructive", 1},
		{"high", "High", "destructive", 2},
		{"warning", "Warning (Intermittent / Margin)", "warn", 3},
		{"low", "Low", "outline", 4},
		{"info", "Info (Best Practice)", "default", 5},
	}
	for _, want := range severities {
		var label, tone string
		var order int
		var builtIn bool
		if err := database.QueryRow(`SELECT label, tone, sort_order, built_in
			FROM troubleshooting_severities WHERE key = $1`, want.key).Scan(&label, &tone, &order, &builtIn); err != nil {
			t.Fatalf("severity %q: %v", want.key, err)
		}
		if label != want.label || tone != want.tone || order != want.order || !builtIn {
			t.Errorf("severity %q = (%q, %q, %d, %t), want (%q, %q, %d, true)", want.key, label, tone, order, builtIn, want.label, want.tone, want.order)
		}
	}
}

func assertRegistryConstraints(t *testing.T, database *sql.DB) {
	t.Helper()
	var foreignKeys int
	if err := database.QueryRow(`SELECT COUNT(*)
		FROM pg_constraint
		WHERE conrelid = 'troubleshooting_guides'::regclass
		  AND contype = 'f'
		  AND confrelid IN ('troubleshooting_categories'::regclass, 'troubleshooting_severities'::regclass)`).Scan(&foreignKeys); err != nil {
		t.Fatalf("query registry foreign keys: %v", err)
	}
	if foreignKeys != 2 {
		t.Fatalf("troubleshooting guide registry foreign keys = %d, want 2", foreignKeys)
	}
	if _, err := database.Exec(`INSERT INTO troubleshooting_guides
		(title, category, severity, created_at, updated_at)
		VALUES ('Invalid category', 'not-registered', 'critical', 't', 't')`); err == nil {
		t.Fatal("unregistered category must be rejected")
	}
	if _, err := database.Exec(`INSERT INTO troubleshooting_guides
		(title, category, severity, created_at, updated_at)
		VALUES ('Invalid severity', 'connection', 'urgent', 't', 't')`); err == nil {
		t.Fatal("unregistered severity must be rejected")
	}
	if _, err := database.Exec(`INSERT INTO troubleshooting_categories
		(key, label, description, sort_order) VALUES ('duplicate-label', 'DUT CONNECTION & SETUP', 'Duplicate', 99)`); err == nil {
		t.Fatal("case-insensitive duplicate category label must be rejected")
	}
	if _, err := database.Exec(`INSERT INTO troubleshooting_severities
		(key, label, description, sort_order, tone) VALUES ('duplicate-label', 'CRITICAL (BLOCKS FLOW)', 'Duplicate', 99, 'outline')`); err == nil {
		t.Fatal("case-insensitive duplicate severity label must be rejected")
	}
	if _, err := database.Exec(`INSERT INTO troubleshooting_categories
		(key, label, description, sort_order) VALUES ('custom-test', 'Custom Test', 'Custom category', 99)`); err != nil {
		t.Fatalf("insert custom category: %v", err)
	}
	var builtIn bool
	if err := database.QueryRow(`SELECT built_in FROM troubleshooting_categories WHERE key = 'custom-test'`).Scan(&builtIn); err != nil {
		t.Fatalf("query custom category: %v", err)
	}
	if builtIn {
		t.Fatal("custom category must default to built_in = false")
	}
}
