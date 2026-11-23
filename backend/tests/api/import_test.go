package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"webprogfndb/internal/api"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
	"webprogfndb/tests/support/workbook"
)

func newImportTestApp(t *testing.T) *fiber.App {
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

// newImportTestAppWithDB exposes the underlying *sql.DB so a test can
// assert schema state directly (used by ADR-0009 device-scope tests to
// confirm no-overwrite semantics on the SQL side).
func newImportTestAppWithDB(t *testing.T) (*fiber.App, *sql.DB) {
	t.Helper()
	d, err := testdb.Open(t)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	if err := db.MigratePostgres(d); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return api.New(d), d
}

// multipartBody builds an upload with a single file field.
func multipartBody(t *testing.T, fieldName, fileName string, content []byte) (string, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("form file: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return w.FormDataContentType(), &buf
}

func postFile(t *testing.T, app *fiber.App, target, contentType string, body *bytes.Buffer) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest("POST", target, bytes.NewReader(body.Bytes()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Host = "example.com"
	req.Header.Set("Content-Type", contentType)
	// Real-workbook imports (~535 rows) exceed fiber's default 1s test
	// timeout on slower disks; give them 30s.
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, string(b)
}

// fixturePath resolves the shared fixture from this package's tests
// (go test runs with the package directory as cwd).
func fixturePath() string {
	return filepath.Join("..", "testdata", workbook.FixtureFile)
}

// fixtureBytes reads the real source workbook, copied into the shared
// testdata directory.
func fixtureBytes(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(fixturePath())
	if err != nil {
		t.Fatalf("fixture missing (%s): %v", fixturePath(), err)
	}
	return b
}

// fixtureCounts derives the fixture's per-table counts through the
// importer's count oracle — the single source every count assertion in
// this package checks against.
func fixtureCounts(t *testing.T) workbook.Counts {
	t.Helper()
	c, err := workbook.FixtureCounts(fixturePath())
	if err != nil {
		t.Fatalf("fixture oracle: %v", err)
	}
	return c
}

// countKeyValues renders the oracle counts as the JSON fragments the
// import/export endpoints emit, e.g. `"functions":<count>`.
func countKeyValues(c workbook.Counts) []string {
	return []string{
		fmt.Sprintf(`"functions":%d`, c.Functions),
		fmt.Sprintf(`"parameter_definitions":%d`, c.Parameters),
		fmt.Sprintf(`"criteria_definitions":%d`, c.Criteria),
		fmt.Sprintf(`"parameter_links":%d`, c.ParameterLinks),
		fmt.Sprintf(`"criteria_links":%d`, c.CriteriaLinks),
		fmt.Sprintf(`"examples":%d`, c.Examples),
	}
}

// syntheticWorkbook builds an edge-case xlsx: an orphan link (function
// absent from function_main), mixed device casing, and the phase typo.
func syntheticWorkbook(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	sheets := map[string][][]string{
		"function_main": {
			{"source", "interface", "name", "description", "convention", "category", "execution_role", "execution_phase"},
			{"BUILT-IN", "RS232", "OPEN", "open port", "1", "communication", "connection", "text_execution"},
		},
		"parameter_main": {
			{"name", "description", "type", "unit", "default_value", "allowed_values", "cardinality", "value_pattern", "required"},
			{"BAUD", "baud rate", "int", "", "9600", "{9600,115200}", "SINGLE", "", "True"},
		},
		"criteria_main": {
			{"name", "description", "type", "unit", "default_value", "allowed_values", "cardinality", "value_pattern", "required"},
			{"POWER", "", "string", "", "", "", "SINGLE", "", "False"},
		},
		"function_parameter": {
			{"name", "function_source", "device_name", "function_name", "source_file", "position", "required"},
			{"BAUD", "BUILT-IN", "GoldenKV", "OPEN", "RD_SETTING.ini", "1", "True"},
			{"BAUD", "BUILT-IN", "GoldenKV", "OPEN", "RD_SETTING.ini", "1", "True"},
			{"GHOST", "DLL", "IOT", "ABSENT_FUNCTION", "", "", "False"},
		},
		"function_criteria": {
			{"name", "function_source", "device_name", "function_name", "source_file", "position", "required"},
			{"POWER", "BUILT-IN", "golden", "OPEN", "", "2", "False"},
		},
		"example": {
			{"function_source", "device_name", "function_name", "script", "meaning"},
			{"BUILT-IN", "iot", "OPEN", "BUILTRS232|OPEN|BAUD=9600|NULL|3000", "opens the port"},
		},
	}
	for name := range sheets {
		f.NewSheet(name)
	}
	for name, rows := range sheets {
		for i, row := range rows {
			for j, v := range row {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
				if err := f.SetCellValue(name, cell, v); err != nil {
					t.Fatalf("set cell: %v", err)
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}

func TestImportRealWorkbookAndIdempotency(t *testing.T) {
	app := newImportTestApp(t)
	data := fixtureBytes(t)

	ct, body := multipartBody(t, "file", "master.xlsx", data)
	resp, res := postFile(t, app, "/api/import", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("import status = %d (body: %s)", resp.StatusCode, res)
	}
	// Counts come from the fixture oracle, not literals: a workbook batch
	// requires zero test edits. ADR-0009 device-scope means ANT and FREQ
	// each carry a WIFI_TEST-scoped and an OSP-scoped definition row, and
	// both must coexist after import.
	c := fixtureCounts(t)
	for _, want := range countKeyValues(c) {
		if !strings.Contains(res, want) {
			t.Fatalf("counts missing %s: %s", want, res)
		}
	}

	// Idempotent: importing the same workbook again keeps counts identical
	// AND content stable — the same normalizations apply deterministically.
	ct, body = multipartBody(t, "file", "master.xlsx", data)
	resp, res = postFile(t, app, "/api/import", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("re-import status = %d (body: %s)", resp.StatusCode, res)
	}
	for _, want := range []string{
		fmt.Sprintf(`"functions":%d`, c.Functions),
		fmt.Sprintf(`"parameter_links":%d`, c.ParameterLinks),
		fmt.Sprintf(`"examples":%d`, c.Examples),
	} {
		if !strings.Contains(res, want) {
			t.Fatalf("re-import counts missing %s: %s", want, res)
		}
	}

	// Content stability: a workbook row's content matches after re-import.
	resp, res = doRequest(t, app, "GET", "/api/functions/1")
	if resp.StatusCode != 200 || !strings.Contains(res, `"created_at"`) {
		t.Fatalf("function 1 readback failed: %s", res)
	}
	first := res

	ct, body = multipartBody(t, "file", "master.xlsx", data)
	if resp, _ := postFile(t, app, "/api/import", ct, body); resp.StatusCode != 200 {
		t.Fatalf("third import failed")
	}
	resp, res = doRequest(t, app, "GET", "/api/functions/1")
	if resp.StatusCode != 200 {
		t.Fatalf("function 1 readback after re-import failed: %s", res)
	}
	// Same natural-key row: description and source content unchanged
	// (only updated_at may differ).
	if stripTimestamps(first) != stripTimestamps(res) {
		t.Fatalf("content drifted across re-import:\nfirst:  %s\nsecond: %s", first, res)
	}

	// Change log lists the runs.
	resp, res = doRequest(t, app, "GET", "/api/import/runs")
	if resp.StatusCode != 200 || !strings.Contains(res, "master.xlsx") {
		t.Fatalf("runs log incomplete: %s", res)
	}
}

// stripTimestamps removes the mutable fields from a function JSON payload
// so two imports can be compared for content equality.
func stripTimestamps(body string) string {
	for _, k := range []string{`"created_at":"…"`, `"updated_at"`, `"id"`} {
		_ = k
	}
	cut := func(s, key string) string {
		i := strings.Index(s, `"`+key+`"`)
		if i < 0 {
			return s
		}
		end := strings.Index(s[i:], ",")
		if end < 0 {
			return s[:i]
		}
		return s[:i] + s[i+end+1:]
	}
	return cut(cut(cut(body, "created_at"), "updated_at"), "id")
}

func TestImportNormalizationsLoggedAndViewable(t *testing.T) {
	app := newImportTestApp(t)

	ct, body := multipartBody(t, "file", "synthetic.xlsx", syntheticWorkbook(t))
	resp, res := postFile(t, app, "/api/import", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("import status = %d (body: %s)", resp.StatusCode, res)
	}
	if !strings.Contains(res, `"run_id":1`) {
		t.Fatalf("run_id missing: %s", res)
	}

	// Per-run view carries the applied normalizations: phase typo +
	// three device-scope casing fixes.
	resp, res = doRequest(t, app, "GET", "/api/import/runs/1")
	if resp.StatusCode != 200 {
		t.Fatalf("run detail status = %d (body: %s)", resp.StatusCode, res)
	}
	for _, want := range []string{
		"phase-typo-text-execution", "device-scope-uppercase",
		"GoldenKV", "GOLDENKV",
	} {
		if !strings.Contains(res, want) {
			t.Fatalf("normalizations missing %s: %s", want, res)
		}
	}
}

func TestImportSurvivesWebEdits(t *testing.T) {
	app := newImportTestApp(t)
	data := fixtureBytes(t)

	ct, body := multipartBody(t, "file", "master.xlsx", data)
	if resp, _ := postFile(t, app, "/api/import", ct, body); resp.StatusCode != 200 {
		t.Fatalf("first import failed")
	}

	webOnly := `{"source": "BUILT-IN", "interface": "CMD_CONSOLE", "name": "WEB_EDITED", "description": "created in UI"}`
	if resp, b := postJSON(t, app, "/api/functions", webOnly); resp.StatusCode != 201 {
		t.Fatalf("web edit failed: %s", b)
	}

	ct, body = multipartBody(t, "file", "master.xlsx", data)
	if resp, _ := postFile(t, app, "/api/import", ct, body); resp.StatusCode != 200 {
		t.Fatalf("re-import failed")
	}

	resp, res := doRequest(t, app, "GET", "/api/functions?search=WEB_EDITED")
	if resp.StatusCode != 200 || !strings.Contains(res, `"name":"WEB_EDITED"`) {
		t.Fatalf("web-edited row did not survive re-import: %s", res)
	}
}

func TestImportResetClearsAndReimports(t *testing.T) {
	app := newImportTestApp(t)
	data := fixtureBytes(t)

	ct, body := multipartBody(t, "file", "master.xlsx", data)
	if resp, _ := postFile(t, app, "/api/import", ct, body); resp.StatusCode != 200 {
		t.Fatalf("first import failed")
	}

	webOnly := `{"source": "BUILT-IN", "interface": "CMD_CONSOLE", "name": "WEB_EDITED"}`
	if resp, _ := postJSON(t, app, "/api/functions", webOnly); resp.StatusCode != 201 {
		t.Fatalf("web edit failed")
	}

	ct, body = multipartBody(t, "file", "master.xlsx", data)
	resp, res := postFile(t, app, "/api/import/reset", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("reset status = %d (body: %s)", resp.StatusCode, res)
	}
	for _, want := range append(countKeyValues(fixtureCounts(t)), `"reset":true`) {
		if !strings.Contains(res, want) {
			t.Fatalf("reset result missing %s: %s", want, res)
		}
	}

	// Reset cleared the web-only row and re-imported deterministically.
	resp, res = doRequest(t, app, "GET", "/api/functions?search=WEB_EDITED")
	if resp.StatusCode != 200 || strings.Contains(res, `"name":"WEB_EDITED"`) {
		t.Fatalf("reset did not clear: %s", res)
	}
}

func TestImportRejectsInvalidHeadersWithoutDeletingFunctions(t *testing.T) {
	app := newImportTestApp(t)
	const functionName = "KEEP_EXISTING"
	if resp, body := postJSON(t, app, "/api/functions", `{"source":"BUILT-IN","interface":"CMD_CONSOLE","name":"`+functionName+`"}`); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create existing function: %d %s", resp.StatusCode, body)
	}

	file := excelize.NewFile()
	defer file.Close()
	for _, sheet := range []string{"function_main", "parameter_main", "criteria_main", "function_parameter", "function_criteria", "example"} {
		if _, err := file.NewSheet(sheet); err != nil {
			t.Fatal(err)
		}
		if err := file.SetCellValue(sheet, "A1", "unrecognized_header"); err != nil {
			t.Fatal(err)
		}
		if err := file.SetCellValue(sheet, "A2", "nonempty data"); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.DeleteSheet("Sheet1"); err != nil {
		t.Fatal(err)
	}
	data, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}

	for _, endpoint := range []string{"/api/import", "/api/import/reset"} {
		ct, body := multipartBody(t, "file", "invalid-headers.xlsx", data.Bytes())
		resp, result := postFile(t, app, endpoint, ct, body)
		if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(result, "sheet function_main: missing required columns: source, interface (or device_name), name") {
			t.Errorf("%s: want 422 with missing headers, got %d %s", endpoint, resp.StatusCode, result)
		}
	}

	resp, result := doRequest(t, app, "GET", "/api/functions?search="+functionName)
	if resp.StatusCode != http.StatusOK || !strings.Contains(result, `"name":"`+functionName+`"`) {
		t.Fatalf("existing function did not survive the rejected reset: %d %s", resp.StatusCode, result)
	}
}

func TestImportStoresOrphansAndCountsUsage(t *testing.T) {
	app := newImportTestApp(t)

	ct, body := multipartBody(t, "file", "synthetic.xlsx", syntheticWorkbook(t))
	resp, res := postFile(t, app, "/api/import", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("import status = %d (body: %s)", resp.StatusCode, res)
	}

	// The orphan link (function ABSENT_FUNCTION is not in function_main)
	// is stored, never dropped (ADR-0003): the parameter definition BAUD
	// shows usage_count 1, and the criteria definition POWER too.
	resp, res = doRequest(t, app, "GET", "/api/parameter-definitions?search=BAUD")
	if resp.StatusCode != 200 || !strings.Contains(res, `"usage_count":1`) {
		t.Fatalf("BAUD usage count wrong: %s", res)
	}
	resp, res = doRequest(t, app, "GET", "/api/criteria-definitions?search=POWER")
	if resp.StatusCode != 200 || !strings.Contains(res, `"usage_count":1`) {
		t.Fatalf("POWER usage count wrong: %s", res)
	}
}

// TestImportPreservesDeviceScopedDefinitions: importing the real workbook
// must NOT overwrite a device-scoped definition with a same-name global
// one (or vice versa) — ADR-0009 keys definitions on (name, device_name).
// The real workbook has both ANT and FREQ scoped to OSP alongside
// a blank-scope (global) sibling; both must survive, and the dictionary
// API must surface them with their distinct device scopes.
func TestImportPreservesDeviceScopedDefinitions(t *testing.T) {
	app, d := newImportTestAppWithDB(t)
	data := fixtureBytes(t)

	ct, body := multipartBody(t, "file", "master.xlsx", data)
	if resp, _ := postFile(t, app, "/api/import", ct, body); resp.StatusCode != 200 {
		t.Fatalf("import failed")
	}

	// Pin the consumer bug being fixed: ANT and FREQ each have exactly
	// two rows after import (WIFI_TEST + OSP). A name-only UNIQUE
	// would have collapsed them on the last-row-wins ON CONFLICT;
	// (name, device_name) keeps both.
	for _, name := range []string{"ANT", "FREQ"} {
		var count int
		if err := d.QueryRow(
			`SELECT COUNT(*) FROM parameter_definitions WHERE name = $1`, name,
		).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		if count != 2 {
			t.Fatalf("%s parameter definitions = %d, want 2 (WIFI_TEST + OSP)", name, count)
		}
	}

	// And the dictionary API surfaces both rows with their distinct
	// scopes so the UI can tell them apart. Semantic, not positional:
	// exactly two ANT rows — one WIFI_TEST-scoped + one OSP-scoped —
	// independent of row order or generated ids. (Search is a substring
	// match, so non-ANT names may appear in the list and are ignored.)
	resp, list := doRequest(t, app, "GET", "/api/parameter-definitions?search=ANT")
	if resp.StatusCode != 200 {
		t.Fatalf("ANT list: %d (body: %s)", resp.StatusCode, list)
	}
	var antList struct {
		Items []struct {
			Name       string `json:"name"`
			DeviceName string `json:"device_name"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(list), &antList); err != nil {
		t.Fatalf("parse ANT list: %v (body: %s)", err, list)
	}
	scopes := map[string]int{}
	for _, it := range antList.Items {
		if it.Name != "ANT" {
			continue
		}
		scopes[it.DeviceName]++
	}
	if len(scopes) != 2 || scopes["WIFI_TEST"] != 1 || scopes["OSP"] != 1 {
		t.Fatalf("ANT rows wrong, want exactly one WIFI_TEST-scoped + one OSP-scoped: %s", list)
	}
}
