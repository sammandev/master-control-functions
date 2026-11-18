# 03: Workbook import — upsert + normalization + change log

**What to build:** A maintainer uploads the knowledge workbook and the app upserts all six tables by natural keys, keeping web-only edits on untouched rows, applying deterministic Normalizations (device-scope casing upper-cased, `text_execution` → `test_execution`), and recording every normalization in a viewable change log. Re-importing the same file produces no duplicates. A reset action truncates and re-imports cleanly. Orphan rows are stored, never dropped (ADR-0003), setting up ticket 05.

**Blocked by:** 02 (Dictionary CRUD).

**Status:** done

- [x] Import endpoint accepts an xlsx upload and populates all six tables by natural keys (Function: source+interface+name; definitions: name; links/examples: their full natural keys)
- [x] Idempotent: importing the same workbook twice leaves row counts and content stable
- [x] Web-edited rows not present in the workbook survive re-import (upsert, not replace)
- [x] Normalizations applied deterministically and recorded in the change-log table (source value → normalized value, row, rule)
- [x] Change-log UI view lists recent import runs and their normalizations
- [x] Reset-from-workbook action clears and re-imports in one step
- [x] Integration tests with the real source workbook (copied into fixtures) plus a synthetic edge-case workbook: duplicate rows, orphans, bad enum values — orphans stored, counts verified

## Comments

**2026-10-01 � implementation notes**

- Migration 0003: `source_file` added to both link tables (workbook column the 0002 schema missed); `examples` recreated with UNIQUE(function_source, function_name, device_name, script) � the natural key ADR-0004's upsert needs; `import_runs` + `import_normalizations` change-log tables.
- Import pipeline: `internal/importer` parses the six sheets with excelize (pure), then upserts everything in ONE transaction (parse errors ? 422 before any DB write; DB errors ? full rollback). Upserts are `INSERT � ON CONFLICT (natural key) DO UPDATE` � workbook wins on match, web-only rows survive. `created_at` is never touched by upserts.
- Normalizations (deterministic, all logged): device-scope casing ? UPPER, `text_execution` ? `test_execution`, plus unparseable number values (`convention`/`position`) are dropped WITH an audit record instead of silently. Empty-name rows are recorded as skipped with their sheet row number.
- Key-less rows (missing source/interface/name) are skipped and reported; orphans (links to absent functions) are stored, never dropped � verified via usage_count assertions.
- API: `POST /api/import` + `POST /api/import/reset` (multipart 'file'), `GET /api/import/runs`, `GET /api/import/runs/:id`. JSON shapes mapped to snake_case for the frontend (store models carry no tags).
- UI: `/import` page � upload, Reset & import (confirm), per-run counts/skipped, expandable normalization log; nav link.
- Real-workbook fixture committed to importer testdata (49KB): parse counts verified 76/58/27/241/104/29 against the ticket-00 analysis AND against raw sheet XML; synthetic edge workbook covers orphan links, duplicate rows (collapse to one), mixed device casing, phase typo.
- Two-axis review fixes: JSON tags on Normalization/SkippedRow + normalization response mapping (frontend rendered undefined before), silent parse-failure swallowing now audited, dead double-parse removed, test helper sheet bug fixed, content-stability assertion added across re-imports (natural-key row compared minus timestamps).
