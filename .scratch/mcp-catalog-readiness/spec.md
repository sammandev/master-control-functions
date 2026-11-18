# Spec: MCP Catalog Readiness

Status: in progress (01–06 resolved; 07–08 need user-run steps or future work)

Make the PostgreSQL catalog the single, reviewable source of truth for `script-audit-mcp` (in `MCP_Server/mcp-program-functions`). Governed by [ADR-0012](../../docs/adr/0012-mcp-catalog-contract.md) and MCP ADR-0006.

## Problem Statement

The MCP composes and validates script lines from this catalog, but whether a Function was usable depended on grammar profiles and key allowlists hard-coded in the MCP. Reviewers could not see or change readiness in the web app. Definitions could be saved with patterns or constraints the MCP cannot interpret, examples were never checked against the line rules, and the MCP used a different database login than the web app.

## Solution

Web-native grammar and key-audit tables keyed by the Function natural key, versioned read-only `mcp_v1_*` views, a readiness engine that mirrors the MCP's rules, a **MCP readiness** tab, stricter definition validation on save, non-blocking example warnings, a read-only health report, and an MCP reader that takes its grammar from the views.

## Results

The source workbook was imported into a disposable local PostgreSQL:

- **Web readiness:** 1 Function is `ready` (BUILT-IN SSH `WAITSTR`) and 295 are `not_ready`.
- **Parity with the MCP:** `list_functions` matched the web readiness for all 296 Functions, with 0 mismatches.
- **Live grammar:** deleting and restoring a grammar row through the API toggled MCP support with no rebuild.
- **Health report:**

  | Finding | Count |
  |---|---|
  | Ambiguous (source, name) pairs | 11 |
  | Definitions outside the vocabulary | 4 |
  | Seeded grammar/audit rows for Functions absent from this workbook | 54 |
  | Encoding/convention mismatch (SSH `CLOSE` is convention 2 but audited raw) | 1 |
  | Audited-Function definitions missing `value_pattern` | 5 |

## Issues

| # | Title | Status |
|---|---|---|
| 01 | [Grammar and key-audit tables](issues/01-grammar-key-audit-tables.md) | resolved |
| 02 | [Versioned `mcp_v1_*` views](issues/02-mcp-v1-views.md) | resolved |
| 03 | [Readiness engine and API](issues/03-readiness-engine-api.md) | resolved |
| 04 | [MCP readiness tab](issues/04-readiness-tab.md) | resolved |
| 05 | [Definition and example checks](issues/05-definition-example-checks.md) | resolved |
| 06 | [MCP reads grammar from the database](issues/06-mcp-database-grammar.md) | resolved |
| 07 | [Health report and read-only MCP role](issues/07-health-report-role.md) | ready-for-human |
| 08 | [Explicit interface on links and examples; catalog revision](issues/08-interface-disambiguation-revision.md) | needs-triage |
