package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// MCP readiness (ADR-0012). The web app owns the audited script grammar the
// script-audit-mcp server composes with, and reports per Function whether the
// MCP can compose it, using the MCP's own field-selection and support rules.

func registerMCPRoutes(app *fiber.App, q *store.Queries) {
	app.Get("/api/mcp-readiness", mcpReadinessSummary(q))
	app.Get("/api/functions/:id/mcp-readiness", functionMCPReadiness(q))
	app.Get("/api/functions/:id/grammar", getGrammar(q))
	app.Put("/api/functions/:id/grammar", putGrammar(q))
	app.Delete("/api/functions/:id/grammar", deleteGrammar(q))
	app.Put("/api/functions/:id/script-keys", putScriptKey(q))
	app.Delete("/api/functions/:id/script-keys/:kind/:name", deleteScriptKey(q))
}

type naturalKey struct{ source, iface, name string }

type catalogDefinition struct {
	Name, DeviceName, Type, Cardinality, ValuePattern, AllowedValues string
}

type catalogLink struct {
	Name, FunctionSource, FunctionName, DeviceName string
	Position                                       sql.NullInt64
	Required                                       bool
}

// mcpCatalog is everything readiness needs, loaded once per request like the
// MCP's own per-request snapshot.
type mcpCatalog struct {
	functions  []store.Function
	grammars   map[naturalKey]store.FunctionGrammar
	audits     map[naturalKey][]store.ScriptKeyAudit
	paramDefs  []catalogDefinition
	critDefs   []catalogDefinition
	paramLinks []catalogLink
	critLinks  []catalogLink
	examples   []store.Example
	nameCounts map[[2]string]int
}

func loadMCPCatalog(ctx context.Context, q *store.Queries) (*mcpCatalog, error) {
	c := &mcpCatalog{
		grammars:   map[naturalKey]store.FunctionGrammar{},
		audits:     map[naturalKey][]store.ScriptKeyAudit{},
		nameCounts: map[[2]string]int{},
	}
	var err error
	if c.functions, err = q.ListAllFunctionsForReadiness(ctx); err != nil {
		return nil, err
	}
	for _, f := range c.functions {
		c.nameCounts[[2]string{f.Source, f.Name}]++
	}
	grammars, err := q.ListAllFunctionGrammars(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range grammars {
		c.grammars[naturalKey{g.FunctionSource, g.FunctionInterface, g.FunctionName}] = g
	}
	audits, err := q.ListAllScriptKeyAudits(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range audits {
		k := naturalKey{a.FunctionSource, a.FunctionInterface, a.FunctionName}
		c.audits[k] = append(c.audits[k], a)
	}
	params, err := q.ListAllParameterDefinitionsForReadiness(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range params {
		c.paramDefs = append(c.paramDefs, catalogDefinition{d.Name, d.DeviceName, d.Type, d.Cardinality, d.ValuePattern, d.AllowedValues})
	}
	criteria, err := q.ListAllCriteriaDefinitionsForReadiness(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range criteria {
		c.critDefs = append(c.critDefs, catalogDefinition{d.Name, d.DeviceName, d.Type, d.Cardinality, d.ValuePattern, d.AllowedValues})
	}
	paramLinks, err := q.ListAllParameterLinks(ctx)
	if err != nil {
		return nil, err
	}
	for _, l := range paramLinks {
		c.paramLinks = append(c.paramLinks, catalogLink{l.Name, l.FunctionSource, l.FunctionName, l.DeviceName, l.Position, l.Required == 1})
	}
	critLinks, err := q.ListAllCriteriaLinks(ctx)
	if err != nil {
		return nil, err
	}
	for _, l := range critLinks {
		c.critLinks = append(c.critLinks, catalogLink{l.Name, l.FunctionSource, l.FunctionName, l.DeviceName, l.Position, l.Required == 1})
	}
	if c.examples, err = q.ListAllExamples(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func loadFunctionMCPCatalog(ctx context.Context, q *store.Queries, f store.Function, scope string) (*mcpCatalog, error) {
	key := naturalKey{f.Source, f.Interface, f.Name}
	c := &mcpCatalog{
		grammars: map[naturalKey]store.FunctionGrammar{},
		audits:   map[naturalKey][]store.ScriptKeyAudit{},
	}
	count, err := q.CountFunctionsSharingNaturalName(ctx, store.CountFunctionsSharingNaturalNameParams{
		Source: f.Source, Name: f.Name,
	})
	if err != nil {
		return nil, err
	}
	c.nameCounts = map[[2]string]int{{f.Source, f.Name}: int(count)}
	grammar, err := q.GetFunctionGrammar(ctx, store.GetFunctionGrammarParams{
		FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
	})
	if err == nil {
		c.grammars[key] = grammar
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	audits, err := q.ListScriptKeyAudits(ctx, store.ListScriptKeyAuditsParams{
		FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
	})
	if err != nil {
		return nil, err
	}
	c.audits[key] = audits

	paramLinks, err := q.ListFunctionParameterLinksForReadiness(ctx, store.ListFunctionParameterLinksForReadinessParams{
		FunctionSource: f.Source, FunctionName: f.Name,
	})
	if err != nil {
		return nil, err
	}
	var paramNames []string
	for _, l := range paramLinks {
		if l.DeviceName == "" || strings.EqualFold(l.DeviceName, scope) {
			c.paramLinks = append(c.paramLinks, catalogLink{l.Name, f.Source, f.Name, l.DeviceName, l.Position, l.Required == 1})
			paramNames = append(paramNames, l.Name)
		}
	}
	if len(paramNames) > 0 {
		namesJSON, _ := json.Marshal(paramNames)
		params, err := q.ListParameterDefinitionsByNameForReadiness(ctx, string(namesJSON))
		if err != nil {
			return nil, err
		}
		for _, d := range params {
			c.paramDefs = append(c.paramDefs, catalogDefinition{d.Name, d.DeviceName, d.Type, d.Cardinality, d.ValuePattern, d.AllowedValues})
		}
	}
	critLinks, err := q.ListFunctionCriteriaLinksForReadiness(ctx, store.ListFunctionCriteriaLinksForReadinessParams{
		FunctionSource: f.Source, FunctionName: f.Name,
	})
	if err != nil {
		return nil, err
	}
	var critNames []string
	for _, l := range critLinks {
		if l.DeviceName == "" || strings.EqualFold(l.DeviceName, scope) {
			c.critLinks = append(c.critLinks, catalogLink{l.Name, f.Source, f.Name, l.DeviceName, l.Position, l.Required == 1})
			critNames = append(critNames, l.Name)
		}
	}
	if len(critNames) > 0 {
		namesJSON, _ := json.Marshal(critNames)
		criteria, err := q.ListCriteriaDefinitionsByNameForReadiness(ctx, string(namesJSON))
		if err != nil {
			return nil, err
		}
		for _, d := range criteria {
			c.critDefs = append(c.critDefs, catalogDefinition{d.Name, d.DeviceName, d.Type, d.Cardinality, d.ValuePattern, d.AllowedValues})
		}
	}
	examples, err := q.ListFunctionExamplesForReadiness(ctx, store.ListFunctionExamplesForReadinessParams{
		FunctionSource: f.Source, FunctionName: f.Name,
	})
	if err != nil {
		return nil, err
	}
	for _, e := range examples {
		c.examples = append(c.examples, store.Example{
			ID: e.ID, FunctionSource: f.Source, FunctionName: f.Name, DeviceName: e.DeviceName, Script: e.Script,
		})
	}
	return c, nil
}

type readinessField struct {
	Name        string `json:"name"`
	DeviceName  string `json:"device_name"`
	Required    bool   `json:"required"`
	Definition  bool   `json:"has_definition"`
	ScriptKey   string `json:"script_key_scope"`
	definition  *catalogDefinition
	positionKey sql.NullInt64
}

type readinessMode struct {
	Ready  bool          `json:"ready"`
	Issues []scriptIssue `json:"issues"`
}

type grammarView struct {
	Encoding  string `json:"encoding"`
	Scope     string `json:"scope"`
	AuditedBy string `json:"audited_by"`
	AuditedAt string `json:"audited_at"`
	Note      string `json:"note"`
}

type readinessReport struct {
	FunctionID int64            `json:"function_id"`
	Source     string           `json:"source"`
	Interface  string           `json:"interface"`
	Name       string           `json:"name"`
	Device     string           `json:"device"`
	Status     string           `json:"status"`
	Grammar    *grammarView     `json:"grammar"`
	Line       readinessMode    `json:"line"`
	Bundle     readinessMode    `json:"bundle"`
	Parameters []readinessField `json:"parameters"`
	Criteria   []readinessField `json:"criteria"`
	Warnings   []scriptIssue    `json:"warnings"`
}

func readinessDevice(f store.Function, device string) string {
	scope := normalizeDeviceName(device)
	if scope == "" {
		return f.Interface
	}
	return scope
}

// Readiness statuses: "ready" composes as a single script line, "bundle_ready"
// only inside INI bundles, "not_ready" neither.
func (c *mcpCatalog) readiness(f store.Function, scope string) readinessReport {
	key := naturalKey{f.Source, f.Interface, f.Name}
	report := readinessReport{
		FunctionID: f.ID, Source: f.Source, Interface: f.Interface, Name: f.Name,
		Device: scope, Warnings: []scriptIssue{},
	}
	var grammar *store.FunctionGrammar
	if g, ok := c.grammars[key]; ok {
		grammar = &g
		report.Grammar = &grammarView{g.Encoding, g.Scope, g.AuditedBy, g.AuditedAt, g.Note}
	}
	audits := map[string]string{}
	for _, a := range c.audits[key] {
		audits[a.Kind+"\x00"+a.Name] = a.Scope
	}

	var resolution []scriptIssue
	params := selectFields(c.paramLinks, c.paramDefs, f, scope, "params", &resolution)
	criteria := selectFields(c.critLinks, c.critDefs, f, scope, "criteria", &resolution)
	for i := range params {
		params[i].ScriptKey = audits["parameter\x00"+params[i].Name]
	}
	for i := range criteria {
		criteria[i].ScriptKey = audits["criterion\x00"+criteria[i].Name]
	}
	report.Parameters, report.Criteria = params, criteria

	for _, bundle := range []bool{false, true} {
		issues := append([]scriptIssue{}, resolution...)
		issues = append(issues, supportIssues(f, grammar, params, criteria, bundle)...)
		mode := readinessMode{Ready: len(issues) == 0, Issues: issues}
		if bundle {
			report.Bundle = mode
		} else {
			report.Line = mode
		}
	}
	switch {
	case report.Line.Ready:
		report.Status = "ready"
	case report.Bundle.Ready:
		report.Status = "bundle_ready"
	default:
		report.Status = "not_ready"
	}

	if n := c.nameCounts[[2]string{f.Source, f.Name}]; n > 1 {
		report.Warnings = append(report.Warnings, scriptIssue{
			Field: "function", Code: "ambiguous_function_key",
			Message: strconv.Itoa(n) + " functions share source " + f.Source + " and name " + f.Name +
				"; their links and examples apply to all of them",
		})
	}
	convention := nullableConvention(f.Convention)
	for _, e := range c.examples {
		if e.FunctionSource != f.Source || e.FunctionName != f.Name ||
			(e.DeviceName != "" && !strings.EqualFold(e.DeviceName, scope)) {
			continue
		}
		for _, issue := range checkExampleAgainst(e.Script, exampleTarget{f.Source, f.Interface, f.Name, convention}) {
			issue.Field = "examples." + strconv.FormatInt(e.ID, 10)
			report.Warnings = append(report.Warnings, issue)
		}
	}
	return report
}

func nullableConvention(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

// selectFields mirrors script-audit-mcp lib.rs select_fields: device-scoped
// links override global ones by name, definitions resolve exact device first,
// then global.
func selectFields(links []catalogLink, defs []catalogDefinition, f store.Function, scope, prefix string, issues *[]scriptIssue) []readinessField {
	var global, scoped []catalogLink
	for _, l := range links {
		if l.FunctionSource != f.Source || l.FunctionName != f.Name {
			continue
		}
		switch {
		case l.DeviceName == "":
			global = append(global, l)
		case strings.EqualFold(l.DeviceName, scope):
			scoped = append(scoped, l)
		}
	}
	selected := map[string]catalogLink{}
	for _, l := range append(global, scoped...) {
		if previous, ok := selected[l.Name]; ok && strings.EqualFold(previous.DeviceName, l.DeviceName) {
			*issues = append(*issues, scriptIssue{Field: prefix + "." + l.Name, Code: "duplicate_link",
				Message: "duplicate link for " + l.Name})
		}
		selected[l.Name] = l
	}
	fields := make([]readinessField, 0, len(selected))
	for _, l := range selected {
		var exact, fallback []catalogDefinition
		for _, d := range defs {
			if d.Name != l.Name {
				continue
			}
			if strings.EqualFold(d.DeviceName, l.DeviceName) {
				exact = append(exact, d)
			} else if d.DeviceName == "" {
				fallback = append(fallback, d)
			}
		}
		candidates := exact
		if len(candidates) == 0 {
			candidates = fallback
		}
		if len(candidates) > 1 {
			*issues = append(*issues, scriptIssue{Field: prefix + "." + l.Name, Code: "ambiguous_definition",
				Message: "ambiguous definition for " + l.Name})
		}
		field := readinessField{Name: l.Name, DeviceName: l.DeviceName, Required: l.Required, positionKey: l.Position}
		if len(candidates) > 0 {
			d := candidates[0]
			field.definition, field.Definition = &d, true
		}
		fields = append(fields, field)
	}
	sort.Slice(fields, func(i, j int) bool {
		a, b := fields[i].positionKey, fields[j].positionKey
		if a.Valid != b.Valid {
			return a.Valid
		}
		if a.Int64 != b.Int64 {
			return a.Int64 < b.Int64
		}
		return fields[i].Name < fields[j].Name
	})
	return fields
}

// supportIssues mirrors script-audit-mcp grammar.rs
// support_issues_with_bundle_keys with the grammar read from the database.
func supportIssues(f store.Function, grammar *store.FunctionGrammar, params, criteria []readinessField, bundle bool) []scriptIssue {
	unsupported := func(field, message string) scriptIssue {
		return scriptIssue{Field: field, Code: "unsupported_grammar", Message: message}
	}
	if grammar == nil {
		return []scriptIssue{unsupported("function", "no audited grammar profile for this function")}
	}
	if !bundle && grammar.Scope != "line" {
		return []scriptIssue{unsupported("function", "grammar is audited for INI bundles only")}
	}
	wantConvention := map[string]int64{"raw": 1, "named": 2}[grammar.Encoding]
	if !f.Convention.Valid || f.Convention.Int64 != wantConvention {
		return []scriptIssue{unsupported("function",
			grammar.Encoding+" grammar requires convention "+strconv.FormatInt(wantConvention, 10))}
	}
	var issues []scriptIssue
	raw := grammar.Encoding == "raw"
	if raw && len(params) != 1 {
		issues = append(issues, unsupported("params", "audited SSH operations require one payload definition"))
	}
	for _, group := range []struct {
		prefix, kind string
		fields       []readinessField
	}{{"params", "parameter", params}, {"criteria", "criterion", criteria}} {
		if raw && len(group.fields) > 1 {
			issues = append(issues, unsupported(group.prefix, "raw profile requires at most one linked definition"))
		}
		for _, field := range group.fields {
			key := group.prefix + "." + field.Name
			if !raw && !(field.ScriptKey == "line" || (bundle && field.ScriptKey == "bundle")) {
				issues = append(issues, unsupported(key, "dictionary-to-script key mapping has not been audited"))
			}
			d := field.definition
			if d == nil {
				issues = append(issues, scriptIssue{key, "missing_definition", "linked definition is missing"})
				continue
			}
			if !validTypes[d.Type] {
				issues = append(issues, scriptIssue{key, "unsupported_type", "unsupported type: " + d.Type})
			}
			if !validCardinalities[d.Cardinality] {
				issues = append(issues, scriptIssue{key, "unsupported_cardinality", "cardinality must be SINGLE or MULTIPLE"})
			}
			if _, err := compileValuePattern(d.ValuePattern, d.Type, d.Cardinality); err != nil {
				issues = append(issues, scriptIssue{key, "unsupported_pattern", err.Error()})
			}
			if _, err := parseAllowed(d.AllowedValues); err != nil {
				issues = append(issues, scriptIssue{key, "unsupported_constraint", err.Error()})
			}
		}
	}
	return issues
}

func functionFromParam(c fiber.Ctx, q *store.Queries) (store.Function, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return store.Function{}, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid function id"})
	}
	f, err := q.GetFunction(c.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return store.Function{}, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		return store.Function{}, serverError(c, "read function", err)
	}
	return f, nil
}

func functionMCPReadiness(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		scope := readinessDevice(f, c.Query("device"))
		catalog, err := loadFunctionMCPCatalog(c.Context(), q, f, scope)
		if err != nil {
			return serverError(c, "load catalog", err)
		}
		return c.JSON(catalog.readiness(f, scope))
	}
}

func mcpReadinessSummary(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		catalog, err := loadMCPCatalog(c.Context(), q)
		if err != nil {
			return serverError(c, "load catalog", err)
		}
		items := make([]fiber.Map, 0, len(catalog.functions))
		counts := map[string]int{"ready": 0, "bundle_ready": 0, "not_ready": 0}
		for _, f := range catalog.functions {
			r := catalog.readiness(f, f.Interface)
			counts[r.Status]++
			encoding, scope := "", ""
			if r.Grammar != nil {
				encoding, scope = r.Grammar.Encoding, r.Grammar.Scope
			}
			items = append(items, fiber.Map{
				"function_id": f.ID, "source": f.Source, "interface": f.Interface, "name": f.Name,
				"status": r.Status, "grammar_encoding": encoding, "grammar_scope": scope,
				"line_issue_count": len(r.Line.Issues), "bundle_issue_count": len(r.Bundle.Issues),
				"warning_count": len(r.Warnings),
			})
		}
		return c.JSON(fiber.Map{"items": items, "counts": counts, "total": len(items)})
	}
}

type grammarBody struct {
	Encoding  string `json:"encoding"`
	Scope     string `json:"scope"`
	AuditedBy string `json:"audited_by"`
	Note      string `json:"note"`
}

type scriptKeyBody struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Scope     string `json:"scope"`
	AuditedBy string `json:"audited_by"`
	Note      string `json:"note"`
}

var grammarScopes = map[string]bool{"line": true, "bundle": true}

func getGrammar(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		var view *grammarView
		g, err := q.GetFunctionGrammar(c.Context(), store.GetFunctionGrammarParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
		})
		switch {
		case err == nil:
			view = &grammarView{g.Encoding, g.Scope, g.AuditedBy, g.AuditedAt, g.Note}
		case !errors.Is(err, sql.ErrNoRows):
			return serverError(c, "read grammar", err)
		}
		keys, err := q.ListScriptKeyAudits(c.Context(), store.ListScriptKeyAuditsParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
		})
		if err != nil {
			return serverError(c, "list script keys", err)
		}
		items := make([]fiber.Map, 0, len(keys))
		for _, k := range keys {
			items = append(items, fiber.Map{
				"kind": k.Kind, "name": k.Name, "scope": k.Scope,
				"audited_by": k.AuditedBy, "audited_at": k.AuditedAt, "note": k.Note,
			})
		}
		return c.JSON(fiber.Map{"grammar": view, "script_keys": items})
	}
}

func putGrammar(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		var b grammarBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		b.Encoding, b.Scope = strings.TrimSpace(b.Encoding), strings.TrimSpace(b.Scope)
		b.AuditedBy = strings.TrimSpace(b.AuditedBy)
		fields := map[string]string{}
		if b.Encoding != "raw" && b.Encoding != "named" {
			fields["encoding"] = "must be raw or named"
		}
		if !grammarScopes[b.Scope] {
			fields["scope"] = "must be line or bundle"
		}
		if b.AuditedBy == "" {
			fields["audited_by"] = "is required"
		}
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)
		g, err := q.UpsertFunctionGrammar(c.Context(), store.UpsertFunctionGrammarParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
			Encoding: b.Encoding, Scope: b.Scope, AuditedBy: b.AuditedBy, AuditedAt: now,
			Note: strings.TrimSpace(b.Note), Now: now,
		})
		if err != nil {
			return serverError(c, "save grammar", err)
		}
		return c.JSON(grammarView{g.Encoding, g.Scope, g.AuditedBy, g.AuditedAt, g.Note})
	}
}

func deleteGrammar(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		affected, err := q.DeleteFunctionGrammar(c.Context(), store.DeleteFunctionGrammarParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
		})
		if err != nil {
			return serverError(c, "delete grammar", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "grammar not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func putScriptKey(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		var b scriptKeyBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		b.Kind, b.Name = strings.TrimSpace(b.Kind), strings.TrimSpace(b.Name)
		b.Scope, b.AuditedBy = strings.TrimSpace(b.Scope), strings.TrimSpace(b.AuditedBy)
		fields := map[string]string{}
		if b.Kind != "parameter" && b.Kind != "criterion" {
			fields["kind"] = "must be parameter or criterion"
		}
		if b.Name == "" {
			fields["name"] = "is required"
		}
		if !grammarScopes[b.Scope] {
			fields["scope"] = "must be line or bundle"
		}
		if b.AuditedBy == "" {
			fields["audited_by"] = "is required"
		}
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)
		k, err := q.UpsertScriptKeyAudit(c.Context(), store.UpsertScriptKeyAuditParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
			Kind: b.Kind, Name: b.Name, Scope: b.Scope, AuditedBy: b.AuditedBy, AuditedAt: now,
			Note: strings.TrimSpace(b.Note), Now: now,
		})
		if err != nil {
			return serverError(c, "save script key", err)
		}
		return c.JSON(fiber.Map{
			"kind": k.Kind, "name": k.Name, "scope": k.Scope,
			"audited_by": k.AuditedBy, "audited_at": k.AuditedAt, "note": k.Note,
		})
	}
}

func deleteScriptKey(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		f, err := functionFromParam(c, q)
		if f.ID == 0 {
			return err
		}
		affected, err := q.DeleteScriptKeyAudit(c.Context(), store.DeleteScriptKeyAuditParams{
			FunctionSource: f.Source, FunctionInterface: f.Interface, FunctionName: f.Name,
			Kind: c.Params("kind"), Name: c.Params("name"),
		})
		if err != nil {
			return serverError(c, "delete script key", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "script key not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
