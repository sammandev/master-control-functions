package api_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
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

func TestDictionaryListDeviceScope(t *testing.T) {
	type definition struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		DeviceName string `json:"device_name"`
	}
	for _, endpoint := range []string{"/api/parameter-definitions", "/api/criteria-definitions"} {
		t.Run(endpoint, func(t *testing.T) {
			app, _ := newDictionaryTestApp(t)
			rows := []definition{
				{Name: "B_ANT", DeviceName: "WIFI_TEST"},
				{Name: "A_ANT"},
				{Name: "A_ANT", DeviceName: "WIFI_TEST"},
				{Name: "A_ANT", DeviceName: "GOLDENKV"},
				{Name: "C_BW", DeviceName: "WIFI_TEST"},
			}
			for i, row := range rows {
				payload := fmt.Sprintf(`{"name":%q,"device_name":%q,"type":"int","cardinality":"SINGLE"}`, row.Name, row.DeviceName)
				resp, body := postJSON(t, app, endpoint, payload)
				if resp.StatusCode != 201 {
					t.Fatalf("seed %s (%s): %d %s", row.Name, row.DeviceName, resp.StatusCode, body)
				}
				if err := json.Unmarshal([]byte(body), &rows[i]); err != nil {
					t.Fatalf("decode seeded definition: %v", err)
				}
			}
			base := endpoint + "?sort=name&page_size=2"
			for _, tc := range []struct {
				name, query string
				page, total int
				items       []definition
			}{
				{"all devices", "", 1, 5, []definition{rows[1], rows[2]}},
				{"scoped first page", "&device=WIFI_TEST", 1, 3, []definition{rows[2], rows[0]}},
				{"scoped last page", "&device=WIFI_TEST&page=2", 2, 3, []definition{rows[4]}},
				{"search within scope", "&device=WIFI_TEST&search=ant", 1, 2, []definition{rows[2], rows[0]}},
				{"global only", "&device=", 1, 1, []definition{rows[1]}},
				{"no matches", "&device=UNUSED", 1, 0, []definition{}},
				{"locate in scope", fmt.Sprintf("&device=WIFI_TEST&definition_id=%d", rows[4].ID), 2, 3, []definition{rows[4]}},
				{"cannot locate outside scope", fmt.Sprintf("&device=WIFI_TEST&page=2&definition_id=%d", rows[1].ID), 2, 3, []definition{rows[4]}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					resp, body := doRequest(t, app, "GET", base+tc.query)
					if resp.StatusCode != 200 {
						t.Fatalf("list: %d %s", resp.StatusCode, body)
					}
					var got struct {
						Items    []definition `json:"items"`
						Page     int          `json:"page"`
						PageSize int          `json:"page_size"`
						Total    int          `json:"total"`
					}
					if err := json.Unmarshal([]byte(body), &got); err != nil {
						t.Fatalf("decode list: %v", err)
					}
					if got.Page != tc.page || got.PageSize != 2 || got.Total != tc.total ||
						got.Items == nil || !slices.Equal(got.Items, tc.items) {
						t.Fatalf("want page=%d page_size=2 total=%d items=%+v, got %s", tc.page, tc.total, tc.items, body)
					}
				})
			}
		})
	}
}

func TestDictionaryListLocatesRowByID(t *testing.T) {
	for _, endpoint := range []string{"/api/parameter-definitions", "/api/criteria-definitions"} {
		t.Run(endpoint, func(t *testing.T) {
			app, _ := newDictionaryTestApp(t)
			var lastID int64
			for _, n := range []string{"ALWAYS_AGC", "APT", "AUTO_CHANGE_CAL_FREQ", "AVERAGES", "BW"} {
				resp, body := postJSON(t, app, endpoint, `{"name": "`+n+`", "type": "int", "cardinality": "SINGLE"}`)
				if resp.StatusCode != 201 {
					t.Fatalf("seed %s failed: %s", n, body)
				}
				var row struct {
					ID int64 `json:"id"`
				}
				if err := json.Unmarshal([]byte(body), &row); err != nil {
					t.Fatalf("decode %s: %v", n, err)
				}
				lastID = row.ID
			}

			// Name-asc puts BW fifth: page 3 at page_size 2.
			resp, body := doRequest(t, app, "GET", fmt.Sprintf("%s?page=1&page_size=2&sort=name&order=asc&definition_id=%d", endpoint, lastID))
			if resp.StatusCode != 200 {
				t.Fatalf("list status = %d (body: %s)", resp.StatusCode, body)
			}
			if !strings.Contains(body, `"page":3`) || !strings.Contains(body, `"name":"BW"`) {
				t.Fatalf("definition_id should open the page holding the row: %s", body)
			}

			_, body = doRequest(t, app, "GET", endpoint+"?page=2&page_size=2&sort=name&order=asc&definition_id=999999")
			if !strings.Contains(body, `"page":2`) || !strings.Contains(body, `"name":"AUTO_CHANGE_CAL_FREQ"`) {
				t.Fatalf("an unknown definition_id should keep the requested page: %s", body)
			}
		})
	}
}

func TestDictionaryPageSortOrder(t *testing.T) {
	for _, kind := range []string{"parameter", "criteria"} {
		t.Run(kind, func(t *testing.T) {
			app, d := newDictionaryTestApp(t)
			if _, err := d.Exec(fmt.Sprintf(`
				INSERT INTO %s_definitions (name, type, device_name, description, created_at, updated_at)
				VALUES ('a', 'int', '', '', '', '2'), ('Z', 'double', '', 'path\q', '', '1'),
				       ('Z', 'int', 'WIFI_TEST', 'pathq', '', '2'), ('Z', 'int', 'IPLAS', '', '', '2'),
				       ('é', 'double', '', '', '', '1'), ('K10', 'int', '', '', '', '1'),
				       ('K2', 'int', '', '', '', '2')`, kind)); err != nil {
				t.Fatal(err)
			}

			endpoint := "/api/" + kind + "-definitions?"
			check := func(t *testing.T, query string, page, size, total int, ids []int64) {
				t.Helper()
				resp, body := doRequest(t, app, "GET", endpoint+query)
				var got struct {
					Page     int `json:"page"`
					PageSize int `json:"page_size"`
					Total    int `json:"total"`
					Items    []struct {
						ID int64 `json:"id"`
					} `json:"items"`
				}
				if err := json.Unmarshal([]byte(body), &got); err != nil {
					t.Fatal(err)
				}
				gotIDs := make([]int64, 0, len(got.Items))
				for _, item := range got.Items {
					gotIDs = append(gotIDs, item.ID)
				}
				if resp.StatusCode != 200 || got.Page != page || got.PageSize != size ||
					got.Total != total || got.Items == nil || !slices.Equal(gotIDs, ids) {
					t.Fatalf("%s: want page=%d size=%d total=%d ids=%v, got %d %s",
						query, page, size, total, ids, resp.StatusCode, body)
				}
			}

			for _, tc := range []struct {
				sort, order string
				ids         []int64
			}{
				{"name", "asc", []int64{6, 7, 2, 3, 4, 1, 5}},
				{"name", "desc", []int64{5, 1, 2, 3, 4, 7, 6}},
				{"type", "asc", []int64{2, 5, 6, 7, 3, 4, 1}},
				{"type", "desc", []int64{6, 7, 3, 4, 1, 2, 5}},
				{"updated_at", "asc", []int64{6, 2, 5, 7, 3, 4, 1}},
				{"updated_at", "desc", []int64{7, 3, 4, 1, 6, 2, 5}},
				{"unsupported", "desc", []int64{1, 2, 3, 4, 5, 6, 7}},
			} {
				t.Run(tc.sort+"/"+tc.order, func(t *testing.T) {
					order := fmt.Sprintf("sort=%s&order=%s&page_size=2", tc.sort, tc.order)
					for i := 0; i < len(tc.ids); i += 2 {
						page := i/2 + 1
						ids := tc.ids[i:min(i+2, len(tc.ids))]
						check(t, fmt.Sprintf("%s&page=%d", order, page), page, 2, 7, ids)
						check(t, fmt.Sprintf("%s&definition_id=%d", order, ids[len(ids)-1]), page, 2, 7, ids)
					}
				})
			}
			check(t, "page_size=2&sort=name&page=9223372036854775807", 9223372036854775807, 2, 7, nil)
			check(t, "page_size=1&search=%25Pa_h%5CQ%25&sort=type&order=desc&definition_id=3&page=2", 2, 1, 1, nil)
			check(t, "page_size=1&search=%25Pa_h%5CQ%25&sort=type&order=desc&definition_id=2&page=2", 1, 1, 1, []int64{2})
			check(t, "page_size=2&search=%25&sort=name", 1, 2, 7, []int64{6, 7})
		})
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

func TestDictionarySortByType(t *testing.T) {
	// Regression: parseListQuery whitelisted sort against functions.go's
	// sortColumns, which has no "type", so sort=type silently fell back
	// to name order and both directions returned the same rows.
	app, _ := newDictionaryTestApp(t)

	for _, base := range []string{"/api/parameter-definitions", "/api/criteria-definitions"} {
		for _, def := range []string{
			`{"name": "ALPHA", "type": "string", "cardinality": "SINGLE"}`,
			`{"name": "BRAVO", "type": "double", "cardinality": "SINGLE"}`,
			`{"name": "CHARLIE", "type": "int", "cardinality": "SINGLE"}`,
		} {
			if resp, body := postJSON(t, app, base, def); resp.StatusCode != 201 {
				t.Fatalf("seed %s: %d %s", base, resp.StatusCode, body)
			}
		}
		for order, want := range map[string][]string{
			"asc":  {"BRAVO", "CHARLIE", "ALPHA"},
			"desc": {"ALPHA", "CHARLIE", "BRAVO"},
		} {
			resp, body := doRequest(t, app, "GET", base+"?sort=type&order="+order)
			if resp.StatusCode != 200 {
				t.Fatalf("GET %s order=%s: status = %d (body: %s)", base, order, resp.StatusCode, body)
			}
			var list struct {
				Items []struct {
					Name string `json:"name"`
				} `json:"items"`
			}
			if err := json.Unmarshal([]byte(body), &list); err != nil {
				t.Fatalf("decode: %v", err)
			}
			got := make([]string, 0, len(list.Items))
			for _, item := range list.Items {
				got = append(got, item.Name)
			}
			if !slices.Equal(got, want) {
				t.Errorf("GET %s?sort=type&order=%s names = %v, want %v", base, order, got, want)
			}
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

func TestDictionaryPageUsageCounts(t *testing.T) {
	for _, kind := range []string{"parameter", "criteria"} {
		t.Run(kind, func(t *testing.T) {
			app, d := newDictionaryTestApp(t)
			if _, err := d.Exec(`
				INSERT INTO functions (source, interface, name, created_at, updated_at)
				VALUES ('DLL', 'WIFI_TEST', 'PRESENT', '', ''),
				       ('DLL', 'IOT', 'PRESENT', '', ''),
				       ('EXT', 'IOT', 'MISSING', '', '')`); err != nil {
				t.Fatal(err)
			}
			if _, err := d.Exec(fmt.Sprintf(`
				INSERT INTO %s_definitions (name, device_name, created_at, updated_at)
				VALUES ('SHARED', '', '', ''), ('SHARED', 'WIFI_TEST', '', ''), ('UNUSED', '', '', '')`, kind)); err != nil {
				t.Fatal(err)
			}
			if _, err := d.Exec(fmt.Sprintf(`
				INSERT INTO %s_links (function_source, function_name, device_name, name, created_at, updated_at)
				VALUES ('DLL', 'PRESENT', '', 'SHARED', '', ''),
				       ('DLL', 'PRESENT', 'IOT', 'SHARED', '', ''),
				       ('DLL', 'MISSING', 'WIFI_TEST', 'SHARED', '', '')`, kind)); err != nil {
				t.Fatal(err)
			}

			type item struct {
				Name   string `json:"name"`
				Device string `json:"device_name"`
				Usage  int64  `json:"usage_count"`
				Orphan int64  `json:"orphan_usage_count"`
			}
			for _, tc := range []struct {
				query string
				want  []item
			}{
				{"?device=&page_size=2", []item{{"SHARED", "", 2, 1}, {"UNUSED", "", 0, 0}}},
				{"?search=SHARED&page=2&page_size=1", []item{{"SHARED", "WIFI_TEST", 1, 1}}},
				{"?search=SHARED&page=3&page_size=1", []item{}},
			} {
				resp, body := doRequest(t, app, "GET", "/api/"+kind+"-definitions"+tc.query)
				if resp.StatusCode != 200 {
					t.Fatalf("%s: status = %d: %s", tc.query, resp.StatusCode, body)
				}
				var got []item
				if err := json.Unmarshal(extractItems(t, body), &got); err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(got, tc.want) {
					t.Errorf("%s: got %+v, want %+v", tc.query, got, tc.want)
				}
			}
		})
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
