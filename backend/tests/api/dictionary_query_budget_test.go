package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"webprogfndb/internal/api"
	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

type queryCounter struct {
	calls atomic.Int64
	rows  atomic.Int64
}

func (c *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	c.calls.Add(1)
	return ctx
}

func (c *queryCounter) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	c.rows.Add(data.CommandTag.RowsAffected())
}

// newTracedApp serves the API over a connection pool whose queries are
// counted, and returns the untraced handle for seeding.
func newTracedApp(t *testing.T) (*fiber.App, *sql.DB, *queryCounter) {
	t.Helper()
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MigratePostgres(d); err != nil {
		t.Fatal(err)
	}

	conn, err := d.Conn(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var config *pgx.ConnConfig
	err = conn.Raw(func(driverConn any) error {
		config = driverConn.(*stdlib.Conn).Conn().Config()
		return nil
	})
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	counter := &queryCounter{}
	config.Tracer = counter
	traced := stdlib.OpenDB(*config)
	t.Cleanup(func() { traced.Close() })
	return api.New(traced), d, counter
}

func TestDictionaryPageQueryBudget(t *testing.T) {
	app, d, counter := newTracedApp(t)

	for _, kind := range []string{"parameter", "criteria"} {
		t.Run(kind, func(t *testing.T) {
			if _, err := d.Exec(fmt.Sprintf(`
				INSERT INTO %s_definitions (name, created_at, updated_at)
				VALUES ('FIRST', '', ''), ('SECOND', '', ''), ('THIRD', '', '')`, kind)); err != nil {
				t.Fatal(err)
			}

			for _, size := range []int{2, 3} {
				before := counter.calls.Load()
				target := fmt.Sprintf("/api/%s-definitions?page_size=%d", kind, size)
				resp, body := doRequest(t, app, "GET", target)
				calls := counter.calls.Load() - before
				if resp.StatusCode != 200 {
					t.Fatalf("page_size=%d: status = %d: %s", size, resp.StatusCode, body)
				}
				var items []json.RawMessage
				if err := json.Unmarshal(extractItems(t, body), &items); err != nil {
					t.Fatal(err)
				}
				if len(items) != size {
					t.Fatalf("page_size=%d: got %d rows, want %d", size, len(items), size)
				}
				if calls != 3 {
					t.Errorf("page_size=%d: got %d SQL queries, want 3", size, calls)
				} else {
					t.Logf("page_size=%d: got 3 SQL queries", size)
				}
			}
		})
	}
}
