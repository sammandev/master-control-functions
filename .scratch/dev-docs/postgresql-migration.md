# PostgreSQL Cutover

The web backend and Script Forge now use PostgreSQL. The transfer command is prepared but has not been run against your databases. Run all database operations yourself on the approved Ubuntu host. The agent must not access that host, connect to either database, or run database tests on the restricted Windows machine.

Target: `10.176.2.139:5432`, database `program_functions`, web application user `ptb`. Rotate the password previously shared in chat before using it. The examples contain no password. Keep secrets in a protected service environment or secret manager, outside the repository.

## 1. Prepare and Check Connectivity

Use Go 1.27.1+, Rust 1.88+, and an existing PostgreSQL server on Ubuntu. PostgreSQL 14+ is recommended. Building the Rust native-TLS client on Ubuntu requires the OpenSSL development headers, `pkg-config`, and a C toolchain. `psql`, `pg_dump`, and `sqlite3` are operator tools for preflight and backups; install them only on an approved off-machine host if needed. RTK wrappers can be removed on Ubuntu if RTK is not installed.

In a private Ubuntu terminal, enter the rotated password without echoing it:

```bash
read -rsp 'Rotated PostgreSQL password: ' PGPASSWORD
printf '\n'
export PGPASSWORD
export DATABASE_URL='postgresql://ptb@10.176.2.139:5432/program_functions?sslmode=verify-full'
export SCRIPT_FORGE_DATABASE_URL="$DATABASE_URL"
```

Both services accept `PGPASSWORD` when the URL has no password. This avoids URL escaping issues with characters such as `#`. Embedded URL passwords must be percent-encoded. Never print a password-bearing URL or commit a credentials file.

For a private certificate authority, set `PGSSLROOTCERT` to its PEM root certificate. The certificate must cover the hostname or IP in the URL; use the server's certificate hostname instead of its IP if required. Both clients default to verified TLS and accept only `verify-full`, `require`, or explicit `disable`; `require` also verifies certificates and hostnames. If the server deliberately provides no TLS on a trusted network, explicitly select `sslmode=disable` in both URLs after accepting that credentials and catalog data will be unencrypted. There is no automatic plaintext fallback.

```bash
rtk proxy psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
  "SELECT version(), current_database(), current_user, current_schema(), current_setting('server_encoding'), has_schema_privilege(current_user, current_schema(), 'CREATE');"
rtk proxy psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
  "SELECT schemaname, tablename FROM pg_tables WHERE schemaname = current_schema();"
```

Confirm the database is `program_functions`, encoding is `UTF8`, the schema is the intended application schema, and `ptb` can create objects there. Database ownership does not always imply CREATE permission on `public`; have the database administrator grant the required privileges if necessary. Start with an empty application schema. The transfer refuses nonempty application tables, even though overwriting the empty target was authorized.

## 2. Build and Rehearse

Transfer both updated repositories to Ubuntu using your normal deployment process. Local edits do not change the deployed services. Keep the old SQLite-backed binaries and their configuration for rollback.

From the web repository root:

```bash
rtk proxy go -C backend mod tidy
rtk proxy go -C backend vet ./...
rtk proxy go -C backend build -o webprogfndb-server ./cmd/server
rtk proxy go -C backend build -tags sqlite_migration -o migrate-sqlite ./cmd/migrate-sqlite
export TEST_DATABASE_URL="$DATABASE_URL"
rtk proxy go -C backend test -tags sqlite_migration ./...
```

Outputs are under `backend/`. Default Go builds contain only the PostgreSQL client; the SQLite driver and transfer command require the off-machine-only `sqlite_migration` tag. Never enable that tag on the restricted Windows host. PostgreSQL integration tests use randomly named schemas and clean up only their own schemas. The tagged transfer regression tests also create disposable SQLite fixtures on the approved test host. Tests check IDs, nullable fields, orphans, Unicode text and device ordering, atomic rollback, retries, and refusal to overwrite existing target rows. Without `TEST_DATABASE_URL`, Go integration tests skip; skips are not cutover proof.

From the MCP repository root:

```bash
rtk cargo fmt --all -- --check
rtk cargo test --all-targets
rtk cargo clippy --all-targets -- -D warnings
rtk cargo build --release
```

The same `TEST_DATABASE_URL` and `PGPASSWORD` must be present. The PostgreSQL fixture checks the actual snapshot's read-only and repeatable-read settings. Rust database fixture tests fail clearly if the test URL is missing. Default features use PostgreSQL; do not enable `legacy-sqlite` unless deliberately verifying the historical reader.

The real UI smoke suite requires a separate disposable PostgreSQL database via `E2E_DATABASE_URL`. Have your administrator provision that test database if needed. Never point E2E at `program_functions`: workbook reset overwrites its catalog. From `frontend/`, build the SPA, set the disposable URL, then run `pnpm test:e2e`. The UI-only mocked configuration does not need a database.

Proceed only after the off-machine tests pass. PostgreSQL version, TLS, source data, and live service behavior have not been verified by the agent.

## 3. Stop Writes and Take Backups

Stop the old web writer and MCP process yourself. Keep them stopped until verification finishes. Identify the SQLite file actually used by the deployed server's `-db` argument; the historical default was `webprogfndb.db` relative to the server's working directory. The older `.bak_before_*` files are not necessarily the current catalog.

Copy that file and any existing matching `-wal` and `-shm` companions to an Ubuntu staging directory after writes stop. Do not take only the main file from a live WAL database. Keep the originals untouched. Create a consistent snapshot from the staging copy using the SQLite backup API on Ubuntu:

```bash
umask 077
rtk proxy sqlite3 /srv/migration/staged/webprogfndb.db \
  ".backup '/srv/migration/catalog-snapshot.db'"
chmod 400 /srv/migration/catalog-snapshot.db
rtk proxy pg_dump --format=custom \
  --file=/srv/migration/program-functions-before.dump "$DATABASE_URL"
```

Use your actual staging paths. The snapshot must have schema version 6. The transfer checks that version and refuses an older schema. If needed, upgrade a staging copy with the retained SQLite-backed server before making the snapshot; do not upgrade or overwrite your sole original backup.

## 4. Transfer and Verify

From the web repository root on Ubuntu:

```bash
rtk proxy ./backend/migrate-sqlite \
  --source /srv/migration/catalog-snapshot.db \
  --confirm-writes-stopped

rtk proxy ./backend/migrate-sqlite \
  --source /srv/migration/catalog-snapshot.db \
  --confirm-writes-stopped --verify-only
```

The destination is selected only by `DATABASE_URL`. The source must already exist and is opened with SQLite read-only mode plus `query_only`. Source and target must use UTF-8; device keys are ordered by equivalent byte collations for deterministic verification. Integrity and aggregate foreign keys are checked before copying. Unknown application tables cause refusal rather than silent omission.

All 11 application tables are copied: functions, both definition dictionaries, both link tables, examples, import runs, import normalizations, sequences, sequence steps, and devices. IDs, scope keys, text timestamps, integer flags, NULLs, and JSON text are retained. Orphan-capable links and examples keep their intentional lack of foreign keys. The old migration ledger is translated to the PostgreSQL version-6 baseline; it is not copied as SQLite migration history.

The command prints a row count and SHA-256 digest for every table. It compares every stored column before committing and restarts identity generators above the copied maximum IDs. Copying and identity restarts share one PostgreSQL transaction. Any mismatch or insert failure rolls back all copied rows. Schema initialization is a separate transaction and may remain after a failed copy; retrying into those empty tables is safe.

Expected final output:

```text
Migration committed: all 11 application tables verified; SQLite source unchanged.
Verification passed: all 11 application tables match the SQLite snapshot.
```

After a successful copy, use `--verify-only`; a second copy refuses the populated target. If the connection fails during commit, the command reports uncertain commit status: verify first instead of blindly retrying. Keep the services stopped during both commands. Save their verification output with the backup records.

## 5. Switch Both Services

Configure the web service's protected environment with `DATABASE_URL` and its password source. Configure the MCP process with `SCRIPT_FORGE_DATABASE_URL` (or `DATABASE_URL`) and its password source. `PGSSLROOTCERT` is shared certificate configuration when needed. Do not leave the old web `-db` flag or MCP `--db` argument in service/client configuration; PostgreSQL startup rejects those old arguments.

Start the new web service, initially with user write access disabled:

```bash
rtk proxy ./backend/webprogfndb-server -addr 127.0.0.1:8080 -web frontend/build
rtk proxy curl --noproxy 127.0.0.1 --fail http://127.0.0.1:8080/health
```

Expect `{"status":"ok","db":"ok"}`. Check device-scoped and Global definitions, orphan reports, examples, sequences, and workbook export. Restart the MCP client with the new binary and environment; verify initialization and all five tools against the migrated catalog. MCP uses stdio, so its stdout is protocol output, not a console interface.

Prefer a separate read-only MCP login. A database administrator can create it in `psql` and set its password interactively:

```sql
CREATE ROLE script_forge LOGIN;
\password script_forge
GRANT CONNECT ON DATABASE program_functions TO script_forge;
GRANT USAGE ON SCHEMA public TO script_forge;
GRANT SELECT ON TABLE public.functions, public.parameter_definitions,
    public.criteria_definitions, public.parameter_links, public.criteria_links,
    public.examples TO script_forge;
```

These commands assume the application schema is `public`; adjust if different. Use that login and its own protected password for `SCRIPT_FORGE_DATABASE_URL`. The MCP never creates tables, migrates data, or writes catalog rows. Its snapshots are read-only repeatable-read transactions; database permissions add an independent write barrier.

Reopen application writes only after the read-only acceptance checks pass. A subsequent CRUD identity check writes new data and begins the PostgreSQL-only operational period.

## Rollback Boundary

Before reopening writes, stop both new processes and restore the retained SQLite-backed binaries and their original configuration. Restart them against the untouched original SQLite dataset. Keep PostgreSQL and both backups; the procedure never deletes the source.

After PostgreSQL accepts new writes, the old SQLite snapshot is stale. Returning to it would lose those changes. Stop writes and plan reconciliation or a PostgreSQL recovery before switching back. There is no automatic SQLite fallback and no implicit destructive cleanup.
