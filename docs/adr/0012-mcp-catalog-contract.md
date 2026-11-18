# MCP Catalog Contract

Accepted on 2026-11-09. `script-audit-mcp` reads this catalog to compose and validate script lines. Its grammar profiles and script-key allowlists were hard-coded in the MCP, so the web app could not show whether a Function was usable, and every newly audited Function needed an MCP release.

Migration 0008 adds two web-native tables: `function_grammars` (encoding `raw`/`named`, scope `line`/`bundle`) and `script_key_audits` (named key per parameter or criterion, scope `line`/`bundle`). Both are keyed by the Function natural key (source, interface, name) rather than `function_id`, because a workbook reset re-creates Function rows; the sequences and devices tables follow the same precedent. Renaming a Function cascades to both tables. Every row records `audited_by`, so audits are attributable. The migration seeds exactly the MCP's former built-in lists, which keeps MCP behavior unchanged.

The MCP reads only the versioned `mcp_v1_*` views. A breaking change requires `mcp_v2_*` views alongside the old ones. Links and examples carry no interface, so the views infer `function_interface` only when (source, name) names exactly one Function; otherwise the value is `''`.

The definition vocabulary (type, cardinality, numeric `required`) has no CHECK constraints. The source workbook itself carries values outside the vocabulary: parameter definition `WAVEFORM_NAME` has a cardinality outside SINGLE/MULTIPLE. Even `NOT VALID` constraints apply to new rows, so they would make every workbook import fail. Instead, saving a definition through the API validates the allowed values, numeric ranges, the default value against them, and `value_pattern` compilation; the readiness report and `docs/sql/mcp-catalog-report.sql` flag imported rows that break the vocabulary.

Example scripts are checked on create and update against the MCP line rules, but the result is returned as `script_issues` warnings and never blocks a save. Historical examples are evidence, and rejecting them would lose information.

`GET /api/mcp-readiness` and the Function **MCP readiness** tab report one of three statuses for each Function: `ready` (line composable), `bundle_ready`, or `not_ready`, each with its issues. The Go rules mirror the MCP's `profile`, `bundle_profile`, `select_fields`, and support checks; an end-to-end check against the imported workbook matched the MCP's `list_functions` exactly.

Deferred: explicit link and example interface columns, a catalog revision marker for MCP caching, and aligning the MCP's database credentials with this database through a SELECT-only role (`docs/sql/mcp-readonly-role.sql`).
