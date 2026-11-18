# Ubuntu Blue-Green Deployment

Run every server command yourself on the approved Ubuntu host. The agent does not connect to it. Database engines, database tests and migration tools must never run on the restricted Windows workstation.

## Layout

The frontend builder uses Bun to install from the canonical pnpm lockfile and run checks/builds with the existing Node 24 tooling. Its temporary Bun lockfile exists only inside the build stage. The Go builder produces the PostgreSQL-client-only server. The final non-root Alpine image contains the Go executable, frontend files and CA certificates. Neither Bun nor Node runs in production.

The explicit Paraglide compilation step runs the installed `@inlang/paraglide-js/bin/run.js` CLI with Node before checks. Do not use `bun x paraglide-js`: the executable belongs to the scoped `@inlang/paraglide-js` dependency, and Bun can instead download the unrelated unscoped package and fail with `could not determine executable to run`. Invoking the installed CLI avoids package auto-install and uses the version selected by the lockfile. A backend stage marked `CANCELED` after this frontend failure does not establish a Go build failure or success.

The frontend stage sets `NODE_USE_ENV_PROXY=1` so Node's built-in `fetch` uses the mounted proxy secrets during explicit Paraglide compilation and Vite builds. Merely mounting `http_proxy` and `https_proxy` does not enable Node proxy support. If plugin downloads fail, Paraglide can report successful compilation while generating empty message declarations; `svelte-check` then reports that `messages/_index.d.ts` is not a module. Restore approved proxy connectivity and rebuild rather than editing generated files. The proxy opt-in is not inherited by the runtime stage, and proxy values remain build secrets.

| Service | Listener | Purpose |
| --- | --- | --- |
| Nginx edge | `31627` | Restricted LAN browser/API entry |
| Blue release | `127.0.0.1:24993` | Backend and static frontend |
| Green release | `127.0.0.1:24994` | Backend and static frontend |
| Existing PostgreSQL | `127.0.0.1:5432` | Stable database, outside Compose |

Containers use Linux host networking so the database remains reachable through loopback without opening PostgreSQL to a Docker subnet. This reduces network isolation; ports remain private, processes are non-root, capabilities are dropped, and filesystems are read-only. No PTB_OT files, ports, containers or networks are changed. Port `39063` is unnecessary because `/api` shares the browser origin.

Each application slot has a 512 MiB/1 CPU limit and a 64 MiB temporary directory. Edge has a 64 MiB/0.5 CPU limit. Both slots stay running after a release for rollback. These are starting limits, not measured capacity guarantees. Logs rotate at 10 MiB, retaining three files per container.

## 1. Prepare the Host

Publish the repository to `/opt/master-control-functions` using your approved transfer method. Include `deploy/`, both source directories and the canonical frontend lockfile. Exclude real `.env` files, `node_modules`, build outputs, graph files, `.scratch`, and database backups. Copy the selected SQLite source separately to a private Ubuntu migration directory. This runbook includes the required transfer steps and does not depend on the removed local documentation directory.

Check tools and ports:

```bash
cd /opt/master-control-functions
docker version
docker compose version
command -v curl flock ss stat
ss -ltn '( sport = :31627 or sport = :24993 or sport = :24994 )'
df -h / /data
docker system df
```

Docker and Compose are already listed in the supplied inventory. Do not reinstall them, relocate Docker storage, remove other services, or run global prune commands. The inventory showed 29 GiB free on `/`; watch build cache and old release images. `/data` had more space, but copying the source there does not move Docker's image storage.

If a selected port is occupied, stop and identify its owner. Replacing an existing native application on the same port needs a separate controlled first cutover. Do not stop PTB_OT or replace the host's global Nginx configuration.

Restrict port `31627` to approved LAN clients through the existing host/network firewall policy before starting edge. Host networking means the edge listener is exposed on the host. This app has no authentication; do not expose it to the Internet. HTTPS and access controls are prerequisites for any later public deployment.

### Build Network Access

BuildKit needs outbound HTTPS for Go modules and the frontend build. The default Go route uses `proxy.golang.org`, checksum service `sum.golang.org`, and module ZIP redirects to `storage.googleapis.com`. Bun also needs the package registry, and the configured Paraglide plugins use `cdn.jsdelivr.net`.

If host checks show connection resets before TLS completes across the Go proxies and checksum service, ask the network administrator for an approved outbound proxy or HTTPS allowlisting. GitHub access alone does not prove the complete Go module graph can download. Keep TLS verification and Go checksum verification enabled; switching to `GOPROXY=direct` still needs Git and access to other module hosts.

For unattended routine releases, create a private build profile once on Ubuntu:

```bash
cd /data/ptb/master-control-functions
if [ ! -e deploy/.env.proxy ]; then
    install -m 600 deploy/.env.proxy.example deploy/.env.proxy
fi
# Edit deploy/.env.proxy privately with the approved proxy URL.
chmod +x deploy/blue-green.sh
./deploy/blue-green.sh
```

Do not overwrite an existing configured profile with the template. The script automatically loads `deploy/.env.proxy` when building a new image; no manual `source` step is required. It refuses a profile whose mode is not `600` or still contains `replace_me`. This is a trusted shell file: do not copy settings from untrusted sources. When present, it replaces inherited `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` and lowercase equivalents for the build only. Omitted proxy settings are unset, although Docker client proxy defaults may still apply. Keep all required proxy exclusions in the profile.

Without that file, export approved `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` settings privately before deployment, as before. Both uppercase and lowercase names work; uppercase takes precedence if both are set in the selected environment. The script forwards nonempty settings to Docker as build secrets exposed only to the network-reaching build steps; proxy values never appear in the script's command arguments, image history, or build cache metadata. Loading the profile happens in a subshell, so its settings do not change the runtime Compose environment. It is separate from `deploy/.env` because Compose loads that file into the application containers at runtime.

For an approved module mirror, set `GOPROXY` in the private profile or export it in the shell. Use a credential-free URL: this custom argument may appear in build metadata. The default remains `https://proxy.golang.org,direct`; comma-separated fallback occurs only on HTTP 404/410, so a connection reset will not trigger direct downloads. An internal mirror must also support checksum verification. Build settings in `deploy/.env` are not loaded into the script's shell.

A `read: connection reset by peer` while fetching module ZIPs means the build's HTTPS connection was reset before compilation. It is not a Go compile error; other `CANCELED` stages are consequences of that failure. The log alone cannot distinguish a firewall, outbound proxy, or remote endpoint reset. In particular, a prepared `.env.proxy` file was ignored by older script versions unless manually sourced. Automatic loading fixes that configuration gap, but an approved, reachable proxy is still required on networks that block direct access. Do not substitute an unapproved public mirror, disable TLS/checksums, or assume repeated retries fix network policy.

BuildKit's cached base layers do not guarantee that `docker run --pull=never` can find a local image tag. For a temporary, database-free connectivity probe on Ubuntu:  

```bash
docker pull golang:1.27.3-alpine
docker run --rm --pull=never --entrypoint sh -w /tmp golang:1.27.3-alpine -c '
  go env GOPROXY GOSUMDB
  timeout 45 go mod download -json github.com/gofiber/fiber/v3@v3.7.0
'
```

The probe uses the container's default networking and does not load the private build profile; BuildKit's network may differ. The image build itself exercises the configured proxy secrets. The proxy address must be reachable inside BuildKit: `127.0.0.1` refers to the build container, not the Ubuntu host. Have the administrator provide an approved reachable address; do not broaden host networking or firewall access implicitly. Transfer the updated build files to Ubuntu and retry the release. A failed image build leaves the active slot and routing unchanged. Public-mirror retries cannot repair a host network policy that blocks every tested mirror.

## 2. Configure Credentials

```bash
cd /opt/master-control-functions
if [ ! -e deploy/.env ]; then
    install -m 600 deploy/.env.example deploy/.env
fi
chmod 600 deploy/.env
```

Edit `deploy/.env` privately on Ubuntu. Set the rotated PostgreSQL password in `PGPASSWORD`, inside single quotes; leave the URL password-free. Keep the file readable only by its owner. Do not send credentials to the agent or print `docker compose config` without `--quiet`. Do not put secrets in frontend build arguments or `VITE_*`/`PUBLIC_*` variables.

The example uses explicit `sslmode=disable` for same-host loopback only. Confirm that the existing server accepts authenticated loopback connections. Never change authentication to `trust`. If TLS is required, use `sslmode=verify-full` with the certificate hostname; a private CA must also be mounted read-only into both slots and referenced by `PGSSLROOTCERT`. Publicly trusted CAs are already in the runtime image.

## 3. Transfer and Verify SQLite First

Keep both the old web writer and any MCP catalog reader stopped throughout the first cutover. Retain their binaries and configuration. Identify the actual current SQLite file from the old service's `-db` argument; the repository's `.bak_before_*` files are historical backups, not proof of a current snapshot. The transfer requires a consistent version-6 snapshot. An older source must be upgraded using a staging copy and the retained SQLite-backed server, never your sole original.

Use the existing PostgreSQL server and authenticated role; no new database engine is needed. If client utilities are absent, have the operator install the approved Ubuntu `psql`, `pg_dump` and `sqlite3` packages. Load the trusted environment privately:

```bash
cd /opt/master-control-functions
set -a
. ./deploy/.env
set +a
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
  "SELECT current_database(), current_user, current_schema(), current_setting('server_encoding'), has_schema_privilege(current_user, current_schema(), 'CREATE');"
psql "$DATABASE_URL" -X -v ON_ERROR_STOP=1 -c \
  "SELECT schemaname, tablename FROM pg_tables WHERE schemaname = current_schema();"
```

Require database `master_control_functions`, encoding `UTF8`, the intended application schema and CREATE permission. The application tables must be empty. Ask the database administrator to resolve permission/TLS problems rather than weakening access rules. Never share password-bearing URLs or database diagnostics without redaction.

After confirming old writes are stopped, stage the selected source and any WAL/SHM companions. Replace `SOURCE` with the actual approved file path; do not use the example literally:

```bash
umask 077
sudo install -d -m 700 -o "$(id -un)" -g "$(id -gn)" \
  /srv/migration /srv/migration/staged
SOURCE=/approved/path/current/catalog.db
cp -- "$SOURCE" /srv/migration/staged/catalog.db
for suffix in -wal -shm; do
    if [ -f "$SOURCE$suffix" ]; then
        cp -- "$SOURCE$suffix" "/srv/migration/staged/catalog.db$suffix"
    fi
done
sqlite3 /srv/migration/staged/catalog.db \
  ".backup '/srv/migration/catalog-snapshot.db'"
chmod 400 /srv/migration/catalog-snapshot.db
pg_dump --format=custom \
  --file=/srv/migration/program-functions-before.dump "$DATABASE_URL"
```

Use new backup filenames for another attempt; do not overwrite retained evidence. Original SQLite files remain untouched. A main-file copy without required WAL data is not a valid snapshot. Keep a separate protected copy of the source and PostgreSQL dump outside the deployment directory.

Go 1.22 from the host's package inventory is too old. Use an approved Go 1.27.3+ toolchain on Ubuntu for the tagged migration build, or compile it in this temporary Go builder container:

```bash
cd /opt/master-control-functions
mkdir -p deploy/state/migration-tools
docker run --rm \
  --user "$(id -u):$(id -g)" \
  -e GOCACHE=/tmp/go-cache -e GOPATH=/tmp/go \
  -v "$PWD/backend:/src:ro" \
  -v "$PWD/deploy/state/migration-tools:/out" \
  -w /src golang:1.27.3-alpine \
  sh -ec 'CGO_ENABLED=0 go build -mod=readonly -tags sqlite_migration -o /out/migrate-sqlite ./cmd/migrate-sqlite'
```

This produces the separate transfer tool containing the legacy SQLite reader on Ubuntu. It is never copied into the application image. Have the administrator provision `master_control_functions_test` as a disposable test database for the same role, or adapt the test-only URL and credentials privately. That role must be able to create/drop isolated test schemas. Run the backend and tagged transfer regressions on Ubuntu before production transfer:

```bash
docker run --rm --network host \
  --user "$(id -u):$(id -g)" --env-file deploy/.env \
  -e 'TEST_DATABASE_URL=postgresql://ptb@127.0.0.1:5432/master_control_functions_test?sslmode=disable' \
  -e GOCACHE=/tmp/go-cache -e GOPATH=/tmp/go \
  -v "$PWD/backend:/src:ro" -w /src golang:1.27.3-alpine \
  sh -ec 'go test -mod=readonly -tags sqlite_migration ./tests/...'
```

Require real passing tests; skipped database fixtures are insufficient. Preserve failing output privately and stop on failure. For TLS/private-CA test connections, mount the CA and update test configuration just as for production. Full UI E2E reset also requires its own disposable database, never `master_control_functions`.

Load the trusted environment without printing it, then run transfer and verification with your actual snapshot path:

```bash
set -a
. ./deploy/.env
set +a
./deploy/state/migration-tools/migrate-sqlite \
  --source /srv/migration/catalog-snapshot.db --confirm-writes-stopped
./deploy/state/migration-tools/migrate-sqlite \
  --source /srv/migration/catalog-snapshot.db --confirm-writes-stopped --verify-only
```

Require both messages before continuing:

```text
Migration committed: all 11 application tables verified; SQLite source unchanged.
Verification passed: all 11 application tables match the SQLite snapshot.
```

Retain the per-table counts/digests with backup records. An uncertain commit requires verification before retrying. The transfer refuses a populated target and preserves Functions, definitions, links, Examples, import history, Sequences and Devices. Do not substitute workbook import for this transfer.

## 4. Deploy the First Release

Before reopening writes, restrict the application to operators and preserve the old SQLite-backed binaries/configuration. The initial data migration needs a controlled write pause. Blue-green switching applies to later compatible application releases.

```bash
cd /opt/master-control-functions
bash deploy/blue-green.sh deploy 20261004-01 --migration-verified
bash deploy/blue-green.sh status
curl --noproxy 127.0.0.1 --fail http://127.0.0.1:31627/health
curl --noproxy 127.0.0.1 --head --fail http://127.0.0.1:31627/import
```

The confirmation flag is an operator assertion that the transfer passed; it does not perform or independently prove migration. The server still applies embedded schema migrations at startup. Back up PostgreSQL and review pending migrations before each release. Health must return `{"status":"ok","db":"ok"}`; `/import` must return HTML. Open `http://<ubuntu-host>:31627`, verify catalog data, navigation, Examples, Sequences and workbook export. Reopen writes only after acceptance.

If the existing MCP reader is also being cut over, configure its PostgreSQL environment and remove its old `--db` argument; do not let it continue reading stale SQLite. Verify initialization and catalog tools before reopening writes. Use an administrator-provisioned read-only PostgreSQL login for that reader. This deployment does not build or manage the separate MCP service.

Before PostgreSQL accepts new writes, migration rollback means stopping the new application/MCP processes and restarting the retained SQLite-backed binaries against their untouched original data/configuration. Keep PostgreSQL and backups. After new PostgreSQL writes, the SQLite snapshot is stale: returning to it loses those writes. Stop and reconcile or restore PostgreSQL through an approved recovery procedure; there is no automatic SQLite fallback.

## Routine Releases and Rollback

After transferring updated source into the Ubuntu checkout, run the script without arguments. Like the PTB_OT deployment entry points, it builds the inactive slot and switches only after health checks. It does not fetch or pull source automatically.

```bash
cd /data/ptb/master-control-functions
./deploy/blue-green.sh
```

No arguments (or `deploy` without a tag) generates a unique UTC timestamp/nanosecond/PID tag and prints the release image and candidate slot. The initial SQLite transfer still requires explicit confirmation: `./deploy/blue-green.sh deploy --migration-verified`, only after the transfer and verification above. Running without arguments never bypasses this safeguard.

Explicit tags remain supported. Use a new tag for changed source. Existing project-owned explicit tags are reused without rebuilding; never retag a release image behind the script. A generated tag collision fails rather than reusing an image.

```bash
cd /opt/master-control-functions
bash deploy/blue-green.sh deploy 20261005-01
bash deploy/blue-green.sh status
bash deploy/blue-green.sh rollback blue
```

Select the known-good slot reported before deployment, rather than blindly using `blue`. Both slots share the same database. Rollback changes application routing only; it does not reverse migrations or restore data. Schema changes must support old and new binaries concurrently. Destructive changes need an explicit maintenance procedure.

The script locks deployments, confirms live/configured routing agrees, and waits for retiring Nginx workers before replacing an inactive slot. It starts the candidate, waits for database-backed health, copies immutable frontend assets into an additive shared store, atomically replaces the upstream file, tests Nginx configuration and gracefully reloads. It verifies the actual responding upstream before success. Failure restores the previous file and, if reload was attempted, verifies the rollback. It never force-recreates a running edge or stops the previous application slot during cutover.

`state/` is private operational state. Preserve it with both slot containers and images. Do not edit upstream files manually during rollout. SIGKILL, host loss or a failing Docker daemon can interrupt recovery; the next release refuses to proceed when configured/live routing disagree. Recover the known-good upstream privately before retrying. Edge configuration/image updates are separate infrastructure maintenance, not covered by routine zero-downtime application deployment.

Old hashed assets remain available to already-open browser tabs. They accumulate, as do release images/build cache. After your rollback retention period, remove only explicitly selected project release images that no retained container references. Asset deletion requires its own retention policy; the script performs none. Never use `docker system prune` on this shared host.

For private diagnostics, inspect `docker logs webprogfndb-blue`, `webprogfndb-green`, or `webprogfndb-edge`. Redact database errors before sharing logs; do not inspect or publish credential-bearing container environment.

## Verification Boundaries

Database-free local checks:

```bash
node --test deploy/tests/blue-green.test.mjs
go -C backend build -mod=readonly ./...
go -C backend vet -mod=readonly ./...
cd frontend
bun run check
bun run build
```

The rollout tests replace Docker and HTTP commands with fakes. They do not start services or access a database. Actual image build, Linux Compose validation, Nginx syntax/runtime, port availability, firewall policy, database transfer, catalog tests and live release/rollback smoke checks remain Ubuntu operator gates. Use `docker compose -f deploy/compose.yml config --quiet` after creating the private environment file. Passes of mocked tests do not establish real zero-downtime behavior or measured performance.

`DEPLOY_TEST_BASH` can select an installed Bash executable for the rollout tests when the default Git Bash launcher/toolchain is unsuitable. It does not affect production deployment.

The Windows `bun --bun run build` check stalled at SSR startup; the Node-backed `bun run build` completed. Docker therefore uses Bun as installer/runner while retaining Node in the builder. That local build also reported Paraglide plugin fetch failures from `cdn.jsdelivr.net`. Confirm the Ubuntu builder can reach the configured plugin URLs and check all locales before accepting its frontend output. Do not assume a successful process exit proves localization completeness.
