# 00: Environment + app skeleton

**What to build:** The project runs as a Go (Fiber) backend serving a built Svelte 5 (SvelteKit, adapter-static) frontend from a single binary, with an openable SQLite database and a migration runner in place. If the Go toolchain is absent, it is installed first under the software-license check. The whole toolchain is proven by a green automated test and a page that loads in a browser.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] Go toolchain verified (or installed with license check passed); `go version` works
- [x] Go module + Fiber server boots and serves the built SPA at the app root
- [x] Health endpoint responds; database file opens and migration runner creates an empty schema version table
- [x] SvelteKit scaffold builds to static assets and a placeholder page renders
- [x] `go test` harness runs an HTTP integration test against a temporary SQLite file and passes
- [x] README section records the dev commands (build, test, run)

## Comments

**2026-10-01 — implementation notes**

- Go license check (software-license-guard): Go/Golang absent from the OA free-software whitelist (87 items scanned) → Step 1 verification against the official source: go.dev/LICENSE confirms BSD 3-Clause (Copyright 2009 The Go Authors) — free for commercial use, no paid edition, no approval process → classified Enterprise free software → install allowed. Downloaded go1.27.1.windows-amd64.zip from go.dev/dl (official source), SHA-256 verified against the published digest (`a3911b5e...`), extracted to `%LOCALAPPDATA%\Programs\go`, user PATH registered. `go version go1.27.1 windows/amd64` confirmed.
- Frontend: SvelteKit scaffold (Svelte 5.56, runes forced) + adapter-static with `fallback: index.html`, `ssr=false` → SPA bundle at `web/build`, served by the Go binary. Placeholder page renders.
- DB: `modernc.org/sqlite` (pure Go, no cgo per ADR-0001), migration runner creates `schema_migrations`, migrations are code-defined, applied in code order inside per-migration transactions, `applied_at` written from Go as RFC3339 (ANSI-compatible — no SQLite-only SQL).
- HTTP seam tests: 8 integration tests in the api package (health, 404-JSON contract with and without SPA bundle registered, SPA root + deep-link fallback, no-bundle behavior) + 3 migration tests, all against temporary SQLite files. `go vet` clean, `svelte-check` 0 errors.
- Live-verified: `/health` → `{"db":"ok","status":"ok"}`; deep link `/functions` → 200 (SPA); unknown `/api/nope` → 404 JSON with bundle present.
- Two-axis code review (Standards + Spec sub-agents) run pre-commit; all actionable findings fixed (strftime removed, `/api/*` 404 contract registered before SPA fallback and now tested in both modes, stdlib errors.As/strings.Contains used, stale comments corrected, unused adapter-auto dep and junk meta removed). Accepted non-fixes: duplicated temp-db test setup across packages (two small call sites), sqlite_master in a driver-pinned test (marked dialect-specific).
