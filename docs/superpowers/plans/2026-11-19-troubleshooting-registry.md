# Troubleshooting Category & Severity Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace fixed troubleshooting category/severity enums with managed, inline-creatable registries while improving the page's shadcn-svelte/Bits UI usability and fixing known state/data-loading defects.

**Architecture:** Append migration `0010` to seed registry tables and preserve guide references behind foreign keys. Add sqlc queries and transactional Fiber endpoints for registry CRUD plus mandatory same-type reassignment on custom-value deletion. Extend the Svelte client with registry metadata, reusable searchable creation controls, a Manage Lists dialog, and page-level state fixes.

**Tech Stack:** Go, Fiber v3, PostgreSQL, sqlc, Svelte 5, SvelteKit, Bits UI 2, Tailwind CSS 4, Vitest, Playwright.

**Spec:** `docs/superpowers/specs/2026-11-19-troubleshooting-registry-design.md`

## Global Constraints

- PostgreSQL is production; add `backend/internal/db/postgres_migrations/0010_troubleshooting_registry.sql` and never edit applied `0009_troubleshooting.sql`.
- Keep legacy SQLite migration/history compatible only where existing migration tests require it; runtime is PostgreSQL.
- Existing keys remain stable: `connection`, `execution`, `measurement`, `production`, `critical`, `warning`, `info`.
- Built-in category/severity rows cannot be deleted.
- Custom deletion requires a different replacement in the same registry and updates every guide before deleting the source in one transaction.
- Labels are trimmed, 1–64 characters, unique case-insensitively; descriptions are at most 240 characters.
- Custom keys are lowercase kebab-case with numeric suffixes on collision; returned responses contain the generated key.
- Severity tones are exactly `destructive`, `warn`, `default`, or `outline`.
- Use existing semantic tokens and shadcn-svelte/Bits UI wrappers; no raw overlay z-index or route-global DevicePicker redesign.
- Database operations against the approved Ubuntu host remain user-run. Local migration tests and static checks are allowed on this workstation.

## Review Focus

- Two labels that slug to the same key must create distinct stable keys and remain selectable by label.
- A deletion replacement missing, self-targeted, built-in, or in the wrong registry must return a structured error and update zero guides.
- Reassignment failure must roll back both guide updates and source deletion.
- More than 100 related guides must remain discoverable; verify paging or the dedicated related endpoint with a 101-guide fixture.
- Step reordering and guide switching must preserve step draft values and per-guide checklist state; test a non-empty middle step moving upward.

---

### Task 1: PostgreSQL Taxonomy Schema and Store Queries

**Files:**
- Create: `backend/internal/db/postgres_migrations/0010_troubleshooting_registry.sql`
- Create: `backend/internal/db/migrations/0008_troubleshooting_registry.sql` if the migration parity tests require it
- Create: `backend/internal/store/queries/troubleshooting_registry.sql`
- Create: `backend/internal/store/troubleshooting_registry.sql.go` by sqlc generation
- Modify: `backend/internal/store/models.go`
- Modify: `backend/internal/store/queries/troubleshooting.sql`
- Test: `backend/internal/db/migrations_test.go`

**Interfaces:**
- Produces catalog columns: `key`, `label`, `description`, `sort_order`, `built_in`, plus severity `tone`.
- Produces sqlc category methods: `ListTroubleshootingCategories`, `GetTroubleshootingCategory`, `InsertTroubleshootingCategory`, `UpdateTroubleshootingCategory`, `DeleteTroubleshootingCategory`, `ReassignTroubleshootingGuideCategories`.
- Produces matching severity methods, plus `GetTroubleshootingRelatedGuides` and taxonomy-aware guide list ordering.

- [ ] **Step 1: Add migration tests that apply `0010` and assert existing guide keys survive with seeded registry rows, guide foreign keys, built-in flags, and severity tones.**
- [ ] **Step 2: Run `go test ./internal/db/...` and verify the new migration assertions fail before `0010` exists.**
- [ ] **Step 3: Create PostgreSQL migration `0010_troubleshooting_registry.sql`, seed the 11 categories and 5 severities from the spec, drop only the old CHECK constraints, and add catalog foreign keys without changing guide values.**
- [ ] **Step 4: Add sqlc registry queries, related-guide query, and join registry `sort_order`/tone into troubleshooting ordering.**
- [ ] **Step 5: Run `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate` from `backend/` and confirm generated models/methods match the Interfaces block.**
- [ ] **Step 6: Run `go test ./internal/db/... ./internal/store/...`; expected: PASS.**
- [ ] **Step 7: Add task changes to git and commit: `feat: add troubleshooting taxonomies`.**

### Task 2: Key Generation and Registry Transaction API

**Files:**
- Create: `backend/internal/api/troubleshooting_registry.go`
- Create: `backend/internal/api/troubleshooting_registry_test.go`
- Modify: `backend/internal/api/troubleshooting.go`
- Modify: `backend/internal/api/server.go`
- Modify: `backend/internal/api/troubleshooting_validation_test.go`
- Test: `backend/tests/api/troubleshooting_registry_test.go`

**Interfaces:**
- Produces `registerTroubleshootingRegistryRoutes(app, q, d)` and endpoints from the spec.
- Produces `slugTroubleshootingKey(label string) string` and collision-safe persistence using numeric suffixes.
- Consumes generated registry queries from Task 1 and existing `validationError` response shape.

- [ ] **Step 1: Write unit tests for lowercase kebab-case generation, punctuation, Unicode fallback, and collisions `"Fixture Check"` → `fixture-check`, then `fixture-check-2`.**
- [ ] **Step 2: Write API tests for blank/duplicate/too-long labels, invalid tone/order, built-in delete, missing/self/wrong-type replacement, and successful atomic reassignment.**
- [ ] **Step 3: Run focused API tests and verify they fail because registry handlers and key generation do not exist.**
- [ ] **Step 4: Implement taxonomy response structs, normalization/validation, list/create/update handlers, and registry route registration.**
- [ ] **Step 5: Implement delete handlers with replacement validation, guide reassignment, and source deletion in one database transaction.**
- [ ] **Step 6: Change guide validation from fixed maps to registry existence checks before POST/PUT transactions. Keep existing required-field and device validation.**
- [ ] **Step 7: Run `go test ./internal/api/... ./tests/api/...`; expected: non-database tests pass and database tests skip without `TEST_DATABASE_URL`.**
- [ ] **Step 8: Commit: `feat: expose troubleshooting registry API`.**

### Task 3: Frontend Taxonomy Client and Unit Tests

**Files:**
- Modify: `frontend/src/lib/troubleshooting.ts`
- Create: `frontend/src/lib/troubleshooting-registry.ts`
- Modify: `frontend/tests/unit/troubleshooting.spec.ts`
- Create: `frontend/tests/unit/troubleshooting-registry.spec.ts`

**Interfaces:**
- Changes guide `category` and `severity` types from closed unions to `string`.
- Produces `TroubleshootingCategory`, `TroubleshootingSeverity`, registry list/create/update/delete request functions, taxonomy label/tone lookup helpers, and safe unknown-key fallbacks.
- Consumes Task 2 route contracts.

- [ ] **Step 1: Write failing client tests for registry request paths, delete `reassign_to`, case-sensitive key preservation, labels from metadata, and neutral fallbacks for unknown keys.**
- [ ] **Step 2: Run `pnpm --filter master-control-functions-frontend exec vitest run tests/unit/troubleshooting*.spec.ts`; expected: FAIL.**
- [ ] **Step 3: Implement typed registry models and CRUD client functions in `troubleshooting-registry.ts`.**
- [ ] **Step 4: Remove closed category/severity unions and replace static-only label/tone functions with metadata-aware helpers that retain safe fallback behavior.**
- [ ] **Step 5: Run the focused Vitest command; expected: all troubleshooting unit tests pass.**
- [ ] **Step 6: Commit: `feat: add troubleshooting registry client`.**

### Task 4: Reusable Inline-Create Registry Picker

**Files:**
- Create: `frontend/src/lib/components/TroubleshootingRegistryPicker.svelte`
- Test: `frontend/tests/components/TroubleshootingRegistryPicker.svelte.spec.ts`
- Add only if missing: shadcn-svelte `field`, `alert`, `empty`, and `skeleton` source under `frontend/src/lib/components/ui/` using the project's package runner and existing `components.json`.

**Interfaces:**
- Consumes ordered registry items and an async create callback.
- Produces accessible 40px select-like picker with local filtering, exact-match detection, `Create "<term>"`, selected-item display, and inline errors.
- Severity mode accepts inline-created tone `outline` as specified; metadata can be edited later.

- [ ] **Step 1: Write component tests proving a matching selection closes, an unmatched term exposes `Create`, successful creation selects the returned key, and failed creation preserves the typed term.**
- [ ] **Step 2: Run the focused component test and verify it fails because the picker does not exist.**
- [ ] **Step 3: Inspect installed UI wrappers and use the official shadcn-svelte/Bits UI composition; add missing wrappers without overwriting custom files.**
- [ ] **Step 4: Implement `TroubleshootingRegistryPicker.svelte` with `Popover`, `Combobox`, accessible search, 40px trigger, larger item rows, and semantic tokens.**
- [ ] **Step 5: Run the component test, `svelte-check`, and focused lint; expected: PASS with 0 Svelte diagnostics.**
- [ ] **Step 6: Commit: `feat: add troubleshooting registry picker`.**

### Task 5: Manage Lists Dialog and Reassignment UX

**Files:**
- Create: `frontend/src/lib/components/TroubleshootingRegistryDialog.svelte`
- Test: `frontend/tests/components/TroubleshootingRegistryDialog.svelte.spec.ts`

**Interfaces:**
- Consumes registry CRUD functions and metadata arrays from Task 3.
- Produces callbacks `onChanged` and `onOpenChange` for page refresh.
- Reassignment deletion always sends a different `reassign_to`; built-in rows expose no Delete action.

- [ ] **Step 1: Write component tests for category/severity tabs, row save, order controls, usage count, hidden built-in delete, and replacement-required custom delete.**
- [ ] **Step 2: Run the focused test and verify it fails.**
- [ ] **Step 3: Implement row-level editing with Tabs, Inputs/Textareas, severity tone Select, Save, Move Up/Down, and toast feedback.**
- [ ] **Step 4: Implement AlertDialog deletion that names the value and moved guide count, requires a replacement, rejects self-reassignment, and closes only after success.**
- [ ] **Step 5: Run component tests and `svelte-check`; expected: PASS.**
- [ ] **Step 6: Commit: `feat: add troubleshooting registry manager`.**

### Task 6: Integrate Registries and Fix Troubleshooting Page Defects

**Files:**
- Modify: `frontend/src/routes/troubleshooting/+page.svelte`
- Modify: `frontend/src/lib/components/TroubleshootingDialog.svelte`
- Modify: `frontend/src/lib/components/TroubleshootingDetail.svelte`
- Modify: `frontend/src/lib/components/DevicePicker.svelte` only for optional page-local comfortable behavior already supported by props; preserve other routes.
- Modify: `frontend/tests/unit/troubleshooting.spec.ts`
- Create: `frontend/tests/e2e/troubleshooting.spec.ts`

**Interfaces:**
- Consumes registry client and picker/dialog components from Tasks 3–5.
- Page owns checklist state by guide ID and stable step drafts by client IDs.

- [ ] **Step 1: Write failing unit helpers/tests for per-guide checklist state updates/reset and stable draft-step identity across `moveStep(1, -1)`.**
- [ ] **Step 2: Integrate registry loading with retryable metadata error feedback that leaves guide content readable. Add **Manage Lists** beside New Guide.**
- [ ] **Step 3: Replace Category/Severity controls in the guide dialog with registry pickers; add visible Device and field labels with 40px controls and `aria-invalid`.**
- [ ] **Step 4: Give top filters visible Category and Device titles and equal 40px controls. Keep DevicePicker changes scoped through its existing `compact`/`comfortable` props.**
- [ ] **Step 5: Lift checklist state to `+page.svelte`, keyed by guide ID, and pass controlled state/actions into `TroubleshootingDetail.svelte`.**
- [ ] **Step 6: Replace `steps`' index key with stable client IDs and move/reorder drafts as complete records.**
- [ ] **Step 7: Fetch all sibling guides through paging or the Task 1 related endpoint; centralize registry label/tone display; fix the duplicate plus and singular/plural wording.**
- [ ] **Step 8: Run unit tests, `svelte-check`, ESLint, and Playwright troubleshooting spec; expected: all focused checks pass.**
- [ ] **Step 9: Commit: `feat: improve troubleshooting registry workflow`.**

### Task 7: Full Verification and Browser UX Audit

**Files:**
- Modify: `docs/superpowers/specs/2026-11-19-troubleshooting-registry-design.md` only if implementation reveals a necessary clarified rule
- Modify: no product files unless a verification failure identifies a defect covered by the spec

- [ ] **Step 1: Run backend checks from `backend/`: `go test ./internal/... ./tests/api/...` and `go vet ./...`; expected: PASS/skip only database-gated cases with no `TEST_DATABASE_URL`.**
- [ ] **Step 2: Run frontend checks: `pnpm --filter master-control-functions-frontend check`, `pnpm --filter master-control-functions-frontend exec eslint .`, and `pnpm --filter master-control-functions-frontend exec vitest run tests/unit tests/components`; expected: 0 errors/warnings and all tests pass.**
- [ ] **Step 3: Build the frontend and backend, then start the local server. Verify schema version `10`, `/health`, registry endpoints, and guide CRUD smoke requests.**
- [ ] **Step 4: Use Svelte MCP `list-sections`, relevant `get-documentation` sections, and `svelte-autofixer` on every modified Svelte component until no issues remain.**
- [ ] **Step 5: In the browser, verify desktop/mobile layouts, visible dropdown titles, 40px alignment, inline category/severity creation, rename, reorder, safe reassignment deletion, checklist persistence, keyboard focus, and no horizontal overflow.**
- [ ] **Step 6: Check browser console and Lighthouse accessibility; expected: no application console errors and no critical accessibility findings.**
- [ ] **Step 7: Run `git diff --check` and review the full diff for scope, generated-file consistency, and migration append-only policy.**
- [ ] **Step 8: Commit remaining verification fixes separately, then request a final whole-branch review before merge.**
