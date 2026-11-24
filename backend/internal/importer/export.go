package importer

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"

	"github.com/xuri/excelize/v2"

	"webprogfndb/internal/store"
)

// Export regenerates the six-sheet knowledge workbook from the database
// (ADR-0004): original sheet names, original column layout, and the
// database's (normalized) values as cell values — orphans included, since
// they exist in the source file too and re-import must round-trip them.

// sheetHeaders is the source file's original column layout per sheet; the
// parser reads these columns by lowercased header name.
//
// parameter_main / criteria_main carry device_name directly after name
// (ADR-0009). Blank device_name on export stays blank — round-trip must
// preserve the global/shared definition rows.
var sheetHeaders = map[string][]string{
	"function_main": {"source", "interface", "name", "description", "convention",
		"category", "execution_role", "execution_phase"},
	"parameter_main": {"name", "device_name", "description", "type", "unit", "default_value",
		"allowed_values", "cardinality", "value_pattern", "required"},
	"criteria_main": {"name", "device_name", "description", "type", "unit", "default_value",
		"allowed_values", "cardinality", "value_pattern", "required"},
	"function_parameter": {"name", "function_source", "device_name",
		"function_name", "source_file", "position", "required"},
	"function_criteria": {"name", "function_source", "device_name",
		"function_name", "source_file", "position", "required"},
	"example": {"function_source", "device_name", "function_name", "script", "meaning"},
}

// boolCell renders the workbook's required flag.
func boolCell(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

// nullableIntCell renders a NULL-able integer as "" or decimal digits.
func nullableIntCell(v sql.NullInt64) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatInt(v.Int64, 10)
}

// nullableStringCell renders a NULL-able text as "" or its value.
func nullableStringCell(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

type exportRow struct {
	id     int64
	name   string
	values []interface{}
}

type ExportCellLimitError struct {
	Sheet, Cell, Field, Name string
	ID                       int64
	Length                   int
}

func (e *ExportCellLimitError) Error() string {
	return fmt.Sprintf("Cannot export workbook: %s record #%d (%.80q), field %s (cell %s), contains %d UTF-16 code units. Excel allows at most %d; the stored value has not changed.",
		e.Sheet, e.ID, e.Name, e.Field, e.Cell, e.Length, excelize.TotalCellChars)
}

// Export reads every table and writes the six-sheet xlsx to w. Rows keep
// their import order (id order), which is the source file's row order.
func Export(ctx context.Context, db *sql.DB, w io.Writer) error {
	q := store.New(db)

	fns, err := q.ListFunctions(ctx, store.ListFunctionsParams{})
	if err != nil {
		return fmt.Errorf("list functions: %w", err)
	}
	params, err := q.ListParameterDefinitions(ctx, store.ListParameterDefinitionsParams{})
	if err != nil {
		return fmt.Errorf("list parameter definitions: %w", err)
	}
	criteria, err := q.ListCriteriaDefinitions(ctx, store.ListCriteriaDefinitionsParams{})
	if err != nil {
		return fmt.Errorf("list criteria definitions: %w", err)
	}
	pLinks, err := q.ListAllParameterLinks(ctx)
	if err != nil {
		return fmt.Errorf("list parameter links: %w", err)
	}
	cLinks, err := q.ListAllCriteriaLinks(ctx)
	if err != nil {
		return fmt.Errorf("list criteria links: %w", err)
	}
	examples, err := q.ListAllExamples(ctx)
	if err != nil {
		return fmt.Errorf("list examples: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()
	rows := map[string][]exportRow{
		"function_main":      functionCells(fns),
		"parameter_main":     parameterDefinitionCells(params),
		"criteria_main":      criteriaDefinitionCells(criteria),
		"function_parameter": parameterLinkCells(pLinks),
		"function_criteria":  criteriaLinkCells(cLinks),
		"example":            exampleCells(examples),
	}
	first := true
	for _, sheet := range sheetOrder {
		if first {
			// NewFile starts with a single sheet; rename it to the first
			// knowledge sheet instead of juggling an extra delete.
			if err := f.SetSheetName("Sheet1", sheet); err != nil {
				return fmt.Errorf("rename sheet %s: %w", sheet, err)
			}
			first = false
		} else if _, err := f.NewSheet(sheet); err != nil {
			return fmt.Errorf("create sheet %s: %w", sheet, err)
		}
		header := sheetHeaders[sheet]
		if err := f.SetSheetRow(sheet, "A1", &header); err != nil {
			return fmt.Errorf("header %s: %w", sheet, err)
		}
		for i, r := range rows[sheet] {
			cellName, err := excelize.CoordinatesToCellName(1, i+2)
			if err != nil {
				return fmt.Errorf("cell name for row %d of %s: %w", i+2, sheet, err)
			}
			for j, value := range r.values {
				text, ok := value.(string)
				if !ok {
					continue
				}
				if n := cellTextLength(text); n > excelize.TotalCellChars {
					cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
					return &ExportCellLimitError{Sheet: sheet, Cell: cell, Field: header[j], Name: r.name, ID: r.id, Length: n}
				}
			}
			if err := f.SetSheetRow(sheet, cellName, &r.values); err != nil {
				return fmt.Errorf("row %d of %s: %w", i+2, sheet, err)
			}
		}
	}
	return f.Write(w)
}

func functionCells(rows []store.Function) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.Name, []interface{}{
			r.Source, r.Interface, r.Name, r.Description,
			nullableIntCell(r.Convention), r.Category, r.ExecutionRole, r.ExecutionPhase,
		}})
	}
	return out
}

func parameterDefinitionCells(rows []store.ParameterDefinition) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.Name, definitionRow(r.Name, r.DeviceName, r.Description, r.Type, r.Unit,
			r.DefaultValue, r.AllowedValues, r.Cardinality, r.ValuePattern, r.Required)})
	}
	return out
}

func criteriaDefinitionCells(rows []store.CriteriaDefinition) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.Name, definitionRow(r.Name, r.DeviceName, r.Description, r.Type, r.Unit,
			r.DefaultValue, r.AllowedValues, r.Cardinality, r.ValuePattern, r.Required)})
	}
	return out
}

// definitionRow renders one parameter_main/criteria_main row in the
// original column layout (both generated types share the shape, so the
// row builder is shared). deviceName is the ADR-0009 scope; blank stays
// blank — round-trip must preserve the global/shared definition rows.
func definitionRow(name, deviceName, description, typ, unit string, defaultValue sql.NullString,
	allowedValues, cardinality, valuePattern string, required int64,
) []interface{} {
	return []interface{}{
		name, deviceName, description, typ, unit,
		nullableStringCell(defaultValue), allowedValues, cardinality, valuePattern,
		boolCell(required != 0),
	}
}

func parameterLinkCells(rows []store.ParameterLink) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.Name, linkRow(r.Name, r.FunctionSource, r.DeviceName, r.FunctionName,
			r.SourceFile, r.Position, r.Required)})
	}
	return out
}

func criteriaLinkCells(rows []store.CriteriaLink) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.Name, linkRow(r.Name, r.FunctionSource, r.DeviceName, r.FunctionName,
			r.SourceFile, r.Position, r.Required)})
	}
	return out
}

// linkRow renders one function_parameter/function_criteria row in the
// original column layout.
func linkRow(name, functionSource, deviceName, functionName, sourceFile string,
	position sql.NullInt64, required int64,
) []interface{} {
	return []interface{}{
		name, functionSource, deviceName, functionName, sourceFile,
		nullableIntCell(position), boolCell(required != 0),
	}
}

func exampleCells(rows []store.Example) []exportRow {
	out := make([]exportRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, exportRow{r.ID, r.FunctionName, []interface{}{
			r.FunctionSource, r.DeviceName, r.FunctionName, r.Script, r.Meaning,
		}})
	}
	return out
}
