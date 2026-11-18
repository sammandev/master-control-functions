# Master-Control-Functions

CRUD web application over the script-definition knowledge. Go (Fiber) backend, Svelte 5 (SvelteKit, adapter-static) frontend, and hosted PostgreSQL.

For the SQLite transfer, credential setup, database verification, and rollback, use [the deployment runbook](deploy/README.md#3-transfer-and-verify-sqlite-first). The agent has prepared the code; the user runs database operations on the approved Ubuntu host.

For the paired backend/frontend Docker release with Bun builds and isolated blue-green switching, use [the Ubuntu deployment runbook](deploy/README.md). Transfer and verify the existing SQLite catalog before the first release; PostgreSQL is currently empty.

See `CONTEXT.md` for domain vocabulary, `docs/adr/` for architectural decisions, and `.scratch/<feature>/` for the spec and tickets of each feature (for example `.scratch/device-vocabulary-registry/`).

## Development

Prerequisites: Go 1.27+ (`go version`), Node 24+ (`node --version`), pnpm 12 (canonical frontend package manager, auto-pinned via the `packageManager` field + corepack). [Bun](https://bun.sh) (optional) can substitute for npm/pnpm as a *script runner* — see below.

```powershell
# Build the frontend (outputs to frontend/build)
cd frontend
pnpm install
pnpm build

# Type-check the frontend
pnpm check
```

The backend requires `DATABASE_URL`, with `PGPASSWORD` or a secret-managed URL supplying the password. Verified TLS is the default; `PGSSLROOTCERT` selects a private CA certificate. Run database-backed commands only on the approved off-machine host:

```bash
rtk proxy go -C backend build -o webprogfndb-server ./cmd/server
export TEST_DATABASE_URL="$DATABASE_URL"
rtk proxy go -C backend test ./...
rtk proxy ./backend/webprogfndb-server -addr 127.0.0.1:8080 -web frontend/build
```

Integration tests create isolated PostgreSQL schemas. Without `TEST_DATABASE_URL` they skip, which does not establish database compatibility. Build, vet, and format checks do not execute database tests and can run on the restricted Windows host.

The SQLite transfer command and its regression fixtures require `-tags sqlite_migration`, used only on the approved Ubuntu host. Normal builds exclude the SQLite engine. The cutover procedure includes the tagged build and test commands.

The server logs the schema version without logging credentials. `GET /health` returns `{"status":"ok","db":"ok"}` when the database is reachable. The old `-db` startup flag is removed.

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

A Playwright suite drives the real UI served by the Go server against a disposable PostgreSQL database seeded through `/api/import/reset`. Set `E2E_DATABASE_URL` explicitly; the suite refuses an absent URL and does not reuse an existing server. Never point it at the production catalog: reset overwrites knowledge rows. Database-backed E2E runs are off-machine only. Happy paths include search, editing links, composition, orphans, sequences, and devices.

```powershell
# One-time setup
cd frontend
pnpm install
pnpm exec playwright install chromium
pnpm build        # the suite serves frontend/build through the Go server

# Run the smoke suite only on the approved host with E2E_DATABASE_URL set
pnpm test:e2e

# Traces are kept for failed runs:
pnpm exec playwright show-trace test-results/<failing-test>/trace.zip
```

### Full Verification

Run on the approved host after setting `TEST_DATABASE_URL` and a separate disposable `E2E_DATABASE_URL`:

```bash
set -e
rtk proxy go -C backend test -tags sqlite_migration ./...
cd frontend
pnpm check
pnpm build
pnpm test:e2e
```

The E2E suite serves `frontend/build`, so the build step keeps it from driving a stale UI bundle.

Performance sanity is pinned in Go: `TestFullDatasetListLatency` imports the real workbook and asserts every read endpoint (lists, detail, links, orphans, facets) answers locally in under a second (they run in low single-digit milliseconds).

## Issue tracker

Issues live as local markdown under `.scratch/<feature>/issues/` (one file per ticket, numbered in dependency order). Work blockers-first: any ticket whose `Blocked by` list is fully resolved can be picked up.
