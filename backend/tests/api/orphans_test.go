package api_test

import (
	"strings"
	"testing"
)

// TestOrphanReportDetection: one orphan per missing-target shape shows up
// in the right group with the right flags, resolved rows do not.
func TestOrphanReportDetection(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// Resolved link: must NOT appear in the report.
	if resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","position":1}`); resp.StatusCode != 201 {
		t.Fatalf("seed resolved link: %d (body: %s)", resp.StatusCode, body)
	}
	// Orphan A: link to a definition that does not exist (importer shape).
	_, err := d.Exec(`INSERT INTO parameter_links
		(function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES ('DLL', 'TX_MULTI_VERIFICATION', '', 'GHOST_PARAM', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed definition orphan: %v", err)
	}
	// Orphan B: link to a function that does not exist (importer shape).
	// The criteria definition exists, so the only missing target is the
	// function.
	if resp, body := postJSON(t, app, "/api/criteria-definitions",
		`{"name":"ANY_CRIT","type":"double"}`); resp.StatusCode != 201 {
		t.Fatalf("seed criteria definition: %d (body: %s)", resp.StatusCode, body)
	}
	_, err = d.Exec(`INSERT INTO criteria_links
		(function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES ('DLL', 'NO_SUCH_FUNCTION', '', 'ANY_CRIT', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed function orphan: %v", err)
	}
	// Orphan C: example whose function does not exist.
	_, err = d.Exec(`INSERT INTO examples
		(function_source, function_name, device_name, script, meaning, created_at, updated_at)
		VALUES ('DLL', 'NO_SUCH_FUNCTION', '', 'DLLWIFI_TEST|F|A=1|NULL|1000|L|', 'meaning', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed example orphan: %v", err)
	}

	resp, body := doRequest(t, app, "GET", "/api/orphans")
	if resp.StatusCode != 200 {
		t.Fatalf("report status = %d (body: %s)", resp.StatusCode, body)
	}

	// Parameter links: exactly the definition orphan (the resolved and
	// function-orphan rows belong to other groups / no group).
	if !strings.Contains(body, `"kind":"parameter-links"`) {
		t.Fatalf("parameter-links group missing: %s", body)
	}
	if !strings.Contains(body, `"name":"GHOST_PARAM"`) {
		t.Fatalf("definition orphan missing: %s", body)
	}
	if !strings.Contains(body, `"definition_exists":false`) ||
		!strings.Contains(body, `"function_exists":true`) {
		t.Fatalf("orphan flags wrong: %s", body)
	}
	// Criteria links: exactly the function orphan. Struct order puts
	// function_exists right after the function key; definition_exists true
	// appears only on this row (the parameter orphan has it false).
	if !strings.Contains(body, `"function_name":"NO_SUCH_FUNCTION","function_exists":false`) {
		t.Fatalf("criteria function orphan missing: %s", body)
	}
	if !strings.Contains(body, `"definition_exists":true`) {
		t.Fatalf("function-orphan flags wrong: %s", body)
	}
	// Examples: exactly the orphan example.
	if !strings.Contains(body, `"kind":"examples"`) || !strings.Contains(body, `"script":"DLLWIFI_TEST|F|A=1|NULL|1000|L|"`) {
		t.Fatalf("example orphan missing: %s", body)
	}
	// The resolved link never appears: its name must not show up in the
	// report at all (it has both targets, so it is not an orphan).
	if strings.Contains(body, `"name":"ANT"`) {
		t.Fatalf("resolved link leaked into report: %s", body)
	}
}

// TestOrphanResolutionFromReport: creating the missing target through the
// regular create endpoints (prefilled from the report row) resolves the
// reference — the row leaves the report and the link flags resolved.
func TestOrphanResolutionFromReport(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// Definition orphan on a real function's link.
	postJSON(t, app, "/api/functions/1/parameter-links", `{"name":"ANT"}`)
	doRequest(t, app, "DELETE", "/api/parameter-definitions/1?force=true")

	// Function orphan on a criteria link. The criteria definition exists,
	// so the ONLY missing target is the function.
	resp, body := postJSON(t, app, "/api/criteria-definitions",
		`{"name":"CRIT_ANY","type":"double"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("seed criteria definition: %d (body: %s)", resp.StatusCode, body)
	}
	_, err := d.Exec(`INSERT INTO criteria_links
		(function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES ('BUILT-IN', 'MISSING_FN', '', 'CRIT_ANY', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, body = doRequest(t, app, "GET", "/api/orphans")
	if !strings.Contains(body, `"name":"ANT"`) || !strings.Contains(body, `"function_name":"MISSING_FN"`) {
		t.Fatalf("orphans not reported before resolution: %s", body)
	}

	// One-action resolution: create the missing definition (name prefilled
	// from the report row) via the dictionary endpoint.
	resp, body = postJSON(t, app, "/api/parameter-definitions",
		`{"name":"ANT","type":"int","required":true}`)
	if resp.StatusCode != 201 {
		t.Fatalf("resolve definition status = %d (body: %s)", resp.StatusCode, body)
	}
	// Create the missing function (source + name prefilled, interface is
	// the Function's own attribute) and note its id from the response.
	resp, body = postJSON(t, app, "/api/functions",
		`{"source":"BUILT-IN","interface":"TELNET","name":"MISSING_FN"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("resolve function status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"id":2`) {
		t.Fatalf("resolved function id unexpected: %s", body)
	}

	// The report is now empty; both links flag resolved in their lists.
	resp, body = doRequest(t, app, "GET", "/api/orphans")
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("orphans still reported after resolution: %s", body)
	}
	_, body = doRequest(t, app, "GET", "/api/functions/1/parameter-links")
	if !strings.Contains(body, `"resolved":true`) {
		t.Fatalf("parameter link not resolved: %s", body)
	}
	_, body = doRequest(t, app, "GET", "/api/functions/2/criteria-links")
	if !strings.Contains(body, `"resolved":true`) {
		t.Fatalf("criteria link not resolved after function creation: %s", body)
	}
}

// TestOrphanReportFromRealWorkbook: after importing the real source
// workbook, the report pins the true invariant — the merged master
// workbook resolves every reference, so all three orphan groups
// count 0.
func TestOrphanReportFromRealWorkbook(t *testing.T) {
	app := newImportTestApp(t)
	data := fixtureBytes(t)
	ct, upload := multipartBody(t, "file", "master.xlsx", data)
	resp, body := postFile(t, app, "/api/import", ct, upload)
	if resp.StatusCode != 200 {
		t.Fatalf("import status = %d (body: %s)", resp.StatusCode, body)
	}

	resp, body = doRequest(t, app, "GET", "/api/orphans")
	if resp.StatusCode != 200 {
		t.Fatalf("report status = %d (body: %s)", resp.StatusCode, body)
	}

	// Exact union counts observed for the merged master workbook (a link
	// may miss both its function and its definition and is counted once).
	// The team's comm-family expansion resolved every orphan; the report
	// is empty for this fixture.
	for _, want := range []string{
		`"kind":"parameter-links","count":0`,
		`"kind":"criteria-links","count":0`,
		`"kind":"examples","count":0`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("report missing %s (body has counts: %s)", want, body)
		}
	}
}

// TestDictionaryOrphanUsage: a definition referenced only by links whose
// functions are missing reports orphan_usage_count alongside usage_count.
func TestDictionaryOrphanUsage(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// One resolved usage (via the function) + one orphan usage.
	postJSON(t, app, "/api/functions/1/parameter-links", `{"name":"ANT"}`)
	_, err := d.Exec(`INSERT INTO parameter_links
		(function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES ('DLL', 'NO_SUCH_FN', '', 'ANT', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed orphan usage: %v", err)
	}

	_, body := doRequest(t, app, "GET", "/api/parameter-definitions")
	if !strings.Contains(body, `"usage_count":2`) || !strings.Contains(body, `"orphan_usage_count":1`) {
		t.Fatalf("usage counts wrong: %s", body)
	}

	// Single-item get carries the same counts.
	_, body = doRequest(t, app, "GET", "/api/parameter-definitions/1")
	if !strings.Contains(body, `"orphan_usage_count":1`) {
		t.Fatalf("get orphan_usage_count wrong: %s", body)
	}
}
