# Troubleshooting Playbook & AI Knowledge Base Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a full-stack Troubleshooting Playbook page (`/troubleshooting`) with step-by-step interactive resolution guides, multi-solution grouping per error code/problem, guided creation/editing, and AI-ready database views (`mcp_v1_troubleshooting_guides`, `mcp_v1_troubleshooting_steps`).

**Architecture:** Add PostgreSQL migration `0009_troubleshooting.sql` (plus legacy SQLite migration `0007_troubleshooting.sql` for parity) with seeded MCV2 troubleshooting guides and `mcp_v1_*` views. Implement sqlc store queries and Fiber REST handlers at `/api/troubleshooting`. Build the Svelte 5 master-detail page at `frontend/src/routes/troubleshooting/+page.svelte` with `TroubleshootingDetail.svelte` and `TroubleshootingDialog.svelte` using existing Shadcn Svelte and Bits UI components.

**Tech Stack:** Go 1.24, Fiber v3, sqlc, PostgreSQL, Svelte 5, SvelteKit, Bits UI (`^2.21.0`), Tailwind CSS v4, Vitest.

**Spec:** `docs/superpowers/specs/2026-11-19-troubleshooting-guide-design.md`

## Global Constraints

- Do not access or run migrations against the approved Ubuntu production host; database-backed Go integration tests remain gated behind `TEST_DATABASE_URL`.
- All UI components must use existing Shadcn Svelte / Bits UI wrappers in `#lib/components/ui/*` (`Button`, `Badge`, `Input`, `Textarea`, `Select`, `Dialog`, `AlertDialog`, `Checkbox`, `Collapsible`, `Pagination`, `Card`) and semantic CSS tokens (`bg-card`, `bg-popover`, `border-border`, `text-foreground`, `text-muted-foreground`, `focus-visible:ring-ring`).
- Multiple guides may share the same `error_code`, `function_name`, or `device_name` (no UNIQUE constraint on `error_code` or `title`); the UI must group/link sibling solutions that share an `error_code` or `function_name` and allow creating an alternate solution pre-filled with the current problem's metadata.
- Device names must be normalized (trimmed and uppercased; `""` = Global) and validated using `normalizeDeviceName` and `validateDeviceName`.

## Review Focus

- Creating multiple guides with the exact same `error_code` and `function_name` must succeed and surface both guides as sibling solutions in the detail view.
- Submitting a guide with 0 steps, blank `title`, blank `symptom`, or a step missing `title` or `instruction` must return structured field validation errors (`400 Bad Request`) without mutating existing steps on `PUT`.
- Reordering steps (Move Up / Move Down) in `TroubleshootingDialog.svelte` and saving via `PUT /api/troubleshooting/:id` must persist the new 1-based `position` sequence.
- Checking off steps in `TroubleshootingDetail.svelte` must track progress per guide ID and reset cleanly when clicking "Reset progress".
- Clicking "Copy AI Context" must format the guide (and any sibling solution summaries for the same `error_code`) into a clean Markdown/JSON diagnostic block and copy it to the clipboard with a toast notification.

---

### Task 1: Database Schema, Seed Data, and Store Queries

**Files:**
- Create: `backend/internal/db/postgres_migrations/0009_troubleshooting.sql`
- Create: `backend/internal/db/migrations/0007_troubleshooting.sql`
- Create: `backend/internal/store/queries/troubleshooting.sql`
- Create: `backend/internal/store/troubleshooting.sql.go`
- Modify: `backend/internal/store/models.go`

**Interfaces:**
- Produces:
  - Tables `troubleshooting_guides` and `troubleshooting_steps`
  - Views `mcp_v1_troubleshooting_guides` and `mcp_v1_troubleshooting_steps`
  - `store.TroubleshootingGuide`, `store.TroubleshootingStep`
  - `(*Queries).InsertTroubleshootingGuide`, `(*Queries).UpdateTroubleshootingGuideMeta`, `(*Queries).DeleteTroubleshootingGuide`, `(*Queries).GetTroubleshootingGuide`, `(*Queries).CountTroubleshootingGuides`, `(*Queries).ListTroubleshootingGuides`, `(*Queries).InsertTroubleshootingStep`, `(*Queries).DeleteTroubleshootingSteps`, `(*Queries).ListTroubleshootingSteps`

- [x] **Step 1: Create `backend/internal/db/postgres_migrations/0009_troubleshooting.sql` and `backend/internal/db/migrations/0007_troubleshooting.sql`**
- [x] **Step 2: Write `backend/internal/store/queries/troubleshooting.sql` and generate/implement `backend/internal/store/troubleshooting.sql.go` + model structs in `backend/internal/store/models.go`**
- [x] **Step 3: Verify Go compilation**

Run: `go test ./internal/db/... ./internal/store/...` (in `backend/`)
Expected: Compiles and passes with 0 errors.

---

### Task 2: Backend REST API & Validation Tests (`/api/troubleshooting`)

**Files:**
- Create: `backend/internal/api/troubleshooting.go`
- Create: `backend/internal/api/troubleshooting_validation_test.go`
- Create: `backend/tests/api/troubleshooting_test.go`
- Modify: `backend/internal/api/server.go:50-65`

**Interfaces:**
- Consumes: `store.Queries` troubleshooting methods from Task 1, `normalizeDeviceName` and `validateDeviceName` from `backend/internal/api/devices.go`.
- Produces:
  - `registerTroubleshootingRoutes(app *fiber.App, q *store.Queries, d *sql.DB)` wired in `api.New`
  - `GET /api/troubleshooting`
  - `GET /api/troubleshooting/:id`
  - `POST /api/troubleshooting`
  - `PUT /api/troubleshooting/:id`
  - `DELETE /api/troubleshooting/:id`

- [x] **Step 1: Write unit validation tests in `backend/internal/api/troubleshooting_validation_test.go` and HTTP integration tests in `backend/tests/api/troubleshooting_test.go`**
- [x] **Step 2: Implement `backend/internal/api/troubleshooting.go` and register routes in `backend/internal/api/server.go`**
- [x] **Step 3: Run backend unit tests**

---

### Task 3: Frontend Troubleshooting Client, Multi-Solution Helper, and Unit Tests

**Files:**
- Create: `frontend/src/lib/troubleshooting.ts`
- Create: `frontend/tests/unit/troubleshooting.spec.ts`

**Interfaces:**
- Produces:
  - Types: `TroubleshootingCategory`, `TroubleshootingSeverity`, `TroubleshootingStep`, `TroubleshootingGuideDetail`, `TroubleshootingGuideListItem`, `TroubleshootingListResult`, `TroubleshootingGuideInput`, `TroubleshootingStepInput`
  - Constants: `TROUBLESHOOTING_CATEGORIES`, `TROUBLESHOOTING_SEVERITIES`
  - Functions: `listTroubleshootingGuides`, `getTroubleshootingGuide`, `createTroubleshootingGuide`, `updateTroubleshootingGuide`, `deleteTroubleshootingGuide`, `findRelatedGuides`, `formatTroubleshootingAiContext`

- [x] **Step 1: Write failing unit tests in `frontend/tests/unit/troubleshooting.spec.ts`**
- [x] **Step 2: Implement `frontend/src/lib/troubleshooting.ts`**
- [x] **Step 3: Run unit tests**

---

### Task 4: Troubleshooting UI Components & Route (`/troubleshooting`)

**Files:**
- Create: `frontend/src/lib/components/TroubleshootingDetail.svelte`
- Create: `frontend/src/lib/components/TroubleshootingDialog.svelte`
- Create: `frontend/src/routes/troubleshooting/+page.svelte`
- Modify: `frontend/src/routes/+layout.svelte:77-86`

**Interfaces:**
- Consumes: `frontend/src/lib/troubleshooting.ts` from Task 3, `#lib/components/DevicePicker.svelte`, and `#lib/components/ui/*` (`Button`, `Badge`, `Input`, `Textarea`, `Select`, `Dialog`, `AlertDialog`, `Checkbox`, `Collapsible`, `Pagination`, `Label`).

- [x] **Step 1: Create `frontend/src/lib/components/TroubleshootingDetail.svelte`**
- [x] **Step 2: Create `frontend/src/lib/components/TroubleshootingDialog.svelte`**
- [x] **Step 3: Create `frontend/src/routes/troubleshooting/+page.svelte` and add `/troubleshooting` to `navItems` in `frontend/src/routes/+layout.svelte`**
- [x] **Step 4: Run `svelte-check`, `eslint`, and unit tests**

Run:
- `pnpm --filter master-control-functions-frontend check`
- `pnpm --filter master-control-functions-frontend exec eslint .`
- `pnpm --filter master-control-functions-frontend exec vitest run tests/unit`
Expected: 0 errors and 0 warnings across all checks; all unit tests pass.
