package api_test

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/store"
)

func uploadGuideTransferBytes(t *testing.T, app *fiber.App, operation, filename string, content []byte) (int, string) {
	t.Helper()
	return uploadGuideTransferWithHeaders(t, app, operation, filename, content, nil)
}

func uploadGuideTransferWithHeaders(t *testing.T, app *fiber.App, operation, filename string, content []byte, headers map[string]string) (int, string) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "http://catalog.example:4174/api/troubleshooting/import/"+operation, &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	result, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, string(result)
}

func TestTroubleshootingGuideImportRequestOrigin(t *testing.T) {
	app, _ := newLinksTestApp(t)
	payload := []byte(`{"guides":[{"title":"Origin test","category":"connection","severity":"high","symptom":"Signal lost","steps":[{"title":"Reset","instruction":"Restart"}]}]}`)
	for _, operation := range []string{"single", "preview", "commit"} {
		t.Run(operation, func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				headers map[string]string
				want    int
			}{
				{"same origin", map[string]string{"Origin": "http://catalog.example:4174", "Sec-Fetch-Site": "same-origin"}, 200},
				{"no browser headers", nil, 200},
				{"cross site fetch", map[string]string{"Sec-Fetch-Site": "cross-site"}, 403},
				{"different origin port", map[string]string{"Origin": "http://catalog.example:4175"}, 403},
				{"untrusted forwarded host", map[string]string{"Origin": "https://untrusted.example", "X-Forwarded-Host": "untrusted.example"}, 403},
				{"opaque origin", map[string]string{"Origin": "null"}, 403},
			} {
				t.Run(tc.name, func(t *testing.T) {
					status, _ := uploadGuideTransferWithHeaders(t, app, operation, "origin.json", payload, tc.headers)
					if status != tc.want {
						t.Fatalf("status = %d, want %d", status, tc.want)
					}
				})
			}
		})
	}
}

func TestTroubleshootingGuideTransferImport(t *testing.T) {
	app, db := newLinksTestApp(t)
	const snippet = `"script_snippet":"    first line\n\tsecond line"`
	create := `{"title":"Transfer create","category":"connection","severity":"high","symptom":"- first\n- second","steps":[{"title":"First","instruction":"Run",` + snippet + `}]}`
	single := `{"guides":[` + create + `]}`
	bad := `{"guides":[` + create + `,{"id":999999,"category":"unknown","severity":"unknown","steps":[{}]}]}`
	assertUpload := func(operation, payload string, want int, fragments ...string) {
		t.Helper()
		status, body := uploadGuideTransferBytes(t, app, operation, "guides.json", []byte(payload))
		if status != want {
			t.Fatalf("%s status = %d, want %d: %s", operation, status, want, body)
		}
		for _, fragment := range fragments {
			if !strings.Contains(body, fragment) {
				t.Fatalf("%s missing %q: %s", operation, fragment, body)
			}
		}
	}
	assertUpload("single", single, 200, `"title":"Transfer create"`, snippet)
	assertUpload("single", bad, 422, "bulk import")
	assertUpload("preview", bad, 200, `"action":"create"`, `"action":"error"`, "unknown category", "unknown severity", "id not found", "title is required", "instruction is required")
	assertUpload("commit", bad, 422, "id not found")
	assertUpload("preview", single, 200, `"action":"create"`)
	_, body := doRequest(t, app, "GET", "/api/troubleshooting?search=Transfer+create")
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("preview or invalid commit changed DB: %s", body)
	}

	failOnWrite := strings.Replace(create, "Transfer create", "Transfer rejected", 1)
	if _, err := db.Exec(`ALTER TABLE troubleshooting_guides ADD CHECK (title <> 'Transfer rejected')`); err != nil {
		t.Fatal(err)
	}
	assertUpload("commit", `{"guides":[`+create+`,`+failOnWrite+`]}`, 500)
	_, body = doRequest(t, app, "GET", "/api/troubleshooting?search=Transfer+create")
	if !strings.Contains(body, `"total":0`) {
		t.Fatalf("commit was not atomic: %s", body)
	}
	assertUpload("commit", single, 200, `"created":1`)
	var id int64
	if err := db.QueryRow(`SELECT id FROM troubleshooting_guides WHERE title = 'Transfer create'`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	update := fmt.Sprintf(`{"guides":[{"id":%d,`, id) + strings.Replace(strings.Replace(create[1:], "First", "Replacement", 1), "Transfer create", "Transfer updated", 1) + `]}`
	assertUpload("preview", update, 200, `"action":"update"`)
	assertUpload("commit", update, 200, `"updated":1`)
	_, body = doRequest(t, app, "GET", fmt.Sprintf("/api/troubleshooting/%d", id))
	if !strings.Contains(body, `"title":"Replacement"`) || strings.Contains(body, `"title":"First"`) {
		t.Fatalf("existing steps were not replaced: %s", body)
	}
	if !strings.Contains(body, snippet) {
		t.Fatalf("import did not preserve script snippet indentation: %s", body)
	}
}

func TestTroubleshootingGuideTransferDownloads(t *testing.T) {
	app, db := newLinksTestApp(t)
	var id int64
	if err := db.QueryRow(`SELECT MIN(id) FROM troubleshooting_guides WHERE error_code = 'ERRCODE:104'`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	request := func(method, url, in string) (*http.Response, []byte) {
		t.Helper()
		req := httptest.NewRequest(method, url, strings.NewReader(in))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		raw, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != 200 || !strings.HasPrefix(res.Header.Get("Content-Disposition"), "attachment;") {
			t.Fatalf("download: %d %s: %s", res.StatusCode, res.Header, raw)
		}
		return res, raw
	}
	_, template := request("GET", "/api/troubleshooting/templates/xlsx", "")
	book, err := excelize.OpenReader(bytes.NewReader(template))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	if rows, err := book.GetRows("Instructions"); err != nil || len(rows) < 3 {
		t.Fatalf("instructions missing: %v %v", rows, err)
	}
	_, jsonTemplate := request("GET", "/api/troubleshooting/templates/json", "")
	if status, body := uploadGuideTransferBytes(t, app, "preview", "template.json", jsonTemplate); status != 200 || !strings.Contains(body, `"action":"create"`) {
		t.Fatalf("JSON template didn't preview as create: %d %s", status, body)
	}
	_, zipped := request("POST", "/api/troubleshooting/export", fmt.Sprintf(`{"format":"both","ids":[%d]}`, id))
	product, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil || len(product.File) != 2 {
		t.Fatalf("ZIP: %v, %v", product, err)
	}
	for _, file := range product.File {
		r, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		status, body := uploadGuideTransferBytes(t, app, "preview", file.Name, data)
		if status != 200 || strings.Count(body, `"action":"update"`) != 1 || !strings.Contains(body, fmt.Sprintf(`"id":%d`, id)) {
			t.Fatalf("round trip %s: %d %s", file.Name, status, body)
		}
	}
	_, matching := request("POST", "/api/troubleshooting/export?error_code=ERRCODE:104", `{"format":"json","scope":"matching"}`)
	if status, body := uploadGuideTransferBytes(t, app, "preview", "matching.json", matching); status != 200 || strings.Count(body, `"action":"update"`) != 2 {
		t.Fatalf("filtered export: %d %s", status, body)
	}
}

func TestTroubleshootingGuideExportCellLimit(t *testing.T) {
	app, db := newLinksTestApp(t)
	const title = "Bench Recovery Guide"
	const endpoint = "/api/troubleshooting/export"
	atLimit := strings.Repeat("🙂", 16383) + "x"
	oversized := atLimit + "x"
	var id int64
	if err := db.QueryRow(`INSERT INTO troubleshooting_guides (title, category, severity, created_at, updated_at)
		VALUES ($1, 'execution', 'high', '', '') RETURNING id`, title).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO troubleshooting_steps (guide_id, position, title, instruction, script_snippet)
		VALUES ($1, 1, 'Restart the controller', 'Run the recovery script.', $2)`, id, atLimit); err != nil {
		t.Fatal(err)
	}
	request := func(format string) (*http.Response, string) {
		t.Helper()
		return postJSON(t, app, endpoint, fmt.Sprintf(`{"format":%q,"ids":[%d]}`, format, id))
	}
	res, body := request("xlsx")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("XLSX at the cell limit: %d %s", res.StatusCode, body)
	}
	book, err := excelize.OpenReader(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer book.Close()
	if got, err := book.GetCellValue("Steps", "E2"); err != nil || got != atLimit {
		t.Fatalf("snippet at the cell limit did not round-trip: %v", err)
	}
	if _, err := db.Exec("UPDATE troubleshooting_steps SET script_snippet = $1 WHERE guide_id = $2", oversized, id); err != nil {
		t.Fatal(err)
	}
	assertLimit := func(format, field string) {
		t.Helper()
		res, body := request(format)
		var problem struct{ Error string }
		if err := json.Unmarshal([]byte(body), &problem); err != nil ||
			res.StatusCode != http.StatusUnprocessableEntity || res.Header.Get("Content-Disposition") != "" ||
			!strings.Contains(problem.Error, title) || !strings.Contains(problem.Error, field) ||
			!strings.Contains(problem.Error, "Export as JSON instead.") {
			t.Errorf("%s limit response for %s: %d %s", format, field, res.StatusCode, body)
		}
	}
	assertLimit("xlsx", "Step 1 / Snippet")
	assertLimit("both", "Step 1 / Snippet")
	res, body = request("json")
	var doc struct {
		Guides []struct {
			Steps []struct {
				Snippet string `json:"script_snippet"`
			}
		}
	}
	if err := json.Unmarshal([]byte(body), &doc); err != nil || res.StatusCode != http.StatusOK ||
		len(doc.Guides) != 1 || len(doc.Guides[0].Steps) != 1 || doc.Guides[0].Steps[0].Snippet != oversized {
		t.Fatalf("JSON did not preserve the oversized snippet: status %d, decode error %v", res.StatusCode, err)
	}
	if _, err := db.Exec("UPDATE troubleshooting_guides SET root_cause = $1 WHERE id = $2", oversized, id); err != nil {
		t.Fatal(err)
	}
	assertLimit("xlsx", "Root Cause")
}

func TestTroubleshootingGuideAllExportOrder(t *testing.T) {
	app, db := newLinksTestApp(t)
	// Exercise registry order independently of either key or title order.
	for _, stmt := range []string{
		`DELETE FROM troubleshooting_guides`,
		`UPDATE troubleshooting_severities SET sort_order = 10 WHERE key = 'critical'`,
		`UPDATE troubleshooting_categories SET sort_order = 20 WHERE key = 'connection'`,
		`INSERT INTO troubleshooting_guides (id, title, category, severity, symptom, created_at, updated_at) VALUES
			(101, 'Zebra', 'execution', 'high', 'Symptom', '', ''),
			(102, 'Alpha', 'connection', 'high', 'Symptom', '', ''),
			(103, 'Aardvark', 'execution', 'critical', 'Symptom', '', ''),
			(104, 'apple', 'execution', 'high', 'Symptom', '', ''),
			(105, 'APPLE', 'execution', 'high', 'Symptom', '', '')`,
		`INSERT INTO troubleshooting_steps (id, guide_id, position, title, instruction, script_snippet, verification) VALUES
			(101, 104, 2, 'Third', 'Instruction 3', '', 'Verified'),
			(103, 104, 1, 'Second', 'Instruction 2', '', ''),
			(102, 104, 1, 'First', 'Instruction 1', E'  Run\n\tCheck', ''),
			(104, 102, 1, 'Other guide', 'Instruction 4', '', '')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	res, all := postJSON(t, app, "/api/troubleshooting/export?search=no-match", `{"format":"json","scope":"all"}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("all export: %d: %s", res.StatusCode, all)
	}
	res, selected := postJSON(t, app, "/api/troubleshooting/export", `{"format":"json","scope":"selected","ids":[104,105,101,102,103]}`)
	if res.StatusCode != http.StatusOK || all != selected {
		t.Fatalf("all export differs from ordered selection (%d):\nall: %s\nselected: %s", res.StatusCode, all, selected)
	}
	if !strings.Contains(all, `"steps": []`) {
		t.Fatalf("expected guides with no steps to export an empty array: %s", all)
	}
}

func TestTroubleshootingGuideMatchingExportOrder(t *testing.T) {
	app, db := newLinksTestApp(t)
	for _, stmt := range []string{
		`DELETE FROM troubleshooting_guides`,
		`UPDATE troubleshooting_severities SET sort_order = 10 WHERE key = 'critical'`,
		`INSERT INTO troubleshooting_guides (id, title, category, severity, error_code, symptom, created_at, updated_at) VALUES
			(201, 'Zebra', 'execution', 'high', 'ERR:7', 'Symptom', '', ''),
			(202, 'Alpha', 'connection', 'high', 'ERR:8', 'Symptom', '', ''),
			(203, 'Aardvark', 'execution', 'critical', 'err:7', 'Symptom', '', ''),
			(204, 'apple', 'execution', 'high', 'ERR:7', 'Symptom', '', ''),
			(205, 'APPLE', 'execution', 'high', 'ERR:7', 'Symptom', '', ''),
			(206, 'Other category', 'connection', 'high', 'ERR:7', 'Symptom', '', ''),
			(207, 'Other error code', 'execution', 'high', 'ERR:9', 'Symptom', '', '')`,
		`INSERT INTO troubleshooting_steps (id, guide_id, position, title, instruction, script_snippet, verification) VALUES
			(201, 204, 2, 'Third', 'Instruction 3', '', 'Verified'),
			(203, 204, 1, 'Second', 'Instruction 2', '', ''),
			(202, 204, 1, 'First', 'Instruction 1', E'  Run\n\tCheck', ''),
			(204, 202, 1, 'Unmatched guide', 'Instruction 4', '', ''),
			(205, 206, 1, 'Unmatched category', 'Instruction 5', '', ''),
			(206, 207, 1, 'Unmatched error code', 'Instruction 6', '', ''),
			(207, 203, 1, 'Lowercase error code', 'Instruction 7', '', '')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	res, matching := postJSON(t, app, "/api/troubleshooting/export?error_code=ERR:7&category=execution", `{"format":"json","scope":"matching"}`)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("matching export: %d: %s", res.StatusCode, matching)
	}
	res, selected := postJSON(t, app, "/api/troubleshooting/export", `{"format":"json","scope":"selected","ids":[204,205,201,203]}`)
	if res.StatusCode != http.StatusOK || matching != selected {
		t.Fatalf("matching export differs from ordered selection (%d):\nmatching: %s\nselected: %s", res.StatusCode, matching, selected)
	}
	steps, err := store.New(db).ListTroubleshootingExportSteps(context.Background(), store.ListTroubleshootingExportStepsParams{
		Category:  sql.NullString{String: "execution", Valid: true},
		ErrorCode: sql.NullString{String: "ERR:7", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 4 {
		t.Fatalf("matching step read returned %d steps, want the 4 steps of matched guides: %+v", len(steps), steps)
	}
	for _, step := range steps {
		if step.GuideID != 203 && step.GuideID != 204 {
			t.Fatalf("matching step read returned a step of guide %d outside the filters", step.GuideID)
		}
	}
	res, body := postJSON(t, app, "/api/troubleshooting/export?search=no-match", `{"format":"json","scope":"matching"}`)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("empty matching export: %d: %s", res.StatusCode, body)
	}
}
