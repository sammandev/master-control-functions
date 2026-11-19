# Troubleshooting Category & Severity Registry Design

## Intent

Extend the Troubleshooting Playbook so categories and severities are managed, searchable vocabulary rather than frontend/backend constants. Users can select an existing value, create a missing value inline, and manage labels without losing guide references. At the same time, improve the troubleshooting page to follow the application's established shadcn-svelte / Bits UI patterns and fix the identified usability defects.

## Goals

- Provide more built-in troubleshooting stages/categories and severity levels.
- Let users create a category or severity while searching in the guide form.
- Provide managed Category and Severity tabs for rename, description, colour, and ordering.
- Preserve existing guides when a custom registry value is renamed.
- Reassign affected guides before deleting a custom registry value.
- Make Category and Device filters equally sized, clearly labelled, and easy to scan.
- Fix the identified troubleshooting UI state and data-loading bugs.
- Keep the feature compatible with PostgreSQL migrations and the existing API/client structure.

## Non-goals

- No user roles or per-user registry permissions are added.
- No bulk import/export screen for registry values is added.
- No historical audit log is added.
- Registry values do not get independent landing pages.
- This work does not redesign the rest of the application.

## Constraints

- PostgreSQL is production; approved host restrictions remain in `AGENTS.md` and `docs/postgresql-migration.md`.
- Add new PostgreSQL migration `0010_troubleshooting_registry.sql`; never edit applied migration `0009_troubleshooting.sql`.
- Keep legacy SQLite migration/history compatible only as required by existing migration tests; runtime uses PostgreSQL.
- Existing category and severity keys (`connection`, `execution`, `measurement`, `production`, `critical`, `warning`, `info`) remain stable.
- Use stable generated keys. Editing a label must not change the key or break guide references.
- Registry labels are unique case-insensitively, trimmed, 1–64 characters, and non-empty.
- Descriptions are optional and limited to 240 characters.
- A built-in registry row cannot be deleted.
- Deleting a custom row requires a different replacement row in the same transaction.
- Deletion reassigns every affected guide before deleting the custom row.
- Severity tone is one of `destructive`, `warn`, `default`, `outline`.
- All controls must use existing shadcn-svelte/Bits UI wrappers, semantic tokens, accessible names, and focus states.
- No raw manual z-index for overlay components and no custom styling that conflicts with app semantic tokens.
- Category and Device controls use 40px trigger height in this page; dropdown option rows remain comfortably tappable.

## Approved Approach

Combine managed registries with inline creation.

The alternatives considered were:

1. **Free text only:** smallest implementation, but causes spelling drift and provides no rename/delete management.
2. **Managed lists only:** consistent, but creating a missing option interrupts guide entry.
3. **Combined managed registries plus inline create:** slightly more work, but supports fast entry, consistent vocabulary, safe rename, and safe deletion.

The combined approach is approved. Deletion uses mandatory reassignment, never silent deletion or orphaned guides.

## Data Model

Create `troubleshooting_categories`:

- `key TEXT PRIMARY KEY`
- `label TEXT NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `sort_order INTEGER NOT NULL`
- `built_in BOOLEAN NOT NULL DEFAULT false`
- unique case-insensitive label index

Create `troubleshooting_severities`:

- `key TEXT PRIMARY KEY`
- `label TEXT NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `tone TEXT NOT NULL CHECK (tone IN ('destructive', 'warn', 'default', 'outline'))`
- `sort_order INTEGER NOT NULL`
- `built_in BOOLEAN NOT NULL DEFAULT false`
- unique case-insensitive label index

`troubleshooting_guides.category` and `troubleshooting_guides.severity` remain text keys and gain foreign keys to the registry tables. Remove the old fixed `CHECK` constraints in migration `0010` before adding foreign keys.

Migration `0010` must seed all registry rows in one migration and migrate constraints without changing guide contents.

### Built-in Categories

1. DUT Connection & Setup (`connection`)
2. Script & Flow Execution (`execution`)
3. Measurement & Criteria (`measurement`)
4. Production Run & Handoff (`production`)
5. Fixture & Hardware (`fixture-hardware`)
6. Instrument & Calibration (`instrument-calibration`)
7. DUT Firmware & Image (`dut-firmware-image`)
8. Network / Shopfloor / MES (`network-shopfloor-mes`)
9. INI & Spec Configuration (`ini-spec-configuration`)
10. Station Environment (`station-environment`)
11. Logging & Reporting (`logging-reporting`)

### Built-in Severities

1. Critical (Blocks Flow) (`critical`, `destructive`)
2. High (`high`, `destructive`)
3. Warning (Intermittent / Margin) (`warning`, `warn`)
4. Low (`low`, `outline`)
5. Info (Best Practice) (`info`, `default`)

The migration retains the existing labels/descriptions for current values where the seed currently documents them.

## Backend API

Add registry store queries and API routes below the existing troubleshooting routes:

- `GET /api/troubleshooting/categories`
- `POST /api/troubleshooting/categories`
- `PUT /api/troubleshooting/categories/:key`
- `DELETE /api/troubleshooting/categories/:key?reassign_to=:key`
- `GET /api/troubleshooting/severities`
- `POST /api/troubleshooting/severities`
- `PUT /api/troubleshooting/severities/:key`
- `DELETE /api/troubleshooting/severities/:key?reassign_to=:key`

List responses return rows ordered by `sort_order, label` and include `guide_count`.

Create derives a lowercase kebab-case key from the label, with a numeric suffix on collision. API responses always return the created key. Built-in updates allow label, description, tone, and order changes but never key or built-in status changes. Inline creation never controls `sort_order`; the backend appends it.

Create and update return field validation errors for blank labels, duplicate labels, length violations, invalid severity tones, and invalid order values. Deletion is transactional: validate replacement row, reject built-in deletion, reject missing/different replacement, update all guides, then delete the registry row. Deletion of an unused custom row still requires an explicit replacement to keep the workflow uniform.

Guide creation/update no longer uses hard-coded maps. It verifies that submitted category and severity keys exist in their registries.

## Frontend Client & State

Change troubleshooting type fields from closed unions to `string` keys. Registry item types contain key, label, description, sort order, built-in status, guide count, and severity tone where applicable.

The client exports typed list/create/update/delete functions for both registries. Guide list filtering remains key-based. Category/severity labels and severity badge variants come from loaded registry metadata, with neutral fallback rendering for unknown or failed metadata.

The page loads registry metadata alongside guides. A metadata failure shows a retryable error but must not make already-loaded guide data unreadable.

## Guide Form UX

Add a reusable searchable registry Select-like control using the established Bits UI combobox/popover composition.

- It displays selected label in a 40px trigger.
- It filters locally and exposes a clear accessible search input.
- If no exact label/key match exists, it shows `Create "<search>"`.
- Category creation immediately creates and selects the category.
- Severity creation immediately creates a default tone `outline`, then selects the value; the user can change tone later in Manage Lists.
- The control shows a short inline error and keeps typed text on create failure.

Use the app's `Label` wrappers and structured error state in the troubleshooting dialog. Field controls must expose `aria-invalid` and field errors accessibly.

Visible labels in the dialog are:

- Testing Stage / Category
- Severity
- Device / Interface Scope

The Category and Device filter controls also have visible titles and 40px triggers. The Device picker itself remains shared; the troubleshooting page supplies page-local comfortable sizing rather than changing other routes globally.

## Manage Lists Dialog

Add **Manage Lists** next to **New Troubleshooting Guide**. Use a Dialog with Tabs for Categories and Severities.

Each tab lists registry rows with:

- drag-free Move Up / Move Down controls (order updates explicitly)
- label input
- description input
- severity tone selector for Severity only
- usage count
- Save action
- Delete action for custom rows

Delete opens an AlertDialog naming the value and usage count. It requires selection of a replacement from the remaining rows. Confirming runs the transactional reassignment and deletion. Built-in rows show no delete action.

A row-level save is preferred over one global unsaved form. This limits validation scope and reduces accidental unrelated changes.

## Troubleshooting Page UX Improvements

- Keep filters in one labelled region with matching 40px controls.
- Preserve selected guide when refreshing filters where that guide is still present; select the first result otherwise.
- Load all sibling guides, not only the first 100. Use paged requests until exhausted or a dedicated related-guide endpoint if page size is globally constrained.
- Keep checklist completion in page-level state keyed by guide ID, not component-local state.
- Give step editor rows stable client IDs; move/reorder the IDs with their fields so inputs never mismatch.
- Remove the duplicated plus from the alternate-solution action.
- Use singular/plural wording for step counts.
- Centralize severity label/tone rendering from registry data.
- Keep the master-detail layout and current app visual language; replace custom empty/error markup only where existing app components make the change straightforward.
- Use toast success/error feedback for registry operations and copy actions.

## Testing

Backend tests prove:

- current fixed categories/severities become valid registry rows after migration.
- invalid or unknown guide category/severity keys return structured `400` errors.
- case-insensitive duplicate labels are rejected.
- inline creation produces a stable key and appends ordering.
- built-in values cannot be deleted.
- custom delete without replacement is rejected.
- custom delete with replacement updates every affected guide and deletes the registry row in one transaction.

Frontend unit tests prove:

- registry labels render from loaded metadata.
- unknown keys have safe fallbacks.
- a no-match search exposes inline create and selects the created row.
- stable step IDs retain values during reordering.
- checklist state survives guide switches and reset affects only the current guide.
- severity badges use registry tones.

Focused Svelte/UI checks prove:

- `svelte-check` reports 0 errors and 0 warnings.
- ESLint passes.
- Vitest unit tests pass.
- Production build passes.
- Browser console has no application errors at `/troubleshooting`.
- At 1280px and mobile width, the filter controls and Manage Lists dialog have no horizontal overflow.

## Migration & Rollback

Runtime migration is additive and registry seeding is idempotent only in the normal one-time migration framework; the migration file must not be rerun manually. The old fixed constraints are removed and foreign keys are installed in the same transaction by the migration runner.

Rollback is database rollback only under the approved `docs/postgresql-migration.md` process. No production rollback is executed by the agent. Application rollback requires restoring the pre-`0010` application/schema pair together because the updated frontend expects registry endpoints.

## Acceptance Criteria

- All current guides continue to load after migration `0010` with unchanged keys, labels, and steps.
- Users can create a missing category and severity from the guide form without leaving the dialog.
- Users can rename, describe, recolour, and reorder registry rows from Manage Lists.
- A custom value used by guides cannot be deleted until a different replacement is chosen; deletion moves all affected guides first.
- Built-in rows cannot be deleted.
- Category, Severity, and Device controls are visibly labelled and Category/Device filters are 40px tall.
- The troubleshooting page has no known state loss, step reordering, sibling loading, or wording defects listed above.
- Automated backend/frontend checks pass, and local browser smoke checks show no console errors or horizontal overflow.
