# 01: Fixture oracle — derive workbook counts, stop pinning literals

**What to build:** One fixture-oracle module in the importer package: opens the importer testdata xlsx, calls the pure `ParseWorkbook`, exposes the six per-table counts. Every test that currently hard-codes the workbook counts switches to asserting the round-trip invariant against the oracle instead of literals. After this ticket, a workbook batch requires zero test edits.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] Oracle added as a non-test file in `internal/importer`: reads the testdata fixture through `ParseWorkbook` and returns `Counts{Functions, Parameters, Criteria, ParameterLinks, CriteriaLinks, Examples}`
- [x] `workbook_test.go` becomes the oracle's consumer (its computed counts become the oracle's shape; its count literals disappear)
- [x] `api/import_test.go` asserts import counts == oracle counts (three pin spots); the row-order-dependent `id:260` pin replaced by the semantic assertion: ANT has exactly 2 rows, one `WIFI_TEST`-scoped + one `OSP`-scoped
- [x] `api/export_test.go` asserts export counts == oracle counts
- [x] `api/orphans_test.go` stale count comment (old join-semantics figures) deleted; assertions keep pinning the true invariant (orphans resolve to 0)
- [x] `e2e/smoke.spec.ts` headline count derived from the API at runtime (extend the existing derive-from-API pattern) — UI total must equal API total, no literal
- [x] `e2e/examples.spec.ts` derives the fixture example count from the API before its +1 seeded example, no literal
- [x] Full pass green: `go -C backend test ./...`, `pnpm check`, `pnpm build`, `pnpm test:e2e`
- [x] Grep confirms no remaining hard-coded workbook counts (`296`, `270`, `572`, `219` literals) in test/e2e files
