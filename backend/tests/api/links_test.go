package api_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"webprogfndb/internal/api"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/db"
	"webprogfndb/tests/support/testdb"
)

// newLinksTestApp boots the app exactly as production wiring does.
func newLinksTestApp(t *testing.T) (*fiber.App, *sql.DB) {
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

// seedFunctionAndParameterDefinition creates one function (id 1 — fresh
// database) and one parameter definition through the public API.
func seedFunctionAndParameterDefinition(t *testing.T, app *fiber.App) {
	t.Helper()
	resp, body := postJSON(t, app, "/api/functions", validFunctionJSON())
	if resp.StatusCode != 201 {
		t.Fatalf("seed function status = %d (body: %s)", resp.StatusCode, body)
	}
	resp, body = postJSON(t, app, "/api/parameter-definitions", validParameterDefinitionJSON())
	if resp.StatusCode != 201 {
		t.Fatalf("seed definition status = %d (body: %s)", resp.StatusCode, body)
	}
}

func TestAttachParameterLinkRoundTrip(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"PHONE-A","position":2,"required":true}`)
	if resp.StatusCode != 201 {
		t.Fatalf("attach status = %d, want 201 (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"resolved":true`) || !strings.Contains(body, `"definition_id":1`) {
		t.Fatalf("attach body not resolved to definition 1: %s", body)
	}

	// List echoes the link and reports the position ordering fields.
	resp, body = doRequest(t, app, "GET", "/api/functions/1/parameter-links")
	if resp.StatusCode != 200 {
		t.Fatalf("list status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"name":"ANT"`) || !strings.Contains(body, `"device_name":"PHONE-A"`) ||
		!strings.Contains(body, `"required":true`) {
		t.Fatalf("list body wrong: %s", body)
	}
}

func TestAttachParameterLinkValidation(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// Attaching by name requires the definition to exist in the dictionary.
	resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"NO_SUCH_PARAM"}`)
	if resp.StatusCode != 422 || !strings.Contains(body, `"name"`) {
		t.Fatalf("unknown definition: status = %d (body: %s)", resp.StatusCode, body)
	}

	// Neither a name nor an inline definition is not attachable.
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links", `{"position":1}`)
	if resp.StatusCode != 422 {
		t.Fatalf("missing name: status = %d (body: %s)", resp.StatusCode, body)
	}

	// Position must be a positive integer when present.
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","position":0}`)
	if resp.StatusCode != 422 {
		t.Fatalf("bad position: status = %d (body: %s)", resp.StatusCode, body)
	}

	// Inline definitions go through the same field validation as the
	// dictionary endpoints.
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links",
		`{"definition":{"name":"NEW_P","type":"bool"}}`)
	if resp.StatusCode != 422 || !strings.Contains(body, `"definition.type"`) {
		t.Fatalf("inline validation: status = %d (body: %s)", resp.StatusCode, body)
	}

	// A contradictory payload (pick AND create) is rejected, not silently
	// resolved in favor of one of them.
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","definition":{"name":"OTHER","type":"string"}}`)
	if resp.StatusCode != 422 || !strings.Contains(body, "not both") {
		t.Fatalf("name+definition status = %d (body: %s)", resp.StatusCode, body)
	}

	// Links of an unknown function are a 404, not an empty list.
	resp, _ = doRequest(t, app, "GET", "/api/functions/999/parameter-links")
	if resp.StatusCode != 404 {
		t.Fatalf("unknown function status = %d, want 404", resp.StatusCode)
	}
}

func TestAttachParameterLinkConflictByNaturalKey(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"PHONE-A"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("first attach status = %d (body: %s)", resp.StatusCode, body)
	}
	// Same function + device + name is the natural key: duplicate is a 409.
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"PHONE-A"}`)
	if resp.StatusCode != 409 {
		t.Fatalf("duplicate attach status = %d, want 409 (body: %s)", resp.StatusCode, body)
	}
	// A different device scope is a different link row (ADR-0002).
	resp, body = postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"PHONE-B"}`)
	if resp.StatusCode != 201 {
		t.Fatalf("other device attach status = %d, want 201 (body: %s)", resp.StatusCode, body)
	}
}

func TestUpdateParameterLink(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, body := postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"","position":1,"required":false}`)
	if resp.StatusCode != 201 {
		t.Fatalf("attach status = %d (body: %s)", resp.StatusCode, body)
	}

	resp, body = doJSON(t, app, "PUT", "/api/functions/1/parameter-links/1",
		`{"device_name":"TABLET-C","position":3,"required":true}`)
	if resp.StatusCode != 200 {
		t.Fatalf("update status = %d (body: %s)", resp.StatusCode, body)
	}
	for _, want := range []string{`"device_name":"TABLET-C"`, `"position":3`, `"required":true`} {
		if !strings.Contains(body, want) {
			t.Fatalf("update body missing %s: %s", want, body)
		}
	}

	// A link of another function is out of scope for this route.
	resp, _ = doJSON(t, app, "PUT", "/api/functions/999/parameter-links/1",
		`{"device_name":"X"}`)
	if resp.StatusCode != 404 {
		t.Fatalf("wrong-function update status = %d, want 404", resp.StatusCode)
	}
}

func TestUpdateLinkDeviceConflict(t *testing.T) {
	for _, noun := range []string{"parameter", "criteria"} {
		t.Run(noun, func(t *testing.T) {
			app, _ := newLinksTestApp(t)
			if resp, body := postJSON(t, app, "/api/functions", validFunctionJSON()); resp.StatusCode != http.StatusCreated {
				t.Fatalf("seed function: status = %d (body: %s)", resp.StatusCode, body)
			}

			path := "/api/functions/1/" + noun + "-links"
			for _, payload := range []string{
				`{"definition":{"name":"POWER","type":"double"},"position":1,"required":true}`,
				`{"name":"POWER","device_name":"PHONE-A","position":2}`,
			} {
				if resp, body := postJSON(t, app, path, payload); resp.StatusCode != http.StatusCreated {
					t.Fatalf("attach %s: status = %d (body: %s)", payload, resp.StatusCode, body)
				}
			}
			_, before := doRequest(t, app, "GET", path)

			resp, body := doJSON(t, app, "PUT", path+"/1", `{"device_name":" phone-a ","position":3}`)
			if resp.StatusCode != http.StatusConflict {
				t.Fatalf("duplicate device update: status = %d, want 409 (body: %s)", resp.StatusCode, body)
			}
			var conflict struct{ Error string }
			want := noun + " link 'POWER' already exists for device scope 'PHONE-A'"
			if err := json.Unmarshal([]byte(body), &conflict); err != nil || conflict.Error != want {
				t.Fatalf("duplicate device update: body = %s, want error %q", body, want)
			}

			resp, after := doRequest(t, app, "GET", path)
			if resp.StatusCode != http.StatusOK || after != before {
				t.Fatalf("link list changed after conflict: status = %d, before = %s, after = %s", resp.StatusCode, before, after)
			}
		})
	}
}

// TestUpdateParameterLinkReplace pins the PUT contract: update is a full
// replace of the editable attributes, so an omitted position becomes NULL
// and an omitted required becomes false.
func TestUpdateParameterLinkReplace(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	postJSON(t, app, "/api/functions/1/parameter-links",
		`{"name":"ANT","device_name":"","position":2,"required":true}`)

	resp, body := doJSON(t, app, "PUT", "/api/functions/1/parameter-links/1",
		`{"device_name":"PHONE-Z"}`)
	if resp.StatusCode != 200 {
		t.Fatalf("update status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"position":null`) || !strings.Contains(body, `"required":false`) {
		t.Fatalf("partial PUT must be a full replace: %s", body)
	}
}

func TestParameterLinkPositionOrdering(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// Attach out of order, plus one link without a position (NULL last).
	for _, b := range []string{
		`{"name":"ANT","position":3}`,
		`{"name":"ANT","device_name":"D2","position":1}`,
		`{"name":"ANT","device_name":"D3","position":2}`,
		`{"name":"ANT","device_name":"D4"}`,
	} {
		if resp, body := postJSON(t, app, "/api/functions/1/parameter-links", b); resp.StatusCode != 201 {
			t.Fatalf("attach %s status = %d (body: %s)", b, resp.StatusCode, body)
		}
	}

	_, body := doRequest(t, app, "GET", "/api/functions/1/parameter-links")
	// Rows ordered by position, NULL position last: ids 2,3,1,4.
	order := []string{`"id":2`, `"id":3`, `"id":1`, `"id":4`}
	last := -1
	for _, idStr := range order {
		idx := strings.Index(body, idStr)
		if idx < 0 || idx < last {
			t.Fatalf("position order wrong around %s (body: %s)", idStr, body)
		}
		last = idx
	}
}

func TestDetachParameterLinkKeepsDefinition(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, _ := postJSON(t, app, "/api/functions/1/parameter-links", `{"name":"ANT","position":1}`)
	if resp.StatusCode != 201 {
		t.Fatalf("attach failed")
	}

	resp, _ = doRequest(t, app, "DELETE", "/api/functions/1/parameter-links/1")
	if resp.StatusCode != 204 {
		t.Fatalf("detach status = %d, want 204", resp.StatusCode)
	}

	// Detaching a link never deletes the definition (CONTEXT.md).
	resp, body := doRequest(t, app, "GET", "/api/parameter-definitions/1")
	if resp.StatusCode != 200 || !strings.Contains(body, `"name":"ANT"`) {
		t.Fatalf("definition must survive detach: %d %s", resp.StatusCode, body)
	}

	_, body = doRequest(t, app, "GET", "/api/functions/1/parameter-links")
	if !strings.Contains(body, `"items":[]`) {
		t.Fatalf("list should be empty after detach: %s", body)
	}
}

func TestOrphanParameterLinkWarningAndEdit(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, _ := postJSON(t, app, "/api/functions/1/parameter-links", `{"name":"ANT","position":1}`)
	if resp.StatusCode != 201 {
		t.Fatalf("attach failed")
	}

	// Delete-with-warning: the definition is referenced, so a plain delete
	// is a 409 carrying the usage count.
	resp, body := doRequest(t, app, "DELETE", "/api/parameter-definitions/1")
	if resp.StatusCode != 409 || !strings.Contains(body, `"usage_count":1`) {
		t.Fatalf("guarded delete status = %d (body: %s)", resp.StatusCode, body)
	}
	// Forced delete succeeds; the link survives as an Orphan (ADR-0003).
	resp, _ = doRequest(t, app, "DELETE", "/api/parameter-definitions/1?force=true")
	if resp.StatusCode != 204 {
		t.Fatalf("forced delete status = %d", resp.StatusCode)
	}

	_, body = doRequest(t, app, "GET", "/api/functions/1/parameter-links")
	if !strings.Contains(body, `"resolved":false`) || !strings.Contains(body, `"definition_id":null`) {
		t.Fatalf("orphan link must render unresolved: %s", body)
	}

	// Orphan links are editable in place: no dictionary lookup on update.
	resp, body = doJSON(t, app, "PUT", "/api/functions/1/parameter-links/1",
		`{"device_name":"LEGACY","position":5,"required":true}`)
	if resp.StatusCode != 200 {
		t.Fatalf("orphan edit status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"resolved":false`) || !strings.Contains(body, `"device_name":"LEGACY"`) {
		t.Fatalf("orphan edit body wrong: %s", body)
	}
}

func TestInlineParameterDefinitionCreation(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// One flow: create the definition and link it, without leaving the page.
	resp, body := postJSON(t, app, "/api/functions/1/parameter-links", `{
		"definition": {"name":"BAUD","type":"int","unit":"bps","required":true},
		"device_name":"PHONE-A",
		"position":1,
		"required":true
	}`)
	if resp.StatusCode != 201 {
		t.Fatalf("inline attach status = %d (body: %s)", resp.StatusCode, body)
	}
	if !strings.Contains(body, `"resolved":true`) {
		t.Fatalf("inline attach should resolve: %s", body)
	}

	// The definition landed in the dictionary with its own id.
	resp, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=BAUD")
	if resp.StatusCode != 200 || !strings.Contains(body, `"name":"BAUD"`) {
		t.Fatalf("inline definition missing from dictionary: %s", body)
	}

	// Inline creation with an existing name is a conflict and must not
	// create a duplicate definition.
	resp, _ = postJSON(t, app, "/api/functions/1/parameter-links", `{
		"definition": {"name":"BAUD","type":"int"},
		"device_name":"PHONE-B"
	}`)
	if resp.StatusCode != 409 {
		t.Fatalf("duplicate inline status = %d, want 409", resp.StatusCode)
	}
	_, body = doRequest(t, app, "GET", "/api/parameter-definitions?search=BAUD")
	if !strings.Contains(body, `"total":1`) {
		t.Fatalf("duplicate inline definition created: %s", body)
	}
}

func TestInlineCreationRollsBackOnLinkConflict(t *testing.T) {
	app, d := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	// An orphan link named STALE exists, exactly as the workbook import
	// leaves one behind: a link row with no matching definition (ADR-0003).
	_, err := d.Exec(`INSERT INTO parameter_links
		(function_source, function_name, device_name, name, required, created_at, updated_at)
		VALUES ('DLL', 'TX_MULTI_VERIFICATION', '', 'STALE', 0, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("seed orphan link: %v", err)
	}

	// Inline-creating a definition named STALE with the same device scope
	// collides with the orphan link's natural key. The definition insert
	// must roll back: either both rows land or neither does.
	resp, _ := postJSON(t, app, "/api/functions/1/parameter-links", `{
		"definition": {"name":"STALE","type":"string"},
		"device_name":""
	}`)
	if resp.StatusCode != 409 {
		t.Fatalf("inline link conflict status = %d, want 409", resp.StatusCode)
	}
	_, body := doRequest(t, app, "GET", "/api/parameter-definitions?search=STALE")
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("conflicted inline definition was not rolled back: %s", body)
	}
}

func TestCriteriaLinksMirrorParameters(t *testing.T) {
	app, _ := newLinksTestApp(t)
	seedFunctionAndParameterDefinition(t, app)

	resp, body := postJSON(t, app, "/api/criteria-definitions", `{
		"name":"CRIT_PWR","type":"double","unit":"dBm","required":true
	}`)
	if resp.StatusCode != 201 {
		t.Fatalf("seed criteria definition status = %d (body: %s)", resp.StatusCode, body)
	}

	// Attach by name.
	resp, body = postJSON(t, app, "/api/functions/1/criteria-links",
		`{"name":"CRIT_PWR","device_name":"PHONE-A","position":1,"required":true}`)
	if resp.StatusCode != 201 || !strings.Contains(body, `"resolved":true`) {
		t.Fatalf("criteria attach status = %d (body: %s)", resp.StatusCode, body)
	}

	// Inline creation in the same flow.
	resp, body = postJSON(t, app, "/api/functions/1/criteria-links", `{
		"definition": {"name":"CRIT_SNR","type":"double","unit":"dB"},
		"position":2
	}`)
	if resp.StatusCode != 201 || !strings.Contains(body, `"resolved":true`) {
		t.Fatalf("criteria inline status = %d (body: %s)", resp.StatusCode, body)
	}

	// Update, then detach: the definition survives.
	resp, body = doJSON(t, app, "PUT", "/api/functions/1/criteria-links/1",
		`{"device_name":"PHONE-B","position":4,"required":false}`)
	if resp.StatusCode != 200 || !strings.Contains(body, `"position":4`) {
		t.Fatalf("criteria update status = %d (body: %s)", resp.StatusCode, body)
	}
	resp, _ = doRequest(t, app, "DELETE", "/api/functions/1/criteria-links/1")
	if resp.StatusCode != 204 {
		t.Fatalf("criteria detach status = %d", resp.StatusCode)
	}
	resp, _ = doRequest(t, app, "GET", "/api/criteria-definitions/1")
	if resp.StatusCode != 200 {
		t.Fatalf("criteria definition must survive detach")
	}

	// Orphan criteria link: force-delete the definition, link flags it.
	resp, _ = doRequest(t, app, "DELETE", "/api/criteria-definitions/2?force=true")
	if resp.StatusCode != 204 {
		t.Fatalf("criteria force delete status = %d", resp.StatusCode)
	}
	_, body = doRequest(t, app, "GET", "/api/functions/1/criteria-links")
	if !strings.Contains(body, `"resolved":false`) {
		t.Fatalf("orphan criteria link must render unresolved: %s", body)
	}
}
