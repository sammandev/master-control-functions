package api_test

import (
	"encoding/json"
	"fmt"
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

func TestFunctionRenameCarriesLinksAndExamples(t *testing.T) {
	for _, mode := range []string{"move", "copy", "merge"} {
		t.Run(mode, func(t *testing.T) {
			app := newFunctionsTestApp(t)
			request := func(method, path, body string, status int) string {
				t.Helper()
				resp, out := doJSON(t, app, method, path, body)
				if resp.StatusCode != status {
					t.Fatalf("%s %s: got %d, want %d: %s", method, path, resp.StatusCode, status, out)
				}
				return out
			}
			create := func(path, body string) { request("POST", path, body, 201) }
			create("/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"RENAME_ME"}`)
			create("/api/parameter-definitions", `{"name":"ANT","type":"int"}`)
			create("/api/criteria-definitions", `{"name":"POWER","type":"double"}`)

			const (
				parameters = "/parameter-links"
				criteria   = "/criteria-links"
				examples   = "/api/examples?function_source=DLL&function_name="
			)
			addAssociations := func(path, name, position, meaning string) {
				create(path+parameters, fmt.Sprintf(`{"name":"ANT","position":%s,"required":true}`, position))
				create(path+criteria, fmt.Sprintf(`{"name":"POWER","position":%s,"required":true}`, position))
				create("/api/examples", fmt.Sprintf(`{"function_source":"DLL","function_name":%q,"script":"DLLWIFI_TEST|RENAME_ME|ANT=1|NULL|1000|NEXT|FAIL","meaning":%q}`, name, meaning))
			}
			addAssociations("/api/functions/1", "RENAME_ME", "4", "original example")

			if mode == "copy" {
				create("/api/functions", `{"source":"DLL","interface":"TELNET","name":"RENAME_ME"}`)
			}
			if mode == "merge" {
				create("/api/functions", `{"source":"DLL","interface":"TELNET","name":"RENAMED"}`)
				addAssociations("/api/functions/2", "RENAMED", "7", "destination example")
			}
			checkItems := func(path, key string, expected ...string) {
				t.Helper()
				out := request("GET", path, "", 200)
				var page struct {
					Items []json.RawMessage `json:"items"`
				}
				if err := json.Unmarshal([]byte(out), &page); err != nil {
					t.Fatalf("decode %s: %v", path, err)
				}
				want := 1
				if key == "" {
					want = 0
				}
				if len(page.Items) != want {
					t.Errorf("%s: got %d items, want %d: %s", path, len(page.Items), want, out)
					return
				}
				if want == 0 {
					return
				}
				for _, field := range append(expected, fmt.Sprintf(`"function_name":%q`, key)) {
					if !strings.Contains(string(page.Items[0]), field) {
						t.Errorf("%s: missing %s: %s", path, field, out)
					}
				}
			}
			checkOrphans := func(stage string) {
				t.Helper()
				out := request("GET", "/api/orphans", "", 200)
				if !strings.Contains(out, `"total":0`) {
					t.Errorf("%s rename: expected zero orphans: %s", stage, out)
				}
			}
			checkOrphans("before")
			request("PUT", "/api/functions/1", `{"source":"DLL","interface":"WIFI_TEST","name":"RENAMED"}`, 200)
			checkOrphans("after")

			id, position, meaning := `"id":1`, `"position":4`, `"meaning":"original example"`
			switch mode {
			case "copy":
				id = `"id":2`
				checkItems("/api/functions/2"+parameters, "RENAME_ME", `"id":1`, position)
				checkItems("/api/functions/2"+criteria, "RENAME_ME", `"id":1`, position)
				checkItems(examples+"RENAME_ME", "RENAME_ME", `"id":1`, meaning)
			case "merge":
				id, position, meaning = `"id":2`, `"position":7`, `"meaning":"destination example"`
				fallthrough
			default:
				checkItems(examples+"RENAME_ME", "")
			}
			checkItems("/api/functions/1"+parameters, "RENAMED", id, `"name":"ANT"`, position, `"required":true`, `"resolved":true`)
			checkItems("/api/functions/1"+criteria, "RENAMED", id, `"name":"POWER"`, position, `"required":true`, `"resolved":true`)
			checkItems(examples+"RENAMED", "RENAMED", id, meaning, `"function_exists":true`)
		})
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

func TestListFunctionsNamePagesMatchGoSort(t *testing.T) {
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MigratePostgres(d); err != nil {
		t.Fatal(err)
	}
	app := api.New(d)
	for _, row := range []struct{ source, device, name string }{
		{"DLL", "RS232", "a"},
		{"DLL", "RS232", "é"},
		{"DLL", "RS232", "Z"},
		{"DLL", "RS232", "B2"},
		{"DLL", "RS232", "B10"},
		{"DLL", "TELNET", "Twin"},
		{"BUILT-IN", "RS232", "Twin"},
		{"DLL", "RS232", "Twin"},
		// Distinct natural keys can compare equal when concatenated.
		{"DLL", "X", "Twin"},
		{"DLLX", "", "Twin"},
		{"DLL", "RS232", "ä"},
	} {
		if _, err := d.Exec(`
			INSERT INTO functions (source, interface, name, created_at, updated_at)
			VALUES ($1, $2, $3, '', '')`, row.source, row.device, row.name); err != nil {
			t.Fatal(err)
		}
	}

	for _, order := range []string{"asc", "desc"} {
		for _, filter := range []string{"", "&search=tWiN&source=DLL"} {
			for _, page := range []string{"1", "2", "3", "4", "5", "6", "7", "999999999999999999999"} {
				url := fmt.Sprintf("/api/functions?sort=name&order=%s&page_size=2&page=%s%s", order, page, filter)
				// An unknown positive locate ID uses Go sorting without moving the page.
				wantResp, want := doRequest(t, app, "GET", url+"&function_id=9223372036854775807")
				gotResp, got := doRequest(t, app, "GET", url)
				if gotResp.StatusCode != http.StatusOK || wantResp.StatusCode != http.StatusOK || got != want {
					t.Errorf("%s: SQL page (%d) %s; Go page (%d) %s", url, gotResp.StatusCode, got, wantResp.StatusCode, want)
				}
			}
		}
	}
}

func TestListFunctionsLocatesRowByID(t *testing.T) {
	app := newFunctionsTestApp(t)
	seedFunctions(t, app)

	_, body := doRequest(t, app, "GET", "/api/functions?search=TX_MULTI_VERIFICATION")
	id := firstIDIn(t, body)

	// Name-asc puts TX_MULTI_VERIFICATION fifth: page 3 at page_size 2.
	resp, body := doRequest(t, app, "GET", fmt.Sprintf("/api/functions?page=1&page_size=2&sort=name&order=asc&function_id=%d", id))
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"page":3`) || !strings.Contains(body, `"name":"TX_MULTI_VERIFICATION"`) {
		t.Fatalf("function_id should open the page holding the row: %s", body)
	}

	_, body = doRequest(t, app, "GET", "/api/functions?page=2&page_size=2&sort=name&order=asc&function_id=999999")
	if !strings.Contains(body, `"page":2`) || !strings.Contains(body, `"name":"OPEN"`) {
		t.Fatalf("an unknown function_id should keep the requested page: %s", body)
	}
}

// firstIDIn returns the id of the first item in a list response body.
func firstIDIn(t *testing.T, body string) int64 {
	t.Helper()
	var list struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &list); err != nil || len(list.Items) == 0 {
		t.Fatalf("no items in list body (%v): %s", err, body)
	}
	return list.Items[0].ID
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

func TestListFunctionsSearchAcrossNameAndDescription(t *testing.T) {
	app := newFunctionsTestApp(t)
	if resp, body := postJSON(t, app, "/api/functions", `{"source":"BUILT-IN","interface":"TEST PROGRAM","name":"ARITHMETIC","description":"Performs arithmetic at 100% using C:\\TEMP"}`); resp.StatusCode != http.StatusCreated {
		t.Fatalf("seed failed: %s", body)
	}

	wantNames(t, app, "?search=%20aRiThMeTiC%20Performs%20", "ARITHMETIC")
	wantNames(t, app, "?search=arithmetic%20%20performs")
	wantNames(t, app, "?search=ARITHMETI_%20Performs")
	wantNames(t, app, "?search=Performs%25")
	wantNames(t, app, "?search=100%25%20using%20C%3A%5CTEMP", "ARITHMETIC")
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
