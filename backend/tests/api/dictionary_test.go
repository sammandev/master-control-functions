package api_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"webprogfndb/internal/api"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

func newDictionaryTestApp(t *testing.T) (*fiber.App, *sql.DB) {
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

func validParameterDefinitionJSON() string {
	return `{
		"name": "ANT",
		"description": "Specifies which antenna(s) to test",
		"type": "int",
		"unit": "",
		"default_value": "1",
		"allowed_values": "{1,2,3,4,7,8,15}",
		"cardinality": "MULTIPLE",
		"value_pattern": "<ant1>,<ant2>,...",
		"required": true
	}`
}

func TestCreateParameterDefinitionRoundTrip(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	resp, body := postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON())
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d, want 201 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name":"ANT"`) || !strings.Contains(body, `"required":true`) {
		t.Fatalf("create body wrong: %s", body)
	}

	resp, body = doRequest(t, app, "GET", "/api/parameter-definitions/1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"cardinality":"MULTIPLE"`) {
		t.Fatalf("get failed: %d %s", resp.StatusCode, body)
	}
}

func TestParameterDefinitionValidation(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	cases := []struct {
		name   string
		body   string
		fields []string // fields that must carry error messages
	}{
		{
			name:   "missing name",
			body:   `{"type": "int"}`,
			fields: []string{"name"},
		},
		{
			name:   "unknown type",
			body:   `{"name": "X", "type": "float64"}`,
			fields: []string{"type"},
		},
		{
			name:   "unknown cardinality",
			body:   `{"name": "X", "cardinality": "MANY"}`,
			fields: []string{"cardinality"},
		},
		{
			name:   "malformed allowed_values set",
			body:   `{"name": "X", "type": "int", "allowed_values": "1,2,3"}`,
			fields: []string{"allowed_values"},
		},
		{
			name:   "allowed_values items violate type",
			body:   `{"name": "X", "type": "int", "allowed_values": "{10,20,abc}"}`,
			fields: []string{"allowed_values"},
		},
		{
			name:   "default not in allowed set",
			body:   `{"name": "X", "type": "int", "allowed_values": "{1,2}", "default_value": "9"}`,
			fields: []string{"default_value"},
		},
		{
			name:   "default violates type",
			body:   `{"name": "X", "type": "int", "default_value": "high"}`,
			fields: []string{"default_value"},
		},
	}
	for _, tc := range cases {
		resp, body := postJSON(t, app, "/api/parameter-definitions", tc.body)
		if resp.StatusCode != 422 {
			t.Fatalf("%s: status = %d, want 422 (body: %s)", tc.name, resp.StatusCode, body)
		}
		var parsed struct {
			Fields map[string]string `json:"fields"`
		}
		if err := json.Unmarshal([]byte(body), &parsed); err != nil {
			t.Fatalf("%s: body not JSON: %s", tc.name, body)
		}
		for _, f := range tc.fields {
			if _, ok := parsed.Fields[f]; !ok {
				t.Fatalf("%s: no field error for %s (got %v)", tc.name, f, parsed.Fields)
			}
		}
	}

	// Valid allowed_values with list[int] type must parse.
	ok := `{"name": "BW", "type": "list[int]", "allowed_values": "{20,40,80}", "default_value": "20"}`
	if resp, body := postJSON(t, app, "/api/parameter-definitions", ok); resp.StatusCode != 201 {
		t.Fatalf("valid list[int] rejected: %d %s", resp.StatusCode, body)
	}
}

func TestParameterDefinitionDuplicateNameRejected(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	if resp, _ := postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON()); resp.StatusCode != 201 {
		t.Fatalf("first create failed")
	}
	resp, body := postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON())
	if resp.StatusCode != 409 {
		t.Fatalf("status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, "already exists") {
		t.Fatalf("conflict message unclear: %s", body)
	}
}

func TestParameterDefinitionUpdateAndNameConflict(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	a := `{"name": "ANT", "type": "int", "allowed_values": "{1,2}"}`
	if resp, body := postJSON(t, app, "/api/parameter-definitions", a); resp.StatusCode != 201 {
		t.Fatalf("create A failed: %s", body)
	}
	b := `{"name": "BW", "type": "int", "allowed_values": "{20,40}"}`
	if resp, body := postJSON(t, app, "/api/parameter-definitions", b); resp.StatusCode != 201 {
		t.Fatalf("create B failed: %s", body)
	}

	update := `{"name": "ANT2", "type": "double", "default_value": "1.5"}`
	resp, body := doJSON(t, app, "PUT", "/api/parameter-definitions/1", update)
	if resp.StatusCode != 200 || !strings.Contains(body, `"type":"double"`) {
		t.Fatalf("update failed: %d %s", resp.StatusCode, body)
	}

	// Rename onto an existing name → 409.
	resp, body = doJSON(t, app, "PUT", "/api/parameter-definitions/1", b)
	if resp.StatusCode != 409 {
		t.Fatalf("name conflict status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
}

// Scope semantics on update (ADR-0009): the PUT body is a full replace —
// a device_name in the body sets the scope, an absent one downgrades the
// definition to Global. The dictionary modal (DictionaryPage.svelte) always
// sends the field, so an ordinary edit never silently strips a scope.
func TestParameterDefinitionUpdateScopeSemantics(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	scoped := `{"name":"ANT","type":"string","cardinality":"SINGLE","device_name":"WIFI_TEST"}`
	if resp, body := postJSON(t, app, "/api/parameter-definitions", scoped); resp.StatusCode != 201 {
		t.Fatalf("create scoped definition failed: %s", body)
	}

	var got struct {
		DeviceName string `json:"device_name"`
	}

	// Editing other fields while sending the same scope keeps it.
	update := `{"name":"ANT","type":"string","cardinality":"SINGLE","device_name":"WIFI_TEST","description":"edited"}`
	resp, body := doJSON(t, app, "PUT", "/api/parameter-definitions/1", update)
	if resp.StatusCode != 200 {
		t.Fatalf("update with scope failed: %d %s", resp.StatusCode, body)
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if got.DeviceName != "WIFI_TEST" {
		t.Fatalf("scope = %q, want WIFI_TEST (body: %s)", got.DeviceName, body)
	}

	// The scope itself is editable: a different device_name moves the row.
	resp, body = doJSON(t, app, "PUT", "/api/parameter-definitions/1",
		`{"name":"ANT","type":"string","cardinality":"SINGLE","device_name":"OSP"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("scope move failed: %d %s", resp.StatusCode, body)
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode scope-move response: %v", err)
	}
	if got.DeviceName != "OSP" {
		t.Fatalf("scope = %q, want OSP (body: %s)", got.DeviceName, body)
	}

	// Absent device_name is a full replace → Global. Explicit downgrade,
	// the contract the fixed modal no longer hits by accident.
	resp, body = doJSON(t, app, "PUT", "/api/parameter-definitions/1",
		`{"name":"ANT","type":"string","cardinality":"SINGLE"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("global downgrade failed: %d %s", resp.StatusCode, body)
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode downgrade response: %v", err)
	}
	if got.DeviceName != "" {
		t.Fatalf("scope = %q, want empty after full-replace update", got.DeviceName)
	}
}

func TestDictionaryListPaginationAndSearch(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	for _, n := range []string{"ALWAYS_AGC", "APT", "AUTO_CHANGE_CAL_FREQ", "AVERAGES", "BW"} {
		body := `{"name": "` + n + `", "type": "int", "description": "test param"}`
		if resp, _ := postJSON(t, app, "/api/parameter-definitions", body); resp.StatusCode != 201 {
			t.Fatalf("seed %s failed", n)
		}
	}

	resp, body := doRequest(t, app, "GET", "/api/parameter-definitions?page=1&page_size=2&sort=name&order=asc")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"total":5`) || !strings.Contains(body, `"page":1`) {
		t.Fatalf("paging metadata wrong: %s", body)
	}
	// Name-asc: ALWAYS_AGC, APT on page 1.
	if !strings.Contains(body, "ALWAYS_AGC") || !strings.Contains(body, "APT") {
		t.Fatalf("page 1 wrong rows: %s", body)
	}

	// Search on description: only AVERAGES contains "AVERAGES" in name.
	_, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=averages")
	if !strings.Contains(body, `"total":1`) || !strings.Contains(body, "AVERAGES") {
		t.Fatalf("search wrong: %s", body)
	}
}

func TestDictionaryUsageCountAndDeleteGuard(t *testing.T) {
	app, d := newDictionaryTestApp(t)

	if resp, body := postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON()); resp.StatusCode != 201 {
		t.Fatalf("create failed: %s", body)
	}

	// Seed a link row referencing ANT (fixture setup: link endpoints land
	// in ticket 04; the usage count is what this ticket must get right).
	_, err := d.Exec(`INSERT INTO parameter_links
		(function_source, function_name, device_name, name, position, required, created_at, updated_at)
		VALUES ('DLL', 'TX_MULTI_VERIFICATION', 'WIFI_TEST', 'ANT', 1, 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed link: %v", err)
	}

	_, body := doRequest(t, app, "GET", "/api/parameter-definitions")
	if !strings.Contains(body, `"usage_count":1`) {
		t.Fatalf("usage_count wrong: %s", body)
	}

	// Delete with usage → 409 with the count, not silent breakage.
	resp, body := doRequest(t, app, "DELETE", "/api/parameter-definitions/1")
	if resp.StatusCode != 409 {
		t.Fatalf("delete status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"usage_count":1`) {
		t.Fatalf("conflict should carry usage_count: %s", body)
	}

	// Forced delete succeeds; the link row survives as an Orphan (ADR-0003).
	resp, body = doRequest(t, app, "DELETE", "/api/parameter-definitions/1?force=true")
	if resp.StatusCode != 204 {
		t.Fatalf("forced delete status = %d (body: %s)", resp.StatusCode, body)
	}
	var count int
	if err := d.QueryRow(`SELECT COUNT(*) FROM parameter_links WHERE name = 'ANT'`).Scan(&count); err != nil {
		t.Fatalf("count links: %v", err)
	}
	if count != 1 {
		t.Fatalf("link row must survive definition deletion, got %d", count)
	}
}

func TestCriteriaDefinitionCRUD(t *testing.T) {
	app, _ := newDictionaryTestApp(t)

	create := `{
		"name": "EXPECTED_RESPONSE",
		"description": "The expected response after sending a command",
		"type": "string",
		"cardinality": "SINGLE",
		"required": true
	}`
	resp, body := postJSON(t, app, "/api/criteria-definitions", create)
	if resp.StatusCode != 201 {
		t.Fatalf("create status = %d (body: %s)", resp.StatusCode, body)
	}

	update := `{"name": "EXPECTED_RESPONSE", "type": "string", "description": "updated"}`
	resp, body = doJSON(t, app, "PUT", "/api/criteria-definitions/1", update)
	if resp.StatusCode != 200 || !strings.Contains(body, `"description":"updated"`) {
		t.Fatalf("update failed: %d %s", resp.StatusCode, body)
	}

	resp, body = doRequest(t, app, "DELETE", "/api/criteria-definitions/1")
	if resp.StatusCode != 204 {
		t.Fatalf("delete status = %d (body: %s)", resp.StatusCode, body)
	}

	resp, _ = doRequest(t, app, "GET", "/api/criteria-definitions/1")
	if resp.StatusCode != 404 {
		t.Fatalf("get after delete = %d, want 404", resp.StatusCode)
	}
}

func TestDictionarySortUnknownColumnDoesNotPanic(t *testing.T) {
	// Regression: the dictionary sort whitelist used to share
	// functions.go's sortColumns, so sort=interface passed the check but
	// had no per-dictionary key function and nil-panicked.
	app, _ := newDictionaryTestApp(t)

	body := `{"name": "ANT", "type": "int", "cardinality": "SINGLE"}`
	if resp, _ := postJSON(t, app, "/api/parameter-definitions", body); resp.StatusCode != 201 {
		t.Fatalf("seed failed")
	}

	for _, q := range []string{"sort=interface", "sort=source", "sort=type", "sort=updated_at&order=desc"} {
		resp, body := doRequest(t, app, "GET", "/api/parameter-definitions?"+q)
		if resp.StatusCode != 200 {
			t.Fatalf("GET %s: status = %d, want 200 (body: %s)", q, resp.StatusCode, body)
		}
	}
}

func TestCriteriaDefinitionValidation(t *testing.T) {
	// Criteria share the parameter validation code path; pin one 422 case
	// per dictionary so the parity is tested, not assumed.
	app, _ := newDictionaryTestApp(t)

	resp, body := postJSON(t, app, "/api/criteria-definitions",
		`{"name": "EVM", "type": "double", "allowed_values": "{-35,abc}", "cardinality": "SINGLE"}`)
	if resp.StatusCode != 422 {
		t.Fatalf("status = %d, want 422 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"allowed_values"`) {
		t.Fatalf("field error for allowed_values missing: %s", body)
	}
}

// TestScopeAwareUsageCount: ADR-0009 makes usage_count scope-aware: a
// scoped definition counts links of its own scope only; a global
// definition counts links whose device scope has no scoped sibling
// (i.e. the links that actually resolve to it under
// exact-scope-then-global). The two ANT rows must report different
// counts when a WIFI_TEST-scoped sibling exists.
func TestScopeAwareUsageCount(t *testing.T) {
	app, d := newDictionaryTestApp(t)

	// Seed one function (the link endpoint requires it).
	if resp, body := postJSON(t, app, "/api/functions", `{"source":"DLL","interface":"WIFI_TEST","name":"TX_MULTI_VERIFICATION"}`); resp.StatusCode != 201 {
		t.Fatalf("function: %s", body)
	}

	// Seed the global ANT row via the dictionary API.
	if resp, body := postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON()); resp.StatusCode != 201 {
		t.Fatalf("global ANT: %s", body)
	}
	// Seed the WIFI_TEST-scoped ANT sibling via the dictionary API so
	// we exercise the validation path too (device_name passes through).
	scoped := `{"name":"ANT","type":"string","cardinality":"SINGLE","device_name":"WIFI_TEST"}`
	if resp, body := postJSON(t, app, "/api/parameter-definitions", scoped); resp.StatusCode != 201 {
		t.Fatalf("scoped ANT: %s", body)
	}

	// Three links: one global, one WIFI_TEST, one IOT. Without the
	// scope-aware count, all three would inflate every ANT row's
	// usage_count; with it, the scoped WIFI_TEST sibling wins for
	// the WIFI_TEST link, so only the IOT link (no scoped sibling)
	// is left for the global ANT.
	links := []struct{ device string }{
		{"WIFI_TEST"},
		{""},
		{"IOT"},
	}
	for _, l := range links {
		body := `{"name":"ANT","device_name":"` + l.device + `","position":1}`
		if resp, b := postJSON(t, app, "/api/functions/1/parameter-links", body); resp.StatusCode != 201 {
			t.Fatalf("link %q: %s", l.device, b)
		}
	}

	// The list endpoint surfaces both rows with their scope-aware
	// usage_count: global ANT excludes the WIFI_TEST link (scoped
	// sibling wins); scoped ANT counts only its own scope's link.
	resp, list := doRequest(t, app, "GET", "/api/parameter-definitions?search=ANT")
	if resp.StatusCode != 200 {
		t.Fatalf("list: %d (body: %s)", resp.StatusCode, list)
	}

	// Both rows must be present with their scope and count. Substring
	// assertions are robust to map-key ordering (Go encoding/json sorts
	// fiber.Map keys alphabetically, so device_name / id / name are
	// not adjacent to usage_count).
	var parsed []struct {
		ID         int64  `json:"id"`
		DeviceName string `json:"device_name"`
		Name       string `json:"name"`
		UsageCount int64  `json:"usage_count"`
	}
	if err := json.Unmarshal(extractItems(t, list), &parsed); err != nil {
		t.Fatalf("parse list: %v (body: %s)", err, list)
	}
	if len(parsed) != 2 {
		t.Fatalf("ANT rows = %d, want 2 (body: %s)", len(parsed), list)
	}
	want := map[string]int64{"": 2, "WIFI_TEST": 1}
	for _, row := range parsed {
		if got, ok := want[row.DeviceName]; !ok || got != row.UsageCount {
			t.Fatalf("%s/%s usage_count = %d, want %d (full: %s)",
				row.Name, row.DeviceName, row.UsageCount, want[row.DeviceName], list)
		}
	}

	// The same scope-aware count must be reflected in the single-item
	// GET endpoint (otherwise the list/GET surfaces diverge).
	for _, want := range []struct{ id, count int64 }{
		{1, 2}, // global ANT
		{2, 1}, // scoped ANT
	} {
		resp, body := doRequest(t, app, "GET", fmt.Sprintf("/api/parameter-definitions/%d", want.id))
		if resp.StatusCode != 200 {
			t.Fatalf("get id %d: %d %s", want.id, resp.StatusCode, body)
		}
		var got struct {
			DeviceName string `json:"device_name"`
			UsageCount int64  `json:"usage_count"`
		}
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatalf("parse id %d: %v", want.id, err)
		}
		if got.UsageCount != want.count {
			t.Fatalf("id %d (%s) usage_count = %d, want %d",
				want.id, got.DeviceName, got.UsageCount, want.count)
		}
	}

	// Removing the WIFI_TEST-scoped ANT sibling migrates the
	// WIFI_TEST link's resolution to global. The global row's count
	// must then include that link (3), and the scoped row's count
	// drops to 0.
	if _, err := d.Exec(`DELETE FROM parameter_definitions WHERE id = 2`); err != nil {
		t.Fatalf("delete scoped ANT: %v", err)
	}
	resp, list2 := doRequest(t, app, "GET", "/api/parameter-definitions?search=ANT")
	if resp.StatusCode != 200 {
		t.Fatalf("list2: %d", resp.StatusCode)
	}
	var rows []struct {
		DeviceName string `json:"device_name"`
		UsageCount int64  `json:"usage_count"`
	}
	if err := json.Unmarshal(extractItems(t, list2), &rows); err != nil {
		t.Fatalf("parse list2: %v", err)
	}
	if len(rows) != 1 || rows[0].DeviceName != "" || rows[0].UsageCount != 3 {
		t.Fatalf("after scoped deleted, want only global ANT usage=3, got %+v", rows)
	}
}

// extractItems pulls the "items" array out of a list response so the
// caller can JSON-decode a typed slice.
func extractItems(t *testing.T, body string) []byte {
	t.Helper()
	var wrapped struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal([]byte(body), &wrapped); err != nil {
		t.Fatalf("extract items: %v", err)
	}
	return []byte(wrapped.Items)
}
