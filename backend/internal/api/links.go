package api

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"webprogfndb/internal/store"
)

// This file implements the function editor's link endpoints (ticket 04):
// list, attach (by picking an existing definition or creating it inline in
// one flow), update of the editable attributes (device scope, position,
// required), and detach. Detaching never deletes the definition; deleting a
// definition that links still reference is guarded by the dictionary
// delete endpoint's warning (ticket 02).
//
// The parameter and criteria dictionaries share one link shape (CONTEXT.md),
// so the handlers are generic over a linkStore adapter. Unlike
// dictionaryRoutes[T] in dictionary.go, this uses a struct of closures: the
// generated link queries have distinct parameter/row types per table, so
// unifying them needs explicit mapping rather than a type parameter.

// linkView is the JSON shape of a link row. resolved/definition_id expose
// the Orphan state (ADR-0003): a link whose definition does not exist is
// kept, rendered with an unresolved warning, and stays editable in place.
type linkView struct {
	ID             int64
	FunctionSource string
	FunctionName   string
	DeviceName     string
	Name           string
	Position       sql.NullInt64
	Required       bool
	SourceFile     string
	Resolved       bool
	DefinitionID   *int64
	CreatedAt      string
	UpdatedAt      string
}

func viewFromJoinedRow(
	id int64, functionSource, functionName, deviceName, name string,
	position sql.NullInt64, required int64, sourceFile string,
	definitionID sql.NullInt64, createdAt, updatedAt string,
) linkView {
	v := linkView{
		ID:             id,
		FunctionSource: functionSource,
		FunctionName:   functionName,
		DeviceName:     deviceName,
		Name:           name,
		Position:       position,
		Required:       required != 0,
		SourceFile:     sourceFile,
		Resolved:       definitionID.Valid,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
	if definitionID.Valid {
		did := definitionID.Int64
		v.DefinitionID = &did
	}
	return v
}

func linkResponse(v linkView) fiber.Map {
	var position *int64
	if v.Position.Valid {
		p := v.Position.Int64
		position = &p
	}
	var defID *int64
	if v.Resolved {
		defID = v.DefinitionID
	}
	return fiber.Map{
		"id":              v.ID,
		"function_source": v.FunctionSource,
		"function_name":   v.FunctionName,
		"device_name":     v.DeviceName,
		"name":            v.Name,
		"position":        position,
		"required":        v.Required,
		"source_file":     v.SourceFile,
		"resolved":        v.Resolved,
		"definition_id":   defID,
		"created_at":      v.CreatedAt,
		"updated_at":      v.UpdatedAt,
	}
}

// sortLinkViews orders rows by position with unordered (NULL) links last,
// then name — the argument order callers see in generated scripts.
func sortLinkViews(items []linkView) {
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.Position.Valid != b.Position.Valid {
			return a.Position.Valid
		}
		if a.Position.Valid && a.Position.Int64 != b.Position.Int64 {
			return a.Position.Int64 < b.Position.Int64
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.ID < b.ID
	})
}

// insertLinkParams carries the validated attach payload.
type insertLinkParams struct {
	FunctionSource string
	FunctionName   string
	DeviceName     string
	Name           string
	Position       sql.NullInt64
	Required       bool
}

// updateLinkAttrs carries the validated update payload: the only editable
// link attributes (ADR-0002 device scope, position, required).
type updateLinkAttrs struct {
	DeviceName string
	Position   sql.NullInt64
	Required   bool
}

// httpError carries a ready-to-render JSON error response from a store
// adapter back to the handler.
type httpError struct {
	status int
	body   fiber.Map
}

func (e *httpError) Error() string {
	if msg, ok := e.body["error"].(string); ok {
		return msg
	}
	return "request failed"
}

// linkStore adapts the two generated link query sets to one interface. The
// *store.Queries parameter lets handlers pass a transaction-scoped Queries
// where atomicity matters (inline creation).
type linkStore struct {
	defExists func(ctx context.Context, q *store.Queries, name, deviceName string) (bool, error)
	list      func(ctx context.Context, q *store.Queries, source, name string) ([]linkView, error)
	get       func(ctx context.Context, q *store.Queries, id int64) (linkView, error)
	insert    func(ctx context.Context, q *store.Queries, p insertLinkParams, now string) (int64, error)
	update    func(ctx context.Context, q *store.Queries, id int64, a updateLinkAttrs, now string) error
	remove    func(ctx context.Context, q *store.Queries, id int64) (int64, error)
	// insertDefinition creates a definition inline; the handler calls it
	// inside its transaction so a failed link insert rolls the definition
	// back: either both rows land or neither does.
	insertDefinition func(ctx context.Context, q *store.Queries, b definitionBody, now string) error
}

// linkBody is the JSON shape of attach requests. Exactly one of name (pick
// an existing definition) or definition (create it inline in the same flow)
// is required. Update requests accept only the editable attributes.
type linkBody struct {
	Name       *string         `json:"name"`
	Definition *definitionBody `json:"definition"`
	DeviceName string          `json:"device_name"`
	Position   *int64          `json:"position"`
	Required   bool            `json:"required"`
}

// validate returns field-level messages; an empty map means valid.
func (b *linkBody) validate() map[string]string {
	fields := map[string]string{}
	// ADR-0002: match the registry's normalization so casing twins
	// cannot diverge from a registered Device.
	b.DeviceName = normalizeDeviceName(b.DeviceName)
	if b.Definition != nil {
		if b.Name != nil && strings.TrimSpace(*b.Name) != "" {
			// A contradictory payload (pick AND create) is rejected, not
			// silently resolved.
			fields["name"] = "provide either name or definition, not both"
		}
		if f := validateDefinition(b.Definition); len(f) > 0 {
			for k, v := range f {
				fields["definition."+k] = v
			}
		}
	} else {
		name := strings.TrimSpace(deref(b.Name))
		b.Name = &name
		if name == "" {
			fields["name"] = "is required (pick an existing definition or provide a definition object)"
		}
	}
	if b.Position != nil && *b.Position < 1 {
		fields["position"] = "must be a positive integer"
	}
	return fields
}

// nullPosition converts the optional JSON position into a nullable column
// value. An absent or null position stores NULL (unordered link).
func (b *linkBody) nullPosition() sql.NullInt64 {
	if b.Position == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *b.Position, Valid: true}
}

// registerLinkRoutes wires the link endpoints for one dictionary under
// /api/functions/:id/<noun>-links. noun ("parameter"/"criteria") also names
// the resource in error messages.
func registerLinkRoutes(app *fiber.App, q *store.Queries, d *sql.DB, s linkStore, noun string) {
	segment := noun + "-links"
	app.Get("/api/functions/:id/"+segment, listLinks(q, s, noun))
	app.Post("/api/functions/:id/"+segment, attachLink(q, d, s, noun))
	app.Put("/api/functions/:id/"+segment+"/:linkId", updateLink(q, s, noun))
	app.Delete("/api/functions/:id/"+segment+"/:linkId", detachLink(q, s, noun))
}

// resolveFunctionKey parses the :id parameter and loads the function's
// natural key. ok is false for a malformed id or an unknown function; both
// render as 404 for the link routes.
func resolveFunctionKey(c fiber.Ctx, q *store.Queries) (source, name string, ok bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return "", "", false
	}
	f, err := q.GetFunction(c.Context(), id)
	if err != nil {
		return "", "", false
	}
	return f.Source, f.Name, true
}

// parseLinkID parses the :linkId parameter.
func parseLinkID(c fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("linkId"), 10, 64)
}

func listLinks(q *store.Queries, s linkStore, noun string) fiber.Handler {
	return func(c fiber.Ctx) error {
		source, name, ok := resolveFunctionKey(c, q)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		items, err := s.list(c.Context(), q, source, name)
		if err != nil {
			return serverError(c, "list "+noun+" links", err)
		}
		sortLinkViews(items)
		out := make([]fiber.Map, 0, len(items))
		for _, v := range items {
			out = append(out, linkResponse(v))
		}
		return c.JSON(fiber.Map{"items": out})
	}
}

func attachLink(q *store.Queries, d *sql.DB, s linkStore, noun string) fiber.Handler {
	return func(c fiber.Ctx) error {
		source, name, ok := resolveFunctionKey(c, q)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		var b linkBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		if fields := b.validate(); len(fields) > 0 {
			return validationError(c, fields)
		}

		now := time.Now().UTC().Format(time.RFC3339)
		p := insertLinkParams{
			FunctionSource: source,
			FunctionName:   name,
			DeviceName:     b.DeviceName,
			Name:           deref(b.Name),
			Position:       b.nullPosition(),
			Required:       b.Required,
		}
		if b.Definition != nil {
			p.Name = b.Definition.Name
		}

		var v linkView
		// Definition insert + link insert are one transaction: an inline
		// creation whose link conflicts must not leave the definition behind.
		err := func() error {
			tx, err := d.BeginTx(c.Context(), nil)
			if err != nil {
				return err
			}
			defer tx.Rollback()
			tq := q.WithTx(tx)

			if b.Definition != nil {
				if err := s.insertDefinition(c.Context(), tq, *b.Definition, now); err != nil {
					return err
				}
			} else {
				exists, err := s.defExists(c.Context(), tq, p.Name, p.DeviceName)
				if err != nil {
					return err
				}
				if !exists {
					return &httpError{
						status: fiber.StatusUnprocessableEntity,
						body: fiber.Map{
							"error":  "validation failed",
							"fields": fiber.Map{"name": "no " + noun + " definition named '" + p.Name + "'"},
						},
					}
				}
			}

			linkID, err := s.insert(c.Context(), tq, p, now)
			if err != nil {
				return err
			}
			v, err = s.get(c.Context(), tq, linkID)
			if err != nil {
				return err
			}
			return tx.Commit()
		}()
		if err != nil {
			if he, ok := err.(*httpError); ok {
				return c.Status(he.status).JSON(he.body)
			}
			if store.IsUniqueViolation(err) {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": noun + " link '" + p.Name + "' already exists for device scope '" + p.DeviceName + "'",
				})
			}
			return serverError(c, "attach "+noun+" link", err)
		}
		return c.Status(fiber.StatusCreated).JSON(linkResponse(v))
	}
}

func updateLink(q *store.Queries, s linkStore, noun string) fiber.Handler {
	return func(c fiber.Ctx) error {
		source, name, ok := resolveFunctionKey(c, q)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		id, err := parseLinkID(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
		}
		var b linkBody
		if err := c.Bind().Body(&b); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
		}
		// The link's name/definition is fixed at attach time; renaming a
		// link means detach + attach. Orphan links are updated the same way
		// as resolved ones: no dictionary lookup here (ADR-0003).
		//
		// Update is a full replace of the editable attributes, matching the
		// PUT contract: an omitted position becomes NULL and an omitted
		// required becomes false. Clients (the link tabs) always send all
		// three; the round-trip is pinned by TestUpdateParameterLinkReplace.
		if b.Name != nil || b.Definition != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "only device_name, position and required are editable on a link; detach and re-attach to change the definition",
			})
		}
		if b.Position != nil && *b.Position < 1 {
			return validationError(c, map[string]string{"position": "must be a positive integer"})
		}

		v, err := s.get(c.Context(), q, id)
		if err != nil || v.FunctionSource != source || v.FunctionName != name {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": noun + " link not found"})
		}

		now := time.Now().UTC().Format(time.RFC3339)
		if err := s.update(c.Context(), q, id, updateLinkAttrs{
			// ADR-0002: match the registry's normalization (see linkBody.validate).
			DeviceName: normalizeDeviceName(b.DeviceName),
			Position:   b.nullPosition(),
			Required:   b.Required,
		}, now); err != nil {
			return serverError(c, "update "+noun+" link", err)
		}
		// Re-read so resolved/definition_id reflect the dictionary as of now.
		v, err = s.get(c.Context(), q, id)
		if err != nil {
			return serverError(c, "reload "+noun+" link", err)
		}
		return c.JSON(linkResponse(v))
	}
}

func detachLink(q *store.Queries, s linkStore, noun string) fiber.Handler {
	return func(c fiber.Ctx) error {
		source, name, ok := resolveFunctionKey(c, q)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "function not found"})
		}
		id, err := parseLinkID(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid link id"})
		}
		v, err := s.get(c.Context(), q, id)
		if err != nil || v.FunctionSource != source || v.FunctionName != name {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": noun + " link not found"})
		}
		if _, err := s.remove(c.Context(), q, id); err != nil {
			return serverError(c, "detach "+noun+" link", err)
		}
		// The definition is intentionally untouched: detaching a link never
		// deletes the definition (CONTEXT.md, user story 11).
		return c.SendStatus(fiber.StatusNoContent)
	}
}

// parameterLinkStore adapts the parameter link and dictionary queries.
func parameterLinkStore() linkStore {
	return linkStore{
		// defExists: a link resolves when an exact-scope or global
		// definition exists (ADR-0009). The store query returns the row
		// we would JOIN to; the body discards it.
		defExists: func(ctx context.Context, q *store.Queries, name, deviceName string) (bool, error) {
			_, err := q.ResolveParameterDefinitionByLink(ctx, store.ResolveParameterDefinitionByLinkParams{
				Name:       name,
				DeviceName: deviceName,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return err == nil, err
		},
		list: func(ctx context.Context, q *store.Queries, source, name string) ([]linkView, error) {
			rows, err := q.ListParameterLinksByFunction(ctx, store.ListParameterLinksByFunctionParams{
				FunctionSource: source,
				FunctionName:   name,
			})
			if err != nil {
				return nil, err
			}
			items := make([]linkView, 0, len(rows))
			for _, r := range rows {
				items = append(items, viewFromJoinedRow(
					r.ID, r.FunctionSource, r.FunctionName, r.DeviceName, r.Name,
					r.Position, r.Required, r.SourceFile, r.DefinitionID,
					r.CreatedAt, r.UpdatedAt,
				))
			}
			return items, nil
		},
		get: func(ctx context.Context, q *store.Queries, id int64) (linkView, error) {
			r, err := q.GetParameterLink(ctx, id)
			if err != nil {
				return linkView{}, err
			}
			return viewFromJoinedRow(
				r.ID, r.FunctionSource, r.FunctionName, r.DeviceName, r.Name,
				r.Position, r.Required, r.SourceFile, r.DefinitionID,
				r.CreatedAt, r.UpdatedAt,
			), nil
		},
		insert: func(ctx context.Context, q *store.Queries, p insertLinkParams, now string) (int64, error) {
			link, err := q.InsertParameterLink(ctx, store.InsertParameterLinkParams{
				FunctionSource: p.FunctionSource,
				FunctionName:   p.FunctionName,
				DeviceName:     p.DeviceName,
				Name:           p.Name,
				Position:       p.Position,
				Required:       boolToInt64(&p.Required),
				CreatedAt:      now,
				UpdatedAt:      now,
			})
			if err != nil {
				return 0, err
			}
			return link.ID, nil
		},
		update: func(ctx context.Context, q *store.Queries, id int64, a updateLinkAttrs, now string) error {
			_, err := q.UpdateParameterLink(ctx, store.UpdateParameterLinkParams{
				DeviceName: a.DeviceName,
				Position:   a.Position,
				Required:   boolToInt64(&a.Required),
				UpdatedAt:  now,
				ID:         id,
			})
			return err
		},
		remove: func(ctx context.Context, q *store.Queries, id int64) (int64, error) {
			return q.DeleteParameterLink(ctx, id)
		},
		insertDefinition: func(ctx context.Context, q *store.Queries, b definitionBody, now string) error {
			_, err := q.InsertParameterDefinition(ctx, store.InsertParameterDefinitionParams{
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			if err != nil && store.IsUniqueViolation(err) {
				return &httpError{
					status: fiber.StatusConflict,
					body:   fiber.Map{"error": "definition '" + b.Name + "' already exists (pick it from the dropdown instead)"},
				}
			}
			return err
		},
	}
}

// criteriaLinkStore adapts the criteria link and dictionary queries.
func criteriaLinkStore() linkStore {
	return linkStore{
		defExists: func(ctx context.Context, q *store.Queries, name, deviceName string) (bool, error) {
			_, err := q.ResolveCriteriaDefinitionByLink(ctx, store.ResolveCriteriaDefinitionByLinkParams{
				Name:       name,
				DeviceName: deviceName,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return err == nil, err
		},
		list: func(ctx context.Context, q *store.Queries, source, name string) ([]linkView, error) {
			rows, err := q.ListCriteriaLinksByFunction(ctx, store.ListCriteriaLinksByFunctionParams{
				FunctionSource: source,
				FunctionName:   name,
			})
			if err != nil {
				return nil, err
			}
			items := make([]linkView, 0, len(rows))
			for _, r := range rows {
				items = append(items, viewFromJoinedRow(
					r.ID, r.FunctionSource, r.FunctionName, r.DeviceName, r.Name,
					r.Position, r.Required, r.SourceFile, r.DefinitionID,
					r.CreatedAt, r.UpdatedAt,
				))
			}
			return items, nil
		},
		get: func(ctx context.Context, q *store.Queries, id int64) (linkView, error) {
			r, err := q.GetCriteriaLink(ctx, id)
			if err != nil {
				return linkView{}, err
			}
			return viewFromJoinedRow(
				r.ID, r.FunctionSource, r.FunctionName, r.DeviceName, r.Name,
				r.Position, r.Required, r.SourceFile, r.DefinitionID,
				r.CreatedAt, r.UpdatedAt,
			), nil
		},
		insert: func(ctx context.Context, q *store.Queries, p insertLinkParams, now string) (int64, error) {
			link, err := q.InsertCriteriaLink(ctx, store.InsertCriteriaLinkParams{
				FunctionSource: p.FunctionSource,
				FunctionName:   p.FunctionName,
				DeviceName:     p.DeviceName,
				Name:           p.Name,
				Position:       p.Position,
				Required:       boolToInt64(&p.Required),
				CreatedAt:      now,
				UpdatedAt:      now,
			})
			if err != nil {
				return 0, err
			}
			return link.ID, nil
		},
		update: func(ctx context.Context, q *store.Queries, id int64, a updateLinkAttrs, now string) error {
			_, err := q.UpdateCriteriaLink(ctx, store.UpdateCriteriaLinkParams{
				DeviceName: a.DeviceName,
				Position:   a.Position,
				Required:   boolToInt64(&a.Required),
				UpdatedAt:  now,
				ID:         id,
			})
			return err
		},
		remove: func(ctx context.Context, q *store.Queries, id int64) (int64, error) {
			return q.DeleteCriteriaLink(ctx, id)
		},
		insertDefinition: func(ctx context.Context, q *store.Queries, b definitionBody, now string) error {
			_, err := q.InsertCriteriaDefinition(ctx, store.InsertCriteriaDefinitionParams{
				Name:          b.Name,
				Description:   deref(b.Description),
				Type:          b.Type,
				Unit:          deref(b.Unit),
				DefaultValue:  nullStrPtr(b.DefaultValue),
				AllowedValues: deref(b.AllowedValues),
				Cardinality:   deref(b.Cardinality),
				ValuePattern:  deref(b.ValuePattern),
				Required:      boolToInt64(b.Required),
				DeviceName:    deref(b.DeviceName),
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			if err != nil && store.IsUniqueViolation(err) {
				return &httpError{
					status: fiber.StatusConflict,
					body:   fiber.Map{"error": "definition '" + b.Name + "' already exists (pick it from the dropdown instead)"},
				}
			}
			return err
		},
	}
}
