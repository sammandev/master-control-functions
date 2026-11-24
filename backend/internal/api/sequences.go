package api

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// Sequence builder (ticket 10, ADR-0005). A Sequence is a web-native named
// aggregate of ordered steps, each referencing a Function by natural key
// (source, interface, name). The API stores steps as submitted — orphan
// references are storable and flagged (ADR-0003); emission re-validates
// every step against the current definitions at request time and joins the
// clean scripts with newlines. Emitted text is never parsed back: steps
// are the source of truth.

// sequenceStepBody is one step on create/PUT. Parameters and criteria are
// name-keyed raw values, stored as JSON objects.
type sequenceStepBody struct {
	FunctionSource    string            `json:"function_source"`
	FunctionInterface string            `json:"function_interface"`
	FunctionName      string            `json:"function_name"`
	Parameters        map[string]string `json:"parameters"`
	Criteria          map[string]string `json:"criteria"`
	TimeoutMs         string            `json:"timeout_ms"`
	Flow              string            `json:"flow"`
	OnFail            string            `json:"on_fail"`
	DeviceName        string            `json:"device_name"`
	Note              string            `json:"note"`
}

// sequenceBody is the write payload. The metadata and step routes each
// use only their own fields.
type sequenceBody struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Steps       []sequenceStepBody `json:"steps"`
}

func (b *sequenceBody) validateMeta() map[string]string {
	fields := map[string]string{}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" {
		fields["name"] = "is required"
	}
	return fields
}

func (b *sequenceBody) validate() map[string]string {
	fields := b.validateMeta()
	b.validateSteps(fields)
	return fields
}

func (b *sequenceBody) validateSteps(fields map[string]string) {
	if len(b.Steps) == 0 {
		fields["steps"] = "at least one step is required"
	}
	for i := range b.Steps {
		s := &b.Steps[i]
		s.FunctionSource = strings.TrimSpace(s.FunctionSource)
		s.FunctionInterface = normalizeDeviceName(s.FunctionInterface)
		s.FunctionName = strings.TrimSpace(s.FunctionName)
		s.DeviceName = normalizeDeviceName(s.DeviceName)
		if s.FunctionSource == "" {
			fields["steps.function_source"] = "step " + strconv.Itoa(i+1) + ": source is required"
		}
		if problem := validateDeviceName(s.FunctionInterface); problem != "" {
			fields["steps.function_interface"] = "step " + strconv.Itoa(i+1) + ": " + problem
		}
		if problem := validateDeviceName(s.DeviceName); problem != "" {
			fields["steps.device_name"] = "step " + strconv.Itoa(i+1) + ": " + problem
		}
		if s.FunctionName == "" {
			fields["steps.function_name"] = "step " + strconv.Itoa(i+1) + ": name is required"
		}
	}
}

// ---- responses -----------------------------------------------------------

type sequenceStepResponse struct {
	ID                int64             `json:"id"`
	Position          int64             `json:"position"`
	FunctionSource    string            `json:"function_source"`
	FunctionInterface string            `json:"function_interface"`
	FunctionName      string            `json:"function_name"`
	Parameters        map[string]string `json:"parameters"`
	Criteria          map[string]string `json:"criteria"`
	TimeoutMs         string            `json:"timeout_ms"`
	Flow              string            `json:"flow"`
	OnFail            string            `json:"on_fail"`
	DeviceName        string            `json:"device_name"`
	Note              string            `json:"note"`
	FunctionExists    bool              `json:"function_exists"`
}

type sequenceResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []sequenceStepResponse `json:"steps"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// parseStoredMap decodes a stored parameters/criteria JSON object. The
// column is only ever written by this package via json.Marshal, so a
// decode failure means real corruption and surfaces as a server error
// rather than being silently swallowed.
func parseStoredMap(raw string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func sequenceStepResponseFromRow(r store.ListSequenceStepsRow) (sequenceStepResponse, error) {
	params, err := parseStoredMap(r.Parameters)
	if err != nil {
		return sequenceStepResponse{}, err
	}
	criteria, err := parseStoredMap(r.Criteria)
	if err != nil {
		return sequenceStepResponse{}, err
	}
	return sequenceStepResponse{
		ID:                r.ID,
		Position:          r.Position,
		FunctionSource:    r.FunctionSource,
		FunctionInterface: r.FunctionInterface,
		FunctionName:      r.FunctionName,
		Parameters:        params,
		Criteria:          criteria,
		TimeoutMs:         r.TimeoutMs,
		Flow:              r.Flow,
		OnFail:            r.OnFail,
		DeviceName:        r.DeviceName,
		Note:              r.Note,
		FunctionExists:    r.FunctionExists,
	}, nil
}

// sequenceDetail loads the sequence and its ordered steps and builds the
// shared response shape for GET, POST, PUT and PATCH.
func sequenceDetail(c fiber.Ctx, q *store.Queries, id int64) (sequenceResponse, error) {
	row, err := q.GetSequence(c.Context(), id)
	if err != nil {
		return sequenceResponse{}, err
	}
	steps, err := q.ListSequenceSteps(c.Context(), id)
	if err != nil {
		return sequenceResponse{}, err
	}
	out := sequenceResponse{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Steps:       make([]sequenceStepResponse, 0, len(steps)),
	}
	for _, s := range steps {
		item, err := sequenceStepResponseFromRow(s)
		if err != nil {
			return sequenceResponse{}, err
		}
		out.Steps = append(out.Steps, item)
	}
	return out, nil
}

// writeSteps replaces a sequence's steps inside the caller's transaction:
// delete all, re-insert in array order with position = index+1.
func writeSteps(ctx fiber.Ctx, tq *store.Queries, sequenceID int64, steps []sequenceStepBody) error {
	if err := tq.DeleteSequenceSteps(ctx.Context(), sequenceID); err != nil {
		return err
	}
	for i, s := range steps {
		if _, err := tq.InsertSequenceStep(ctx.Context(), store.InsertSequenceStepParams{
			SequenceID:        sequenceID,
			Position:          int64(i + 1),
			FunctionSource:    s.FunctionSource,
			FunctionInterface: s.FunctionInterface,
			FunctionName:      s.FunctionName,
			Parameters:        marshalStoredMap(s.Parameters),
			Criteria:          marshalStoredMap(s.Criteria),
			TimeoutMs:         strings.TrimSpace(s.TimeoutMs),
			Flow:              strings.TrimSpace(s.Flow),
			OnFail:            strings.TrimSpace(s.OnFail),
			DeviceName:        strings.TrimSpace(s.DeviceName),
			Note:              s.Note,
		}); err != nil {
			return err
		}
	}
	return nil
}

// marshalStoredMap encodes a name→value map as the stored JSON object;
// nil becomes {}.
func marshalStoredMap(m map[string]string) string {
	if m == nil {
		m = map[string]string{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		// map[string]string marshalling cannot fail; belt and braces.
		return "{}"
	}
	return string(b)
}

// ---- handlers ------------------------------------------------------------

// registerSequenceRoutes wires /api/sequences endpoints.
func registerSequenceRoutes(app *fiber.App, q *store.Queries, d *sql.DB) {
	app.Get("/api/sequences", listSequences(q))
	app.Post("/api/sequences", createSequence(q, d))
	app.Get("/api/sequences/:id", getSequence(q))
	app.Put("/api/sequences/:id", updateSequence(q, d, sequenceUpdateAll))
	app.Patch("/api/sequences/:id", updateSequence(q, d, sequenceUpdateMeta))
	app.Put("/api/sequences/:id/steps", updateSequence(q, d, sequenceUpdateSteps))
	app.Delete("/api/sequences/:id", deleteSequence(q))
	app.Post("/api/sequences/:id/emit", emitSequence(q))
}

// sequenceSortColumns whitelists the list's sortable columns; any other
// sort value keeps the default (name, id) order. Text columns use
// naturalCompare, matching the page's former client-side sort.
var sequenceSortColumns = map[string]func(a, b store.ListSequencesRow) int{
	"name":        func(a, b store.ListSequencesRow) int { return naturalCompare(a.Name, b.Name) },
	"description": func(a, b store.ListSequencesRow) int { return naturalCompare(a.Description, b.Description) },
	"steps":       func(a, b store.ListSequencesRow) int { return cmp.Compare(a.StepCount, b.StepCount) },
	"updated_at":  func(a, b store.ListSequencesRow) int { return strings.Compare(a.UpdatedAt, b.UpdatedAt) },
}

// naturalCompare orders text case-insensitively with digit runs compared
// by value, so "Run 2" sorts before "Run 10".
func naturalCompare(a, b string) int {
	a, b = strings.ToLower(a), strings.ToLower(b)
	for a != "" && b != "" {
		da, db := digitPrefix(a), digitPrefix(b)
		if da != "" && db != "" {
			na, nb := strings.TrimLeft(da, "0"), strings.TrimLeft(db, "0")
			if c := cmp.Or(cmp.Compare(len(na), len(nb)), strings.Compare(na, nb)); c != 0 {
				return c
			}
			a, b = a[len(da):], b[len(db):]
			continue
		}
		ra, sa := utf8.DecodeRuneInString(a)
		rb, sb := utf8.DecodeRuneInString(b)
		if c := cmp.Compare(ra, rb); c != 0 {
			return c
		}
		a, b = a[sa:], b[sb:]
	}
	return cmp.Compare(len(a), len(b))
}

func digitPrefix(s string) string {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i]
}

func listSequences(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		var search sql.NullString
		if s := strings.TrimSpace(c.Query("search")); s != "" {
			if p := likePattern(s); p != "" {
				search = sql.NullString{String: p, Valid: true}
			}
		}
		total, err := q.CountSequences(c.Context(), search)
		if err != nil {
			return serverError(c, "count sequences", err)
		}
		lq := parseListQuery(c)
		start := int64(math.MaxInt64)
		if lq.page-1 <= math.MaxInt64/lq.pageSize {
			start = (lq.page - 1) * lq.pageSize
		}
		compare := sequenceSortColumns[c.Query("sort")]
		params := store.ListSequencesParams{Search: search}
		if compare == nil {
			params.PageLimit = sql.NullInt64{Int64: lq.pageSize, Valid: true}
			params.PageOffset = start
		}
		rows, err := q.ListSequences(c.Context(), params)
		if err != nil {
			return serverError(c, "list sequences", err)
		}
		// Custom sorts operate on all matches, keeping the SQL order
		// (name, id) as the tiebreak.
		if compare != nil {
			desc := strings.EqualFold(c.Query("order"), "desc")
			slices.SortStableFunc(rows, func(a, b store.ListSequencesRow) int {
				if desc {
					return compare(b, a)
				}
				return compare(a, b)
			})
			start = min(start, int64(len(rows)))
			end := min(start+lq.pageSize, int64(len(rows)))
			rows = rows[start:end]
		}
		items := make([]fiber.Map, 0, len(rows))
		for _, r := range rows {
			items = append(items, fiber.Map{
				"id":          r.ID,
				"name":        r.Name,
				"description": r.Description,
				"step_count":  r.StepCount,
				"created_at":  r.CreatedAt,
				"updated_at":  r.UpdatedAt,
			})
		}
		return c.JSON(fiber.Map{
			"items":     items,
			"total":     total,
			"page":      lq.page,
			"page_size": lq.pageSize,
		})
	}
}

func createSequence(q *store.Queries, d *sql.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var b sequenceBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)

		var id int64
		// Sequence row + steps are one transaction: a step insert failure
		// must not leave a stepless sequence behind.
		err := func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)
			seq, err := tq.InsertSequence(c.Context(), store.InsertSequenceParams{
				Name:        b.Name,
				Description: b.Description,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			if err != nil {
				return err
			}
			id = seq.ID
			if err := writeSteps(c, tq, id, b.Steps); err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "a sequence named '" + b.Name + "' already exists",
				})
			}
			return serverError(c, "create sequence", err)
		}
		out, err := sequenceDetail(c, q, id)
		if err != nil {
			return serverError(c, "read created sequence", err)
		}
		return c.Status(fiber.StatusCreated).JSON(out)
	}
}

func getSequence(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		out, err := sequenceDetail(c, q, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sequence not found"})
			}
			return serverError(c, "read sequence", err)
		}
		return c.JSON(out)
	}
}

type sequenceUpdateMode int

const (
	sequenceUpdateAll sequenceUpdateMode = iota
	sequenceUpdateMeta
	sequenceUpdateSteps
)

func updateSequence(q *store.Queries, d *sql.DB, mode sequenceUpdateMode) fiber.Handler {
	writeMeta := mode != sequenceUpdateSteps
	replaceSteps := mode != sequenceUpdateMeta
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		var b sequenceBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		fields := map[string]string{}
		if writeMeta {
			fields = b.validateMeta()
		}
		if replaceSteps {
			b.validateSteps(fields)
		}
		if len(fields) > 0 {
			return validationError(c, fields)
		}
		now := time.Now().UTC().Format(time.RFC3339)

		err = func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)
			var affected int64
			if writeMeta {
				affected, err = tq.UpdateSequenceMeta(c.Context(), store.UpdateSequenceMetaParams{
					Name:        b.Name,
					Description: b.Description,
					UpdatedAt:   now,
					ID:          id,
				})
			} else {
				// Lock the parent before replacing steps, without rewriting its metadata.
				affected, err = tq.TouchSequence(c.Context(), store.TouchSequenceParams{
					UpdatedAt: now,
					ID:        id,
				})
			}
			if err != nil {
				return err
			}
			if affected == 0 {
				return sql.ErrNoRows
			}
			if replaceSteps {
				if err := writeSteps(c, tq, id, b.Steps); err != nil {
					return err
				}
			}
			return tx.Commit()
		}()
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sequence not found"})
			}
			if writeMeta && store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "a sequence named '" + b.Name + "' already exists",
				})
			}
			return serverError(c, "update sequence", err)
		}
		out, err := sequenceDetail(c, q, id)
		if err != nil {
			return serverError(c, "read updated sequence", err)
		}
		return c.JSON(out)
	}
}

func deleteSequence(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		affected, err := q.DeleteSequence(c.Context(), id)
		if err != nil {
			return serverError(c, "delete sequence", err)
		}
		if affected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sequence not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// emitSequence re-validates every stored step against the current
// definitions and returns the newline-joined script of the clean steps
// plus a per-step result array. Orphaned (function gone) and invalid steps
// are flagged, never silently dropped (ADR-0003 spirit, ADR-0005).
func emitSequence(q *store.Queries) fiber.Handler {
	return func(c fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
		}
		steps, err := q.ListSequenceSteps(c.Context(), id)
		if err != nil {
			return serverError(c, "list sequence steps", err)
		}
		if len(steps) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "sequence not found"})
		}

		type stepEmit struct {
			StepID   int64             `json:"step_id"`
			Position int64             `json:"position"`
			Resolved bool              `json:"resolved"`
			Script   string            `json:"script"`
			Errors   map[string]string `json:"errors"`
		}
		type functionData struct {
			fn       store.Function
			params   []composeField
			criteria []composeField
		}
		functions := make(map[store.GetFunctionByNaturalKeyParams]functionData)
		results := make([]stepEmit, 0, len(steps))
		clean := make([]string, 0, len(steps))
		for _, s := range steps {
			item := stepEmit{StepID: s.ID, Position: s.Position, Errors: map[string]string{}}
			params, err := parseStoredMap(s.Parameters)
			if err != nil {
				return serverError(c, "parse step parameters", err)
			}
			criteria, err := parseStoredMap(s.Criteria)
			if err != nil {
				return serverError(c, "parse step criteria", err)
			}

			if !s.FunctionExists {
				item.Errors["function"] = "function '" + s.FunctionName +
					"' with source '" + s.FunctionSource +
					"' and interface '" + s.FunctionInterface + "' not found"
				results = append(results, item)
				continue
			}
			item.Resolved = true

			key := store.GetFunctionByNaturalKeyParams{
				Source:    s.FunctionSource,
				Interface: s.FunctionInterface,
				Name:      s.FunctionName,
			}
			data, ok := functions[key]
			if !ok {
				fn, err := q.GetFunctionByNaturalKey(c.Context(), key)
				if err != nil {
					return serverError(c, "read step function", err)
				}
				pFields, cFields, err := composeFieldsFor(c, q, fn)
				if err != nil {
					return serverError(c, "list composer fields", err)
				}
				data = functionData{fn: fn, params: pFields, criteria: cFields}
				functions[key] = data
			}
			script, fieldErrors := validateAndEmit(data.fn, data.params, data.criteria, composeRequest{
				Parameters: params,
				Criteria:   criteria,
				TimeoutMs:  s.TimeoutMs,
				Flow:       s.Flow,
				OnFail:     s.OnFail,
			})
			if len(fieldErrors) > 0 {
				item.Errors = fieldErrors
			} else {
				item.Script = script
				clean = append(clean, script)
			}
			results = append(results, item)
		}
		return c.JSON(fiber.Map{
			// The joined emission is "emitted text" per CONTEXT.md — "script"
			// is reserved for a single Function's string.
			"text":  strings.Join(clean, "\n"),
			"steps": results,
		})
	}
}
