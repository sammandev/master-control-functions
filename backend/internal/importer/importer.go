package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"webprogfndb/internal/store"
)

// RunResult summarizes one import run.
type RunResult struct {
	RunID          int64           `json:"run_id"`
	FileName       string          `json:"file_name"`
	Reset          bool            `json:"reset"`
	Counts         map[string]int  `json:"counts"`
	Skipped        []SkippedRow    `json:"skipped"`
	Normalizations []Normalization `json:"normalizations"`
}

// Import upserts the workbook into the database in a single transaction
// (ADR-0004). With reset=true the six entity tables are emptied first, in
// the same transaction. Rows missing natural-key fields are skipped and
// reported, never fatal — the workbook is allowed to be incomplete
// (ADR-0003). Every distinct non-blank device scope in the imported
// content is auto-registered in the devices registry (ADR-0013); the
// registry is never deleted by an import.
func Import(ctx context.Context, db *sql.DB, wb *Workbook, fileName string, reset bool) (*RunResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	q := store.New(tx)
	res := &RunResult{
		FileName:       fileName,
		Reset:          reset,
		Counts:         map[string]int{},
		Skipped:        wb.Skipped,
		Normalizations: wb.Normalizations,
	}

	if reset {
		// Links and examples first; nothing references them via FKs, but
		// this keeps the reset order obvious.
		for _, del := range []func(context.Context) error{
			q.DeleteAllParameterLinks, q.DeleteAllCriteriaLinks, q.DeleteAllExamples,
			q.DeleteAllParameterDefinitions, q.DeleteAllCriteriaDefinitions, q.DeleteAllFunctions,
		} {
			if err := del(ctx); err != nil {
				return nil, err
			}
		}
	}

	for _, r := range wb.Functions {
		if r.Source == "" {
			res.Skipped = append(res.Skipped, SkippedRow{Table: "functions", RowKey: r.Name, Reason: "missing source"})
			continue
		}
		if _, err := q.UpsertFunctionNatural(ctx, store.UpsertFunctionNaturalParams{
			Source:         r.Source,
			Interface:      r.Interface,
			Name:           r.Name,
			Description:    r.Description,
			Convention:     nullInt64(r.Convention),
			Category:       r.Category,
			ExecutionRole:  r.ExecutionRole,
			ExecutionPhase: r.ExecutionPhase,
			CreatedAt:      now,
			UpdatedAt:      now,
		}); err != nil {
			return nil, fmt.Errorf("function %s: %w", r.Name, err)
		}
		res.Counts["functions"]++
	}

	for _, r := range wb.Parameters {
		if _, err := q.UpsertParameterDefinition(ctx, parameterDefinitionParams(r, now)); err != nil {
			return nil, fmt.Errorf("parameter definition %s: %w", r.Name, err)
		}
		res.Counts["parameter_definitions"]++
	}
	for _, r := range wb.Criteria {
		if _, err := q.UpsertCriteriaDefinition(ctx, criteriaDefinitionParams(r, now)); err != nil {
			return nil, fmt.Errorf("criteria definition %s: %w", r.Name, err)
		}
		res.Counts["criteria_definitions"]++
	}

	for _, r := range wb.ParameterLinks {
		if r.FunctionSource == "" {
			res.Skipped = append(res.Skipped, SkippedRow{Table: "parameter_links", RowKey: r.FunctionName + "/" + r.Name, Reason: "missing function_source"})
			continue
		}
		if _, err := q.UpsertParameterLink(ctx, store.UpsertParameterLinkParams{
			FunctionSource: r.FunctionSource,
			FunctionName:   r.FunctionName,
			DeviceName:     r.DeviceName,
			Name:           r.Name,
			SourceFile:     r.SourceFile,
			Position:       nullInt64(r.Position),
			Required:       boolInt(r.Required),
			CreatedAt:      now,
			UpdatedAt:      now,
		}); err != nil {
			return nil, fmt.Errorf("parameter link %s/%s: %w", r.FunctionName, r.Name, err)
		}
		res.Counts["parameter_links"]++
	}
	for _, r := range wb.CriteriaLinks {
		if r.FunctionSource == "" {
			res.Skipped = append(res.Skipped, SkippedRow{Table: "criteria_links", RowKey: r.FunctionName + "/" + r.Name, Reason: "missing function_source"})
			continue
		}
		if _, err := q.UpsertCriteriaLink(ctx, store.UpsertCriteriaLinkParams{
			FunctionSource: r.FunctionSource,
			FunctionName:   r.FunctionName,
			DeviceName:     r.DeviceName,
			Name:           r.Name,
			SourceFile:     r.SourceFile,
			Position:       nullInt64(r.Position),
			Required:       boolInt(r.Required),
			CreatedAt:      now,
			UpdatedAt:      now,
		}); err != nil {
			return nil, fmt.Errorf("criteria link %s/%s: %w", r.FunctionName, r.Name, err)
		}
		res.Counts["criteria_links"]++
	}

	for _, r := range wb.Examples {
		if r.Script == "" {
			res.Skipped = append(res.Skipped, SkippedRow{Table: "examples", RowKey: r.FunctionName, Reason: "empty script"})
			continue
		}
		if _, err := q.UpsertExample(ctx, store.UpsertExampleParams{
			FunctionSource: r.FunctionSource,
			FunctionName:   r.FunctionName,
			DeviceName:     r.DeviceName,
			Script:         r.Script,
			Meaning:        r.Meaning,
			CreatedAt:      now,
			UpdatedAt:      now,
		}); err != nil {
			return nil, fmt.Errorf("example %s: %w", r.FunctionName, err)
		}
		res.Counts["examples"]++
	}

	// Auto-populate the device registry from the imported content (ADR-0013):
	// every distinct non-blank device scope across the five scope tables
	// becomes a registered device. Idempotent; never deletes, so manual
	// registrations and prior runs survive both plain imports and resets
	// (TestResetLeavesDevicesRegistered).
	if n, err := q.RegisterImportedDeviceScopes(ctx, now); err != nil {
		return nil, fmt.Errorf("register device scopes: %w", err)
	} else if n > 0 {
		res.Counts["devices"] = int(n)
	}

	// Run + normalizations are recorded inside the same transaction: an
	// import that fails leaves no change-log trace.
	countsJSON, err := json.Marshal(res.Counts)
	if err != nil {
		return nil, err
	}
	skippedJSON, err := json.Marshal(res.Skipped)
	if err != nil {
		return nil, err
	}
	run, err := q.InsertImportRun(ctx, store.InsertImportRunParams{
		FileName:   fileName,
		Reset:      boolInt(reset),
		StartedAt:  now,
		FinishedAt: time.Now().UTC().Format(time.RFC3339),
		Counts:     string(countsJSON),
		Skipped:    string(skippedJSON),
	})
	if err != nil {
		return nil, err
	}
	res.RunID = run.ID
	for _, n := range wb.Normalizations {
		if err := q.InsertImportNormalization(ctx, store.InsertImportNormalizationParams{
			RunID:           run.ID,
			TableName:       n.Table,
			RowKey:          n.RowKey,
			ColumnName:      n.Column,
			SourceValue:     n.Source,
			NormalizedValue: n.Normalized,
			Rule:            n.Rule,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func parameterDefinitionParams(r DefinitionRow, now string) store.UpsertParameterDefinitionParams {
	return store.UpsertParameterDefinitionParams{
		Name:          r.Name,
		Description:   r.Description,
		Type:          r.Type,
		Unit:          r.Unit,
		DefaultValue:  nullString(r.DefaultValue),
		AllowedValues: r.AllowedValues,
		Cardinality:   r.Cardinality,
		ValuePattern:  r.ValuePattern,
		Required:      boolInt(r.Required),
		DeviceName:    r.DeviceName,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func criteriaDefinitionParams(r DefinitionRow, now string) store.UpsertCriteriaDefinitionParams {
	return store.UpsertCriteriaDefinitionParams{
		Name:          r.Name,
		Description:   r.Description,
		Type:          r.Type,
		Unit:          r.Unit,
		DefaultValue:  nullString(r.DefaultValue),
		AllowedValues: r.AllowedValues,
		Cardinality:   r.Cardinality,
		ValuePattern:  r.ValuePattern,
		Required:      boolInt(r.Required),
		DeviceName:    r.DeviceName,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func nullInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
