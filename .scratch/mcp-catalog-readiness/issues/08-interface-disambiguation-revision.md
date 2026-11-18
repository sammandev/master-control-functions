# 08: Explicit interface on links and examples; catalog revision

**What to build:** Links and examples identify their Function only by (source, name). Eleven names are shared across interfaces, for example BUILT-IN `OPEN`/`CLOSE` across SSH, TELNET, RS232 and CMD_CONSOLE. Add an explicit `function_interface` column filled on import where the name is unambiguous. Ambiguous rows are resolved in the UI, and the MCP filters on the new column. Also expose a catalog revision, such as the latest `updated_at` or a counter, so the MCP can cache snapshots.

**Blocked by:** 02

**Status:** needs-triage

- [ ] Workbook import/export contract for the interface column decided (header or inference)
- [ ] Migration, UI to resolve ambiguous rows, `mcp_v2_*` views
- [ ] MCP filters links and examples by interface
- [ ] Revision marker exposed and consumed by MCP caching
