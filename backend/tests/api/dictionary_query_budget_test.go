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
				beforeRows := counter.rows.Load()
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
				if rows := counter.rows.Load() - beforeRows; rows != 1+2*int64(size) {
					t.Errorf("page_size=%d: got %d SQL result rows, want %d (total, page, usage)", size, rows, 1+2*size)
				}
			}

			beforeRows := counter.rows.Load()
			resp, body := doRequest(t, app, "GET", fmt.Sprintf("/api/%s-definitions?page_size=1&sort=name&definition_id=3", kind))
			var located struct {
				Page  int `json:"page"`
				Items []struct {
					ID int64 `json:"id"`
				} `json:"items"`
			}
			if err := json.Unmarshal([]byte(body), &located); err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 || located.Page != 3 || len(located.Items) != 1 || located.Items[0].ID != 3 {
				t.Fatalf("locate: %d %s", resp.StatusCode, body)
			}
			if rows := counter.rows.Load() - beforeRows; rows > 4 {
				t.Errorf("locate: got %d SQL result rows, want at most 4 (total, position, page, usage)", rows)
			}

			for _, tc := range []struct {
				query string
				total int
			}{
				{"search=no-match&definition_id=3", 0},
				{"sort=name&order=desc&page_size=2&page=5", 3},
			} {
				beforeCalls, beforeRows := counter.calls.Load(), counter.rows.Load()
				resp, body := doRequest(t, app, "GET", fmt.Sprintf("/api/%s-definitions?%s", kind, tc.query))
				var got struct {
					Total int               `json:"total"`
					Items []json.RawMessage `json:"items"`
				}
				if err := json.Unmarshal([]byte(body), &got); err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != 200 || got.Total != tc.total || got.Items == nil || len(got.Items) != 0 {
					t.Fatalf("%s: got %d %s", tc.query, resp.StatusCode, body)
				}
				if calls, rows := counter.calls.Load()-beforeCalls, counter.rows.Load()-beforeRows; calls != 1 || rows != 1 {
					t.Errorf("%s: got %d SQL queries and %d rows, want one count query returning one row", tc.query, calls, rows)
				}
			}
		})
	}
}
