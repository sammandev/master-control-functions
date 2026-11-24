package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/xuri/excelize/v2"
)

type troubleshootingTransferGuide struct {
	ID *int64 `json:"id,omitempty"`
	troubleshootingBody
	issues []string
}

type troubleshootingTransfer struct {
	Guides []troubleshootingTransferGuide `json:"guides"`
	issues []string
}

var transferGuideHeaders = []string{"guide_ref", "id", "title", "category", "severity", "symptom", "root_cause", "error_code", "device_name", "function_name"}
var transferStepHeaders = []string{"guide_ref", "step_no", "title", "instruction", "script_snippet", "verification"}

var transferGuideFieldLabels = []string{"Guide Reference", "ID", "Title", "Category", "Severity", "Symptom", "Root Cause", "Error Code", "Device", "Function"}
var transferStepFieldLabels = []string{"Guide Reference", "Step Number", "Title", "Instruction", "Snippet", "Verification"}

type troubleshootingXLSXCellLimitError struct {
	guide string
	field string
}

func (e *troubleshootingXLSXCellLimitError) Error() string {
	return fmt.Sprintf("Guide %s: %s exceeds Excel's 32,767-character cell limit. Export as JSON instead.", e.guide, e.field)
}

func parseTroubleshootingTransfer(r io.Reader, format string) (troubleshootingTransfer, error) {
	switch format {
	case "json":
		d := json.NewDecoder(r)
		d.DisallowUnknownFields()
		var doc troubleshootingTransfer
		var input struct {
			Guides []struct {
				ID json.RawMessage `json:"id"`
				troubleshootingBody
			} `json:"guides"`
		}
		if err := d.Decode(&input); err != nil {
			return doc, fmt.Errorf("invalid JSON: %w", err)
		}
		if d.Decode(new(any)) != io.EOF || input.Guides == nil {
			return doc, errors.New("expected one JSON object containing a guides array")
		}
		doc.Guides = make([]troubleshootingTransferGuide, 0, len(input.Guides))
		for _, item := range input.Guides {
			guide := troubleshootingTransferGuide{troubleshootingBody: item.troubleshootingBody}
			if item.ID != nil {
				var id int64
				if err := json.Unmarshal(item.ID, &id); err != nil || id <= 0 {
					guide.issues = append(guide.issues, "id must be a positive integer")
				} else {
					guide.ID = &id
				}
			}
			doc.Guides = append(doc.Guides, guide)
		}
		return doc, nil
	case "xlsx":
		return readTroubleshootingXLSX(r)
	default:
		return troubleshootingTransfer{}, errors.New("choose a .json or .xlsx file")
	}
}

func transferSheet(book *excelize.File, name string, expected []string) ([][]string, map[string]int, error) {
	i, err := book.GetSheetIndex(name)
	if err != nil || i < 0 {
		return nil, nil, fmt.Errorf("missing %s sheet", name)
	}
	rows, err := book.GetRows(name, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, nil, fmt.Errorf("%s sheet: %w", name, err)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("%s sheet is missing columns: %s", name, strings.Join(expected, ", "))
	}
	cols := make(map[string]int)
	var problems []string
	for i, header := range rows[0] {
		header = strings.TrimSpace(header)
		if header == "" {
			continue
		}
		if _, exists := cols[header]; exists {
			problems = append(problems, fmt.Sprintf("repeats column %q", header))
		} else if !slices.Contains(expected, header) {
			problems = append(problems, fmt.Sprintf("unknown column %q", header))
		}
		cols[header] = i
	}
	for _, header := range expected {
		if _, exists := cols[header]; !exists {
			problems = append(problems, fmt.Sprintf("missing column %q", header))
		}
	}
	if len(problems) > 0 {
		return nil, nil, fmt.Errorf("%s sheet: %s", name, strings.Join(problems, "; "))
	}
	return rows[1:], cols, nil
}

func transferCell(row []string, cols map[string]int, name string) string {
	if i, exists := cols[name]; exists && i < len(row) {
		return row[i]
	}
	return ""
}

func emptyTransferRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func readTroubleshootingXLSX(r io.Reader) (troubleshootingTransfer, error) {
	doc := troubleshootingTransfer{Guides: []troubleshootingTransferGuide{}}
	book, err := excelize.OpenReader(r, excelize.Options{UnzipSizeLimit: 32 << 20})
	if err != nil {
		return doc, fmt.Errorf("invalid XLSX: %w", err)
	}
	defer book.Close()
	guideRows, cols, guideErr := transferSheet(book, "Guides", transferGuideHeaders)
	stepRows, stepCols, stepErr := transferSheet(book, "Steps", transferStepHeaders)
	if guideErr != nil || stepErr != nil {
		return doc, errors.Join(guideErr, stepErr)
	}
	refs := make(map[string]int)
	for i, row := range guideRows {
		if emptyTransferRow(row) {
			continue
		}
		cell := func(name string) string { return transferCell(row, cols, name) }
		guide := troubleshootingTransferGuide{troubleshootingBody: troubleshootingBody{
			Title: cell("title"), Category: cell("category"), Severity: cell("severity"), Symptom: cell("symptom"),
			RootCause: cell("root_cause"), ErrorCode: cell("error_code"), DeviceName: cell("device_name"), FunctionName: cell("function_name"),
		}}
		if raw := strings.TrimSpace(cell("id")); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
				guide.ID = &id
			} else {
				guide.issues = append(guide.issues, fmt.Sprintf("Guides row %d: id must be a positive integer", i+2))
			}
		}
		ref := strings.TrimSpace(cell("guide_ref"))
		switch previous, exists := refs[ref]; {
		case ref == "":
			guide.issues = append(guide.issues, fmt.Sprintf("Guides row %d: missing guide_ref", i+2))
		case exists:
			issue := fmt.Sprintf("Guides row %d: duplicate guide_ref %q", i+2, ref)
			doc.Guides[previous].issues = append(doc.Guides[previous].issues, issue)
			guide.issues = append(guide.issues, issue)
		default:
			refs[ref] = len(doc.Guides)
		}
		doc.Guides = append(doc.Guides, guide)
	}
	steps := make(map[int]map[int]troubleshootingStepBody)
	for i, row := range stepRows {
		if emptyTransferRow(row) {
			continue
		}
		cell := func(name string) string { return transferCell(row, stepCols, name) }
		ref := strings.TrimSpace(cell("guide_ref"))
		index, exists := refs[ref]
		if !exists {
			doc.issues = append(doc.issues, fmt.Sprintf("Steps row %d: orphan guide_ref %q", i+2, ref))
			continue
		}
		guide := &doc.Guides[index]
		number, err := strconv.Atoi(strings.TrimSpace(cell("step_no")))
		if err != nil || number < 1 {
			guide.issues = append(guide.issues, fmt.Sprintf("Steps row %d: step_no must be a positive integer", i+2))
			continue
		}
		if steps[index] == nil {
			steps[index] = make(map[int]troubleshootingStepBody)
		}
		if _, duplicate := steps[index][number]; duplicate {
			guide.issues = append(guide.issues, fmt.Sprintf("Steps row %d: duplicate step_no %d", i+2, number))
			continue
		}
		steps[index][number] = troubleshootingStepBody{
			Title: cell("title"), Instruction: cell("instruction"), ScriptSnippet: cell("script_snippet"), Verification: cell("verification"),
		}
	}
	for i, numbered := range steps {
		order := make([]int, 0, len(numbered))
		for n := range numbered {
			order = append(order, n)
		}
		slices.Sort(order)
		for _, n := range order {
			doc.Guides[i].Steps = append(doc.Guides[i].Steps, numbered[n])
		}
	}
	return doc, nil
}

func writeTroubleshootingTransfer(doc troubleshootingTransfer, format string, instructions [][]string) ([]byte, error) {
	if format == "json" {
		return json.MarshalIndent(doc, "", "  ")
	}
	if format != "xlsx" {
		return nil, errors.New("unknown export format")
	}
	book := excelize.NewFile()
	defer book.Close()
	if err := book.SetSheetName("Sheet1", "Guides"); err != nil {
		return nil, err
	}
	if _, err := book.NewSheet("Steps"); err != nil {
		return nil, err
	}
	writeRow := func(sheet string, row int, values []string, guide *troubleshootingTransferGuide, step int) error {
		for col, value := range values {
			if len(utf16.Encode([]rune(value))) <= excelize.TotalCellChars {
				continue
			}
			if guide == nil {
				return fmt.Errorf("%s row %d column %d exceeds the Excel cell character limit", sheet, row, col+1)
			}
			var field string
			if step > 0 {
				field = fmt.Sprintf("Step %d / %s", step, transferStepFieldLabels[col])
			} else {
				field = transferGuideFieldLabels[col]
			}
			title := []rune(guide.Title)
			if len(title) > 80 {
				name := fmt.Sprintf("%q", string(title[:80])+"…")
				if guide.ID != nil {
					name += fmt.Sprintf(" (ID %d)", *guide.ID)
				}
				return &troubleshootingXLSXCellLimitError{guide: name, field: field}
			}
			return &troubleshootingXLSXCellLimitError{guide: fmt.Sprintf("%q", guide.Title), field: field}
		}
		cell, err := excelize.CoordinatesToCellName(1, row)
		if err != nil {
			return err
		}
		return book.SetSheetRow(sheet, cell, &values)
	}
	if err := writeRow("Guides", 1, transferGuideHeaders, nil, 0); err != nil {
		return nil, err
	}
	if err := writeRow("Steps", 1, transferStepHeaders, nil, 0); err != nil {
		return nil, err
	}
	stepRow := 2
	for i, g := range doc.Guides {
		ref, id := strconv.Itoa(i+1), ""
		if g.ID != nil {
			id = strconv.FormatInt(*g.ID, 10)
		}
		if err := writeRow("Guides", i+2, []string{ref, id, g.Title, g.Category, g.Severity, g.Symptom, g.RootCause, g.ErrorCode, g.DeviceName, g.FunctionName}, &g, 0); err != nil {
			return nil, err
		}
		for n, s := range g.Steps {
			if err := writeRow("Steps", stepRow, []string{ref, strconv.Itoa(n + 1), s.Title, s.Instruction, s.ScriptSnippet, s.Verification}, &g, n+1); err != nil {
				return nil, err
			}
			stepRow++
		}
	}
	if instructions != nil {
		if _, err := book.NewSheet("Instructions"); err != nil {
			return nil, err
		}
		for i, row := range instructions {
			if err := writeRow("Instructions", i+1, row, nil, 0); err != nil {
				return nil, err
			}
		}
	}
	data, err := book.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}
