# Spec: Web Program Functions Database

Status: done

CRUD web application over the script-definition knowledge workbook. Go (Fiber) + Svelte 5, SQLite starter database. Governed by `CONTEXT.md` (glossary) and ADR-0001..0004.

## Problem Statement

Test engineers maintain the definitions of test-script functions — their parameters, verification criteria, and usage examples — in a six-sheet Excel workbook. The workbook is large, structurally fragile, hard to search, and full of inconsistent references: many parameter/criteria links point at functions or dictionary entries that do not exist. Editing is manual and error-prone, validation (allowed values, cardinality, required flags) exists only as text, and producing correct example scripts requires memorizing the grammar.

## Solution

A fast, single-binary web application backed by SQLite that gives every entity of the knowledge workbook first-class CRUD: functions, parameter definitions, criteria definitions, the links binding them, and examples. It imports the workbook as a repeatable idempotent upsert (never dropping data, flagging orphans instead), validates edits against the definitions' own constraints, composes example scripts through a guided builder, and exports the full six-sheet workbook for round-trip fidelity. No login in v1; internal/trusted network tool.

## User Stories

1. As a test engineer, I want to browse all functions in a searchable, sortable list, so that I can find a function without opening Excel.
2. As a test engineer, I want to filter functions by source, interface, category, execution role, and execution phase, so that I can narrow hundreds of entries to the relevant few.
3. As a test engineer, I want to open a function and see its full definition — description, interface, convention, role, phase — so that I understand what it does before using it.
4. As a test engineer, I want the function detail page to show its parameter links, criteria links, and examples as tabs, so that everything about one function is on one screen.
5. As a test engineer, I want to create, edit, and delete functions, so that the knowledge base stays current.
6. As a test engineer, I want parameter links on a function to be ordered by position, so that callers know the argument order.
7. As a test engineer, I want to mark a parameter link as required or optional, so that scripts can be validated for completeness.
8. As a test engineer, I want to add an existing parameter definition to a function by picking from a searchable dropdown, so that dictionary terms are reused, not retyped.
9. As a test engineer, I want to create a new parameter definition inline while linking it to a function, so that I am not forced to interrupt my flow to visit another page.
10. As a test engineer, I want to link criteria to a function the same way, so that verification thresholds stay attached to the functions that verify them.
11. As a test engineer, I want to detach a parameter or criteria link from a function without deleting the definition itself, so that dictionary entries outlive their use.
12. As a knowledge-base maintainer, I want dictionary pages for all parameter definitions and criteria definitions, so that I can audit vocabulary independently of functions.
13. As a knowledge-base maintainer, I want to edit a parameter definition's type, unit, default, allowed values, cardinality, value pattern, and required flag, so that constraints stay accurate.
14. As a test engineer, I want form fields for a parameter value to be validated against its allowed values at edit time, so that invalid data never enters the database.
15. As a test engineer, I want a parameter marked MULTIPLE to accept multiple values in forms, so that cardinality is enforced by the UI, not by my memory.
16. As a test engineer, I want to see which functions, links, and examples reference a parameter or criteria definition, so that I can judge the impact of editing it.
17. As a test engineer, I want orphaned links and examples visibly flagged as "unresolved" wherever they appear, so that incomplete data is impossible to miss.
18. As a knowledge-base maintainer, I want an orphan report view listing every unresolved reference, so that I can clean the knowledge base systematically.
19. As a knowledge-base maintainer, I want to create the missing function or definition directly from an orphan row, so that fixing the reference takes one action.
20. As a test engineer, I want to import the workbook at any time, so that the web database reflects the latest shared file.
21. As a knowledge-base maintainer, I want the import to be an idempotent upsert by natural keys, so that repeated imports don't create duplicates.
22. As a knowledge-base maintainer, I want rows I edited in the web UI that are untouched by the workbook to survive re-import, so that the web database is not blindly overwritten.
23. As a knowledge-base maintainer, I want import-time normalizations (e.g. casing fixes, typo mapping) recorded in a change log, so that the divergence from the raw file is auditable.
24. As a knowledge-base maintainer, I want a "reset database from workbook" action, so that I can start over cleanly after experiments.
25. As a knowledge-base maintainer, I want to export the complete six-sheet workbook in the original column layout, so that the Excel artifact can be regenerated from the web database.
26. As a test engineer, I want the export to be re-importable without data loss, so that export and import form a trustworthy round trip.
27. As a test engineer, I want a script composer on a function that generates a form from its parameter links, so that I don't have to memorize argument grammar.
28. As a test engineer, I want the composer to validate my inputs against allowed values, cardinality, and required flags before producing a script, so that generated scripts are correct by construction.
29. As a test engineer, I want the composer to emit the pipe-delimited script string (source+interface prefix, function, parameter assignments, criteria or NULL, timeout, trailing segments), so that I can copy it into test programs.
30. As a test engineer, I want a composer-produced script to be saved directly as an example of that function, so that good scripts accumulate as documentation.
31. As a test engineer, I want examples to show their function, device scope, script text, and meaning together, so that examples are self-explanatory.
32. As a test engineer, I want to create, edit, and delete examples, so that usage documentation stays current.
33. As a test engineer, I want list pages to search-as-I-type and paginate, so that navigation stays fast with hundreds of rows.
34. As a test engineer, I want the app to feel fast (sub-second page transitions, no full reloads), so that daily use is not a chore.
35. As a test engineer, I want required fields, invalid values, and missing references to produce clear inline error messages, so that I can fix input without guessing.
36. As a knowledge-base maintainer, I want deleted definitions to warn me when links still reference them, so that I don't break the knowledge base by accident.
37. As a test engineer, I want the composer's deferred multi-function sequence mode to be considered in the data model, so that building it later does not force a rewrite.
38. As a developer, I want the database layer to avoid SQLite-specific SQL, so that migrating to PostgreSQL later is a config change, not a rewrite.

## Implementation Decisions

- **Architecture**: one Go binary serving a Fiber HTTP API and the built SvelteKit static bundle; SQLite database file alongside the binary. SvelteKit uses adapter-static; all routing is client-side.
- **Data layer**: `modernc.org/sqlite` (pure Go, no cgo) with `sqlc` generating typed queries from hand-written SQL. All SQL ANSI-compatible (ADR-0001): no `AUTOINCREMENT`, no SQLite-only functions, ISO-8601 timestamps, connection-level `PRAGMA`s isolated from query code. PostgreSQL + `pgx` migration later = driver swap + sqlc engine config change.
- **Schema**: six entity tables mirroring the workbook (functions, parameter definitions, criteria definitions, parameter links, criteria links, examples) plus an import/normalization change-log table. Functions keyed by (source, interface, name). Link rows join functions by (function_source, function_name) and carry device scope as a free-text attribute (ADR-0002). Link tables carry no database-level foreign keys to functions or definitions, so orphan rows are storable (ADR-0003); API-layer validation applies to non-orphan data.
- **Orphan policy**: import keeps every row; the API flags unresolved references; the UI renders orphans as first-class records with a warning state and offers inline creation of the missing target (ADR-0003).
- **Import**: idempotent upsert by natural keys; workbook rows update matching database rows, untouched rows survive (ADR-0004). Deterministic normalizations at import (device-scope casing upper-cased, `text_execution` → `test_execution` phase mapping) recorded in the change log. A reset action truncates and re-imports.
- **Export**: full six-sheet snapshot in the source workbook's original column layout, byte-shape-compatible so it re-imports cleanly. Excel I/O via a Go xlsx library (excelize-class).
- **Script composer**: form is generated from the function's parameter links (position order, required enforcement, allowed-values validation, cardinality MULTI- vs SINGLE-value inputs). Output grammar (working assumption, DLL-sourced): `<SOURCE><INTERFACE>|<FUNCTION>|<KEY=VAL&...>|<CRITERIA or NULL>|<timeout-ms>|<label>|<flags>`. Emitted scripts can be saved as examples. Sequence builder deferred; data model does not preclude it.
- **API contract**: REST JSON. Resources: functions, parameter-definitions, criteria-definitions, parameter-links, criteria-links, examples, plus import (upload), export (download), reset, change-log, and orphans report endpoints. List endpoints accept search, enum filters, sort, pagination parameters. Validation errors return field-level messages.
- **Frontend**: Svelte 5 runes + Tailwind with an accessible component set; function-centric navigation — dashboard/search → function detail with Parameters/Criteria/Examples tabs; dictionary pages; orphan report; composer as a dialog or dedicated view. Instant search with debounced server queries.
- **Validation source of truth**: allowed values stored as a set string (`{a,b,c}`) are parsed to a set for validation; SINGLE/MULTIPLE cardinality and required flags drive form behavior; type constrains numeric parsing (int/double/list[int]/string).
- **Environment precondition**: Go toolchain must be installed (currently absent; Node 24/npm 11/Python 3.13 present). Go install must pass the software-license check first.

## Testing Decisions

- **Good test** = asserts external behavior only: HTTP request → response status/body, plus observable database state afterward. No assertion on internal call graphs or store internals.
- **Primary seam**: the HTTP API. Go integration tests spin the Fiber app against a temporary SQLite file and exercise every endpoint: CRUD round-trips, filter/sort/pagination contracts, validation rejections (field-level error shape), orphan flagging, import upsert idempotency (import twice → no duplicates), normalization change-log entries, and reset.
- **Round-trip determinism test**: import the source workbook (copied into test fixtures), export, and compare exported sheets cell-by-cell against the source. This single test pins import fidelity, export fidelity, and schema mapping together.
- **Composer tests**: through the API — generate a form spec for a known function, submit valid/invalid parameter sets, assert emitted script strings and saved examples.
- **E2E smoke (secondary)**: Playwright over three or four happy paths (search a function, edit a parameter link, compose-and-save a script, view orphan report). Thin on purpose; wiring bugs only.
- **Prior art**: none — greenfield. Standard Go testing with httptest is the pattern; fixtures are the real source workbook plus small synthetic workbooks for edge cases (duplicates, orphans, bad enum values).

## Out of Scope

- Authentication, users, roles, audit trail of UI edits
- Automatic two-way workbook↔database sync (export/import are manual actions)
- Multi-function sequence builder implementation (design-noted only, story 37)
- Filtered or partial exports beyond the full snapshot
- Multi-user concurrency beyond last-write-wins
- Deployment hardening (TLS, reverse proxy), dark mode, localization
- Migrating to PostgreSQL (only the exit path is kept clean)

## Further Notes

- Sequence builder (story 37): shipped after the spec as ticket 10 with ADR-0005; the Out-of-Scope deferral above is superseded by that decision.

- Verified workbook data facts driving the design: 25/63 parameter-links and 8/22 criteria-links reference absent functions; 58/241 linked parameter names and 23/104 linked criteria names are absent from dictionaries; 16/29 examples are orphaned; `function_main` is unique on (source, interface, name); `device_name` is case-inconsistent and is not the interface.
- The trailing-segment semantics of example scripts (criteria / timeout-ms / label / flags) are a working assumption from data inspection; composer segment order follows it and must be re-confirmed against real test-program usage before the composer ships.
- Repo state at spec time: greenfield; documentation only (`AGENTS.md`, `CONTEXT.md`, `docs/adr/0001..0004`, `docs/agents/*`).
