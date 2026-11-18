package api_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// seedFilterFunctions creates four functions across distinct enum values
// (two share the source DLL) through the API.
func seedFilterFunctions(t *testing.T, app *fiber.App) {
	t.Helper()
	post := func(body string) {
		if resp, b := postJSON(t, app, "/api/functions", body); resp.StatusCode != 201 {
			t.Fatalf("seed %s: %d (body: %s)", body, resp.StatusCode, b)
		}
	}
	post(`{"source":"DLL","interface":"WIFI_TEST","name":"FN_WIFI","category":"CAT_A","execution_role":"ROLE_A","execution_phase":"PHASE_A"}`)
	post(`{"source":"DLL","interface":"IPLAS","name":"FN_IPLAS","category":"CAT_B","execution_role":"ROLE_B","execution_phase":"PHASE_B"}`)
	post(`{"source":"BUILT-IN","interface":"TELNET","name":"FN_BUILTIN","category":"CAT_C","execution_role":"ROLE_C","execution_phase":"PHASE_C"}`)
	post(`{"source":"BUILT-IN","interface":"GPIB","name":"FN_ETHER","category":"CAT_D","execution_role":"ROLE_D","execution_phase":"PHASE_D"}`)
}

func itoa(id int64) string { return strconv.FormatInt(id, 10) }

// listBody is the functions list envelope the tests need.
type listBody struct {
	Total int64 `json:"total"`
	Items []struct {
		ID int64 `json:"id"`
	} `json:"items"`
}

func listBodyAt(t *testing.T, app *fiber.App, query string) listBody {
	t.Helper()
	resp, body := doRequest(t, app, "GET", "/api/functions"+query)
	if resp.StatusCode != 200 {
		t.Fatalf("list %s: %d (body: %s)", query, resp.StatusCode, body)
	}
	var lb listBody
	if err := json.Unmarshal([]byte(body), &lb); err != nil {
		t.Fatalf("decode %s: %v (body: %s)", query, err, body)
	}
	return lb
}

// wantNames asserts the list at query returns exactly the named functions.
func wantNames(t *testing.T, app *fiber.App, query string, names ...string) {
	t.Helper()
	lb := listBodyAt(t, app, query)
	if lb.Total != int64(len(names)) || len(lb.Items) != len(names) {
		t.Fatalf("%s: got %d/%d rows, want %v (body total %d)", query, len(lb.Items), lb.Total, names, lb.Total)
	}
	for _, item := range lb.Items {
		resp, body := doRequest(t, app, "GET", "/api/functions/"+itoa(item.ID))
		if resp.StatusCode != 200 {
			t.Fatalf("detail %d: %d", item.ID, resp.StatusCode)
		}
		ok := false
		for _, n := range names {
			if strings.Contains(body, `"name":"`+n+`"`) {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("%s: id %d is none of %v (body: %s)", query, item.ID, names, body)
		}
	}
}

// TestFunctionEnumFilters: every enum column filters server-side and an
// unknown value yields an empty list.
func TestFunctionEnumFilters(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFilterFunctions(t, app)

	wantNames(t, app, "?source=BUILT-IN", "FN_BUILTIN", "FN_ETHER")
	wantNames(t, app, "?source=DLL", "FN_WIFI", "FN_IPLAS")
	wantNames(t, app, "?interface=WIFI_TEST", "FN_WIFI")
	wantNames(t, app, "?interface=TELNET", "FN_BUILTIN")
	wantNames(t, app, "?category=CAT_A", "FN_WIFI")
	wantNames(t, app, "?execution_role=ROLE_B", "FN_IPLAS")
	wantNames(t, app, "?execution_phase=PHASE_D", "FN_ETHER")
	wantNames(t, app, "?source=GHOST") // unknown value → 0 rows
}

// TestFunctionFilterComposesWithSearchSortPaging: enum filters AND with
// search, and compose with sort + pagination without corrupting either.
func TestFunctionFilterComposesWithSearchSortPaging(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFilterFunctions(t, app)

	// Filter + search: both DLL functions share "FN" but only the IPLAS
	// one matches the search text.
	wantNames(t, app, "?source=DLL&search=IPLAS", "FN_IPLAS")

	// Sort desc: FN_WIFI is the largest name.
	lb := listBodyAt(t, app, "?sort=name&order=desc&page_size=1")
	if len(lb.Items) != 1 || lb.Items[0].ID == 0 {
		t.Fatalf("desc page 1: %+v", lb)
	}
	if resp, body := doRequest(t, app, "GET", "/api/functions/"+itoa(lb.Items[0].ID)); resp.StatusCode != 200 ||
		!strings.Contains(body, `"name":"FN_WIFI"`) {
		t.Fatalf("desc page 1 should be FN_WIFI: %d (body: %s)", resp.StatusCode, body)
	}

	// Sort asc over source=DLL: FN_IPLAS then FN_WIFI, so page 2 (size 1)
	// is FN_WIFI.
	lb = listBodyAt(t, app, "?source=DLL&sort=name&order=asc&page_size=1&page=2")
	if len(lb.Items) != 1 {
		t.Fatalf("asc page 2: %+v", lb)
	}
	if resp, body := doRequest(t, app, "GET", "/api/functions/"+itoa(lb.Items[0].ID)); resp.StatusCode != 200 ||
		!strings.Contains(body, `"name":"FN_WIFI"`) {
		t.Fatalf("asc page 2 should be FN_WIFI: %d (body: %s)", resp.StatusCode, body)
	}

	// Two enum filters AND.
	wantNames(t, app, "?source=BUILT-IN&interface=TELNET", "FN_BUILTIN")
}

// TestFunctionFacets: the facet endpoint returns the distinct enum values.
func TestFunctionFacets(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFilterFunctions(t, app)

	resp, body := doRequest(t, app, "GET", "/api/functions/facets")
	if resp.StatusCode != 200 {
		t.Fatalf("facets: %d (body: %s)", resp.StatusCode, body)
	}
	for _, want := range []string{
		`"sources":["BUILT-IN","DLL"]`,
		`"interfaces":["GPIB","IPLAS","TELNET","WIFI_TEST"]`,
		`"categories":["CAT_A","CAT_B","CAT_C","CAT_D"]`,
		`"execution_roles":["ROLE_A","ROLE_B","ROLE_C","ROLE_D"]`,
		`"execution_phases":["PHASE_A","PHASE_B","PHASE_C","PHASE_D"]`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("facets missing %s (body: %s)", want, body)
		}
	}
}
