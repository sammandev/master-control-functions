# 05: Definition and example checks

**What to build:** Saving a definition uses the MCP's allowed-value rules (sets, ranges, character classes). A numeric range requires a numeric type, the default value must satisfy the rule, and a non-empty `value_pattern` must compile. Example create and update return `script_issues` warnings from the MCP line rules without blocking the save.

**Blocked by:** None.

**Status:** resolved

- [x] `scriptrules.go` ports the MCP's pattern, allowed-value and example rules, with unit tests
- [x] `validateDefinition` returns 422 on rule violations; imported rows are never rejected
- [x] Example responses carry `script_issues`
