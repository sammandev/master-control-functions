package api_test

import (
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// seedTwoFunctionsWithExamples creates two functions (ids 1 and 2) and one
// example each, through the API.
func seedTwoFunctionsWithExamples(t *testing.T, app *fiber.App) {
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"TELNET","name":"FN_ONE"}`); resp.StatusCode != 201 {
		t.Fatalf("seed function 1: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"WIFI_TEST","name":"FN_TWO"}`); resp.StatusCode != 201 {
		t.Fatalf("seed function 2: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"FN_ONE","device_name":"WIFI_TEST","script":"DLLWIFI_TEST|F|A=1|NULL|1000|NEXT|FAIL","meaning":"resets the DUT"}`); resp.StatusCode != 201 {
		t.Fatalf("seed example 1: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"FN_TWO","device_name":"IOT","script":"DLLWIFI_TEST|G|B=2|NULL|1000|QUIT|FAIL","meaning":"locks the tester"}`); resp.StatusCode != 201 {
		t.Fatalf("seed example 2: %d (body: %s)", resp.StatusCode, body)
	}
}

func TestExampleFunctionNames(t *testing.T) {
	app, d := newLinksTestApp(t)
	const path = "/api/examples/function-names"
	if resp, body := doRequest(t, app, "GET", path); resp.StatusCode != 200 || body != `{"items":[]}` {
		t.Fatalf("empty names: %d (body: %s)", resp.StatusCode, body)
	}

	seedTwoFunctionsWithExamples(t, app)
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"TELNET","name":"NO_EXAMPLES"}`); resp.StatusCode != 201 {
		t.Fatalf("seed function without examples: %d (body: %s)", resp.StatusCode, body)
	}
	_, err := d.Exec(`INSERT INTO examples
		(function_source, function_name, device_name, script, meaning, created_at, updated_at)
		VALUES
			('EXE', 'FN_ONE', '', 'duplicate name script', 'duplicate name meaning', '', ''),
			('DLL', 'ORPHAN_FN', 'TELNET', 'orphan script', 'orphan meaning', '', '')`)
	if err != nil {
		t.Fatalf("seed duplicate and orphan names: %v", err)
	}

	if resp, body := doRequest(t, app, "GET", path); resp.StatusCode != 200 ||
		body != `{"items":["FN_ONE","FN_TWO","ORPHAN_FN"]}` {
		t.Fatalf("distinct example names: %d (body: %s)", resp.StatusCode, body)
	}
}

// TestExamplesCRUDRoundTrip: create, get, full-replace update, delete.
func TestExamplesCRUDRoundTrip(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"WIFI_TEST","script":"DLLWIFI_TEST|F|A=1|NULL|1000|NEXT|FAIL","meaning":"resets the DUT"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"id":1`) || !strings.Contains(body, `"function_exists":true`) {
		t.Fatalf("create response wrong: %s", body)
	}

	resp, body = doRequest(t, app, "GET", "/api/examples/1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"script":"DLLWIFI_TEST|F|A=1|NULL|1000|NEXT|FAIL"`) {
		t.Fatalf("get: %d (body: %s)", resp.StatusCode, body)
	}

	// Update replaces device scope, script and meaning; the function
	// reference is immutable.
	resp, body = doJSON(t, app, "PUT", "/api/examples/1",
		`{"device_name":"IOT","script":"DLLWIFI_TEST|F|A=2|NULL|1000|NEXT|FAIL","meaning":"updated meaning"}`)
	if resp.StatusCode != 200 ||
		!strings.Contains(body, `"device_name":"IOT"`) ||
		!strings.Contains(body, `"meaning":"updated meaning"`) ||
		!strings.Contains(body, `"function_name":"TX_MULTI_VERIFICATION"`) {
		t.Fatalf("update: %d (body: %s)", resp.StatusCode, body)
	}

	resp, _ = doRequest(t, app, "DELETE", "/api/examples/1")
	if resp.StatusCode != 204 {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	resp, _ = doRequest(t, app, "GET", "/api/examples/1")
	if resp.StatusCode != 404 {
		t.Fatalf("get after delete: %d", resp.StatusCode)
	}
}

// TestExamplesValidation: empty script and unknown function references are
// rejected; the natural key (function, device, script) cannot duplicate.
func TestExamplesValidation(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"","script":"   ","meaning":""}`); resp.StatusCode != 422 {
		t.Fatalf("empty script: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"NO_SUCH_FN","device_name":"","script":"S","meaning":""}`); resp.StatusCode != 422 {
		t.Fatalf("unknown function: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"WIFI_TEST","script":"DLLWIFI_TEST|F|A=1|NULL|1000|NEXT|FAIL","meaning":""}`); resp.StatusCode != 201 {
		t.Fatalf("seed: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"WIFI_TEST","script":"DLLWIFI_TEST|F|A=1|NULL|1000|NEXT|FAIL","meaning":""}`); resp.StatusCode != 409 {
		t.Fatalf("duplicate: %d (body: %s)", resp.StatusCode, body)
	}
}

// TestExamplesFunctionFilterAndSearch: function filter narrows to that
// function's examples; search matches script or meaning text.
func TestExamplesFunctionFilterAndSearch(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedTwoFunctionsWithExamples(t, app)

	// Function filter.
	_, body := doRequest(t, app, "GET", "/api/examples?function_name=FN_ONE")
	if !strings.Contains(body, `"meaning":"resets the DUT"`) || strings.Contains(body, `"meaning":"locks the tester"`) {
		t.Fatalf("function filter wrong: %s", body)
	}
	// Device scope filter.
	_, body = doRequest(t, app, "GET", "/api/examples?device=IOT")
	if !strings.Contains(body, `"meaning":"locks the tester"`) || strings.Contains(body, `"meaning":"resets the DUT"`) {
		t.Fatalf("device filter wrong: %s", body)
	}
	// Search hits script text.
	_, body = doRequest(t, app, "GET", "/api/examples?search=G|B%3D2")
	if !strings.Contains(body, `"meaning":"locks the tester"`) || strings.Contains(body, `"meaning":"resets the DUT"`) {
		t.Fatalf("script search wrong: %s", body)
	}
	// Search hits meaning text.
	_, body = doRequest(t, app, "GET", "/api/examples?search=resets")
	if !strings.Contains(body, `"meaning":"resets the DUT"`) || strings.Contains(body, `"meaning":"locks the tester"`) {
		t.Fatalf("meaning search wrong: %s", body)
	}
	// Pagination envelope present.
	_, body = doRequest(t, app, "GET", "/api/examples")
	if !strings.Contains(body, `"total":2`) {
		t.Fatalf("total wrong: %s", body)
	}
}

func TestExamplesSearchByFunctionLabel(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedTwoFunctionsWithExamples(t, app)

	resp, body := doRequest(t, app, "GET", "/api/examples?search=dll%20fn_one")
	if resp.StatusCode != 200 || !strings.Contains(body, `"total":1`) ||
		!strings.Contains(body, `"meaning":"resets the DUT"`) ||
		strings.Contains(body, `"meaning":"locks the tester"`) {
		t.Fatalf("function label search: %d (body: %s)", resp.StatusCode, body)
	}
}

func TestExamplesLocatePageByID(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedTwoFunctionsWithExamples(t, app)

	// Device ordering puts example 2 (IOT) before example 1 (WIFI_TEST).
	resp, body := doRequest(t, app, "GET", "/api/examples?page=1&page_size=1&example_id=1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"page":2`) ||
		!strings.Contains(body, `"total":2`) || !strings.Contains(body, `"id":1`) ||
		strings.Contains(body, `"id":2`) {
		t.Fatalf("located page: %d (body: %s)", resp.StatusCode, body)
	}

	resp, body = doRequest(t, app, "GET", "/api/examples?device=IOT&page=1&page_size=1&example_id=1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"page":1`) ||
		!strings.Contains(body, `"total":1`) || !strings.Contains(body, `"id":2`) ||
		strings.Contains(body, `"id":1`) {
		t.Fatalf("excluded example changed the requested page or filters: %d (body: %s)", resp.StatusCode, body)
	}
}

// TestExamplesDeviceParamPresentButEmpty: an absent device param means no
// filter, while a present-but-empty one filters blank-scope (global)
// examples — /api/devices/_global reuses this endpoint for its examples tab.
func TestExamplesDeviceParamPresentButEmpty(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedTwoFunctionsWithExamples(t, app)
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"FN_ONE","device_name":"","script":"DLLGLOBAL|F|C=3|NULL|1000|NEXT|FAIL","meaning":"global fallback"}`); resp.StatusCode != 201 {
		t.Fatalf("seed blank example: %d (body: %s)", resp.StatusCode, body)
	}

	// Absent param: no filter, all three examples.
	_, body := doRequest(t, app, "GET", "/api/examples")
	if !strings.Contains(body, `"total":3`) {
		t.Fatalf("no filter total wrong: %s", body)
	}
	// Present but empty: blank scope only.
	_, body = doRequest(t, app, "GET", "/api/examples?device=")
	if !strings.Contains(body, `"total":1`) ||
		!strings.Contains(body, `"meaning":"global fallback"`) ||
		strings.Contains(body, `"meaning":"resets the DUT"`) {
		t.Fatalf("blank-scope filter wrong: %s", body)
	}
}

// TestOrphanExampleFlaggedAndSurvives: examples whose function is missing
// (importer shape, or left behind by a function deletion) stay in place,
// are flagged function_exists:false in list and get, and validate that
// deleting the function does not drop them (ADR-0003).
func TestOrphanExampleFlaggedAndSurvives(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// Example 1 references the seeded function through the API.
	if resp, body := postJSON(t, app, "/api/examples",
		`{"function_source":"DLL","function_name":"TX_MULTI_VERIFICATION","device_name":"","script":"DLLWIFI_TEST|I|D=4|NULL|1000|NEXT|FAIL","meaning":"linked example"}`); resp.StatusCode != 201 {
		t.Fatalf("seed linked example: %d (body: %s)", resp.StatusCode, body)
	}

	// Importer shape: example referencing a function that does not exist.
	_, err := d.Exec(`INSERT INTO examples
		(function_source, function_name, device_name, script, meaning, created_at, updated_at)
		VALUES ('DLL', 'NO_SUCH_FN', '', 'DLLWIFI_TEST|H|C=3|NULL|1000|NEXT|FAIL', 'orphan meaning', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed orphan: %v", err)
	}

	_, body := doRequest(t, app, "GET", "/api/examples")
	if !strings.Contains(body, `"function_exists":false`) {
		t.Fatalf("orphan not flagged in list: %s", body)
	}
	_, body = doRequest(t, app, "GET", "/api/examples/2")
	if !strings.Contains(body, `"function_exists":false`) {
		t.Fatalf("orphan not flagged in get: %s", body)
	}

	// Deleting the referenced function must keep its examples (orphans).
	resp, _ := doRequest(t, app, "DELETE", "/api/functions/1")
	if resp.StatusCode != 204 {
		t.Fatalf("delete function: %d", resp.StatusCode)
	}
	resp, body = doRequest(t, app, "GET", "/api/examples/1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"function_exists":false`) {
		t.Fatalf("example of deleted function not flagged: %d (body: %s)", resp.StatusCode, body)
	}
}
