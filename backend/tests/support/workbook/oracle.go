package workbook

import (
	"bytes"
	"fmt"
	"os"

	"webprogfndb/internal/importer"
)

// FixtureFile is the testdata workbook the count oracle derives from.
const FixtureFile = "Master_Control_V2_Script_Definition_Knowledge.xlsx"

// Counts holds the per-table row counts of a parsed workbook.
type Counts struct {
	Functions      int
	Parameters     int
	Criteria       int
	ParameterLinks int
	CriteriaLinks  int
	Examples       int
}

// CountsFrom derives the per-table counts from an already-parsed workbook.
func CountsFrom(wb *importer.Workbook) Counts {
	return Counts{
		Functions:      len(wb.Functions),
		Parameters:     len(wb.Parameters),
		Criteria:       len(wb.Criteria),
		ParameterLinks: len(wb.ParameterLinks),
		CriteriaLinks:  len(wb.CriteriaLinks),
		Examples:       len(wb.Examples),
	}
}

// FixtureCounts opens the workbook at path (the shared testdata
// fixture), parses it through the same pure ParseWorkbook the import
// path uses, and returns its per-table counts. Tests assert against
// this oracle instead of pinned literals, so a workbook batch requires
// zero test edits. Callers resolve path relative to their own package
// directory - go test's working directory.
func FixtureCounts(path string) (Counts, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Counts{}, fmt.Errorf("read fixture: %w", err)
	}
	wb, err := importer.ParseWorkbook(bytes.NewReader(b))
	if err != nil {
		return Counts{}, fmt.Errorf("parse fixture: %w", err)
	}
	return CountsFrom(wb), nil
}
