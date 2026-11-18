package api_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"webprogfndb/internal/api"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

// newFunctionsTestApp boots the app with migrations (including the
// functions schema) exactly as production wiring does.
func newFunctionsTestApp(t *testing.T) *fiber.App {
	t.Helper()
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return api.New(d)
}

// doJSON issues a JSON-body request through the Fiber test harness.
func doJSON(t *testing.T, app *fiber.App, method, target, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Host = "example.com"
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, string(b)
}

func postJSON(t *testing.T, app *fiber.App, target, body string) (*http.Response, string) {
	t.Helper()
	return doJSON(t, app, "POST", target, body)
}

func validFunctionJSON() string {
	return `{
		"source": "DLL",
		"interface": "WIFI_TEST",
		"name": "TX_MULTI_VERIFICATION",
		"description": "Measure the Wi-Fi quality of antennas",
		"convention": 2,
		"category": "wifi_test",
		"execution_role": "measurement",
		"execution_phase": "test_execution"
	}`
}

func TestCreateFunctionRoundTrip(t *testing.T) {
	app := newFunctionsTestApp(t)

	resp, body := postJSON(t, app, "/api/functions", validFunctionJSON())
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d, want 201 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"id":1`) {
		t.Fatalf("created body missing id: %s", body)
	}

	resp, body = doRequest(t, app, "GET", "/api/functions/1")
	if resp.StatusCode != 200 {
		t.Fatalf("get status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name":"TX_MULTI_VERIFICATION"`) {
		t.Fatalf("get body wrong: %s", body)
	}
	if !strings.Contains(body, `"created_at"`) {
		t.Fatalf("created_at missing: %s", body)
	}
}

func TestCreateFunctionValidationFieldErrors(t *testing.T) {
	app := newFunctionsTestApp(t)

	missing := `{"source": "DLL", "interface": "WIFI_TEST"}`
	resp, body := postJSON(t, app, "/api/functions", missing)
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name"`) {
		t.Fatalf("field error for name missing: %s", body)
	}

	badSource := `{"source": "MAGIC", "interface": "WIFI_TEST", "name": "X"}`
	resp, body = postJSON(t, app, "/api/functions", badSource)
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"source"`) {
		t.Fatalf("field error for source missing: %s", body)
	}
}

func TestCreateFunctionDuplicateKeyRejected(t *testing.T) {
	app := newFunctionsTestApp(t)

	if resp, _ := postJSON(t, app, "/api/functions", validFunctionJSON()); resp.StatusCode != 201 {
		t.Fatalf("first create failed")
	}
	resp, body := postJSON(t, app, "/api/functions", validFunctionJSON())
	if resp.StatusCode != 409 {
		t.Fatalf("status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, "already exists") {
		t.Fatalf("conflict message unclear: %s", body)
	}
}

func TestFunctionCRUDRoundTrip(t *testing.T) {
	app := newFunctionsTestApp(t)

	resp, body := postJSON(t, app, "/api/functions", validFunctionJSON())
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d (body: %s)", resp.StatusCode, body)
	}

	update := `{
		"source": "DLL",
		"interface": "WIFI_TEST",
		"name": "TX_MULTI_VERIFICATION",
		"description": "Updated description",
		"category": "rf",
		"execution_role": "verification",
		"execution_phase": "post_test"
	}`
	resp, body = doJSON(t, app, "PUT", "/api/functions/1", update)
	if resp.StatusCode != 200 {
		t.Fatalf("update status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"description":"Updated description"`) {
		t.Fatalf("update not applied: %s", body)
	}

	resp, body = doRequest(t, app, "DELETE", "/api/functions/1")
	if resp.StatusCode != 204 {
		t.Fatalf("delete status = %d (body: %s)", resp.StatusCode, body)
	}

	resp, _ = doRequest(t, app, "GET", "/api/functions/1")
	if resp.StatusCode != 404 {
		t.Fatalf("get after delete status = %d, want 404", resp.StatusCode)
	}

	resp, _ = doRequest(t, app, "DELETE", "/api/functions/1")
	if resp.StatusCode != 404 {
		t.Fatalf("re-delete status = %d, want 404", resp.StatusCode)
	}
}

func TestUpdateFunctionKeyConflict(t *testing.T) {
	app := newFunctionsTestApp(t)

	a := `{"source": "BUILT-IN", "interface": "RS232", "name": "OPEN"}`
	if resp, _ := postJSON(t, app, "/api/functions", a); resp.StatusCode != 201 {
		t.Fatalf("create A failed")
	}
	b := `{"source": "BUILT-IN", "interface": "TELNET", "name": "OPEN"}`
	if resp, _ := postJSON(t, app, "/api/functions", b); resp.StatusCode != 201 {
		t.Fatalf("create B failed")
	}

	// Rename B onto A's key → conflict.
	conflict := `{"source": "BUILT-IN", "interface": "RS232", "name": "OPEN", "description": "x"}`
	resp, body := doJSON(t, app, "PUT", "/api/functions/2", conflict)
	if resp.StatusCode != 409 {
		t.Fatalf("status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
}

func seedFunctions(t *testing.T, app *fiber.App) {
	t.Helper()
	seeds := []string{
		`{"source": "BUILT-IN", "interface": "RS232", "name": "CLOSE", "description": "close port"}`,
		`{"source": "BUILT-IN", "interface": "RS232", "name": "OPEN", "description": "open port"}`,
		`{"source": "BUILT-IN", "interface": "TELNET", "name": "CLOSE", "description": "close session"}`,
		`{"source": "DLL", "interface": "WIFI_TEST", "name": "TX_MULTI_VERIFICATION", "description": "measure wifi quality"}`,
		`{"source": "DLL", "interface": "DMM", "name": "READ_VOLTAGE", "description": "read voltage"}`,
	}
	for _, s := range seeds {
		if resp, body := postJSON(t, app, "/api/functions", s); resp.StatusCode != 201 {
			t.Fatalf("seed failed: %s → %s", s, body)
		}
	}
}

func TestListFunctionsPaginationContract(t *testing.T) {
	app := newFunctionsTestApp(t)
	seedFunctions(t, app)

	resp, body := doRequest(t, app, "GET", "/api/functions?page=1&page_size=2&sort=name&order=asc")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"total":5`) {
		t.Fatalf("total missing: %s", body)
	}
	if !strings.Contains(body, `"page":1`) || !strings.Contains(body, `"page_size":2`) {
		t.Fatalf("paging metadata missing: %s", body)
	}
	// Sorted by name asc: CLOSE/RS232, CLOSE/TELNET first (tiebreak by
	// interface), then OPEN, READ_VOLTAGE, TX_MULTI_VERIFICATION.
	if !strings.Contains(body, `"name":"CLOSE"`) {
		t.Fatalf("expected CLOSE first: %s", body)
	}
	if strings.Count(body, `"name":"CLOSE"`) != 2 {
		t.Fatalf("expected both CLOSE rows on page 1: %s", body)
	}

	// Name-asc order: CLOSE, CLOSE, OPEN, READ_VOLTAGE, TX_MULTI_VERIFICATION.
	// Page 2 (rows 3-4): OPEN, READ_VOLTAGE.
	resp, body = doRequest(t, app, "GET", "/api/functions?page=2&page_size=2&sort=name&order=asc")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	if !strings.Contains(body, `"name":"OPEN"`) || !strings.Contains(body, `"name":"READ_VOLTAGE"`) {
		t.Fatalf("page 2 should contain OPEN and READ_VOLTAGE: %s", body)
	}
	// Page 3 (row 5): TX_MULTI_VERIFICATION.
	resp, body = doRequest(t, app, "GET", "/api/functions?page=3&page_size=2&sort=name&order=asc")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	if !strings.Contains(body, `"name":"TX_MULTI_VERIFICATION"`) {
		t.Fatalf("page 3 should contain TX_MULTI_VERIFICATION: %s", body)
	}
}

func TestListFunctionsSearch(t *testing.T) {
	app := newFunctionsTestApp(t)
	seedFunctions(t, app)

	resp, body := doRequest(t, app, "GET", "/api/functions?search=clos")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("search should match 2 CLOSE rows: %s", body)
	}
	if strings.Contains(body, "READ_VOLTAGE") {
		t.Fatalf("non-matching row leaked: %s", body)
	}

	// Description search: "wifi quality" only in TX_MULTI_VERIFICATION.
	_, body = doRequest(t, app, "GET", "/api/functions?search=wifi%20quality")
	if !strings.Contains(body, `"total":1`) {
		t.Fatalf("description search should match 1 row: %s", body)
	}
}

func TestListFunctionsFilter(t *testing.T) {
	app := newFunctionsTestApp(t)
	seedFunctions(t, app)

	_, body := doRequest(t, app, "GET", "/api/functions?source=DLL")
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("source filter should match 2 rows: %s", body)
	}

	_, body = doRequest(t, app, "GET", "/api/functions?execution_phase=nothing")
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("unknown phase should match 0 rows: %s", body)
	}
}
