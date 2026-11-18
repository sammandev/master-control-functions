# 04: MCP readiness tab

**What to build:** A Function detail tab that shows the readiness status for a chosen device scope. On it, reviewers can audit or remove the grammar and audit or remove each named key (Line, Bundle or Remove). Every change records **Audited by**.

**Blocked by:** 03

**Status:** resolved

- [x] `McpReadinessPanel.svelte` and the `mcp.ts` client; the tab is wired into `FunctionDetail.svelte`
- [x] `npm run check`, prettier and eslint are clean
- [ ] Browser smoke test and e2e coverage on the approved host
