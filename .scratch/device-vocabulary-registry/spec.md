# Device vocabulary, device registry, and device page flow

Status: done

## Completion Record

Status reconciled on 2026-10-15 with all eight implementation tickets marked done. Implementation and commit references remain in the individual [tickets](issues/), including the completed [tracker/docs hygiene ticket](issues/08-tracker-docs-hygiene.md).

The requirements below preserve the original feature record. This status update does not claim a new database-backed test run, database access, or deployment verification; those checks remain operator-run off-machine under the current repository restrictions.

## Problem Statement

The app calls the device dimension "scope" everywhere — nav item, page titles, column heads, URLs (`/scopes`), and API paths (`/api/scopes`) — but users think in devices. The forced translation between the UI's vocabulary and the user's vocabulary makes the device browser feel foreign, and the read-only device pages dead-end: the active tab is invisible in the URL, nothing links back from a Function to the devices it serves, and a brand-new device cannot be reserved — it only exists once some link, definition, or Example row references its name, so "create a device, then populate it" is impossible.

## Solution

Rename the user-facing vocabulary from "Scope" to "Device" across the entire UI, routes, and API (`/devices`, `/api/devices`), keeping "(Global)" for the blank value and "device scope" only as the internal row-attribute concept. Add a web-native devices registry (Sequences precedent, ADR-0005) so a Device can be created and exist — empty — before it has content: creation normalizes casing to uppercase, rejects duplicates, and registered-but-empty devices survive import/reset. The device list shows registry ∪ derived names. The device page becomes a proper navigation hub: deep-linkable tabs, clickable count columns from the list, definition-creation shortcuts (device prefilled), and a function→device back-path on Function pages. Registered devices with zero references can be deleted. Paper trail already in place: ADR-0010 and the updated CONTEXT.md glossary.

## User Stories

1. As a knowledge-worker, I want the navigation item and pages to say "Devices", so that I never have to translate "scope" into "device" in my head.
2. As a knowledge-worker, I want device pages to live under `/devices`, so that URLs match what I see on screen.
3. As a knowledge-worker, I want the devices API under `/api/devices`, so that the API surface uses the same vocabulary as the UI.
4. As a knowledge-worker, I want the blank device value to keep rendering as "(Global)", so that shared definitions stay recognizable as Global.
5. As a knowledge-worker, I want form labels and column heads across the app to say "Device" (not "Device scope"), so that every screen uses one consistent word for the same attribute.
6. As a knowledge-worker, I want the Orphans page, Examples page, Links panel, Compose dialog, and Sequence editor to use the new vocabulary, so that no screen still speaks the old one.
7. As a knowledge-worker, I want a "Create device" action on the devices page, so that I can reserve a new device name before any data exists for it.
8. As a knowledge-worker, I want creation to normalize casing to uppercase automatically, so that `Osp` and `OSP` can never become two different devices.
9. As a knowledge-worker, I want a clear error when creating a device whose name already exists (registered or derived), so that I don't accidentally think I created a second device.
10. As a knowledge-worker, I want created devices to survive workbook import/reset, so that my registered devices are not wiped when data is re-imported.
11. As a knowledge-worker, I want the devices list to include both registered and data-derived devices, so that imported/legacy data still appears without re-registration.
12. As a knowledge-worker, I want each device page to show Functions, Parameters, Criteria, and Examples tabs, so that everything belonging to a device is in one place.
13. As a knowledge-worker, I want the active tab reflected in the URL, so that I can share or bookmark a link straight to a device's parameters.
14. As a knowledge-worker, I want browser back/forward to work across tab switches, so that navigation behaves the way the rest of the web does.
15. As a knowledge-worker, I want the count columns on the devices list to be clickable, so that I can jump straight to a device's Parameters tab, for example, without clicking through.
16. As a knowledge-worker, I want a path from a Function page to the devices its links serve, so that I can answer "which devices use this function?" in one click.
17. As a knowledge-worker, I want to create a parameter or criteria definition directly from a device page with the device prefilled, so that I can populate a new device without detouring to the dictionary and re-typing the name.
18. As a knowledge-worker, I want the device page's Functions tab to stay link-driven with guidance when empty, so that Function ownership follows the data model (links on Function pages) rather than a parallel creation path.
19. As a knowledge-worker, I want to delete a registered device that has no references, so that typo'd or abandoned registrations don't clutter the list.
20. As a knowledge-worker, I want deletion blocked (with a clear error) for a device that any link, definition, or Example still references, so that I can't orphan data through the devices page.
21. As a knowledge-worker, I want dictionary "Device" values to deep-link into the device browser, so that I can pivot from a definition to its device context.
22. As a knowledge-worker, I want a registered empty device to appear in the list with zero counts, so that I can see what I've reserved and what still needs populating.
23. As a knowledge-worker, I want the glossary and ADRs to record the vocabulary change, so that future readers understand why the UI says "Device" while the schema says `device_name`.

## Implementation Decisions

- **Vocabulary decision (ADR-0010)**: "Device" is the user-facing noun everywhere; "Device scope" survives only as the internal row-attribute concept; blank value = Global, rendered "(Global)".
- **Rename depth**: UI text + frontend routes (`/devices`, `/devices/[name]`) + backend API routes (`/api/devices`) + trivial identifier renames on both sides. No DB column changes (columns already `device_name`); no DB migration for the rename itself.
- **Devices registry**: new `devices` table (unique `device_name`, timestamps) via an appended migration (append-only per migration policy, ANSI-compatible SQL per ADR-0001). Web-native per the Sequences precedent (ADR-0005): not part of the workbook's six sheets, untouched by import/reset, not exported.
- **Device list = registry ∪ derived**: the devices endpoint unions registered names with distinct `device_name` values from link, definition, and Example rows; counts remain computed from data. Global (blank) is never a registry entry — not registrable, not deletable — and remains the reserved first row of the device list, as before.
- **Create semantics**: uppercase normalization before insert (ADR-0002 rule); 409-style rejection when the name already exists registered or derived.
- **Delete semantics**: allowed only when no link, definition, or Example row references the name; otherwise rejected with a clear error.
- **Population model**: parameter/criteria creation reuses the existing dictionary modal with the device prefilled; Functions remain link-driven from Function pages (Function identity is device-independent, ADR-0009).
- **Tab deep-linking**: active tab encoded in the URL query; browser history integrates with tab switches; tab counts on the list page link to the target device with the tab preselected.
- **Function→device back-path**: a Function page's links' device values become links into the device browser.
- **Deep-links**: dictionary Device values link into the device browser (existing behavior, updated URL).
- **Paper trail**: ADR-0010 (already written) and CONTEXT.md glossary update (already written) are part of this feature's record; E2E and dictionary deep-link updates ship in the same change set.

## Testing Decisions

- **Good test**: asserts external behavior through the UI or HTTP surface only — never internal function names, file layout, or SQL text. Vocabulary assertions check what the user sees (headings, labels, URLs), not source strings.
- **Seams (both existing, no new seams)**:
  - Go integration tests driving HTTP handlers against a fresh temp SQLite DB (migrations included) — registry CRUD, normalization, duplicate rejection, delete-empty rule, list = registry ∪ derived, `/api/devices` contract. Prior art: existing scopes/parameter-definition API tests (e.g. the scope-semantics update tests).
  - Playwright E2E against the real binary + workbook-seeded DB — renamed nav/pages/URLs/labels, create-device-through-UI, deep-linkable tabs, clickable counts, function→device back-path, delete empty device, dictionary deep-links. Prior art: the existing scopes browser spec (rewritten for the new vocabulary), dialogs spec, smoke spec.
- **Migration safety**: exercised implicitly — integration and E2E runs start from fresh databases and apply all migrations.
- **Rename completeness gate**: the E2E vocabulary test asserts the old wording is gone from the surfaces it drives (headings, nav, column heads, form labels it already interacts with).

## Out of Scope

- Renaming existing devices (cascading rename of links/definitions/Examples) — separate future feature.
- Foreign keys or any integrity enforcement on `device_name` (free-text model stays; Orphan rule ADR-0003 unchanged).
- A devices sheet in the workbook, or devices in Export/import round-trip (registry is web-native).
- Creating or editing Functions from the device page.
- Device descriptions/metadata beyond the name (registry stores name + timestamps only).
- Devices entity changes to the Orphan report's resolution logic (exact-scope-then-global unchanged).

## Further Notes

- ADR-0010 (`docs/adr/0010-device-vocabulary-and-registry.md`) and the CONTEXT.md glossary update are already written and considered part of this feature's done state.
- The registry table's relationship to import/reset mirrors Sequences: reset re-seeds workbook-derived tables only; `devices` rows survive.
- Casing: UI continues to display stored values verbatim; normalization applies at creation time only.
- Global (blank) is not a Device and never appears in the devices registry; it remains a reserved view (existing "(Global)" handling).
