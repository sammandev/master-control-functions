package testdb

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"webprogfndb/internal/db"
)

func Open(t *testing.T) (*sql.DB, error) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration test: set TEST_DATABASE_URL on the authorized off-machine test host")
	}
	admin, err := db.OpenPostgres(dsn)
	if err != nil {
		return nil, err
	}
	token := make([]byte, 12)
	if _, err := rand.Read(token); err != nil {
		admin.Close()
		return nil, err
	}
	schema := "wpf_test_" + hex.EncodeToString(token)
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec("CREATE SCHEMA " + identifier); err != nil {
		admin.Close()
		return nil, err
	}
	connectionURL, err := url.Parse(dsn)
	if err != nil {
		admin.Exec("DROP SCHEMA " + identifier + " CASCADE")
		admin.Close()
		return nil, fmt.Errorf("invalid TEST_DATABASE_URL")
	}
	options := connectionURL.Query()
	options.Set("search_path", schema)
	connectionURL.RawQuery = options.Encode()
	database, err := db.OpenPostgres(connectionURL.String())
	if err != nil {
		admin.Exec("DROP SCHEMA " + identifier + " CASCADE")
		admin.Close()
		return nil, err
	}
	t.Cleanup(func() {
		database.Close()
		if _, err := admin.Exec("DROP SCHEMA " + identifier + " CASCADE"); err != nil {
			t.Errorf("clean up isolated test schema: %v", err)
		}
		admin.Close()
	})
	return database, nil
}
