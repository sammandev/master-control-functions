# Master-Control-Functions

CRUD web application over the script-definition knowledge. Go (Fiber) backend, Svelte 5 (SvelteKit, adapter-static) frontend, and hosted PostgreSQL.

For the SQLite transfer, credential setup, database verification, and rollback, use [the deployment runbook](deploy/README.md#3-transfer-and-verify-sqlite-first). The agent has prepared the code; the user runs database operations on the approved Ubuntu host.

For the paired backend/frontend Docker release with Bun builds and isolated blue-green switching, use [the Ubuntu deployment runbook](deploy/README.md). Before the first production release, transfer and verify the existing SQLite catalog into the approved target PostgreSQL database.

See `CONTEXT.md` for domain vocabulary, `docs/adr/` for architectural decisions, and `.scratch/<feature>/` for the spec and tickets of each feature (for example `.scratch/device-vocabulary-registry/`).

## Local Development

Prerequisites: Go 1.27+ (`go version`), Node 24+ (`node --version`), pnpm 12.6+ (`pnpm --version`), and a development PostgreSQL database.

### 1. Start PostgreSQL and configure the backend

The Go backend does not load `.env` files automatically. A local values file may contain the individual `DB_*` values, but `DATABASE_URL` must use Go/pgx syntax:

```text
postgresql://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable
```

The existing `backend/.env.dev` is suitable only as a source of local `DB_*` components. Its `postgresql+psycopg://` `DATABASE_URL` is for SQLAlchemy and is rejected by the Go backend.

From the repository root in PowerShell, load the local component values and build a process-local Go URL:

```powershell
Get-Content backend/.env.dev | ForEach-Object {
    $line = $_.Trim()
    if ($line -and -not $line.StartsWith('#') -and $line -match '^([A-Za-z_][A-Za-z0-9_]*)=(.*)$') {
        $value = $matches[2].Trim().Trim('"').Trim("'")
        [Environment]::SetEnvironmentVariable($matches[1], $value, 'Process')
    }
}
$user = [Uri]::EscapeDataString($env:DB_USER)
$password = [Uri]::EscapeDataString($env:DB_PASSWORD)
$database = [Uri]::EscapeDataString($env:DB_NAME)
$env:DATABASE_URL = "postgresql://${user}:${password}@$($env:DB_HOST):$($env:DB_PORT)/${database}?sslmode=disable"
```

Use `sslmode=disable` only for an existing local/loopback PostgreSQL connection. Do not print, commit, or share a credential-bearing `DATABASE_URL`. The `.env*` pattern is gitignored except for examples.

### 2. Start the backend

Keep this terminal open:

```powershell
go -C backend run -mod=readonly ./cmd/server -addr 127.0.0.1:8088
```

The server applies embedded PostgreSQL migrations at startup. On this workstation, `DATABASE_URL` must point to the development catalog only. Check `http://127.0.0.1:8088/health`; a healthy backend returns `{"status":"ok","db":"ok"}`.

### 3. Start the frontend

In a second terminal:

```powershell
cd frontend
pnpm install --frozen-lockfile
pnpm dev --host 127.0.0.1 --port 5173 --strictPort
```

Open `http://127.0.0.1:5173`. Vite proxies `/api` to `http://127.0.0.1:8088`. Set `BACKEND_URL` before `pnpm dev` only when the API runs at a different address.

### 4. Run local checks

```powershell
go -C backend test ./...
go -C backend vet ./...
pnpm --dir frontend check
pnpm --dir frontend exec eslint .
pnpm --dir frontend exec vitest run --project server --project client
```

For database-backed Go tests, set `TEST_DATABASE_URL` to a disposable PostgreSQL database. The tests create and drop isolated schemas, but do not use the production catalog:

```powershell
$env:TEST_DATABASE_URL = $env:DATABASE_URL
go -C backend test ./internal/api ./internal/db ./tests/api -count=1 -v -run Troubleshooting
```

Full Playwright E2E tests reset data through `/api/import/reset`, so use a separate disposable database for `E2E_DATABASE_URL`; do not point that variable at the normal development catalog. The SQLite transfer suite still requires the `sqlite_migration` tag and must follow the approved [deployment runbook](deploy/README.md).

The server logs the schema version without logging credentials. The old `-db` startup flag is removed.

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

A Playwright suite drives the real UI served by the Go server against a disposable PostgreSQL database seeded through `/api/import/reset`. Set `E2E_DATABASE_URL` explicitly; the suite refuses an absent URL and does not reuse an existing server. Never point it at the production or normal development catalog: reset overwrites knowledge rows.

```powershell
# One-time setup
cd frontend
pnpm install --frozen-lockfile
pnpm exec playwright install chromium
pnpm build

# In PowerShell, point this at a separate disposable database.
$env:E2E_DATABASE_URL = 'postgresql://USER:PASSWORD@HOST:PORT/DISPOSABLE_DATABASE?sslmode=disable'
pnpm test:e2e

# Failed-run traces:
pnpm exec playwright show-trace test-results/<failing-test>/trace.zip
```

For the production-equivalent SQLite transfer gate, use the approved Ubuntu host and the `deploy/README.md` procedure.

### Full Verification

Local database and frontend gates:

```powershell
$env:TEST_DATABASE_URL = $env:DATABASE_URL
go -C backend test ./...
pnpm --dir frontend check
pnpm --dir frontend exec eslint .
pnpm --dir frontend exec vitest run --project server --project client
pnpm --dir frontend build
```

For the full reset-based Playwright run, add a separate disposable `E2E_DATABASE_URL` and run `pnpm --dir frontend test:e2e`. For SQLite migration coverage on the approved host:

```bash
set -e
go -C backend test -tags sqlite_migration ./...
```

The E2E suite serves `frontend/build`, so the build step keeps it from driving a stale UI bundle.

Performance sanity is pinned in Go: `TestFullDatasetListLatency` imports the real workbook and asserts every read endpoint (lists, detail, links, orphans, facets) answers locally in under a second (they run in low single-digit milliseconds).

## Issue tracker

Issues live as local markdown under `.scratch/<feature>/issues/` (one file per ticket, numbered in dependency order). Work blockers-first: any ticket whose `Blocked by` list is fully resolved can be picked up.
