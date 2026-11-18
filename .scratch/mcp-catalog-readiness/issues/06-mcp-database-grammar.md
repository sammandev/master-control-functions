# 06: MCP reads grammar from the database

**What to build:** The MCP's PostgreSQL reader detects `mcp_v1_functions`. When the views are present, it reads all tables through them and takes the grammar profiles and named-key allowlists from the catalog instead of `grammar.rs`. Otherwise it falls back to the built-in profiles.

**Blocked by:** 02

**Status:** resolved

- [x] `AuditedGrammar` on `Function`, skipped in tool output and schemas
- [x] `profile`, `bundle_profile` and the key checks use the catalog when it provides grammar; the convention checks stay in code
- [x] Unit tests plus a PostgreSQL fixture test; 148/148 MCP tests pass, and clippy is clean
- [x] Live check: deleting and restoring `WAITSTR`'s grammar through the web API toggled MCP support
- [x] MCP ADR-0006 and the README section
