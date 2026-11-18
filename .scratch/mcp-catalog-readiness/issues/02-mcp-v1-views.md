# 02: Versioned `mcp_v1_*` views

**What to build:** Add read-only views that make up the MCP contract. `mcp_v1_functions` gains `grammar_encoding` and `grammar_scope`. The definition, link and example views add `function_interface`, inferred only when (source, name) is unambiguous. `mcp_v1_script_key_audits` lists the key audits. A breaking change requires new `mcp_v2_*` views alongside the old ones.

**Blocked by:** 01

**Status:** resolved

- [x] Seven `mcp_v1_*` views in migration 0008
- [x] Ambiguous or orphaned rows report `function_interface = ''`
- [x] Integration test covering inference only when the name is unambiguous
