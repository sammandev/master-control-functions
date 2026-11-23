// Package importer parses the knowledge workbook and upserts its content
// into SQLite (ADR-0004). Parsing is pure; persistence happens inside one
// transaction so an import is all-or-nothing.
package importer

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Workbook holds the six sheets of the knowledge workbook, mapped to
// domain rows. Normalizations and skipped rows are recorded during
// parsing so the change log can audit every divergence from the raw file.
type Workbook struct {
	Functions      []FunctionRow
	Parameters     []DefinitionRow
	Criteria       []DefinitionRow
	ParameterLinks []LinkRow
	CriteriaLinks  []LinkRow
	Examples       []ExampleRow

	Normalizations []Normalization
	Skipped        []SkippedRow
}

type FunctionRow struct {
	Source, Interface, Name, Description    string
	Convention                              *int64
	Category, ExecutionRole, ExecutionPhase string
}

type DefinitionRow struct {
	Name, Description, Type, Unit, DefaultValue, AllowedValues string
	Cardinality, ValuePattern                                  string
	Required                                                   bool
	// DeviceName is the definition's device scope (ADR-0009). Blank means
	// "global/shared" — the fallback for any link whose device scope has no
	// exact-match definition row.
	DeviceName string
}

type LinkRow struct {
	Name, FunctionSource, DeviceName, FunctionName, SourceFile string
	Position                                                   *int64
	Required                                                   bool
}

type ExampleRow struct {
	FunctionSource, DeviceName, FunctionName, Script, Meaning string
}

// Normalization is one deterministic source-value fix (CONTEXT.md).
type Normalization struct {
	Table      string `json:"table"`
	RowKey     string `json:"row_key"`
	Column     string `json:"column"`
	Source     string `json:"source"`
	Normalized string `json:"normalized"`
	Rule       string `json:"rule"`
}

// SkippedRow records a row that could not be imported and why.
type SkippedRow struct {
	Table  string `json:"table"`
	RowKey string `json:"row_key"`
	Reason string `json:"reason"`
}

// sheetNames lists the workbook sheets in a fixed processing order.
var sheetOrder = []string{
	"function_main", "parameter_main", "criteria_main",
	"function_parameter", "function_criteria", "example",
}

var requiredHeaders = map[string][]string{
	"function_main":      {"source", "interface", "name"},
	"parameter_main":     {"name"},
	"criteria_main":      {"name"},
	"function_parameter": {"name", "function_source", "function_name"},
	"function_criteria":  {"name", "function_source", "function_name"},
	"example":            {"function_name", "script"},
}

// record accumulates the audit trail while mapping sheet rows.
type record struct {
	norms   []Normalization
	skipped []SkippedRow
}

func (rc *record) norm(table, rowKey, column, source, normalized, rule string) {
	rc.norms = append(rc.norms, Normalization{
		Table: table, RowKey: rowKey, Column: column,
		Source: source, Normalized: normalized, Rule: rule,
	})
}

func (rc *record) skip(table, rowKey, reason string) {
	rc.skipped = append(rc.skipped, SkippedRow{Table: table, RowKey: rowKey, Reason: reason})
}

// ParseWorkbook reads an xlsx stream and maps every sheet to rows,
// applying Normalizations deterministically along the way.
func ParseWorkbook(r io.Reader) (*Workbook, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("open workbook: %w", err)
	}
	defer f.Close()

	wb := &Workbook{}
	rc := &record{}
	for _, sheet := range sheetOrder {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, fmt.Errorf("sheet %s: %w", sheet, err)
		}
		var header []string
		var data [][]string
		if len(rows) > 0 {
			header, data = rows[0], rows[1:]
		}
		m := indexHeader(header)
		if err := validateHeaders(sheet, m); err != nil {
			return nil, err
		}
		switch sheet {
		case "function_main":
			wb.Functions = parseFunctions(data, m, rc)
		case "parameter_main":
			wb.Parameters = parseDefinitions(data, m, "parameter_definitions", rc)
		case "criteria_main":
			wb.Criteria = parseDefinitions(data, m, "criteria_definitions", rc)
		case "function_parameter":
			wb.ParameterLinks = parseLinks(data, m, "parameter_links", rc)
		case "function_criteria":
			wb.CriteriaLinks = parseLinks(data, m, "criteria_links", rc)
		case "example":
			wb.Examples = parseExamples(data, m, rc)
		}
	}
	wb.Normalizations = rc.norms
	wb.Skipped = rc.skipped
	return wb, nil
}

func validateHeaders(sheet string, header map[string]int) error {
	var missing []string
	for _, col := range requiredHeaders[sheet] {
		if _, ok := header[col]; ok {
			continue
		}
		if sheet == "function_main" && col == "interface" {
			if _, ok := header["device_name"]; ok {
				continue
			}
			col = "interface (or device_name)"
		}
		missing = append(missing, col)
	}
	if len(missing) > 0 {
		return fmt.Errorf("sheet %s: missing required columns: %s", sheet, strings.Join(missing, ", "))
	}
	return nil
}

// indexHeader maps lowercased column names to their column index.
func indexHeader(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func cell(row []string, m map[string]int, col string) string {
	i, ok := m[col]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// parseBool reads the workbook's required flag ("True"/"true"/"1").
func parseBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true
	}
	return false
}

// parseNullableInt reads "10" or "10.0" (Excel float formatting).
func parseNullableInt(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("not a number: %q", s)
	}
	v := int64(f)
	return &v, nil
}

// normalizePhase fixes the known execution-phase typo.
func normalizePhase(phase string) (string, bool) {
	if phase == "text_execution" {
		return "test_execution", true
	}
	return phase, false
}

// parseFunctions maps the function_main sheet, normalizing the known
// execution-phase typo (text_execution → test_execution).
func parseFunctions(data [][]string, m map[string]int, rc *record) []FunctionRow {
	out := make([]FunctionRow, 0, len(data))
	interfaceCol := "interface"
	if _, exists := m[interfaceCol]; !exists {
		interfaceCol = "device_name"
	}
	for i, row := range data {
		name := cell(row, m, "name")
		if name == "" {
			rc.skip("functions", fmt.Sprintf("(row %d)", i+2), "missing name")
			continue
		}
		r := FunctionRow{
			Source:      cell(row, m, "source"),
			Interface:   cell(row, m, interfaceCol),
			Name:        name,
			Description: cell(row, m, "description"),
			Category:    cell(row, m, "category"),
		}
		if conv, err := parseNullableInt(cell(row, m, "convention")); err == nil {
			r.Convention = conv
		} else if raw := cell(row, m, "convention"); raw != "" {
			rc.norm("functions", r.Source+"/"+r.Interface+"/"+r.Name, "convention",
				raw, "(empty)", "unparseable-number-dropped")
		}
		r.ExecutionRole = cell(row, m, "execution_role")
		r.ExecutionPhase = cell(row, m, "execution_phase")
		if fixed, changed := normalizePhase(r.ExecutionPhase); changed {
			rc.norm("functions", r.Source+"/"+r.Interface+"/"+r.Name, "execution_phase",
				r.ExecutionPhase, fixed, "phase-typo-text-execution")
			r.ExecutionPhase = fixed
		}
		out = append(out, r)
	}
	return out
}

// parseDefinitions maps parameter_main / criteria_main. The device_name
// column carries the scope (ADR-0009); blank = global definition. Casing
// follows the ADR-0002 uppercase rule; deviations are recorded in the
// change log under rule "device-scope-uppercase", matching the rule used
// for link and example device scopes.
func parseDefinitions(data [][]string, m map[string]int, table string, rc *record) []DefinitionRow {
	out := make([]DefinitionRow, 0, len(data))
	for _, row := range data {
		name := cell(row, m, "name")
		if name == "" {
			continue
		}
		// TrimSpace joins the ToUpper rule (ADR-0002): a padded workbook
		// cell must not diverge from the API's canonical device values.
		device := strings.TrimSpace(cell(row, m, "device_name"))
		if fixed := strings.ToUpper(device); device != fixed {
			rc.norm(table, name, "device_name",
				device, fixed, "device-scope-uppercase")
			device = fixed
		}
		out = append(out, DefinitionRow{
			Name:          name,
			Description:   cell(row, m, "description"),
			Type:          cell(row, m, "type"),
			Unit:          cell(row, m, "unit"),
			DefaultValue:  cell(row, m, "default_value"),
			AllowedValues: cell(row, m, "allowed_values"),
			Cardinality:   cell(row, m, "cardinality"),
			ValuePattern:  cell(row, m, "value_pattern"),
			Required:      parseBool(cell(row, m, "required")),
			DeviceName:    device,
		})
	}
	return out
}

// parseLinks maps function_parameter / function_criteria, normalizing
// device-scope casing (ADR-0002: UPPER-CASE is canonical).
func parseLinks(data [][]string, m map[string]int, table string, rc *record) []LinkRow {
	out := make([]LinkRow, 0, len(data))
	deviceColumn := "device_name"
	if _, exists := m[deviceColumn]; !exists && table == "parameter_links" {
		deviceColumn = "interface"
	}
	for i, row := range data {
		name := cell(row, m, "name")
		fn := cell(row, m, "function_name")
		if name == "" || fn == "" {
			rc.skip(table, fmt.Sprintf("(row %d)", i+2), "missing name or function_name")
			continue
		}
		r := LinkRow{
			Name:           name,
			FunctionSource: cell(row, m, "function_source"),
			FunctionName:   fn,
			SourceFile:     cell(row, m, "source_file"),
			Required:       parseBool(cell(row, m, "required")),
		}
		device := strings.TrimSpace(cell(row, m, deviceColumn))
		if fixed := strings.ToUpper(device); device != fixed {
			rc.norm(table, r.FunctionSource+"/"+r.FunctionName+"/"+name, "device_name",
				device, fixed, "device-scope-uppercase")
			device = fixed
		}
		r.DeviceName = device
		if pos, err := parseNullableInt(cell(row, m, "position")); err == nil {
			r.Position = pos
		} else if raw := cell(row, m, "position"); raw != "" {
			rc.norm(table, r.FunctionSource+"/"+r.FunctionName+"/"+name, "position",
				raw, "(empty)", "unparseable-number-dropped")
		}
		out = append(out, r)
	}
	return out
}

// parseExamples maps the example sheet with the same device-scope rule.
func parseExamples(data [][]string, m map[string]int, rc *record) []ExampleRow {
	out := make([]ExampleRow, 0, len(data))
	for i, row := range data {
		fn := cell(row, m, "function_name")
		if fn == "" {
			rc.skip("examples", fmt.Sprintf("(row %d)", i+2), "missing function_name")
			continue
		}
		r := ExampleRow{
			FunctionName: fn,
			Script:       cell(row, m, "script"),
			Meaning:      cell(row, m, "meaning"),
		}
		device := strings.TrimSpace(cell(row, m, "device_name"))
		if fixed := strings.ToUpper(device); device != fixed {
			rc.norm("examples", r.FunctionName, "device_name",
				device, fixed, "device-scope-uppercase")
			device = fixed
		}
		r.DeviceName = device
		r.FunctionSource = cell(row, m, "function_source")
		out = append(out, r)
	}
	return out
}
