# Backend Dependency Audit

Date: 2026-11-10. Scope: Go backend dependencies and the configured Go builder,
not the frontend, OS/container packages, or a full application penetration test.

## Deployment Clarification And Update Status

The operator confirmed that the application is an internal-only HTTP tool and
intentionally uses PostgreSQL `sslmode=disable`. Preserve that explicit choice;
this dependency-maintenance task does not introduce HTTPS or require a company
CA. Browser HTTP and PostgreSQL transport are separate connections. Internal-only
access reduces exposure but does not prevent malformed-workbook denial of service
or provide encryption on either connection.

Finding 2 does **not** apply to certificate verification on the confirmed
plaintext database connection. Retain the Go patch as toolchain maintenance and
protection for any other TLS paths, not as a demonstrated database exploit.
Finding 4 remains conditional on reliance on the affected authentication-policy
settings and a rogue server; plaintext alone does not establish exploitation.
The missing simple-protocol batch call path in finding 3 is unchanged.

### Applied Security Updates

The operator authorized a process-local `http://localhost:3128` proxy for command
execution. Trusted Go downloads now succeed. VS Code settings and persistent
proxy environment variables were not changed; certificate verification and Go
checksum verification remain enabled. The earlier download block is resolved.

| Component | Original audit | Updated selection |
|---|---|---|
| Go minimum / Docker builder | 1.27.1 | 1.27.3 |
| Excelize | 2.11.0 | 2.11.1 |
| pgx | 5.9.2 | 5.12.0 |
| compress | 1.19.2 | 1.20.4 |
| Brotli | 1.2.2 | 1.3.0 |
| Fiber | v2 2.52.15 | v3 3.7.0 |

Go 1.27.3 was downloaded into the user toolchain cache, not installed over the
system Go executable. The module's Go directive requires that patched minimum.
Deployment and migration-builder examples also use 1.27.3.

The existing user changes to transitive dependencies were integrated, and a
follow-up migration moved all application handlers, static file serving, and API
tests from deprecated Fiber v2 (`v2.52.16`) to Fiber v3 (`v3.7.0`). fasthttp 1.78.0 uses
`github.com/molecule-man/go-brrr` for Brotli support. The patched andybalholm
Brotli 1.3.0 selection is explicitly retained for the dependency graph even
though it no longer contributes a package to the server binary.

### Remaining Compatible Updates

A separate validation phase updated SQLite to 1.62.0, go-humanize to 1.1.0,
go-runewidth to 0.0.31, go-strftime to 1.1.0, x/sync to 0.25.0, x/sys to
0.50.0, and x/text to 0.44.0. SQLite selected libc 1.80.1. Excelize selected
mscfb 1.0.9 and its compatible efp pseudo-version. Fiber v3.7.0 selects
gofiber/schema 1.9.3, gofiber/utils/v2 2.6.4, philhofer/fwd 1.2.1, and
tinylib/msgp 1.6.5.

Do not equate every `go list -m -u all` entry with an application update that must
be forced. Parent-selected dependencies without application packages, dependency
test/tool modules, and SQLite's libc coupling retain their upstream selections.
For example, libc 1.80.2 is available but was not forced beyond SQLite's 1.80.1
selection. Fiber v3 migration is complete and Fiber v2 is removed from the module
graph. No selected-module retractions were reported by
`go list -m -mod=mod -u -retracted all`.

### Scanner Results

Scanner: govulncheck 1.9.0. Database: `https://vuln.go.dev`, last updated
2026-11-05 23:49:57 UTC when checked. These results cover the database available
at scan time, not unknown issues or container OS packages.

The pre-update Linux server scan under Go 1.27.3 reported **10 reachable
Excelize vulnerabilities**, five additional imported-package findings, and one
module-only advisory. This baseline used the user's partially updated graph,
not the original audit snapshot or the old deployed Go compiler.

After the security phase, the Linux server scan reported **zero reachable
vulnerabilities and zero imported-package findings**. One module-only advisory
remains: [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932), the unmaintained
`golang.org/x/crypto/openpgp` packages. The selected x/crypto 0.58.0 is the latest
stable release observed; OpenPGP is not imported by the application. This is not
silenced or represented as a clean module-level audit. A development
pseudo-version is not forced solely to remove an unused-package warning.

### Final Verification

| Check | Result |
|---|---|
| `go mod verify` | All modules verified |
| `go build -mod=vendor ./...` | Passed |
| `go vet -mod=vendor ./...` | Passed |
| Database-free API and importer tests, vendor mode | 19 passed across two packages |
| Same tests, `-mod=readonly` | 19 passed |
| Fixture rewrite/parse regression | Parsed records and normalization results preserved |
| Oversized row after valid header regression | Returns `excelize.ErrMaxRows` |
| Linux/amd64 server and tagged migration cross-builds, `-mod=readonly`, CGO disabled | Passed; binary metadata confirms Go 1.27.3 and updated module selections |
| Linux migration-tagged `go vet -mod=vendor -tags sqlite_migration ./...` | Passed |
| Linux `govulncheck -show verbose ./...`, vendor mode | Exit 0; zero reachable and zero imported-package findings; one module-only OpenPGP warning |
| Linux `govulncheck -show verbose -tags sqlite_migration ./...`, vendor mode | Same result; SQLite 1.62.0 included |
| Docker builder wiring tests | 4 passed, including Go minimum/builder alignment |
| Full database-free deployment mock suite | 7 passed, 20 failed; Windows Git Bash subprocesses timed out (`spawnSync ... bash.exe ETIMEDOUT`) |

The full deployment mock suite is not green and is not reported as passing.
No deployment-script behavior was changed to work around that failure. The
focused builder tests pass independently. Docker is unavailable on this device,
so no container image build is claimed.

The generated vendor directory was refreshed locally but remains gitignored.
The portable changes are the manifest/checksums, regression tests, builder pin,
and documentation. The report remains local and ignored at the operator's chosen
location; unrelated worktree changes are not staged or committed.

No Ubuntu host, PostgreSQL connection, backend startup, or database migration was
performed. PostgreSQL-backed API tests and database export/import roundtrips
remain operator-run on a disposable database before deployment. Cross-builds do
not verify runtime database behavior. The deployed binary has not been replaced.

### Earlier Offline Workstation Baseline

The operator requested preserving the default intranet proxy. Follow-up checks
used process-local `GOPROXY=off` and `GOTOOLCHAIN=local`, with no network requests,
proxy changes, TLS exceptions, database connections, or backend startup.

- Offline build: `go build -mod=vendor ./...` passed.
- Offline static analysis: `go vet -mod=vendor ./...` passed.
- Database-free baseline:
  `go test -mod=vendor -count=1 ./internal/api ./tests/importer` passed:
  17 tests across two packages.
- Linux/amd64, `CGO_ENABLED=0` cross-builds passed for `./cmd/server` and
  `-tags sqlite_migration ./cmd/migrate-sqlite`. Neither executable was run.
  Temporary validation binaries were removed.
- Binary build metadata confirms the baseline used local Go 1.27.2 and the
  existing vulnerable dependency versions, not the proposed updates.
- Cache lookup could not resolve the proposed pgx, Fiber v2, Brotli, or compress
  releases offline. Excelize 2.11.1 has cached metadata but no module archive.
  No cached Go 1.27.3 toolchain or vulnerability-scanner module was found in the
  checked default download-cache locations.

These earlier results preceded the applied updates above. Do not interpret that
offline baseline as remediation. The Ubuntu host was not accessed.

## Original Audit Record

The following inventory, findings, and limitations describe the pre-update
snapshot. The applied-update sections above supersede its pending recommendations
and download limitations. Historical line references can shift after changes.

## Summary

- [go.mod](../backend/go.mod) declared 38 external modules: 4 direct and 34 indirect.
- All four direct dependencies have newer releases. Public module metadata and
  upstream release notes identify security fixes and one confirmed deprecation.
- The clearest application-level exposure is Excelize's workbook-import parser.
  Other findings below have configuration or call-path conditions.
- **This is not a completed govulncheck scan or an exhaustive vulnerability
  count.** Automated scanning, full retraction checks, and function-level
  reachability remain blocked by module-download certificate/network failures.
- No dependencies, application code, or deployment settings were changed.
  Existing worktree edits were left untouched. No Ubuntu host or database was
  accessed.

## Findings

Severity is upstream severity where available, otherwise review triage.
Confidence measures evidence for the stated finding, including its limitations;
it does not assert that every module advisory is exploitable in production.
Rows group findings by remediation, not by individual CVE count.

| # | Severity | File | Lines | Vulnerability | Confidence |
|---|----------|------|-------|---------------|------------|
| 1 | 🟠 HIGH | [workbook.go](../backend/internal/importer/workbook.go#L99) | 99-108 | Excelize 2.11.0: unbounded row iteration and additional malformed-workbook parser advisories; import code calls OpenReader and GetRows | 9/10 |
| 2 | 🟠 HIGH | [Dockerfile](../deploy/Dockerfile#L3) | 3, 31 | Go 1.27.1 lacks the compromised-intermediate certificate denylist; conditional TLS certificate-verification exposure | 8/10 |
| 3 | 🟠 HIGH | [go.mod](../backend/go.mod#L7) | 7 | pgx 5.9.2: SQL injection in non-default simple-protocol SendBatch; no application SendBatch calls found | 9/10 |
| 4 | 🟠 HIGH | [go.mod](../backend/go.mod#L7) | 7 | pgx authentication/channel-binding policy bypass under specific connection settings; verify-full mitigates server impersonation | 8/10 |
| 5 | 🟠 HIGH | [go.mod](../backend/go.mod#L22) | 22 | compress 1.19.2: malformed zstd panic advisory; affected packages occur in the server graph, untrusted decode reachability unconfirmed | 7/10 |
| 6 | ⚪ LOW | [go.mod](../backend/go.mod#L13) | 13 | Brotli 1.2.2: malformed input can panic or loop; untrusted decode reachability unconfirmed | 8/10 |

### 1. Excelize: prioritize workbook import

[CVE-2026-107212 / GHSA-jw42-f3rr-4cc3](https://github.com/qax-os/excelize/security/advisories/GHSA-jw42-f3rr-4cc3)
describes excessive CPU consumption when `GetRows` encounters an invalid row
number after a valid row. The upstream advisory assigns CVSS 3.1 score 7.5.
[ParseWorkbook](../backend/internal/importer/workbook.go#L98) uses the implicated
`OpenReader` and `GetRows` APIs. This is source-level applicability evidence,
not a malicious-workbook execution against this application.

The [2.11.1 release notes](https://github.com/qax-os/excelize/releases/tag/v2.11.1)
also list fixes for shared-string/style index panics, malformed encrypted
workbooks, excessive decryption allocation/spin counts, malformed ZIP metadata,
and other parser/formula edge cases. Not every listed API is used by this app.
For example, the
[shared-string advisory](https://github.com/qax-os/excelize/security/advisories/GHSA-48hm-4h8j-58fg)
explains an additional streaming-path panic.

**Update:** Excelize 2.11.0 -> 2.11.1. Test successful imports, malformed-input
errors, normalization, and export/import roundtrips. Do not equate an HTTP upload
size limit with protection against small files that expand or cause long loops.

### 2. Go standard-library TLS verification

[GO-2026-6809 / CVE-2026-111943](https://pkg.go.dev/vuln/GO-2026-6809)
adds a denylist for compromised intermediate certificates. The configured
production builder uses Go 1.27.1; the fixed release in this line is **1.27.3**.
The installed local Go 1.27.2 is also below that fix.

[OpenPostgres](../backend/internal/db/db.go#L25) maps absent or `require` SSL mode to
`verify-full`, making certificate verification relevant for TLS-enabled
connections. Exposure still depends on trust-store/certificate conditions and
the actual connection configuration. Explicit `sslmode=disable` does not use this
TLS path; it is not a recommended workaround.

**Update:** build with Go 1.27.3 or later and rebuild/redeploy the binary. Changing
only the runtime Alpine image does not replace the Go standard library embedded
in the executable. Actual deployed binary/toolchain versions were not inspected.

### 3-4. pgx: two conditional advisories

The [5.12.0 changelog](https://github.com/jackc/pgx/blob/v5.12.0/CHANGELOG.md)
documents two high-severity fixes:

- [GHSA-j38q-h3ph-jjpq](https://github.com/jackc/pgx/security/advisories/GHSA-j38q-h3ph-jjpq):
  simple-protocol batches can allow SQL injection when one query ends in an
  unterminated line comment and the following query contains an attacker-controlled
  string argument. Extended protocol is unaffected. No application `SendBatch`
  or simple-protocol configuration references were found in backend Go source.
- [GHSA-95fj-3fxw-g62g](https://github.com/jackc/pgx/security/advisories/GHSA-95fj-3fxw-g62g):
  `require_auth` / `channel_binding=require` policies may be bypassed when a
  rogue server skips authentication. This does not bypass TLS certificate
  verification. This app normally enforces `verify-full` but also allows an
  explicit `disable` mode. Production connection settings were not inspected.

**Update:** pgx 5.9.2 -> 5.12.0 even though exploitation is not established here.
One concrete compatibility change: custom `BatchResults` implementations,
including mocks, must implement the new `Tx` method. No such application
implementations were found. Test connection policies, transactions, query
parameters, scanning, and cancellation.

### 5. klauspost/compress

[GO-2026-6865 / CVE-2026-116591](https://pkg.go.dev/vuln/GO-2026-6865)
and [GHSA-fpmh-frq8-7rj8](https://github.com/klauspost/compress/security/advisories/GHSA-fpmh-frq8-7rj8)
describe a crafted zstd frame causing a panic. Upstream CVSS 4.0 score is 8.7.
The Go advisory includes the installed 1.19.2 in its affected range and fixes it
in **1.20.2**. The Linux server dependency graph includes `fse` and `huff0`;
this alone does not establish that attacker-controlled zstd input reaches them.

**Update:** 1.19.2 -> at least 1.20.2; latest observed is 1.20.4.
The separate older `s2.NewDict` issue is fixed before the installed version and
the `s2` package was absent from the observed production graph.

### 6. Brotli

[GHSA-c9gr-gmg4-2vqv](https://github.com/andybalholm/brotli/security/advisories/GHSA-c9gr-gmg4-2vqv)
affects versions before **1.2.7** and describes malformed-input panic/infinite
loop behavior. The maintainer classifies it as low severity. This is a Go
denial-of-service issue, not the native Brotli buffer-overflow/RCE described by
the related native-library advisory.

**Update:** 1.2.2 -> at least 1.2.7; latest observed is 1.3.0.
Confirm whether request decompression or another untrusted-input path uses the
decoder; a module import alone is not proof of exposure.

## Outdated Dependencies

Versions below are those observed in the public Go module proxy/release metadata
on the audit date. These are individual available versions, not a tested,
mutually compatible lock graph. Do not replace every version mechanically.

### Direct dependencies and toolchain

| Dependency | Current | Latest observed | Expected project effect |
|---|---|---|---|
| Go builder | 1.27.1 | 1.27.3 | Security patch; rebuild binary |
| Fiber v2 | 2.52.15 | 2.52.16 | Same-major patch; v2 officially deprecated |
| pgx v5 | 5.9.2 | 5.12.0 | Security fixes; BatchResults interface change |
| Excelize v2 | 2.11.0 | 2.11.1 | Security patch; malformed files may now return errors |
| modernc SQLite | 1.60.1 | 1.62.0 | Migration-only dependency; test SQLite transfer fixtures |

SQLite is absent from the normal server graph and present in the
`sqlite_migration`-tagged migration command graph. Do not treat it as the
production database driver. Production PostgreSQL itself was not audited.

### Indirect dependencies with newer versions

| Module | Current | Latest observed |
|---|---|---|
| github.com/andybalholm/brotli | 1.2.2 | 1.3.0 |
| github.com/dustin/go-humanize | 1.0.1 | 1.1.0 |
| github.com/gofiber/fiber/v3 | 3.5.0 | 3.7.0 |
| github.com/gofiber/schema | 1.8.3 | 1.9.3 |
| github.com/gofiber/utils/v2 | 2.4.1 | 2.6.4 |
| github.com/jackc/puddle/v2 | 2.2.2 | 2.2.3 |
| github.com/klauspost/compress | 1.19.2 | 1.20.4 |
| github.com/mattn/go-colorable | 0.1.15 | 0.1.16 |
| github.com/mattn/go-runewidth | 0.0.16 | 0.0.31 |
| github.com/ncruces/go-strftime | 1.0.0 | 1.1.0 |
| github.com/philhofer/fwd | 1.2.0 | 1.2.1 |
| github.com/richardlehane/mscfb | 1.0.7 | 1.0.9 |
| github.com/rivo/uniseg | 0.2.0 | 0.4.7 |
| github.com/tinylib/msgp | 1.6.4 | 1.6.5 |
| github.com/valyala/fasthttp | 1.73.0 | 1.78.0 |
| github.com/xuri/efp | 0.0.1 | 0.0.2 tagged; Excelize 2.11.1 requires a newer 0.0.3 pseudo-version |
| golang.org/x/crypto | 0.54.0 | 0.58.0 |
| golang.org/x/net | 0.57.0 | 0.62.0 |
| golang.org/x/sync | 0.23.0 | 0.25.0 |
| golang.org/x/sys | 0.48.0 | 0.50.0 |
| golang.org/x/text | 0.40.0 | 0.44.0 |
| modernc.org/libc | 1.77.1 | 1.80.2 |

The other 12 declared modules did not show a newer version in the checked
metadata. In particular, the installed `xuri/nfp` pseudo-version is newer than
its latest stable tag; do not "update" it backwards to that tag.

## Deprecation and Breaking Changes

### Fiber v2 -> v3: migration completed

The [v2.52.16 module declaration](https://proxy.golang.org/github.com/gofiber/fiber/v2/@v/v2.52.16.mod)
explicitly deprecates v2 in favor of v3. Deprecation is not itself a vulnerability
and does not prove a specific end-of-support date.

The [official v3 migration guide](https://docs.gofiber.io/next/whats_new/)
documents source and behavioral changes:

- Handler contexts change from `*fiber.Ctx` to the `fiber.Ctx` interface.
- `BodyParser` becomes `Bind().Body`; other binding APIs and some tags change.
- Middleware imports, configuration, and access to middleware-provided values
  change; revalidate error handling, uploads, request limits, routing, and tests.

The backend has since migrated its handlers, binding calls, middleware imports,
and Fiber context types to v3. The application dependency is now
`github.com/gofiber/fiber/v3` v3.7.0; the previous v2 dependency is removed.
The migration was validated with backend build, vet, and test runs. This does
not establish deployment or production-traffic compatibility.
Fiber v2.52.16 also fixes a CORS wildcard-origin edge case, but the inspected
server graph did not include Fiber's CORS middleware.

### Other update risks

- [pgx 5.12.0](https://github.com/jackc/pgx/blob/v5.12.0/CHANGELOG.md):
  custom `BatchResults.Tx` requirement, stricter authentication, pool cancellation
  changes, and query/type-handling fixes. Same major does not guarantee no
  behavior changes.
- [Excelize 2.11.1](https://github.com/qax-os/excelize/releases/tag/v2.11.1):
  Go 1.26 minimum; stricter malformed-file handling can change error results.
- SQLite 1.62.0, fasthttp 1.78.0, and the checked latest `golang.org/x/*` updates
  declare Go 1.26 minimums. pgx 5.12.0 requires Go 1.25. The configured 1.27 line
  satisfies these, but must still receive its own security patch.
- `golang.org/x/*`, `uniseg`, `go-runewidth`, and other pre-1.0 modules do not
  provide a v1+ major-version compatibility promise.
- SQLite/libc and Excelize/parser transitive dependencies should follow their
  owning module's graph. Prefer supported upstream selections over arbitrary
  direct pinning or replacement.
- This workstation has a local, gitignored [vendor/modules.txt](../backend/vendor/modules.txt).
  The original review incorrectly described it as committed.
  Updating only the manifest or checksums leaves offline/vendor-mode builds
  inconsistent. Regenerate and review vendored changes in any implementation.
- No complete module-retraction clearance is claimed: the normal Go metadata
  workflow could not finish.

## Checks and Limitations

Executed by the review:

| Check | Outcome |
|---|---|
| `go version` | Local Go 1.27.2, Windows/amd64 |
| Inspect manifest, vendored module inventory, Docker builder | Builder Go 1.27.1; 38 declared external modules |
| Linux/amd64 `go list -deps -mod=vendor` for server | Dependency graph available; not a function-level vulnerability scan |
| Tagged migration command dependency listing | SQLite isolated to migration-tagged graph |
| `go run golang.org/x/vuln/cmd/govulncheck@latest -version` | Failed downloading tool: certificate parse error involving a negative serial number |
| `go list -m -mod=readonly all` | Module fetch failed certificate validation |
| `go list -m -u all` | Default vendor mode prevents normal update lookup |
| Public Go proxy, Go vulnerability DB, upstream release/advisory checks | Available through web retrieval; used for this report |
| Supplementary shell OSV batch/update queries | Blocked by unavailable configured proxy; direct HTTPS retry also failed |
| Repository modification | Report only; no code or dependency upgrades |

No certificate-validation bypass was used. No malicious-input exploit tests,
dependency-upgrade build/tests, deployed-image scan, or production connection
tests were performed.

## Recommended Order

1. Restore trusted module-download/scanner connectivity. Run `govulncheck` against
   the Linux production target and separately the migration build tag. If scanning
   under a newer local compiler, also inspect the production binary/toolchain:
   local source analysis alone can miss an older deployed standard library.
2. Security-first change: patch Go, Excelize, pgx, Brotli, and compress. Review
   the Fiber v3 migration as part of the current dependency graph. Resolve the
   graph with Go tooling and regenerate vendor; do not run a blanket unreviewed
   upgrade.
3. Verify unit/API/import/export tests, build Linux server and migration targets,
   compare dependency graphs, and repeat vulnerability scanning. PostgreSQL-backed
   validation remains operator-run on the approved host.
4. Update remaining same-major dependencies in coherent groups with regression
   tests, including fasthttp and pre-1.0 modules.

This report's original dependency findings and recommendations reflect the
reviewed graph at that time. The Fiber v3 migration has since been implemented
and validated locally; this report does not claim that all proposed dependency
updates have been tested together or deployed.
