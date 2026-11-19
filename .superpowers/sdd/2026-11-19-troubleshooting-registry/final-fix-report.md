# Final Review Fix Report

## Summary

Completed all six Important findings at HEAD `b4c9c02` without expanding scope:

1. Added atomic category/severity reorder endpoints. Both sort-order writes run in one PostgreSQL transaction; the client now sends one reorder request rather than two independent PUTs.
2. Registry manager rows reconcile refreshed metadata by stable key while retaining unsaved label, description, and tone drafts for other rows.
3. Metadata refresh only advances to list/detail reconciliation when both category and severity refreshes succeed; failed refreshes preserve the existing guide UI and show current errors.
4. Category/severity key creation retries primary-key collisions in a bounded loop, while unique-label violations remain field conflicts.
5. Registry list counts now use grouped sqlc queries; mutation responses use key-scoped count queries and never load the troubleshooting guide catalog.
6. Guide form errors are displayed for category and severity, and title, device, category, severity, symptom, step title, and step instruction carry exact `aria-invalid`/`aria-describedby` wiring.

## TDD evidence

### RED

Backend count-source regression:

```text
$ go test ./internal/api -count=1 -run TroubleshootingRegistryCountsUseGroupedQueries
--- FAIL: TestTroubleshootingRegistryCountsUseGroupedQueries
    troubleshooting_registry_test.go:122: registry counts must not load or count the full troubleshooting guide catalog
FAIL
```

Frontend focused regression run:

```text
$ pnpm exec vitest run --project client tests/components/TroubleshootingRegistryDialog.svelte.spec.ts tests/components/TroubleshootingDialog.svelte.spec.ts
Test Files  2 failed (2)
Tests       4 failed | 6 passed (10)
```

Expected RED failures proved:

- reorder still made two fetches;
- metadata rerender erased another row's unsaved draft;
- guide controls lacked `aria-invalid` and error references.

The Playwright metadata-preservation scenario verifies the final page guard and passed after the production fix.

### GREEN

Focused frontend regressions:

```text
$ pnpm exec vitest run --project client tests/components/TroubleshootingRegistryDialog.svelte.spec.ts tests/components/TroubleshootingDialog.svelte.spec.ts tests/components/TroubleshootingRegistryPicker.svelte.spec.ts
Test Files  3 passed (3)
Tests       16 passed (16)
```

Focused backend (with database-gated cases compiled/skipped when no URL is configured):

```text
$ go test ./internal/api ./internal/db ./tests/api -count=1 -v -run Troubleshooting
PASS
ok webprogfndb/internal/api
ok webprogfndb/internal/db
ok webprogfndb/tests/api
```

Troubleshooting desktop/mobile Playwright:

```text
$ pnpm exec playwright test --config playwright.ui.config.ts tests/e2e/troubleshooting.spec.ts
Running 2 tests using 1 worker
2 passed (6.9s)
```

## Verification

Backend:

```text
$ go test ./... -count=1
ok webprogfndb/internal/api
ok webprogfndb/internal/db
ok webprogfndb/tests/api
ok webprogfndb/tests/db
ok webprogfndb/tests/importer

$ go vet ./...
(no output; exit 0)
```

Frontend:

```text
$ pnpm exec vitest run --project server --project client
Test Files  12 passed (12)
Tests       49 passed (49)

$ pnpm check
svelte-check found 0 errors and 0 warnings

$ pnpm exec eslint <modified frontend source and tests>
(no output; exit 0)

$ pnpm exec prettier --check <modified frontend source and tests>
All matched files use Prettier code style!
```

Svelte MCP process:

- `list-sections` was run before edits.
- Fetched the Svelte 5 state, derived/effect, binding, keyed-list, testing, and accessibility guidance relevant to this wave.
- Final autofixer checks returned **0 issues**. The manager reconciliation excerpt also returned three advisory `$effect`/`$derived` suggestions; Svelte check itself remained at 0 errors and 0 warnings, and the stable-key reconciliation behavior is regression-tested.

Graph/checks:

```text
$ graphify update .
Graph has 6424 nodes
Rebuilt: 6424 nodes, 11783 edges, 537 communities

$ git diff --check
(no whitespace errors; only existing LF/CRLF conversion warnings)
```

## Database-gated coverage

`TEST_DATABASE_URL` is unset in this worktree, so PostgreSQL integration tests were skipped only with this exact reason:

```text
PostgreSQL integration test: set TEST_DATABASE_URL on the authorized off-machine test host
```

Added DB-gated tests for:

- rollback of the second reorder write preserving both original order values;
- concurrent different-label slug collisions selecting the base key and numeric suffix;
- existing registry/count/mutation behavior and guide reconciliation.

No approved Ubuntu or production host was accessed.

## Final concerns

- The two new PostgreSQL rollback/concurrency tests compile but could not execute without `TEST_DATABASE_URL`.
- Graphify warns that its optional SQL parser is unavailable, an existing analyzer limitation, and that graph labels can be refreshed separately.
