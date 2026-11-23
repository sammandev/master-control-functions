# Backend

Go/Fiber REST API backed by PostgreSQL. The server can also serve the built frontend; the Ubuntu setup below uses a separate Nginx frontend site.

The [paired Docker deployment](../deploy/README.md) packages the backend and Bun-built frontend together, with isolated blue-green application slots behind a stable Nginx edge. Its first release requires a verified SQLite transfer. The native-host systemd setup below remains an alternative.

## Requirements

- Go 1.27.3 or newer, matching [go.mod](go.mod).
- An existing PostgreSQL database on the approved Ubuntu host. This guide does not install a database engine.
- Nginx and systemd for the Ubuntu deployment.

Database connections, backend startup, migrations, and database-backed tests must run on the approved off-machine host. On the restricted Windows workstation, use build, vet, and explicitly database-free validation only. The server applies embedded schema migrations at startup; back up the database and review the [PostgreSQL cutover procedure](../docs/postgresql-migration.md) before deployment. Startup does not transfer the old SQLite data.

## Database Configuration

From this directory on Ubuntu, create a private environment file for an interactive development session:

```bash
cp .env.example .env
chmod 600 .env
```

Edit [.env.example](.env.example)'s copied values in `.env`: replace `replace_me` with the current password and confirm the existing role (`ptb`), database (`master_control_functions`) and port (`5432`). URL-encode reserved characters in credentials. Keep the URL quoted and retain the `KEY="value"` format; do not add shell commands or `export` to the file.

`localhost` means the machine running the Go process. These defaults assume PostgreSQL and the backend run directly on the same Ubuntu host, outside containers. PostgreSQL must accept authenticated loopback TCP connections for that role/database. Have the database operator check its existing `listen_addresses` and `pg_hba.conf`; do not change them to `trust` or expose port 5432 to make this work.

`sslmode=disable` explicitly permits plaintext on same-host loopback only. If PostgreSQL requires TLS, use `sslmode=verify-full`, a trusted certificate valid for `localhost`, and `&sslrootcert=/path/to/ca.crt` for a private CA. For an approved remote-database topology, replace the host with its certificate hostname and require verified TLS. Never use the loopback plaintext example for a network connection.

Go does not read `.env` automatically. In Bash, load the trusted file into the current process environment:

```bash
set -a
. ./.env
set +a
```

Do not commit real environment files or share connection URLs containing passwords. Rotate any previously shared database password before deployment. Production uses a separate root-owned environment file described below.

## Local Build And Development

Run these database-free checks from `backend/` on Windows or Ubuntu:

```sh
go build -mod=readonly ./...
go vet -mod=readonly ./...
```

### MCP Readiness Parity

The MCP repository provides `deploy/mcp-parity.py` for checking the web
readiness summary against `script-audit-mcp` using this repository's actual
PostgreSQL migrations. The user must run it on the approved host with
`TEST_DATABASE_URL` set to a disposable test database; it creates and removes
an isolated schema. Never use the production catalog. From the MCP repository:

```bash
export TEST_DATABASE_URL='<disposable test database URL>'
python deploy/mcp-parity.py --web-repo /path/to/Web_Program_Functions_Database
```

The agent does not access the approved host or run this database-backed check.

### Offline Dependency-Maintenance Baseline

When approved network access cannot fetch Go modules, use an existing local
vendor tree without changing the workstation's proxy. This directory is
gitignored, not committed. Regenerate it with `go mod vendor` after dependency
changes while approved downloads are available. In a separate PowerShell terminal:

```powershell
$env:GOPROXY = 'off'
$env:GOTOOLCHAIN = 'local'
go build -mod=vendor ./...
go vet -mod=vendor ./...
go test -mod=vendor -count=1 ./internal/api ./tests/importer
```

These settings affect only that terminal; close it to discard them. The selected tests are database-free. A passing baseline verifies the existing vendored versions, not the proposed security updates or absence of vulnerabilities. Run the full database-backed suite separately with an explicit disposable `TEST_DATABASE_URL`.

An offline upgrade requires the complete approved module archives and checksums
in the Go cache, plus the required Go toolchain. Metadata files alone are not
enough. Keep manifests, checksums, and vendor synchronized with Go tooling; do not
change version numbers alone or hand-patch vendor to bypass unavailable downloads.
Run vulnerability scanning once the scanner and advisory data are available.

### Security Update Validation

The security update uses Excelize 2.11.1, pgx 5.12.0, compress 1.20.4, and Fiber
v3.7.0 with Go 1.27.3. Keep the Go builder in the
[Dockerfile](../deploy/Dockerfile) and the migration-builder commands aligned.
Application handlers and static middleware are migrated from deprecated Fiber v2
to Fiber v3 (`v3.7.0`).

Remaining compatible updates include SQLite 1.62.0 and its selected libc 1.80.1,
go-humanize 1.1.0, go-runewidth 0.0.31, go-strftime 1.1.0, x/sync 0.25.0,
x/sys 0.50.0, and x/text 0.44.0, plus Fiber v3's transitive requirements
(gofiber/schema 1.9.3, gofiber/utils/v2 2.6.4, philhofer/fwd 1.2.1, and
tinylib/msgp 1.6.5). Do not force SQLite's libc dependency to an
unrelated version merely because a newer release exists.

The Linux production and migration-tagged govulncheck scans found no reachable
or imported-package vulnerabilities with the updated selections. The unused
OpenPGP packages in x/crypto 0.58.0 still produce the module-only
[GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) warning; it is not suppressed.
These results apply to the advisory database checked on 2026-11-10, not future
advisories or OS/container packages.

fasthttp 1.78.0 uses go-brrr for Brotli support. The module graph still explicitly
selects patched andybalholm/brotli 1.3.0 even though no package from that module
is compiled into the server. `go mod tidy` can remove this unused security pin;
if maintaining it, reapply `go get github.com/andybalholm/brotli@v1.3.0` after
tidying and regenerate vendor.

The workbook regression tests cover fixture rewrite/parse preservation and
rejection of an oversized row after a valid header. These are database-free;
they do not replace the operator-run database export/import roundtrip.
Keep explicit `sslmode=disable` unchanged for the existing approved deployment.

On the approved Ubuntu host, load the operator-managed environment and run the server as described in the deployment section. The embedded migrations apply before the listener starts. The `-web` path is relative to `backend/`; if the directory is absent, the server reports API-only mode.

On `SIGINT` or `SIGTERM`, the server stops accepting connections and waits for in-flight requests to finish. `-shutdown-timeout` sets the maximum wait (default `45s`). Keep it below the supervisor's stop grace period: Compose `stop_grace_period` is `60s` and the systemd default `TimeoutStopSec` is `90s`.

Do not start the backend on the restricted Windows workstation. Live API workflows and migrations require the approved Ubuntu database host.

## Device / Interface Contract

Device, Device Name and Interface are the same catalog. Function payloads retain `interface`, and links, definitions and Examples retain `device_name`, for compatibility. Blank values remain valid and mean Global. The Device index includes Function interfaces as well as existing linked content and registered Devices.

`PUT /api/devices/:device` accepts `{"device_name":"NEW_NAME"}`. Names are trimmed and uppercased. Both registered and derived Devices can be renamed, including populated Devices. One transaction updates Function interfaces, link and definition scopes, Example scopes and sequence references. IDs, registry creation timestamps and saved Example script text are preserved. Global cannot be renamed. Invalid names return 422, a missing Device returns 404, and an existing target name returns 409 without merging data.

`DELETE /api/devices/:device` requires a JSON body `{"confirmation":"DELETE DEVICE_NAME"}` matching the stored name exactly (case and whitespace included). The Devices directory offers deletion for populated and empty Devices, with a typed confirmation dialog. This intentionally replaces the former empty-only delete contract; deploy frontend and backend together.

Deletion is permanent and transactional: it removes the registration, Functions whose `interface` is the Device, scoped Parameter/Criteria definitions, scoped links and Examples, Function grammar/key audits, and Sequence steps whose Function interface or Device scope matches. Links and Examples in other scopes belonging exclusively to deleted Functions are also removed. Because those rows have no Function interface, other-scope rows are preserved when another Function has the same `(source, name)`. Global definitions, other Devices' Functions, and Sequence containers survive; surviving steps retain IDs/order and are renumbered. Saved text and import history are not rewritten. Global cannot be deleted (422); invalid confirmation returns 422, malformed/missing JSON returns 400, and a missing Device returns 404.

Before release, run the Device API regressions in [devices_test.go](tests/api/devices_test.go) and [device_delete_test.go](tests/api/device_delete_test.go) against a disposable PostgreSQL database on the approved Ubuntu host.

### Device Type And Workbook Headers

Create Devices with `POST /api/devices` and `{"device_name":"RS232","source":"BUILT-IN"}`. The UI labels `source` as Type and offers the same `BUILT-IN` and `DLL` values as Function Source. Older clients may omit `source`. The catalog returns a single `source` plus a `sources` array; for untyped registrations and content-derived Devices it infers sources from Function interfaces, parameter/criteria links and Examples. Multiple inferred sources are reported together; Devices with no source-bearing content remain unassigned. An explicitly registered Type takes precedence over inference. Rename preserves Type, and import/reset preserves existing registry metadata. Registry-only metadata is not exported into the six-sheet workbook.

Both Import and Reset & import accept `function_parameter.device_name`, including blank Global scopes. Legacy `function_parameter.interface` is accepted only when the canonical header is absent; a blank canonical value never falls back to the legacy cell. Export always writes `function_parameter.device_name`. The `function_main.interface` header is unchanged.

Review [0007_device_source.sql](internal/db/postgres_migrations/0007_device_source.sql) before deployment. It adds a nullable, enum-checked column without changing existing rows. Apply it through normal backend startup on the approved host before using the updated frontend, then run the Device and import/export regressions against a disposable PostgreSQL database there. For an application rollback, retain this column and its data: the older application ignores it. Do not drop the column or alter existing registrations as part of rollback.

## MCP Catalog Contract

`script-audit-mcp` reads this catalog through the read-only `mcp_v1_*` views added by [0008_mcp_catalog.sql](internal/db/postgres_migrations/0008_mcp_catalog.sql) ([ADR-0012](../docs/adr/0012-mcp-catalog-contract.md)). The migration adds `function_grammars` and `script_key_audits`, keyed by the Function natural key, and seeds them with the MCP's former built-in profiles, so MCP behavior is unchanged at deployment. It changes no existing rows. Import, reset and export leave both tables untouched; Function renames cascade to them.

- `GET /api/mcp-readiness` lists every Function's status (`ready`, `bundle_ready`, `not_ready`) with its issue counts.
- `GET /api/functions/:id/mcp-readiness?device=` returns one Function's status with issues, warnings and selected fields.
- `GET`/`PUT`/`DELETE /api/functions/:id/grammar` reads, writes or removes the grammar. A PUT body is `{"encoding":"raw|named","scope":"line|bundle","audited_by":"...","note":"..."}`.
- `PUT /api/functions/:id/script-keys` takes `{"kind":"parameter|criterion","name":"...","scope":"line|bundle","audited_by":"..."}`.
- `DELETE /api/functions/:id/script-keys/:kind/:name` removes one key audit.

Example create and update responses include `script_issues` warnings from the MCP line rules; they never block a save. Saving a definition checks its allowed values, ranges, default value and `value_pattern`. Imported rows are not rejected. [mcp-catalog-report.sql](../docs/sql/mcp-catalog-report.sql) lists catalog problems read-only. [mcp-readonly-role.sql](../docs/sql/mcp-readonly-role.sql) creates a SELECT-only MCP login; point the MCP's `SCRIPT_AUDIT_DATABASE_URL` at this database with that role.

For rollback, the older application ignores the new tables and views. Keep them, because they hold audit history.

## Ubuntu Deployment

All commands in this section are operator-run on Ubuntu. Place the repository at `/opt/master-control-functions` using the approved deployment process. Keep database backups, real environment files and private working directories out of the deployed source. Back up existing application configuration before replacing it; leave unrelated services and Nginx sites unchanged.

The selected random high-port defaults are:

| Component      | Address           | Purpose                                 |
| -------------- | ----------------- | --------------------------------------- |
| PostgreSQL     | `localhost:5432`  | Same-host database                      |
| Go backend     | `127.0.0.1:24993` | Private Nginx upstream                  |
| Backend Nginx  | `39063`           | Separate HTTP API entry point           |
| Frontend Nginx | `31627`           | Browser application and same-origin API |

Before first deployment, check for existing listeners:

```bash
ss -ltn '( sport = :24993 or sport = :39063 or sport = :31627 )'
```

Availability has not been checked on the server. If a selected port is occupied by another service, choose a free one and update the service `-addr` plus both Nginx upstreams/listeners as applicable. Keep Go bound to loopback.

### Build And Secrets

Build a Linux binary on Ubuntu; a Windows executable cannot run there:

```bash
cd /opt/master-control-functions/backend
go build -mod=readonly -o webprogfndb-server ./cmd/server
```

Create the service account if it does not already exist:

```bash
sudo adduser --system --group --no-create-home webprogfndb
```

For a new deployment, create the production environment file without overwriting an existing one:

```bash
sudo install -d -m 700 -o root -g root /etc/master-control-functions
if [ ! -e /etc/master-control-functions/backend.env ]; then
    sudo install -m 600 -o root -g root .env.example /etc/master-control-functions/backend.env
fi
sudoedit /etc/master-control-functions/backend.env
```

Set the real production `DATABASE_URL` in that file. The same localhost/TLS rules apply. The systemd manager reads it before starting the non-root process; Nginx never reads it. Ensure the service user can traverse the deployment directories and execute the binary:

```bash
sudo -u webprogfndb test -x /opt/master-control-functions/backend/webprogfndb-server
```

### Backend Service

Create `/etc/systemd/system/master-control-functions-backend.service` with:

```ini
[Unit]
Description=Web Program Functions Database backend
After=network.target postgresql.service

[Service]
Type=simple
User=webprogfndb
Group=webprogfndb
WorkingDirectory=/opt/master-control-functions/backend
EnvironmentFile=/etc/master-control-functions/backend.env
ExecStart=/opt/master-control-functions/backend/webprogfndb-server -addr 127.0.0.1:24993 -web /opt/master-control-functions/frontend/build
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

After confirming the database is ready and startup migrations are approved:

```bash
sudo systemd-analyze verify /etc/systemd/system/master-control-functions-backend.service
sudo systemctl daemon-reload
sudo systemctl enable --now master-control-functions-backend
sudo systemctl status master-control-functions-backend --no-pager
curl --fail http://127.0.0.1:24993/health
```

Health should return `{"status":"ok","db":"ok"}`. If the unit already existed and its binary or environment changed, restart it with `sudo systemctl restart master-control-functions-backend`; `enable --now` does not restart an already-running service.

### Backend Nginx

[nginx.conf](nginx.conf) is a site template, included inside Nginx's existing `http` context. It listens on port `39063` and proxies to `127.0.0.1:24993`. Install it without replacing the global Nginx configuration:

```bash
sudo install -m 644 nginx.conf /etc/nginx/sites-available/master-control-functions-backend
```

For first activation only, create the site symlink:

```bash
sudo ln -s /etc/nginx/sites-available/master-control-functions-backend /etc/nginx/sites-enabled/master-control-functions-backend
```

Test the complete configuration before reloading. If Nginx is not yet running, start it instead of reloading:

```bash
sudo nginx -t
sudo systemctl reload nginx
curl --fail http://127.0.0.1:39063/health
```

The external API URL is `http://<ubuntu-host>:39063/api/...`. The browser normally uses the frontend site's same-origin proxy, so opening backend port `39063` through the firewall is optional. Restrict both HTTP listeners to approved clients. These templates provide no TLS or authentication; do not expose the CRUD/import endpoints to the public Internet without those controls. Ports `24993` and `5432` must remain private. Configure the [frontend site](../frontend/README.md#ubuntu-deployment) next.

## Tests

Tests live under [tests/](tests/). Database-backed tests use `TEST_DATABASE_URL`; without it, integration cases skip and do not prove PostgreSQL compatibility. The approved operator runs them on Ubuntu against a separate disposable database and never uses the production URL as `TEST_DATABASE_URL`. SQLite transfer fixtures require the `sqlite_migration` build tag and use the same approved off-machine boundary.

On the restricted Windows workstation, run only the explicitly database-free compile, vet, importer, type-check, lint, unit-test, and mocked-UI checks documented in this repository.

## Troubleshooting

- `DATABASE_URL is required`: export the file for an interactive run, or check the systemd `EnvironmentFile` path and restart the service after changes.
- Database authentication/TLS failures: check the existing role, URL encoding, loopback access rules and certificate policy with the database operator. Do not print a credential-bearing URL in logs or support messages.
- Nginx `502`: check the backend unit and its loopback port; inspect `sudo journalctl -u master-control-functions-backend -n 50 --no-pager` privately and redact secrets before sharing.
- Upload `413`: Nginx caps request bodies at 4 MiB. Review the backend limit too before raising either site's `client_max_body_size`.
