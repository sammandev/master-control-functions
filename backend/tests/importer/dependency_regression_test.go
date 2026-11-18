package importer_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/importer"
)

func TestWorkbookDependencyRoundTrip(t *testing.T) {
	input := openFixture(t)
	want, err := importer.ParseWorkbook(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	file, err := excelize.OpenReader(input)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	output, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	got, err := importer.ParseWorkbook(output)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("workbook rewrite changed parsed records or normalization results")
	}
}

func TestParseWorkbookRejectsOversizedRowAfterHeader(t *testing.T) {
	file := excelize.NewFile()
	defer file.Close()
	if err := file.SetSheetName("Sheet1", "function_main"); err != nil {
		t.Fatal(err)
	}
	original, err := file.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(original.Bytes()), int64(original.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var malformed bytes.Buffer
	writer := zip.NewWriter(&malformed)
	for _, entry := range reader.File {
		if entry.Name != "xl/worksheets/sheet1.xml" {
			if err := writer.Copy(entry); err != nil {
				t.Fatal(err)
			}
			continue
		}
		sheet, err := writer.Create(entry.Name)
		if err != nil {
			t.Fatal(err)
		}
		// A valid first row exercises the Columns path, not Next's existing guard.
		_, err = io.WriteString(sheet, `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>source</t></is></c></row><row r="1048577"><c r="A1048577"><v>1</v></c></row></sheetData></worksheet>`)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = importer.ParseWorkbook(&malformed)
	if !errors.Is(err, excelize.ErrMaxRows) {
		t.Fatalf("got %v, want %v", err, excelize.ErrMaxRows)
	}
}
