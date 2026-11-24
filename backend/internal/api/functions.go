package api

import (
	"database/sql"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// validSources is the Source vocabulary from CONTEXT.md: where a Function
// is implemented.
var validSources = map[string]bool{
	"BUILT-IN": true,
	"DLL":      true,
}

// sortColumns whitelists the sortable list fields. Anything else falls back
// to the default (id) order rather than erroring, so a stray value never
// breaks list rendering.
var sortColumns = map[string]func(a, b store.Function) int{
	"name":       cmpBy(func(f store.Function) string { return f.Name }),
	"interface":  cmpBy(func(f store.Function) string { return f.Interface }),
	"source":     cmpBy(func(f store.Function) string { return f.Source }),
	"updated_at": cmpBy(func(f store.Function) string { return f.UpdatedAt }),
}

func cmpBy(key func(store.Function) string) func(a, b store.Function) int {
	return func(a, b store.Function) int {
		if key(a) < key(b) {
			return -1
		}
		if key(a) > key(b) {
			return 1
		}
		// Deterministic tiebreak on the natural key, matching the schema's
		// UNIQUE (source, interface, name) order.
		if a.Source+a.Interface+a.Name < b.Source+b.Interface+b.Name {
			return -1
		}
		if a.Source+a.Interface+a.Name > b.Source+b.Interface+b.Name {
			return 1
		}
		return 0
	}
}

// registerFunctionRoutes mounts static paths before :id so they are not shadowed.
func registerFunctionRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Post("/api/functions", createFunction(q))
	app.Get("/api/functions", listFunctions(q))
	app.Get("/api/functions/facets", functionFacets(q))
	app.Get("/api/functions/options", listFunctionOptions(q))
	app.Get("/api/functions/:id", getFunction(q))
	app.Put("/api/functions/:id", updateFunction(q, d))
	app.Delete("/api/functions/:id", deleteFunction(q))
}

type functionOption struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Interface string `json:"interface"`
}

func listFunctionOptions(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		rows, err := q.ListFunctionOptions(c.Context())
		if err != nil {
			return serverError(c, "list function options", err)
		}
		items := make([]functionOption, 0, len(rows))
		for _, row := range rows {
			items = append(items, functionOption(row))
		}
		return c.JSON(fiber.Map{"items": items})
	}
}

// functionFacets serves the distinct enum values for the functions list
// filter bar (ticket 09).
func functionFacets(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		sources, err := q.DistinctFunctionSources(c.Context())
		if err != nil {
			return serverError(c, "list sources", err)
		}
		interfaces, err := q.DistinctFunctionInterfaces(c.Context())
		if err != nil {
			return serverError(c, "list interfaces", err)
		}
		categories, err := q.DistinctFunctionCategories(c.Context())
		if err != nil {
			return serverError(c, "list categories", err)
		}
		roles, err := q.DistinctFunctionExecutionRoles(c.Context())
		if err != nil {
			return serverError(c, "list execution roles", err)
		}
		phases, err := q.DistinctFunctionExecutionPhases(c.Context())
		if err != nil {
			return serverError(c, "list execution phases", err)
		}
		return c.JSON(fiber.Map{
			"sources":          sources,
			"interfaces":       interfaces,
			"categories":       categories,
			"execution_roles":  roles,
			"execution_phases": phases,
		})
	}
}

// functionBody is the JSON shape of a Function on create/update.
type functionBody struct {
	Source         string  `json:"source"`
	Interface      string  `json:"interface"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	Convention     *int64  `json:"convention"`
	Category       *string `json:"category"`
	ExecutionRole  *string `json:"execution_role"`
	ExecutionPhase *string `json:"execution_phase"`
}

// validate returns field-level messages; an empty map means valid.
func (b *functionBody) validate() map[string]string {
	fields := map[string]string{}
	b.Source = strings.TrimSpace(b.Source)
	b.Interface = normalizeDeviceName(b.Interface)
	b.Name = strings.TrimSpace(b.Name)
	if b.Source == "" {
		fields["source"] = "is required"
	} else if !validSources[b.Source] {
		fields["source"] = "must be BUILT-IN or DLL"
	}
	if problem := validateDeviceName(b.Interface); problem != "" {
		fields["interface"] = problem
	}
	if b.Name == "" {
		fields["name"] = "is required"
	}
	return fields
}

func validationError(c fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"error":  "validation failed",
		"fields": fields,
	})
}

// isUniqueViolation moved to the store package (store.IsUniqueViolation):
// it matches driver-specific error text, which is dialect-coupled logic
// that belongs at the data layer per ADR-0001's PostgreSQL exit path.

func functionResponse(f store.Function) fiber.Map {
	var convention *int64
	if f.Convention.Valid {
		v := f.Convention.Int64
		convention = &v
	}
	return fiber.Map{
		"id":              f.ID,
		"source":          f.Source,
		"interface":       f.Interface,
		"name":            f.Name,
		"description":     f.Description,
		"convention":      convention,
		"category":        f.Category,
		"execution_role":  f.ExecutionRole,
		"execution_phase": f.ExecutionPhase,
		"created_at":      f.CreatedAt,
		"updated_at":      f.UpdatedAt,
	}
}

// deref returns *p or the empty string, for optional string fields.
func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func createFunction(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		var b functionBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)
		f, err := q.InsertFunction(c.Context(), store.InsertFunctionParams{
			Source:         b.Source,
			Interface:      b.Interface,
			Name:           b.Name,
			Description:    deref(b.Description),
			Convention:     nullInt64Value(b.Convention),
			Category:       deref(b.Category),
			ExecutionRole:  deref(b.ExecutionRole),
			ExecutionPhase: deref(b.ExecutionPhase),
			CreatedAt:      now,
			UpdatedAt:      now,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				return conflict(c, b)
			}
			return serverError(c, "create function", err)
		}
		return c.Status(fiber.StatusCreated).JSON(functionResponse(f))
	}
}

func conflict(c fiber.Ctx, b functionBody) error {
	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"error": "function '" + b.Source + "/" + b.Interface + "/" + b.Name + "' already exists",
	})
}

func serverError(c fiber.Ctx, action string, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": action + ": " + err.Error(),
	})
}

// likePattern wraps the search term for a substring LIKE match. PostgreSQL
// uses backslash as LIKE's default escape character, so a literal backslash
// is doubled. User-supplied % is stripped; _ remains a single-character wildcard.
func likePattern(search string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(search), "%", "")
	if cleaned == "" {
		return ""
	}
	return "%" + strings.ReplaceAll(cleaned, `\`, `\\`) + "%"
}

// listQuery holds the parsed list query-string parameters.
type listQuery struct {
	search   string
	filters  map[string]*string
	sortCol  string
	sortDesc bool
	page     int64
	pageSize int64
}

func parseListQuery(c fiber.Ctx) listQuery {
	lq := listQuery{search: c.Query("search"), filters: map[string]*string{}, page: 1, pageSize: 20}
	for _, f := range []string{"source", "interface", "category", "execution_role", "execution_phase"} {
		if v := strings.TrimSpace(c.Query(f)); v != "" {
			lq.filters[f] = &v
		}
	}
	// Each list handler checks sortCol against its own sort keys.
	lq.sortCol = strings.TrimSpace(c.Query("sort"))
	lq.sortDesc = strings.EqualFold(c.Query("order"), "desc")
	if p, err := strconv.ParseInt(c.Query("page"), 10, 64); err == nil && p > 0 {
		lq.page = p
	}
	if ps, err := strconv.ParseInt(c.Query("page_size"), 10, 64); err == nil && ps > 0 {
		lq.pageSize = ps
	}
	if lq.pageSize > 100 {
		lq.pageSize = 100
	}
	return lq
}

func (lq *listQuery) storeParams() store.ListFunctionsParams {
	return store.ListFunctionsParams{
		Search:         nullStrPtr(&lq.search),
		Source:         nullStrPtr(lq.filters["source"]),
		Interface:      nullStrPtr(lq.filters["interface"]),
		Category:       nullStrPtr(lq.filters["category"]),
		ExecutionRole:  nullStrPtr(lq.filters["execution_role"]),
		ExecutionPhase: nullStrPtr(lq.filters["execution_phase"]),
	}
}

func (lq *listQuery) countParams() store.CountFunctionsParams {
	return store.CountFunctionsParams{
		Search:         nullStrPtr(&lq.search),
		Source:         nullStrPtr(lq.filters["source"]),
		Interface:      nullStrPtr(lq.filters["interface"]),
		Category:       nullStrPtr(lq.filters["category"]),
		ExecutionRole:  nullStrPtr(lq.filters["execution_role"]),
		ExecutionPhase: nullStrPtr(lq.filters["execution_phase"]),
	}
}

// locateRow moves lq to the page holding the row whose id is in the param
// query, so a client can open a just-saved row wherever it sorts. An absent
// or unknown id keeps the requested page.
func locateRow[T any](c fiber.Ctx, lq *listQuery, param string, rows []T, idOf func(T) int64) {
	id, err := strconv.ParseInt(c.Query(param), 10, 64)
	if err != nil || id <= 0 {
		return
	}
	if index := slices.IndexFunc(rows, func(row T) bool { return idOf(row) == id }); index >= 0 {
		lq.page = int64(index)/lq.pageSize + 1
	}
}

func listFunctions(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		lq := parseListQuery(c)

		total, err := q.CountFunctions(c.Context(), lq.countParams())
		if err != nil {
			return serverError(c, "count functions", err)
		}
		items, err := q.ListFunctions(c.Context(), lq.storeParams())
		if err != nil {
			return serverError(c, "list functions", err)
		}

		// Sort and paginate in Go: the knowledge base is workbook-sized
		// (hundreds of rows), and SQL-level dynamic ORDER BY is not worth
		// the dialect risk (ADR-0001).
		if cmp := sortColumns[lq.sortCol]; cmp != nil {
			sort.SliceStable(items, func(i, j int) bool {
				c := cmp(items[i], items[j])
				if lq.sortDesc {
					return c > 0
				}
				return c < 0
			})
		}
		locateRow(c, &lq, "function_id", items, func(f store.Function) int64 { return f.ID })
		start := (lq.page - 1) * lq.pageSize
		if start > int64(len(items)) {
			start = int64(len(items))
		}
		end := start + lq.pageSize
		if end > int64(len(items)) {
			end = int64(len(items))
		}
		page := items[start:end]

		out := make([]fiber.Map, 0, len(page))
		for _, f := range page {
			out = append(out, functionResponse(f))
		}
		return c.JSON(fiber.Map{
			"items":     out,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

func getFunction(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid function id"})
		}
		f, err := q.GetFunction(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		return c.JSON(functionResponse(f))
	}
}

func updateFunction(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid function id"})
		}
		var b functionBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}

		current, err := q.GetFunction(c.Context(), id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}

		now := time.Now().UTC().Format(time.RFC3339)
		keyChanged := b.Source != current.Source ||
			b.Interface != current.Interface ||
			b.Name != current.Name

		// Key change + field update are two writes: run them in one
		// transaction so a mid-way failure can't leave a half-updated row.
		var f store.Function
		err = func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)
			if keyChanged {
				if _, err := tq.UpdateFunctionKey(c.Context(), store.UpdateFunctionKeyParams{
					ID:        id,
					Source:    b.Source,
					Interface: b.Interface,
					Name:      b.Name,
					UpdatedAt: now,
				}); err != nil {
					return err
				}
				rename := store.RenameFunctionGrammarParams{
					NewSource: b.Source, NewInterface: b.Interface, NewName: b.Name,
					UpdatedAt: now,
					OldSource: current.Source, OldInterface: current.Interface, OldName: current.Name,
				}
				if err := tq.RenameFunctionGrammar(c.Context(), rename); err != nil {
					return err
				}
				if err := tq.RenameScriptKeyAudits(c.Context(), store.RenameScriptKeyAuditsParams(rename)); err != nil {
					return err
				}
			}
			f, err = tq.UpdateFunction(c.Context(), store.UpdateFunctionParams{
				ID:             id,
				Description:    deref(b.Description),
				Convention:     nullInt64Value(b.Convention),
				Category:       deref(b.Category),
				ExecutionRole:  deref(b.ExecutionRole),
				ExecutionPhase: deref(b.ExecutionPhase),
				UpdatedAt:      now,
			})
			if err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			if store.IsUniqueViolation(err) {
				return conflict(c, b)
			}
			return serverError(c, "update function", err)
		}
		return c.JSON(functionResponse(f))
	}
}

// nullInt64 converts an optional JSON integer into a sql.NullInt64 for
// the generated nullable-column parameter.
func nullInt64Value(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func deleteFunction(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid function id"})
		}
		affected, err := q.DeleteFunction(c.Context(), id)
		if err != nil {
			return serverError(c, "delete function", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
