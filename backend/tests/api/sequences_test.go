package api_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Sequences API tests (ticket 10). The HTTP seam is the primary test seam;
// every assertion is on request → response status/body per spec.

const seqStepA = `{"function_source":"DLL","function_interface":"TELNET","function_name":"FN_A","parameters":{"FREQ":"10"},"criteria":{},"timeout_ms":"1000","flow":"NEXT","on_fail":"FAIL","device_name":"RACK1","note":"first step"}`
const seqStepB = `{"function_source":"DLL","function_interface":"GPIB","function_name":"FN_B","parameters":{},"criteria":{},"timeout_ms":"2000","flow":"QUIT","on_fail":"FAIL","device_name":"","note":""}`

// seedSequenceFunctions creates two functions: FN_A (TELNET, one required
// FREQ parameter whose allowed values are {10,20}) and FN_B (GPIB, no
// links) — through the public API.
func seedSequenceFunctions(t *testing.T, app *fiber.App) {
	t.Helper()
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"TELNET","name":"FN_A"}`); resp.StatusCode != 201 {
		t.Fatalf("seed FN_A: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"definition":{"name":"FREQ","type":"int","allowed_values":"{10,20}","cardinality":"SINGLE"},"required":true,"position":1,"device_name":""}`); resp.StatusCode != 201 {
		t.Fatalf("seed FREQ link: %d (body: %s)", resp.StatusCode, body)
	}
	if resp, body := postJSON(t, app, "/api/functions",
		`{"source":"DLL","interface":"GPIB","name":"FN_B"}`); resp.StatusCode != 201 {
		t.Fatalf("seed FN_B: %d (body: %s)", resp.StatusCode, body)
	}
}

// TestSequenceCRUDRoundTrip: create with ordered steps, list with search,
// get, full-replace PUT (reorder + edit), delete.
func TestSequenceCRUDRoundTrip(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	resp, body := postJSON(t, app, "/api/sequences",
		`{"name":"bringup","description":"d1","steps":[`+seqStepA+`,`+seqStepB+`]}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name":"bringup"`) || !strings.Contains(body, `"function_exists":true`) {
		t.Fatalf("create response wrong: %s", body)
	}

	// GET detail: steps ordered by position, stored values round-trip.
	resp, body = doRequest(t, app, "GET", "/api/sequences/1")
	if resp.StatusCode != 200 ||
		!strings.Contains(body, `"device_name":"RACK1"`) ||
		!strings.Contains(body, `"note":"first step"`) ||
		!strings.Contains(body, `"parameters":{"FREQ":"10"}`) {
		t.Fatalf("get: %d (body: %s)", resp.StatusCode, body)
	}
	firstPos := strings.Index(body, `"function_name":"FN_A"`)
	secondPos := strings.Index(body, `"function_name":"FN_B"`)
	if firstPos == -1 || secondPos == -1 || firstPos > secondPos {
		t.Fatalf("steps out of order: %s", body)
	}

	// List + search.
	resp, body = doRequest(t, app, "GET", "/api/sequences?search=brin")
	if resp.StatusCode != 200 || !strings.Contains(body, `"total":1`) {
		t.Fatalf("list search: %d (body: %s)", resp.StatusCode, body)
	}

	// PUT full-replace: swap order and edit the note; step ids may change,
	// ordering must follow the array.
	putBody := `{"name":"bringup","description":"d2","steps":[` +
		strings.Replace(seqStepB, `"note":""`, `"note":"now first"`, 1) + `,` + seqStepA + `]}`
	resp, body = doJSON(t, app, "PUT", "/api/sequences/1", putBody)
	if resp.StatusCode != 200 || !strings.Contains(body, `"description":"d2"`) {
		t.Fatalf("put: %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = doRequest(t, app, "GET", "/api/sequences/1")
	if strings.Index(body, `"function_name":"FN_B"`) > strings.Index(body, `"function_name":"FN_A"`) {
		t.Fatalf("reorder not applied: %s", body)
	}
	if !strings.Contains(body, `"note":"now first"`) {
		t.Fatalf("note edit lost: %s", body)
	}

	resp, _ = doRequest(t, app, "DELETE", "/api/sequences/1")
	if resp.StatusCode != 204 {
		t.Fatalf("delete: %d", resp.StatusCode)
	}
	resp, _ = doRequest(t, app, "GET", "/api/sequences/1")
	if resp.StatusCode != 404 {
		t.Fatalf("get after delete: %d", resp.StatusCode)
	}
}

// TestSequenceValidation: name required, at least one step required, the
// sequence name is unique, and a failed PUT leaves the stored rows alone.
func TestSequenceValidation(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	resp, body := postJSON(t, app, "/api/sequences", `{"name":"","steps":[]}`)
	if resp.StatusCode != 422 {
		t.Fatalf("empty create: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name"`) || !strings.Contains(body, `"steps"`) {
		t.Fatalf("field errors missing: %s", body)
	}

	if resp, body := postJSON(t, app, "/api/sequences",
		`{"name":"bringup","steps":[`+seqStepA+`]}`); resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	resp, body = postJSON(t, app, "/api/sequences",
		`{"name":"bringup","steps":[`+seqStepB+`]}`)
	if resp.StatusCode != 409 {
		t.Fatalf("duplicate name: %d (body: %s)", resp.StatusCode, body)
	}

	// Orphan steps are storable: no existence check at the API (ADR-0003).
	resp, body = postJSON(t, app, "/api/sequences",
		`{"name":"ghosts","steps":[{"function_source":"DLL","function_interface":"USB","function_name":"MISSING","parameters":{},"criteria":{},"timeout_ms":"1","flow":"NEXT","on_fail":"FAIL","device_name":"","note":""}]}`)
	if resp.StatusCode != 201 {
		t.Fatalf("orphan step create: %d (body: %s)", resp.StatusCode, body)
	}

	// Failed PUT (no steps) must not disturb the stored sequence.
	resp, _ = doJSON(t, app, "PUT", "/api/sequences/1", `{"name":"bringup","steps":[]}`)
	if resp.StatusCode != 422 {
		t.Fatalf("empty put: %d", resp.StatusCode)
	}
	resp, body = doRequest(t, app, "GET", "/api/sequences/1")
	if !strings.Contains(body, `"function_name":"FN_A"`) {
		t.Fatalf("steps changed by failed put: %s", body)
	}
}

// TestSequenceEmitPerStepFlags: emission joins clean steps with newlines;
// an invalid step is flagged with field errors and a missing function with
// resolved=false — clean steps still emit.
func TestSequenceEmitPerStepFlags(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	body := `{"name":"mix","steps":[` +
		seqStepA + `,` +
		strings.Replace(seqStepA, `"FREQ":"10"`, `"FREQ":"30"`, 1) + `,` +
		seqStepB + `,
		{"function_source":"DLL","function_interface":"USB","function_name":"GHOST","parameters":{},"criteria":{},"timeout_ms":"5","flow":"NEXT","on_fail":"FAIL","device_name":"","note":""}]}`
	resp, body := postJSON(t, app, "/api/sequences", body)
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	resp, body = doRequest(t, app, "POST", "/api/sequences/1/emit")
	if resp.StatusCode != 200 {
		t.Fatalf("emit: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"text":"DLLTELNET|FN_A|FREQ=10|NULL|1000|NEXT|FAIL\nDLLGPIB|FN_B||NULL|2000|QUIT|FAIL"`) {
		t.Fatalf("joined emission wrong: %s", body)
	}
	if !strings.Contains(body, `"resolved":true`) || !strings.Contains(body, `"resolved":false`) {
		t.Fatalf("per-step flags missing: %s", body)
	}
	if !strings.Contains(body, "must be one of") {
		t.Fatalf("invalid step not flagged: %s", body)
	}
	if !strings.Contains(body, "GHOST") || !strings.Contains(body, "not found") {
		t.Fatalf("missing function step not flagged: %s", body)
	}
}

// TestSequenceStepOrphanAfterFunctionDelete: deleting a referenced function
// leaves the step in place, flagged unresolved (ADR-0003).
func TestSequenceStepOrphanAfterFunctionDelete(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	if resp, body := postJSON(t, app, "/api/sequences",
		`{"name":"orph","steps":[`+seqStepA+`]}`); resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	if resp, _ := doRequest(t, app, "DELETE", "/api/functions/1"); resp.StatusCode != 204 {
		t.Fatalf("delete function: %d", resp.StatusCode)
	}

	resp, body := doRequest(t, app, "GET", "/api/sequences/1")
	if resp.StatusCode != 200 {
		t.Fatalf("get: %d", resp.StatusCode)
	}
	if !strings.Contains(body, `"function_exists":false`) {
		t.Fatalf("step not flagged orphan: %s", body)
	}

	resp, body = doRequest(t, app, "POST", "/api/sequences/1/emit")
	if resp.StatusCode != 200 {
		t.Fatalf("emit: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"resolved":false`) || !strings.Contains(body, "not found") {
		t.Fatalf("emit did not flag orphan step: %s", body)
	}
	if strings.Contains(body, `"text":"DLLTELNET`) {
		t.Fatalf("orphan step must not emit: %s", body)
	}
}

// TestSequenceEmitRejectsSegmentDelimiters: a stored step whose Flow or
// On-fail holds '|' is flagged instead of emitting an extra segment; clean
// steps still emit.
func TestSequenceEmitRejectsSegmentDelimiters(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	body := `{"name":"pipes","steps":[` +
		seqStepA + `,` +
		strings.Replace(seqStepA, `"flow":"NEXT"`, `"flow":"NEXT|EXTRA"`, 1) + `,` +
		strings.Replace(seqStepB, `"on_fail":"FAIL"`, `"on_fail":"FAIL|EXTRA"`, 1) + `]}`
	if resp, body := postJSON(t, app, "/api/sequences", body); resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	resp, body := doRequest(t, app, "POST", "/api/sequences/1/emit")
	if resp.StatusCode != 200 {
		t.Fatalf("emit: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"text":"DLLTELNET|FN_A|FREQ=10|NULL|1000|NEXT|FAIL"`) {
		t.Fatalf("joined emission wrong: %s", body)
	}
	for _, want := range []string{
		`"flow":"must not contain '|' or a line break"`,
		`"on_fail":"must not contain '|' or a line break"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("step not flagged with %s: %s", want, body)
		}
	}
}

// TestResetLeavesSequencesUntouched: the workbook reset truncates the six
// entity tables only (ADR-0005); stored sequences survive.
func TestResetLeavesSequencesUntouched(t *testing.T) {
	app, _ := newLinksTestApp(t)

	// Orphan step: no function exists yet, and the workbook won't create
	// one — the sequence must survive reset regardless.
	resp, body := postJSON(t, app, "/api/sequences",
		`{"name":"web-native","steps":[`+seqStepA+`]}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create: %d (body: %s)", resp.StatusCode, body)
	}

	ct, upload := multipartBody(t, "file", "wb.xlsx", fixtureBytes(t))
	resp2, body2 := postFile(t, app, "/api/import/reset", ct, upload)
	if resp2.StatusCode != 200 {
		t.Fatalf("reset: %d (body: %s)", resp2.StatusCode, body2)
	}
	resp, body = doRequest(t, app, "GET", "/api/sequences/1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"name":"web-native"`) {
		t.Fatalf("sequence lost after reset: %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"timeout_ms":"1000"`) {
		t.Fatalf("step lost after reset: %s", body)
	}
}

func TestSequenceListDefaultPaging(t *testing.T) {
	app, _ := newLinksTestApp(t)
	check := func(query string, page, total int64, want ...[2]int64) {
		t.Helper()
		resp, body := doRequest(t, app, "GET", "/api/sequences?"+query)
		var got struct {
			Items []struct {
				ID        int64 `json:"id"`
				StepCount int64 `json:"step_count"`
			} `json:"items"`
			Total    int64 `json:"total"`
			Page     int64 `json:"page"`
			PageSize int64 `json:"page_size"`
		}
		err := json.Unmarshal([]byte(body), &got)
		if resp.StatusCode != 200 || err != nil || got.Items == nil ||
			got.Total != total || got.Page != page || got.PageSize != 1 {
			t.Fatalf("%s: unexpected response: %d %s (decode: %v)", query, resp.StatusCode, body, err)
		}
		var items [][2]int64
		for _, r := range got.Items {
			items = append(items, [2]int64{r.ID, r.StepCount})
		}
		if !slices.Equal(items, want) {
			t.Errorf("%s: (id, step_count) = %v, want %v", query, items, want)
		}
	}
	check("page_size=1", 1, 0)
	seedSequenceFunctions(t, app)
	for _, body := range []string{
		`{"name":"Bravo","description":"needle","steps":[` + seqStepB + `]}`,
		`{"name":"Alpha","description":"elsewhere","steps":[` + seqStepB + `]}`,
		`{"name":"Zulu","description":"NEEDLE","steps":[` + seqStepB + `,` + seqStepB + `]}`,
	} {
		if resp, response := postJSON(t, app, "/api/sequences", body); resp.StatusCode != 201 {
			t.Fatalf("create: %d (body: %s)", resp.StatusCode, response)
		}
	}
	check("page_size=1", 1, 3, [2]int64{2, 1})
	check("page_size=1&page=2", 2, 3, [2]int64{1, 1})
	check("page_size=1&page=3", 3, 3, [2]int64{3, 2})
	check("page_size=1&page=2&search=Needle", 2, 2, [2]int64{3, 2})
	check("page_size=1&page=3&search=Needle", 3, 2)

	for _, sort := range []string{"", "&sort=name"} {
		resp, body := doRequest(t, app, "GET", "/api/sequences?page=9223372036854775807"+sort)
		if resp.StatusCode != 200 || !strings.Contains(body, `"items":[]`) {
			t.Errorf("large page%s: want an empty page, got %d %s", sort, resp.StatusCode, body)
		}
	}
}

// TestSequenceListSortsBeforePaging: sort/order apply to the whole
// filtered list before the page is cut, so page 1 of a Steps-descending
// sort holds the longest sequence even when its name sorts last.
func TestSequenceListSortsBeforePaging(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedSequenceFunctions(t, app)

	for _, seq := range []struct {
		name  string
		steps int
	}{{"alpha", 1}, {"Bravo", 2}, {"charlie", 3}} {
		steps := strings.TrimSuffix(strings.Repeat(seqStepB+",", seq.steps), ",")
		if resp, body := postJSON(t, app, "/api/sequences",
			`{"name":"`+seq.name+`","steps":[`+steps+`]}`); resp.StatusCode != 201 {
			t.Fatalf("create %s: %d (body: %s)", seq.name, resp.StatusCode, body)
		}
	}

	for _, tc := range []struct {
		query string
		want  string
	}{
		{"sort=steps&order=desc&page_size=1", "charlie"},
		{"sort=steps&order=desc&page_size=1&page=2", "Bravo"},
		{"sort=steps&order=asc&page_size=1", "alpha"},
		{"sort=name&order=desc&page_size=1", "charlie"},
		{"sort=name&order=desc&page_size=1&page=2", "Bravo"},
		{"sort=bogus&order=desc&page_size=1", "alpha"},
	} {
		resp, body := doRequest(t, app, "GET", "/api/sequences?"+tc.query)
		if resp.StatusCode != 200 || !strings.Contains(body, `"total":3`) ||
			!strings.Contains(body, `"name":"`+tc.want+`"`) {
			t.Errorf("%s: want %s on the page: %d (body: %s)", tc.query, tc.want, resp.StatusCode, body)
		}
	}
}
