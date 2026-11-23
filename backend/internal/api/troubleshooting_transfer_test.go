package api

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"webprogfndb/internal/store"
)

type transferTestLookups struct{}

func (transferTestLookups) GetTroubleshootingCategory(_ context.Context, key string) (store.TroubleshootingCategory, error) {
	if key != "connection" {
		return store.TroubleshootingCategory{}, sql.ErrNoRows
	}
	return store.TroubleshootingCategory{}, nil
}

func (transferTestLookups) GetTroubleshootingSeverity(_ context.Context, key string) (store.TroubleshootingSeverity, error) {
	if key != "high" {
		return store.TroubleshootingSeverity{}, sql.ErrNoRows
	}
	return store.TroubleshootingSeverity{}, nil
}

func (transferTestLookups) GetTroubleshootingGuide(_ context.Context, id int64) (store.TroubleshootingGuide, error) {
	if id != 42 {
		return store.TroubleshootingGuide{}, sql.ErrNoRows
	}
	return store.TroubleshootingGuide{ID: id}, nil
}

func TestTroubleshootingTransferPreview(t *testing.T) {
	guide := `"title":"Cable","category":"connection","severity":"high","symptom":"- first\n- second","steps":[{"title":"Check","instruction":"  Check cable\n\n  Reconnect  "}]`
	doc, err := parseTroubleshootingTransfer(strings.NewReader(`{"guides":[{`+guide+`},{"id":42,`+guide+`},{"id":99,"category":"bad","severity":"bad","steps":[{}]},{}]}`), "json")
	if err != nil {
		t.Fatal(err)
	}
	preview, err := previewTroubleshootingTransfer(context.Background(), transferTestLookups{}, &doc)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Valid || len(preview.Rows) != 4 || preview.Rows[0].Action != "create" || preview.Rows[1].Action != "update" || preview.Rows[2].Action != "error" || preview.Rows[3].Action != "error" {
		t.Fatalf("bad preview: %+v", preview)
	}
	reasons := strings.Join(preview.Rows[2].Reasons, " ")
	for _, want := range []string{"unknown category", "unknown severity", "id not found", "title is required", "instruction is required"} {
		if !strings.Contains(reasons, want) {
			t.Errorf("missing %q in %q", want, reasons)
		}
	}
	if !strings.Contains(strings.Join(preview.Rows[3].Reasons, " "), "at least one resolution step") {
		t.Fatalf("missing required step: %+v", preview.Rows[3])
	}
	if got := doc.Guides[0].Steps[0].Instruction; got != "  Check cable\n\n  Reconnect  " {
		t.Fatalf("validation changed multiline text: %q", got)
	}
}

func TestTroubleshootingTransferCombinedDownload(t *testing.T) {
	doc, err := parseTroubleshootingTransfer(strings.NewReader(`{"guides":[{"id":42,"title":"Guide","steps":[{"title":"Step","instruction":"Run"}]}]}`), "json")
	if err != nil {
		t.Fatal(err)
	}
	data, err := zipTroubleshootingTransfer(doc)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(z.File) != 2 {
		t.Fatalf("ZIP: %v %v", z, err)
	}
	want, _ := json.Marshal(doc)
	for _, entry := range z.File {
		f, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		got, err := parseTroubleshootingTransfer(f, strings.TrimPrefix(entry.Name, "troubleshooting-guides."))
		if err := f.Close(); err != nil && err != io.EOF {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(got)
		if err != nil || !bytes.Equal(raw, want) {
			t.Fatalf("%s round trip: %s (%v)", entry.Name, raw, err)
		}
	}
}
