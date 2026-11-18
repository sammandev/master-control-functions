# 03: Readiness engine and API

**What to build:** A Go readiness engine that mirrors the MCP's `profile`, `bundle_profile`, `select_fields` and support checks. It returns `ready`, `bundle_ready` or `not_ready` with line and bundle issues, warnings and the selected fields. It is exposed through the list, per-Function, grammar and script-key routes.

**Blocked by:** 01

**Status:** resolved

- [x] `GET /api/mcp-readiness` and `GET /api/functions/:id/mcp-readiness?device=`
- [x] `GET`/`PUT`/`DELETE /api/functions/:id/grammar`, `PUT /api/functions/:id/script-keys`, `DELETE /api/functions/:id/script-keys/:kind/:name`
- [x] Warnings for ambiguous keys and for examples that fail the line rules
- [x] The source workbook gives the same results as the MCP's `list_functions` (296/296)
