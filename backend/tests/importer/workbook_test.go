package importer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/importer"
	"webprogfndb/tests/support/workbook"
)

// writeWorkbook builds an xlsx with the six knowledge sheets from
// row-maps, writing header + data rows per sheet.
func writeWorkbook(t *testing.T, path string, sheets map[string][][]string) {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	for _, sheet := range []string{
		"function_main", "parameter_main", "criteria_main",
		"function_parameter", "function_criteria", "example",
	} {
		if _, ok := sheets[sheet]; !ok {
			continue
		}
		f.NewSheet(sheet)
		for i, row := range sheets[sheet] {
			for j, v := range row {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
				if err := f.SetCellValue(sheet, cell, v); err != nil {
					t.Fatalf("set cell: %v", err)
				}
			}
		}
	}
	// excelize creates a default "Sheet1" that no knowledge sheet uses.
	if err := f.DeleteSheet("Sheet1"); err != nil {
		t.Fatalf("delete default sheet: %v", err)
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}
}

func openFixture(t *testing.T) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "testdata", workbook.FixtureFile))
	if err != nil {
		t.Fatalf("fixture missing: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// buildSyntheticWorkbook returns sheets exercising the edge cases the
// ticket names: orphans (links to absent functions), duplicate rows,
// bad enum-ish values, mixed device casing, and the phase typo.
func buildSyntheticWorkbook(t *testing.T) *importer.Workbook {
	f := writeAndOpen(t, map[string][][]string{
		"function_main": {
			{"source", "interface", "name", "description", "convention", "category", "execution_role", "execution_phase"},
			{"BUILT-IN", "RS232", "OPEN", "open port", "1", "communication", "connection", "text_execution"},
		},
		"parameter_main": {
			{"name", "description", "type", "unit", "default_value", "allowed_values", "cardinality", "value_pattern", "required"},
			{"BAUD", "baud rate", "int", "", "9600", "{9600,115200}", "SINGLE", "", "True"},
		},
		"criteria_main": {
			{"name", "description", "type", "unit", "default_value", "allowed_values", "cardinality", "value_pattern", "required"},
			{"POWER", "expected power", "string", "", "", "", "SINGLE", "", "False"},
		},
		"function_parameter": {
			{"name", "function_source", "device_name", "function_name", "source_file", "position", "required"},
			{"BAUD", "BUILT-IN", "GoldenKV", "OPEN", "RD_SETTING.ini", "1", "True"},
			{"GHOST", "DLL", "IOT", "ABSENT_FUNCTION", "", "", "False"},
		},
		"function_criteria": {
			{"name", "function_source", "device_name", "function_name", "source_file", "position", "required"},
			{"POWER", "BUILT-IN", "golden", "OPEN", "", "2", "False"},
		},
		"example": {
			{"function_source", "device_name", "function_name", "script", "meaning"},
			{"BUILT-IN", "iot", "OPEN", "BUILTRS232|OPEN|BAUD=9600|NULL|3000", "opens the port"},
		},
	})
	wb, err := importer.ParseWorkbook(f)
	if err != nil {
		t.Fatalf("parse synthetic workbook: %v", err)
	}
	return wb
}

func writeAndOpen(t *testing.T, sheets map[string][][]string) *os.File {
	t.Helper()
	path := filepath.Join(t.TempDir(), "synthetic.xlsx")
	writeWorkbook(t, path, sheets)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open synthetic: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func TestParseWorkbookRealFixture(t *testing.T) {
	f := openFixture(t)
	wb, err := importer.ParseWorkbook(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// The oracle derives the fixture's counts through ParseWorkbook
	// itself; this round-trip pins the parser↔oracle agreement that every
	// other count assertion in the repo (api tests, e2e) relies on. No
	// literals here: a workbook batch requires zero test edits.
	want, err := workbook.FixtureCounts(filepath.Join("..", "testdata", workbook.FixtureFile))
	if err != nil {
		t.Fatalf("oracle: %v", err)
	}
	if got := workbook.CountsFrom(wb); got != want {
		t.Fatalf("counts = %+v, want oracle %+v", got, want)
	}

	// Device-scoped definitions: ANT and FREQ each have a WIFI_TEST-scoped
	// row and an OSP-scoped row with incompatible semantics (ADR-0009).
	// Parsing must preserve both, not collapse them.
	var antScopes, freqScopes []string
	for _, p := range wb.Parameters {
		switch p.Name {
		case "ANT":
			antScopes = append(antScopes, p.DeviceName)
		case "FREQ":
			freqScopes = append(freqScopes, p.DeviceName)
		}
	}
	if len(antScopes) != 2 {
		t.Fatalf("ANT rows = %d, want 2 (scopes: %v)", len(antScopes), antScopes)
	}
	if len(freqScopes) != 2 {
		t.Fatalf("FREQ rows = %d, want 2 (scopes: %v)", len(freqScopes), freqScopes)
	}
	wantANT := map[string]bool{"WIFI_TEST": true, "OSP": true}
	for _, s := range antScopes {
		if !wantANT[s] {
			t.Fatalf("ANT scope %q not expected (want WIFI_TEST or OSP)", s)
		}
	}
	wantFREQ := map[string]bool{"WIFI_TEST": true, "OSP": true}
	for _, s := range freqScopes {
		if !wantFREQ[s] {
			t.Fatalf("FREQ scope %q not expected (want WIFI_TEST or OSP)", s)
		}
	}
}

func TestParseWorkbookParameterDeviceHeaders(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		header, row []string
		want        string
	}{
		{name: "canonical", header: []string{"name", "function_source", "device_name", "function_name"}, row: []string{"ANT", "DLL", " wifi_test ", "TX"}, want: "WIFI_TEST"},
		{name: "normalized header", header: []string{"name", "function_source", " DEVICE_NAME ", "function_name"}, row: []string{"ANT", "DLL", " wifi_test ", "TX"}, want: "WIFI_TEST"},
		{name: "legacy", header: []string{"name", "function_source", "interface", "function_name"}, row: []string{"ANT", "DLL", " wifi_test ", "TX"}, want: "WIFI_TEST"},
		{name: "canonical wins", header: []string{"name", "function_source", "device_name", "interface", "function_name"}, row: []string{"ANT", "DLL", "OSP", "WIFI_TEST", "TX"}, want: "OSP"},
		{name: "canonical blank", header: []string{"name", "function_source", "device_name", "interface", "function_name"}, row: []string{"ANT", "DLL", "", "WIFI_TEST", "TX"}, want: ""},
		{name: "legacy blank", header: []string{"name", "function_source", "interface", "function_name"}, row: []string{"ANT", "DLL", "", "TX"}, want: ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			file := writeAndOpen(t, map[string][][]string{
				"function_main":      {{"source", "interface", "name"}},
				"parameter_main":     {{"name", "device_name"}},
				"criteria_main":      {{"name", "device_name"}},
				"function_parameter": {testCase.header, testCase.row},
				"function_criteria":  {{"name", "function_source", "device_name", "function_name"}},
				"example":            {{"function_source", "device_name", "function_name", "script", "meaning"}},
			})
			parsed, err := importer.ParseWorkbook(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(parsed.ParameterLinks) != 1 || parsed.ParameterLinks[0].DeviceName != testCase.want {
				t.Fatalf("parameter scope = %+v, want %q", parsed.ParameterLinks, testCase.want)
			}
		})
	}
}

func TestParseWorkbookNormalizations(t *testing.T) {
	wb := buildSyntheticWorkbook(t)

	var phaseFixed bool
	for _, fn := range wb.Functions {
		if fn.Name == "OPEN" && fn.ExecutionPhase == "test_execution" {
			phaseFixed = true
		}
	}
	if !phaseFixed {
		t.Fatalf("execution_phase typo not normalized: %+v", wb.Functions)
	}

	var deviceUpper int
	for _, n := range wb.Normalizations {
		if n.Rule == "device-scope-uppercase" {
			deviceUpper++
		}
	}
	// GoldenKV→GOLDENKV, golden→GOLDEN, iot→IOT — three link/example rows.
	if deviceUpper != 3 {
		t.Fatalf("device-scope normalizations = %d, want 3 (%+v)", deviceUpper, wb.Normalizations)
	}
	if len(wb.CriteriaLinks) != 1 || wb.CriteriaLinks[0].DeviceName != "GOLDEN" {
		t.Fatalf("criteria link device not normalized: %+v", wb.CriteriaLinks)
	}
}
