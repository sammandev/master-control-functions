package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/importer"
)

// importFixture imports the real source workbook with reset.
func importFixture(t *testing.T, app *fiber.App) string {
	t.Helper()
	ct, body := multipartBody(t, "file", "master.xlsx", fixtureBytes(t))
	resp, res := postFile(t, app, "/api/import/reset", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("import: %d (body: %s)", resp.StatusCode, res)
	}
	return res
}

// exportWorkbookBytes downloads the workbook via the export endpoint.
func exportWorkbookBytes(t *testing.T, app *fiber.App) ([]byte, *http.Response) {
	t.Helper()
	resp, b := doRequest(t, app, "GET", "/api/export")
	return []byte(b), resp
}

// TestExportRoundTripCellByCell: import the source workbook → export →
// the exported file must carry the same sheet names, the same header
// rows cell-by-cell, and the same (normalized) rows.
func TestExportRoundTripCellByCell(t *testing.T) {
	app := newImportTestApp(t)
	importFixture(t, app)

	xlsx, resp := exportWorkbookBytes(t, app)
	if resp.StatusCode != 200 {
		t.Fatalf("export: %d (body: %s)", resp.StatusCode, string(xlsx))
	}
	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "spreadsheetml") {
		t.Fatalf("content type: %s", got)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.Contains(got, "attachment") {
		t.Fatalf("disposition: %s", got)
	}

	exported, err := importer.ParseWorkbook(bytes.NewReader(xlsx))
	if err != nil {
		t.Fatalf("parse exported: %v", err)
	}
	source, err := importer.ParseWorkbook(bytes.NewReader(fixtureBytes(t)))
	if err != nil {
		t.Fatalf("parse source: %v", err)
	}

	// Headers and sheet names, cell-by-cell, straight from both files.
	wantSheets := []string{"function_main", "parameter_main", "criteria_main",
		"function_parameter", "function_criteria", "example"}
	sf, err := excelize.OpenReader(bytes.NewReader(fixtureBytes(t)))
	if err != nil {
		t.Fatalf("open source: %v", err)
	}
	defer sf.Close()
	ef, err := excelize.OpenReader(bytes.NewReader(xlsx))
	if err != nil {
		t.Fatalf("open exported: %v", err)
	}
	defer ef.Close()
	if got := ef.GetSheetList(); !reflect.DeepEqual(got, wantSheets) {
		t.Fatalf("sheet list: %v", got)
	}
	for _, sheet := range wantSheets {
		want, err := sf.GetRows(sheet)
		if err != nil || len(want) == 0 {
			t.Fatalf("source sheet %s: %v", sheet, err)
		}
		got, err := ef.GetRows(sheet)
		if err != nil || len(got) == 0 {
			t.Fatalf("exported sheet %s: %v", sheet, err)
		}
		if sheet == "function_parameter" {
			for index, header := range want[0] {
				if header == "interface" {
					want[0][index] = "device_name"
				}
			}
			if len(got[0]) < 3 || got[0][2] != "device_name" {
				t.Fatalf("parameter scope header: %q", got[0])
			}
		}
		if !reflect.DeepEqual(got[0], want[0]) {
			t.Fatalf("%s header:\n got: %q\nwant: %q", sheet, got[0], want[0])
		}
	}

	// Row content: ParseWorkbook already applied the import normalizations
	// (device-scope upper-case, phase typo, dropped unparseable numbers)
	// and skipped unimportable rows — that parsed state IS the database
	// state, so a faithful export re-parses to the same rows.
	compareRows(t, "functions", source.Functions, exported.Functions,
		func(i int, a, b importer.FunctionRow) {
			t.Fatalf("functions[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})
	compareRows(t, "parameters", source.Parameters, exported.Parameters,
		func(i int, a, b importer.DefinitionRow) {
			t.Fatalf("parameters[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})
	compareRows(t, "criteria", source.Criteria, exported.Criteria,
		func(i int, a, b importer.DefinitionRow) {
			t.Fatalf("criteria[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})
	compareRows(t, "parameter links", source.ParameterLinks, exported.ParameterLinks,
		func(i int, a, b importer.LinkRow) {
			t.Fatalf("parameter links[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})
	compareRows(t, "criteria links", source.CriteriaLinks, exported.CriteriaLinks,
		func(i int, a, b importer.LinkRow) {
			t.Fatalf("criteria links[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})
	compareRows(t, "examples", source.Examples, exported.Examples,
		func(i int, a, b importer.ExampleRow) {
			t.Fatalf("examples[%d]:\n got: %+v\nwant: %+v", i, b, a)
		})

	// The exported file carries normalized values: parsing it again must
	// produce no new normalizations and no skipped rows.
	if len(exported.Normalizations) != 0 || len(exported.Skipped) != 0 {
		t.Fatalf("exported workbook not clean: norms=%v skipped=%v",
			exported.Normalizations, exported.Skipped)
	}
}

// compareRows asserts equal length + pairwise equality of a row slice.
func compareRows[T any](t *testing.T, what string, want []T, got []T, fail func(i int, a, b T)) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: exported %d rows, source %d", what, len(got), len(want))
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			fail(i, want[i], got[i])
		}
	}
}

// TestExportReimportIdempotent: importing the exported workbook changes
// nothing — same counts, no skips, no normalizations, and a second export
// matches the first.
func TestParameterDeviceHeaderImportExportCompatibility(t *testing.T) {
	for _, endpoint := range []string{"/api/import", "/api/import/reset"} {
		for _, header := range []string{"device_name", "interface", "canonical blank"} {
			t.Run(endpoint+"/"+header, func(t *testing.T) {
				app := newImportTestApp(t)
				if resp, body := postJSON(t, app, "/api/devices", `{"device_name":"MANUAL_DEVICE","source":"BUILT-IN"}`); resp.StatusCode != 201 {
					t.Fatalf("register Type: %d %s", resp.StatusCode, body)
				}
				file, err := excelize.OpenReader(bytes.NewReader(syntheticWorkbook(t)))
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				wantScope := "GOLDENKV"
				if header == "canonical blank" {
					wantScope = ""
					for cell, value := range map[string]string{"C2": "", "C3": "", "H1": "interface", "H2": "WIFI_TEST", "H3": "WIFI_TEST"} {
						if err := file.SetCellValue("function_parameter", cell, value); err != nil {
							t.Fatal(err)
						}
					}
				} else if err := file.SetCellValue("function_parameter", "C1", header); err != nil {
					t.Fatal(err)
				}
				buffer, err := file.WriteToBuffer()
				if err != nil {
					t.Fatal(err)
				}
				contentType, body := multipartBody(t, "file", "scope.xlsx", buffer.Bytes())
				if resp, result := postFile(t, app, endpoint, contentType, body); resp.StatusCode != 200 {
					t.Fatalf("import: %d %s", resp.StatusCode, result)
				}
				exportedBytes, resp := exportWorkbookBytes(t, app)
				if resp.StatusCode != 200 {
					t.Fatalf("export: %d", resp.StatusCode)
				}
				exportedFile, err := excelize.OpenReader(bytes.NewReader(exportedBytes))
				if err != nil {
					t.Fatal(err)
				}
				defer exportedFile.Close()
				if got, err := exportedFile.GetCellValue("function_parameter", "C1"); err != nil || got != "device_name" {
					t.Fatalf("export scope header: %q %v", got, err)
				}
				parsed, err := importer.ParseWorkbook(bytes.NewReader(exportedBytes))
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, link := range parsed.ParameterLinks {
					if link.Name == "BAUD" {
						found = true
						if link.DeviceName != wantScope {
							t.Fatalf("exported scope: %+v, want %q", link, wantScope)
						}
					}
				}
				if !found {
					t.Fatal("parameter link missing from export")
				}
				_, result := doRequest(t, app, "GET", "/api/devices")
				var index struct {
					Items []struct {
						DeviceName string `json:"device_name"`
						Source     string `json:"source"`
					} `json:"items"`
				}
				if err := json.Unmarshal([]byte(result), &index); err != nil {
					t.Fatal(err)
				}
				found = false
				for _, item := range index.Items {
					if item.DeviceName == "MANUAL_DEVICE" {
						found = item.Source == "BUILT-IN"
					}
				}
				if !found {
					t.Fatalf("Type lost during import/reset: %s", result)
				}
			})
		}
	}
}

func TestExportReimportIdempotent(t *testing.T) {
	app := newImportTestApp(t)
	importFixture(t, app)

	xlsx1, _ := exportWorkbookBytes(t, app)

	ct, body := multipartBody(t, "file", "export.xlsx", xlsx1)
	resp, res := postFile(t, app, "/api/import", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("re-import: %d (body: %s)", resp.StatusCode, res)
	}
	// Counts come from the fixture oracle (workbook batches edit the
	// workbook, not this test); re-import must also stay clean.
	for _, want := range append(countKeyValues(fixtureCounts(t)), `"skipped":null`, `"normalizations":null`) {
		if !strings.Contains(res, want) {
			t.Fatalf("re-import result missing %s: %s", want, res)
		}
	}

	xlsx2, _ := exportWorkbookBytes(t, app)
	first, err := importer.ParseWorkbook(bytes.NewReader(xlsx1))
	if err != nil {
		t.Fatalf("parse export 1: %v", err)
	}
	second, err := importer.ParseWorkbook(bytes.NewReader(xlsx2))
	if err != nil {
		t.Fatalf("parse export 2: %v", err)
	}
	if !reflect.DeepEqual(first.Functions, second.Functions) ||
		!reflect.DeepEqual(first.Parameters, second.Parameters) ||
		!reflect.DeepEqual(first.Criteria, second.Criteria) ||
		!reflect.DeepEqual(first.ParameterLinks, second.ParameterLinks) ||
		!reflect.DeepEqual(first.CriteriaLinks, second.CriteriaLinks) ||
		!reflect.DeepEqual(first.Examples, second.Examples) {
		t.Fatalf("export changed after re-import — the re-import was not idempotent")
	}
}

// TestExportIncludesOrphans: orphan link rows and orphan examples exist in
// the source file, so the export must carry them faithfully — with the
// normalized device scope, exactly as the database stores them.
func TestExportIncludesOrphans(t *testing.T) {
	app := newImportTestApp(t)
	ct, body := multipartBody(t, "file", "synthetic.xlsx", syntheticWorkbook(t))
	resp, res := postFile(t, app, "/api/import/reset", ct, body)
	if resp.StatusCode != 200 {
		t.Fatalf("import: %d (body: %s)", resp.StatusCode, res)
	}

	xlsx, _ := exportWorkbookBytes(t, app)
	exported, err := importer.ParseWorkbook(bytes.NewReader(xlsx))
	if err != nil {
		t.Fatalf("parse exported: %v", err)
	}

	// The GHOST→ABSENT_FUNCTION link (position empty, source_file empty)
	// must survive, with device scope upper-cased as imported.
	var ghost *importer.LinkRow
	for i := range exported.ParameterLinks {
		l := exported.ParameterLinks[i]
		if l.Name == "GHOST" {
			ghost = &l
		}
	}
	if ghost == nil {
		t.Fatalf("orphan GHOST link not exported: %+v", exported.ParameterLinks)
	}
	if ghost.FunctionSource != "DLL" || ghost.FunctionName != "ABSENT_FUNCTION" ||
		ghost.DeviceName != "IOT" || ghost.Position != nil || ghost.SourceFile != "" {
		t.Fatalf("orphan link values wrong: %+v", ghost)
	}

	// The criteria link with lowercase device scope exports normalized.
	if len(exported.CriteriaLinks) != 1 || exported.CriteriaLinks[0].DeviceName != "GOLDEN" {
		t.Fatalf("criteria links wrong: %+v", exported.CriteriaLinks)
	}

	// The example with lowercase device scope exports normalized.
	if len(exported.Examples) != 1 || exported.Examples[0].DeviceName != "IOT" {
		t.Fatalf("examples wrong: %+v", exported.Examples)
	}
}

// TestExportEmptyDatabase: exporting a fresh database yields a valid
// workbook with the six sheets and headers but no rows.
func TestExportEmptyDatabase(t *testing.T) {
	app := newImportTestApp(t)
	xlsx, resp := exportWorkbookBytes(t, app)
	if resp.StatusCode != 200 {
		t.Fatalf("export: %d", resp.StatusCode)
	}
	wb, err := importer.ParseWorkbook(bytes.NewReader(xlsx))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(wb.Functions)+len(wb.Parameters)+len(wb.Criteria)+
		len(wb.ParameterLinks)+len(wb.CriteriaLinks)+len(wb.Examples) != 0 {
		t.Fatalf("empty database exported rows: %+v", wb)
	}
}
