# 08: Export — six-sheet round-trip

**What to build:** A maintainer downloads the full knowledge workbook regenerated from the database: six sheets in the source file's original column layout, byte-shape-compatible, re-importable without loss (ADR-0004). The round-trip is proven by a determinism test: import the source workbook → export → compare cell-by-cell.

**Blocked by:** 03 (Workbook import).

**Status:** done

- [x] Export endpoint produces the six-sheet xlsx in the original column layout with original sheet names and header rows
- [x] Orphan rows are exported faithfully (they exist in the source file too) rather than dropped
- [x] Normalized values (casing, phase typo mapping) are what get exported — export reflects database state
- [x] UI: export button (download) on the admin/import area
- [x] Round-trip determinism integration test: import fixture workbook → export → cell-by-cell comparison passes
- [x] Re-importing the exported file is idempotent (no new rows, no changed rows)

Annotations:
- "Cell-by-cell": header rows compared raw via excelize GetRows; data rows compared field-by-field after ParseWorkbook (parsed source state IS the expected DB state — skipped rows are not DB state, so they are excluded on both sides by design).
- "Byte-shape-compatible" operationalized as the determinism + idempotency tests: fresh excelize file, all cells as strings, original number formats/styles not reproduced.
- All six sheets export in id order (ListAll* queries) = source row order on first import; after UI edits, later rows append at the end (DB order), which is inherent to regeneration from the database.
