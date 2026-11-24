package api

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestTroubleshootingTransferRoundTrip(t *testing.T) {
	raw := `{"guides":[{"id":42,"title":"Network","category":"connection","severity":"high","symptom":"1. Timeout\n2. Reset","steps":[{"title":"Reset","instruction":"- Exit\n- Reconnect"}]}]}`
	doc, err := parseTroubleshootingTransfer(strings.NewReader(raw), "json")
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Guides[0]; got.ID == nil || *got.ID != 42 || got.Steps[0].Instruction != "- Exit\n- Reconnect" {
		t.Fatalf("decode: %+v", got)
	}
	for _, format := range []string{"json", "xlsx"} {
		t.Run(format, func(t *testing.T) {
			data, err := writeTroubleshootingTransfer(doc, format, nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := parseTroubleshootingTransfer(bytes.NewReader(data), format)
			if err != nil || !reflect.DeepEqual(doc, got) {
				t.Fatalf("round trip: %+v, %v", got, err)
			}
		})
	}
}

func TestTroubleshootingTransferJSONInvalidIDIsNotCreate(t *testing.T) {
	raw := `{"guides":[{"id":null},{"id":0},{"id":"unknown"},{}]}`
	doc, err := parseTroubleshootingTransfer(strings.NewReader(raw), "json")
	if err != nil {
		t.Fatal(err)
	}
	for i, guide := range doc.Guides[:3] {
		if len(guide.issues) != 1 || !strings.Contains(guide.issues[0], "id must be a positive integer") {
			t.Errorf("guide %d missing invalid id error: %+v", i+1, guide)
		}
	}
	if len(doc.Guides[3].issues) != 0 || doc.Guides[3].ID != nil {
		t.Fatalf("absent id should be valid for create: %+v", doc.Guides[3])
	}
}

func TestTroubleshootingTransferXLSXRejectsUnexpectedHeaders(t *testing.T) {
	for _, tc := range []struct {
		sheet     string
		cell      string
		newHeader string
		want      []string
	}{
		{"Guides", "B1", "guide_id", []string{`missing column "id"`, `unknown column "guide_id"`}},
		{"Steps", "E1", "snippet", []string{`missing column "script_snippet"`, `unknown column "snippet"`}},
		{"Steps", "F1", "", []string{`missing column "verification"`}},
		{"Guides", "K1", "extra_field", []string{`unknown column "extra_field"`}},
	} {
		t.Run(tc.sheet+"/"+tc.cell, func(t *testing.T) {
			data, err := writeTroubleshootingTransfer(troubleshootingTransfer{}, "xlsx", nil)
			if err != nil {
				t.Fatal(err)
			}
			book, err := excelize.OpenReader(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			defer book.Close()
			if err := book.SetCellStr(tc.sheet, tc.cell, tc.newHeader); err != nil {
				t.Fatal(err)
			}
			file, err := book.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			_, err = parseTroubleshootingTransfer(file, "xlsx")
			if err == nil {
				t.Fatal("workbook with a malformed header row was accepted")
			}
			for _, want := range append(tc.want, tc.sheet) {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
		})
	}
}

func TestTroubleshootingTransferXLSXReferenceErrors(t *testing.T) {
	book := excelize.NewFile()
	t.Cleanup(func() { _ = book.Close() })
	if err := book.SetSheetName("Sheet1", "Guides"); err != nil {
		t.Fatal(err)
	}
	if _, err := book.NewSheet("Steps"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []struct {
		sheet string
		rows  [][]string
	}{
		{"Guides", [][]string{transferGuideHeaders, {"alpha", "", "First", "connection", "high", "symptom"}, {"alpha", "", "Duplicate", "connection", "high", "symptom"}}},
		{"Steps", [][]string{transferStepHeaders, {"alpha", "2", "Second", "second"}, {"alpha", "1", "First", "first"}, {"orphan", "1", "Orphan", "orphan"}}},
	} {
		for i, row := range table.rows {
			cell, _ := excelize.CoordinatesToCellName(1, i+1)
			if err := book.SetSheetRow(table.sheet, cell, &row); err != nil {
				t.Fatal(err)
			}
		}
	}
	raw, err := book.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	doc, err := parseTroubleshootingTransfer(raw, "xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(doc.Guides[1].issues, " "), "duplicate guide_ref") ||
		!strings.Contains(strings.Join(doc.issues, " "), "orphan guide_ref") {
		t.Fatalf("missing reference errors: %+v", doc)
	}
	if steps := doc.Guides[0].Steps; len(steps) != 2 || steps[0].Title != "First" || steps[1].Title != "Second" {
		t.Fatalf("steps not sorted: %+v", steps)
	}
}
