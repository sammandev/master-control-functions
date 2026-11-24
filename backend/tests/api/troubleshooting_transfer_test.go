package api_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"
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
