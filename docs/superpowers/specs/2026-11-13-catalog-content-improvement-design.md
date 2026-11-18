# Catalog Content Improvement from HH5K Reference Scripts — Design

Date: 2026-11-13. Status: chosen under autopilot (user unavailable); user approval happens at import time.

## Intent
Improve the function, parameter/criteria, and example content of the production
catalog (web app at http://10.176.2.139:31627) using the HH5K program scripts,
the general test-scripts guide, and the HH5K test plan in
`MCP_Server/mcp-program-functions/.dev-reference`, so MCP clients and engineers get
accurate, evidence-backed definitions and real usage examples.

## Baseline (read-only snapshot 2026-11-13)
- 282 functions, 261 parameter defs, 70 criteria defs, 559 parameter links, 191 criteria links, 207 examples; 0 orphans.
- HH5K uses 126 distinct (source, device, function) keys; 29 are missing from the catalog;
  46 catalog functions used by HH5K have no example; 53 functions have no parameter links.

## Approaches
1. **Delta workbook + diff report (chosen).** Export the workbook, produce a six-sheet workbook containing
   only new/changed rows, and a human-readable diff. User imports it via the app (non-reset upsert).
   Reviewable before any production change; rollback = re-import the original export.
2. Live API writes. Faster, but writes production without review and is harder to roll back. Rejected.
3. Full-workbook reset import. Rewrites everything; riskier than a delta. Rejected.

## Goals
- Add missing functions that HH5K actually uses, with parameter/criteria links and examples.
- Add verbatim HH5K examples (with concrete "meaning") to used functions lacking examples.
- Correct or sharpen descriptions, parameter/criteria definitions and links where HH5K evidence contradicts or refines them.

## Non-goals
- No MCP grammar/readiness changes (needs separate audit).
- No `convention` changes on existing functions (affects MCP grammar); mismatches are reported only.
- No renaming/vocabulary normalization of category/role/phase values; no deletions.
- No direct production writes by the agent.

## Constraints
- Upsert overwrites every column: changed rows must carry all columns (unchanged values copied).
- Natural keys: function (source, interface, name); definition (name, device_name);
  link (function_source, function_name, device_name, name); example (function_source, function_name, device_name, script).
- Every change cites evidence (program/file/section/key or guide section).
- Examples are verbatim active HH5K lines; never invented.
