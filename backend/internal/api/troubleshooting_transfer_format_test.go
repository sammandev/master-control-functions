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
		{"Guides", [][]string{{"guide_ref", "title", "category", "severity", "symptom"}, {"alpha", "First", "connection", "high", "symptom"}, {"alpha", "Duplicate", "connection", "high", "symptom"}}},
		{"Steps", [][]string{{"guide_ref", "step_no", "title", "instruction"}, {"alpha", "2", "Second", "second"}, {"alpha", "1", "First", "first"}, {"orphan", "1", "Orphan", "orphan"}}},
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
