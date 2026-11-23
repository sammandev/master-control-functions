# Master-Control-Functions

CRUD web application over the script-definition knowledge. Go (Fiber) backend, Svelte 5 (SvelteKit, adapter-static) frontend, and hosted PostgreSQL.

For the SQLite transfer, credential setup, database verification, and rollback, use [the deployment runbook](deploy/README.md#3-transfer-and-verify-sqlite-first). The agent has prepared the code; the user runs database operations on the approved Ubuntu host.

For the paired backend/frontend Docker release with Bun builds and isolated blue-green switching, use [the Ubuntu deployment runbook](deploy/README.md). Before the first production release, transfer and verify the existing SQLite catalog into the approved target PostgreSQL database.

See `CONTEXT.md` for domain vocabulary, `docs/adr/` for architectural decisions, and `.scratch/<feature>/` for the spec and tickets of each feature (for example `.scratch/device-vocabulary-registry/`).

## Local Development

Prerequisites: Go 1.27+ (`go version`), Node 24+ (`node --version`), and pnpm 12.6+ (`pnpm --version`).

The restricted Windows workstation may run build, vet, type-check, lint, unit-test, and mocked-UI checks. Do not install or connect to a database, start the database-backed server, or run database-backed tests here. Database operations and migrations are operator-run on the approved Ubuntu host.

Run the database-free checks from the repository root:

```powershell
go -C backend build ./...
go -C backend vet ./...
go -C backend test ./internal/api ./tests/importer -run .
pnpm --dir frontend check
pnpm --dir frontend exec eslint .
pnpm --dir frontend exec vitest run --project server
pnpm --dir frontend build
```

For live API workflows, use an operator-managed Ubuntu backend and the approved connection procedure in [backend/README.md](backend/README.md). Do not start the Go server locally with a database connection.

Database-backed Go tests, reset-based Playwright E2E tests, schema migrations, and SQLite transfer checks use disposable databases on the approved Ubuntu host only. See the [deployment runbook](deploy/README.md).

### Bun as an optional script runner (ADR-0007)

pnpm is the canonical installer — `pnpm-lock.yaml` is the only lockfile, and the `packageManager` field pins pnpm 12. Every script can also be run through bun (MIT, [bun.sh](https://bun.sh)); `bun run` shells out to the package.json scripts, so vite still executes under Node:

```powershell
bun run dev      # vite dev server
bun run build    # vite build
bun run check    # svelte-check
```

Bun as a local *installer* is intentionally unsupported: mixing `pnpm-lock.yaml` with a bun lockfile invites lockfile drift and conflicting `node_modules` layouts. The isolated Docker builder is an exception: it derives a temporary Bun lockfile from the canonical pnpm lockfile and uses Bun to run checks/builds with the existing Node tooling. It does not change the workstation's dependencies or add a second repository lockfile.

### Go toolchain notes

If `go` is not on PATH: it is installed at `%LOCALAPPDATA%\Programs\go\bin` (official go.dev distribution, BSD-3-Clause). Add that directory to your user PATH, or invoke `go` by full path.

### Database migrations

PostgreSQL schema changes are embedded `.sql` files under `backend/internal/db/postgres_migrations`, applied in filename order and recorded in `schema_migrations`. The baseline starts at version 6, matching the complete legacy catalog. Append new migrations; never edit an applied one. The original SQLite migrations remain unchanged under `backend/internal/db/migrations` for transfer fixtures and historical reference. See ADR-0012 for the cutover decision.

sqlc (v1.31.1, run via `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate` from `backend/`) regenerates the PostgreSQL store from `internal/store/queries/*.sql` and the PostgreSQL migrations. Keep device queries in their existing small files and regenerate after query changes; do not hand-edit generated `*.sql.go` output. Nullable search filters use explicit PostgreSQL text casts.

### End-to-end tests

A Playwright suite drives the real UI served by the Go server against a disposable PostgreSQL database seeded through `/api/import/reset`. The approved Ubuntu operator sets `E2E_DATABASE_URL`; the suite refuses an absent URL, does not reuse an existing server, and must never point at the production catalog. Run this suite only on the approved host; failed traces are retained under `frontend/test-results/`.

For the production-equivalent SQLite transfer gate, use the approved Ubuntu host and the `deploy/README.md` procedure.

### Full Verification

Run database-free frontend and compile gates on the restricted workstation. The operator runs the full database-backed suite on the approved Ubuntu host with separate disposable `TEST_DATABASE_URL` and `E2E_DATABASE_URL` values:

```bash
set -e
go -C backend test -tags sqlite_migration ./...
pnpm --dir frontend check
pnpm --dir frontend build
pnpm --dir frontend test:e2e
```

The E2E suite serves `frontend/build`, so the build step keeps it from driving a stale UI bundle.

Performance sanity is pinned in Go: `TestFullDatasetListLatency` imports the real workbook and asserts every read endpoint (lists, detail, links, orphans, facets) answers locally in under a second (they run in low single-digit milliseconds).

## Issue tracker

Issues live as local markdown under `.scratch/<feature>/issues/` (one file per ticket, numbered in dependency order). Work blockers-first: any ticket whose `Blocked by` list is fully resolved can be picked up.
