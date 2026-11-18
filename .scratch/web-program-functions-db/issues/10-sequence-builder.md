# 10: Sequence builder (multi-function sequences)

**What to build:** First-class stored Sequences per ADR-0005: a unique-name aggregate with ordered Sequence steps, each step referencing a Function by natural key (source, interface, name) with parameter/criteria values stored as JSON TEXT, per-step timeout_ms/flow/on_fail, device scope, and an optional note. Dedicated `/sequences` page (list + detail with ordered step cards, reorder, whole-sequence preview + copy-all); the composer on a Function page gains an "add step to sequence" action (pick existing sequence or create new).

- [x] Schema: migration 0004 (ANSI SQL) — `sequences(id PK, name UNIQUE NOT NULL, description, created_at, updated_at)`; `sequence_steps(id PK, sequence_id FK→sequences ON DELETE CASCADE, position INT, function_source, function_interface, function_name, parameters TEXT, criteria TEXT, timeout_ms, flow, on_fail, device_name, note)`. No FK from steps to functions (orphan-able, mirror link tables). FK precedent exists in 0003 (`import_runs`), so the aggregate-internal cascade is used.
- [x] API: `GET /api/sequences` (search + Go-side pagination like other lists), `POST /api/sequences`, `GET /api/sequences/:id`, `PUT /api/sequences/:id` (full replace, one tx), `DELETE /api/sequences/:id`, `POST /api/sequences/:id/emit`.
- [x] Rules: unique sequence name (409); minimum 1 step (422); same Function may appear in multiple steps; step `position` = array index+1 (reorder = renumber via PUT); no existence check on step function references at API level (orphan steps storable per ADR-0003); deleting a referenced Function leaves steps orphaned and flagged — the same treatment link rows get (no blocking, no warning dialog).
- [x] Emission (server-owned, single source of truth with ticket 07): each step re-validated live via the same `validateAndEmit` pure function; steps joined with newlines; response carries per-step `{resolved, script, errors}` — unresolved/invalid steps flagged, clean steps still emit. Emitted text is never parsed back; steps are the source of truth.
- [x] Reset behavior: "reset database from workbook" leaves sequences untouched (importer truncates the six entity tables only; pinned by `TestResetLeavesSequencesUntouched`).
- [x] UI: `/sequences` list page (search, totals); sequence detail — header rename/delete, per-step cards (orphan badge, values as chips, edit form with `name=value` line editing for parameters/criteria, up/down reorder buttons instead of drag, remove step), live whole-sequence preview with per-step flags, copy-all; composer dialog gains "Add to sequence" (existing picker or create-new, device scope prefilled from links).
- [x] Tests: Go integration — CRUD round-trip, unique-name 409, min-1-step 422, emit per-step flags, orphan flag after function delete, failed-PUT atomicity, reset survival. E2E happy path: seed two-step sequence → emit → reorder → composer "Add to sequence" appends (3 steps).

**Blocked by:** None.

**Status:** done

Annotations:
- `validateAndEmit` reused unchanged for per-step emission — sequence emission is literally the composer's pure function applied per step, then newline-joined.
- Step edit form edits name-keyed values as `name=value` lines — never a pipe-grammar parse (steps are the source of truth, ADR-0005).
- Reorder uses ↑/↓ buttons, not drag — same capability, less code, keyboard accessible.
- sqlc v1.31.1 (MIT) run via `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate`.
- db_test's pinned schema version bumped 3 → 4 with the new migration.
