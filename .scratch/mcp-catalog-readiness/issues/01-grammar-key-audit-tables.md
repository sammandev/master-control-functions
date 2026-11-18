# 01: Grammar and key-audit tables

**What to build:** Add the web-native tables `function_grammars` (encoding `raw`/`named`, scope `line`/`bundle`) and `script_key_audits` (parameter or criterion key, scope `line`/`bundle`). Both are keyed by Function natural key and require `audited_by`. Seed them with the MCP's former built-in profiles so behavior does not change.

**Blocked by:** None.

**Status:** resolved

- [x] Migration `0008_mcp_catalog.sql`: both tables are created and seeded with 12 grammars and their key audits (17 for `TX_MULTI_VERIFICATION`)
- [x] Function rename and Device rename carry the rows along in the same transaction
- [x] Import, reset and export leave both tables untouched
- [x] Integration tests: seed counts, the Function rename cascade and the Device rename cascade

## Comments

- Natural key instead of `function_id`, because a workbook reset re-creates Function rows.
- No CHECK constraints on the definition vocabulary: the source workbook already violates it (`WAVEFORM_NAME` cardinality), and even `NOT VALID` constraints break every import.
