package api_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type registryItem struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	SortOrder  int64  `json:"sort_order"`
	BuiltIn    bool   `json:"built_in"`
	Tone       string `json:"tone"`
	GuideCount *int64 `json:"guide_count"`
}

type registryList struct {
	Items []registryItem `json:"items"`
}

func registryItems(t *testing.T, body string) []registryItem {
	t.Helper()
	var out registryList
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode registry list: %v (%s)", err, body)
	}
	return out.Items
}

func registryItemByKey(t *testing.T, items []registryItem, key string) registryItem {
	t.Helper()
	for _, item := range items {
		if item.Key == key {
			return item
		}
	}
	t.Fatalf("registry key %q not found in %+v", key, items)
	return registryItem{}
}

func TestTroubleshootingRegistryValidationAndKeyGeneration(t *testing.T) {
	app, _ := newLinksTestApp(t)

	for _, tc := range []struct {
		name    string
		path    string
		payload string
		field   string
	}{
		{name: "blank category", path: "/api/troubleshooting/categories", payload: `{"label":"   "}`, field: "label"},
		{name: "too long category", path: "/api/troubleshooting/categories", payload: fmt.Sprintf(`{"label":%q}`, strings.Repeat("x", 65)), field: "label"},
		{name: "duplicate category", path: "/api/troubleshooting/categories", payload: `{"label":"dut connection & setup"}`, field: "label"},
		{name: "invalid severity tone", path: "/api/troubleshooting/severities", payload: `{"label":"High Risk","tone":"urgent"}`, field: "tone"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := postJSON(t, app, tc.path, tc.payload)
			if resp.StatusCode != 422 {
				t.Fatalf("status = %d, want 422 (%s)", resp.StatusCode, body)
			}
			if !strings.Contains(body, `"`+tc.field+`"`) {
				t.Fatalf("missing %q field error: %s", tc.field, body)
			}
		})
	}

	resp, emptyDescriptionBody := postJSON(t, app, "/api/troubleshooting/categories", `{"label":"No Description Category"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create with optional description: %d (%s)", resp.StatusCode, emptyDescriptionBody)
	}

	resp, defaultToneBody := postJSON(t, app, "/api/troubleshooting/severities", `{"label":"Default Tone Severity"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create severity with default tone: %d (%s)", resp.StatusCode, defaultToneBody)
	}
	if !strings.Contains(defaultToneBody, `"tone":"outline"`) {
		t.Fatalf("new severity tone = %s, want outline", defaultToneBody)
	}

	resp, firstBody := postJSON(t, app, "/api/troubleshooting/categories", `{"label":"Foo Bar"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create first colliding key: %d (%s)", resp.StatusCode, firstBody)
	}
	resp, secondBody := postJSON(t, app, "/api/troubleshooting/categories", `{"label":"Foo-Bar"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create second colliding key: %d (%s)", resp.StatusCode, secondBody)
	}
	if !strings.Contains(firstBody, `"key":"foo-bar"`) || !strings.Contains(secondBody, `"key":"foo-bar-2"`) {
		t.Fatalf("collision keys not suffixed: first=%s second=%s", firstBody, secondBody)
	}

	for _, tc := range []struct{ path, payload string }{
		{path: "/api/troubleshooting/categories/foo-bar", payload: `{"label":"Foo Bar","sort_order":-1}`},
		{path: "/api/troubleshooting/categories/foo-bar", payload: `{"label":"Foo Bar","sort_order":0}`},
		{path: "/api/troubleshooting/severities/critical", payload: `{"sort_order":0}`},
	} {
		resp, body := doJSON(t, app, "PUT", tc.path, tc.payload)
		if resp.StatusCode != 422 || !strings.Contains(body, `"sort_order"`) {
			t.Fatalf("invalid update order for %s: %d (%s)", tc.path, resp.StatusCode, body)
		}
	}
}

func TestTroubleshootingRegistryListIncludesCountsAndStableUpdates(t *testing.T) {
	app, _ := newLinksTestApp(t)

	resp, categoryBody := doRequest(t, app, "GET", "/api/troubleshooting/categories")
	if resp.StatusCode != 200 {
		t.Fatalf("list categories: %d (%s)", resp.StatusCode, categoryBody)
	}
	categories := registryItems(t, categoryBody)
	measurement := registryItemByKey(t, categories, "measurement")
	if measurement.GuideCount == nil || *measurement.GuideCount == 0 {
		t.Fatalf("measurement guide_count should include seeded guides: %+v", measurement)
	}
	for i := 1; i < len(categories); i++ {
		if categories[i-1].SortOrder > categories[i].SortOrder ||
			(categories[i-1].SortOrder == categories[i].SortOrder && strings.ToLower(categories[i-1].Label) > strings.ToLower(categories[i].Label)) {
			t.Fatalf("categories not ordered by sort_order, label: %+v", categories)
		}
	}

	resp, severityBody := doRequest(t, app, "GET", "/api/troubleshooting/severities")
	if resp.StatusCode != 200 {
		t.Fatalf("list severities: %d (%s)", resp.StatusCode, severityBody)
	}
	severities := registryItems(t, severityBody)
	if critical := registryItemByKey(t, severities, "critical"); critical.GuideCount == nil || *critical.GuideCount != 4 {
		t.Fatalf("critical guide_count should be 4 seeded guides: %s", severityBody)
	}
	if high := registryItemByKey(t, severities, "high"); high.GuideCount == nil || *high.GuideCount != 0 {
		t.Fatalf("high guide_count should be present and equal 0: %s", severityBody)
	}

	resp, updatedBody := doJSON(t, app, "PUT", "/api/troubleshooting/severities/critical", `{"label":"Critical Priority","description":"Updated","sort_order":1,"tone":"destructive","key":"renamed","built_in":false}`)
	if resp.StatusCode != 200 {
		t.Fatalf("update severity: %d (%s)", resp.StatusCode, updatedBody)
	}
	if !strings.Contains(updatedBody, `"key":"critical"`) || !strings.Contains(updatedBody, `"built_in":true`) {
		t.Fatalf("update changed stable fields: %s", updatedBody)
	}
}

func TestTroubleshootingRegistryDeleteValidationAndAtomicReassignment(t *testing.T) {
	app, d := newLinksTestApp(t)
	resp, categoryBody := postJSON(t, app, "/api/troubleshooting/categories", `{"label":"API Fixture Category"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create category: %d (%s)", resp.StatusCode, categoryBody)
	}
	resp, severityBody := postJSON(t, app, "/api/troubleshooting/severities", `{"label":"API Fixture Severity","tone":"warn"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create severity: %d (%s)", resp.StatusCode, severityBody)
	}

	guide := `{
		"title":"Registry Transaction Fixture","category":"api-fixture-category","severity":"api-fixture-severity",
		"device_name":"WIFI_TEST","symptom":"Fixture symptom","steps":[{"title":"Inspect","instruction":"Inspect it"}]
	}`
	resp, guideBody := postJSON(t, app, "/api/troubleshooting", guide)
	if resp.StatusCode != 201 {
		t.Fatalf("create guide: %d (%s)", resp.StatusCode, guideBody)
	}
	var createdGuide struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(guideBody), &createdGuide); err != nil {
		t.Fatalf("decode guide: %v (%s)", err, guideBody)
	}

	for _, tc := range []struct{ path, field string }{
		{path: "/api/troubleshooting/categories/connection?reassign_to=measurement", field: "key"},
		{path: "/api/troubleshooting/categories/api-fixture-category", field: "reassign_to"},
		{path: "/api/troubleshooting/categories/api-fixture-category?reassign_to=api-fixture-category", field: "reassign_to"},
		{path: "/api/troubleshooting/categories/api-fixture-category?reassign_to=critical", field: "reassign_to"},
		{path: "/api/troubleshooting/categories/api-fixture-category?reassign_to=missing", field: "reassign_to"},
	} {
		resp, body := doRequest(t, app, "DELETE", tc.path)
		if resp.StatusCode != 422 || !strings.Contains(body, `"`+tc.field+`"`) {
			t.Fatalf("delete %s = %d, want 422 field %s (%s)", tc.path, resp.StatusCode, tc.field, body)
		}
	}

	resp, body := doRequest(t, app, "DELETE", "/api/troubleshooting/categories/api-fixture-category?reassign_to=measurement")
	if resp.StatusCode != 204 {
		t.Fatalf("successful category reassignment: %d (%s)", resp.StatusCode, body)
	}
	resp, body = doRequest(t, app, "GET", fmt.Sprintf("/api/troubleshooting/%d", createdGuide.ID))
	if resp.StatusCode != 200 || !strings.Contains(body, `"category":"measurement"`) {
		t.Fatalf("guide category not reassigned: %d (%s)", resp.StatusCode, body)
	}

	if _, err := d.Exec(`CREATE FUNCTION fail_troubleshooting_reassignment() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.severity = 'api-fixture-severity' THEN RAISE EXCEPTION 'forced reassignment failure'; END IF;
  RETURN NEW;
END
$$`); err != nil {
		t.Fatalf("create rollback trigger function: %v", err)
	}
	if _, err := d.Exec(`CREATE TRIGGER fail_troubleshooting_reassignment_trigger
BEFORE UPDATE OF severity ON troubleshooting_guides
FOR EACH ROW EXECUTE FUNCTION fail_troubleshooting_reassignment()`); err != nil {
		t.Fatalf("create rollback trigger: %v", err)
	}
	resp, body = doRequest(t, app, "DELETE", "/api/troubleshooting/severities/api-fixture-severity?reassign_to=warning")
	if resp.StatusCode != 500 {
		t.Fatalf("forced reassignment failure: %d (%s)", resp.StatusCode, body)
	}
	resp, severityList := doRequest(t, app, "GET", "/api/troubleshooting/severities")
	if resp.StatusCode != 200 {
		t.Fatalf("list severities after rollback: %d (%s)", resp.StatusCode, severityList)
	}
	_ = registryItemByKey(t, registryItems(t, severityList), "api-fixture-severity")
	resp, body = doRequest(t, app, "GET", fmt.Sprintf("/api/troubleshooting/%d", createdGuide.ID))
	if resp.StatusCode != 200 || !strings.Contains(body, `"severity":"api-fixture-severity"`) {
		t.Fatalf("guide changed after rollback: %d (%s)", resp.StatusCode, body)
	}
}

func TestTroubleshootingRegistryReorderRollsBackBothRows(t *testing.T) {
	app, d := newLinksTestApp(t)

	originalOrders := map[string]int64{}
	rows, err := d.Query(`SELECT key, sort_order FROM troubleshooting_categories WHERE key IN ('connection', 'execution')`)
	if err != nil {
		t.Fatalf("read category orders: %v", err)
	}
	for rows.Next() {
		var key string
		var order int64
		if err := rows.Scan(&key, &order); err != nil {
			t.Fatalf("scan category order: %v", err)
		}
		originalOrders[key] = order
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("close category order rows: %v", err)
	}

	if _, err := d.Exec(`CREATE FUNCTION fail_troubleshooting_reorder() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.key = 'execution' THEN RAISE EXCEPTION 'forced reorder failure'; END IF;
  RETURN NEW;
END
$$`); err != nil {
		t.Fatalf("create reorder trigger function: %v", err)
	}
	if _, err := d.Exec(`CREATE TRIGGER fail_troubleshooting_reorder_trigger
BEFORE UPDATE OF sort_order ON troubleshooting_categories
FOR EACH ROW EXECUTE FUNCTION fail_troubleshooting_reorder()`); err != nil {
		t.Fatalf("create reorder trigger: %v", err)
	}

	resp, body := postJSON(t, app, "/api/troubleshooting/categories/reorder", `{"key":"connection","with_key":"execution"}`)
	if resp.StatusCode != 500 {
		t.Fatalf("forced reorder failure = %d, want 500 (%s)", resp.StatusCode, body)
	}
	for key, want := range originalOrders {
		var got int64
		if err := d.QueryRow(`SELECT sort_order FROM troubleshooting_categories WHERE key = $1`, key).Scan(&got); err != nil {
			t.Fatalf("read %s after rollback: %v", key, err)
		}
		if got != want {
			t.Fatalf("%s sort_order = %d after rollback, want %d", key, got, want)
		}
	}
}

func TestTroubleshootingRegistryConcurrentSlugAllocation(t *testing.T) {
	app, _ := newLinksTestApp(t)

	start := make(chan struct{})
	results := make(chan registryItem, 2)
	errors := make(chan error, 2)
	for _, label := range []string{"Concurrent Collision", "Concurrent-Collision"} {
		label := label
		go func() {
			<-start
			resp, body := postJSON(t, app, "/api/troubleshooting/categories", fmt.Sprintf(`{"label":%q}`, label))
			if resp.StatusCode != 201 {
				errors <- fmt.Errorf("create %q: %d (%s)", label, resp.StatusCode, body)
				return
			}
			var item registryItem
			if err := json.Unmarshal([]byte(body), &item); err != nil {
				errors <- fmt.Errorf("decode %q: %w (%s)", label, err, body)
				return
			}
			results <- item
		}()
	}
	close(start)

	keys := map[string]bool{}
	for range 2 {
		select {
		case err := <-errors:
			t.Fatal(err)
		case item := <-results:
			if len(item.Key) > 64 {
				t.Fatalf("key %q exceeds 64 bytes", item.Key)
			}
			keys[item.Key] = true
		}
	}
	if !keys["concurrent-collision"] || !keys["concurrent-collision-2"] {
		t.Fatalf("concurrent keys = %v, want base and -2 suffix", keys)
	}
}

func TestTroubleshootingGuideRegistryValidationPrecedesWrites(t *testing.T) {
	app, _ := newLinksTestApp(t)
	resp, body := postJSON(t, app, "/api/troubleshooting", `{
		"title":"Unknown Registry Keys","category":"not-a-category","severity":"not-a-severity",
		"device_name":"WIFI_TEST","symptom":"symptom",
		"steps":[{"title":"Check","instruction":"Check it"}]
	}`)
	if resp.StatusCode != 400 || !strings.Contains(body, `"category"`) || !strings.Contains(body, `"severity"`) {
		t.Fatalf("unknown registry validation: %d (%s)", resp.StatusCode, body)
	}

	resp, body = postJSON(t, app, "/api/troubleshooting", `{
		"title":"Registry Preservation","category":"connection","severity":"critical",
		"device_name":"WIFI_TEST","symptom":"known symptom",
		"steps":[{"title":"Original Step","instruction":"Keep this step"}]
	}`)
	if resp.StatusCode != 201 {
		t.Fatalf("create guide for update validation: %d (%s)", resp.StatusCode, body)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatalf("decode created guide: %v (%s)", err, body)
	}
	resp, body = doJSON(t, app, "PUT", fmt.Sprintf("/api/troubleshooting/%d", created.ID), `{
		"title":"Should Not Persist","category":"not-a-category","severity":"not-a-severity",
		"device_name":"WIFI_TEST","symptom":"new symptom",
		"steps":[{"title":"Replacement Step","instruction":"Do not write this"}]
	}`)
	if resp.StatusCode != 400 {
		t.Fatalf("unknown registry update validation: %d (%s)", resp.StatusCode, body)
	}
	resp, body = doRequest(t, app, "GET", fmt.Sprintf("/api/troubleshooting/%d", created.ID))
	if resp.StatusCode != 200 || !strings.Contains(body, `"title":"Registry Preservation"`) || !strings.Contains(body, `"Original Step"`) {
		t.Fatalf("guide changed after invalid registry update: %d (%s)", resp.StatusCode, body)
	}
}
