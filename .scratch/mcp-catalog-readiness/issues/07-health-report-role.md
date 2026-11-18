# 07: Health report and read-only MCP role

**What to build:** A read-only SQL report covering:

- ambiguous names;
- orphans;
- links without definitions;
- vocabulary violations;
- grammar and audit rows without a Function;
- encoding/convention mismatches;
- missing `value_pattern`.

Also a SELECT-only MCP role on the `mcp_v1_*` views, so the MCP uses the same database as the web app.

**Blocked by:** 01, 02

**Status:** ready-for-human

- [x] `docs/sql/mcp-catalog-report.sql`, validated against the imported source workbook
- [x] `docs/sql/mcp-readonly-role.sql`, validated against a disposable database
- [ ] User: after deploying migration 0008, run the report against production and triage the findings, including SSH `CLOSE` being convention 2 while it is audited raw
- [ ] User: create the role and point the MCP's `SCRIPT_AUDIT_DATABASE_URL` at the web database; it currently uses `script_forge`/`program_functions`
- [ ] User: rerun the 22-case MCP reference matrix against the aligned database
