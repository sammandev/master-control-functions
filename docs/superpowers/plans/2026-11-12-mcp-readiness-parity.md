# MCP Readiness Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add regression protection that compares web line readiness with MCP composition support and tests the MCP against the web application's actual PostgreSQL migrations and views.

**Architecture:** Keep both runtime implementations and their existing unit fixtures. Add a migration-backed Rust PostgreSQL test fixture, then an opt-in cross-repository integration runner that builds the Go API, starts it against the same temporary schema, and compares its readiness endpoint with the MCP engine's `list_functions` result.

**Tech Stack:** Go, Rust, Python standard library, PostgreSQL, existing Fiber API and Rust MCP catalog.

**Spec:** [2026-11-12-mcp-readiness-parity-design.md](../specs/2026-11-12-mcp-readiness-parity-design.md)

## Global Constraints

- “Integration checks must require an explicit `TEST_DATABASE_URL` and isolate their objects in a temporary schema. They must not default to a deployment or production URL.”
- Database-backed test runs are user-run on the approved host; the agent does not access that host.
- “Preserve the existing fast unit tests and their synthetic fixtures.”
- The comparison is keyed by source, interface, and name; `ready` maps to `composition_supported`, while `bundle_ready` does not.

## Review Focus

- Missing `TEST_DATABASE_URL` or migration-directory input must fail clearly before starting the API or leaving a schema behind; test the runner's required-argument validation.
- `bundle_ready` must not be reported as line composition support; include a bundle-only seeded function and assert MCP support is false.
- Duplicate `(source, name)` Functions across interfaces must remain distinct in parity matching; seed same-name Functions and compare full natural keys.
- Link/example interface inference must return an interface only for a unique Function; assert unique and ambiguous cases through the real views.
- A catalog field or grammar discrepancy must fail parity even if total Function counts match; compare each Function's readiness rather than aggregate counts.

---

### Task 1: Test MCP against the real web catalog contract

**Files:**
- Modify: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\MCP_Server\mcp-program-functions\src\postgres.rs`
- Modify: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\MCP_Server\mcp-program-functions\README.md`
- Source migrations: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\Web_Program_Functions_Database\backend\internal\db\postgres_migrations\*.sql`

**Interfaces:**
- Consumes: `TEST_DATABASE_URL` and `WEB_POSTGRES_MIGRATIONS_DIR`.
- Produces: a test-only fixture that applies the sorted SQL migrations in a unique PostgreSQL schema and drops that schema on cleanup.

- [ ] **Step 1: Add a failing PostgreSQL contract test**

Add a test named `postgres_snapshot_reads_real_mcp_v1_views` that uses the web migration directory, inserts fixtures through the real tables, snapshots the catalog, and asserts:
  - audited grammar and key rows are read from `mcp_v1_functions` and `mcp_v1_script_key_audits`;
  - parameter/criteria definitions and links map correctly;
  - unique link/example keys infer their interface, while ambiguous keys yield an empty interface.

- [ ] **Step 2: Run the targeted test and verify it fails before the fixture exists**

Run from the MCP repository:

```bash
export TEST_DATABASE_URL='<disposable test database URL>'
export WEB_POSTGRES_MIGRATIONS_DIR='<web repo>/backend/internal/db/postgres_migrations'
cargo test postgres_snapshot_reads_real_mcp_v1_views
```

Expected: the new test fails because the current fixture creates simplified schema/views instead of applying the web migrations.

- [ ] **Step 3: Implement the migration-backed fixture in `src/postgres.rs`**

Create a separate fixture constructor for migration-backed tests. Read only `.sql` files from `WEB_POSTGRES_MIGRATIONS_DIR`, sort by migration version, execute them in the fixture's isolated schema, and retain the current synthetic fixture for fast existing tests. Report missing/invalid paths and SQL errors; do not silently fall back to `tests/fixtures/schema.sql`.

- [ ] **Step 4: Run the real-view test and existing PostgreSQL tests**

The user runs the targeted test on the approved host, then:

```bash
cargo test postgres_snapshot_
```

Expected: the actual-view test and existing snapshot tests pass; the new test verifies ambiguous interface inference through the real views.

- [ ] **Step 5: Document the migration-backed test inputs**

Document that the user must run these database-backed tests on the approved host with `TEST_DATABASE_URL` pointing to a disposable test database and `WEB_POSTGRES_MIGRATIONS_DIR` pointing at the web repository's PostgreSQL migration directory. The agent must not connect to that host.

### Task 2: Add cross-repository readiness parity runner

**Files:**
- Create: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\MCP_Server\mcp-program-functions\deploy\mcp-parity.py`
- Modify: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\MCP_Server\mcp-program-functions\src\postgres.rs`
- Modify: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\MCP_Server\mcp-program-functions\README.md`
- Modify: `C:\Users\Samuel_Halomoan\Desktop\opencode_project\Web_Program_Functions_Database\backend\README.md`

**Interfaces:**
- Consumes: `TEST_DATABASE_URL`, `WEB_POSTGRES_MIGRATIONS_DIR`, and a temporary Go API server binary built by the runner.
- Produces: an opt-in ignored Rust integration test that compares per-Function web readiness with `Engine::list_functions` against one migration-backed schema.

- [ ] **Step 1: Add a failing ignored parity integration test**

Add `postgres_readiness_matches_mcp_list_functions` in `src/postgres.rs`. Build a migration-backed fixture with line-ready, bundle-only, and not-ready Functions; start the API binary supplied through `WEB_API_BINARY` on loopback; query `/api/mcp-readiness`; compare each API item keyed by `(source, interface, name)` to `Engine::list_functions(None)`. Assert `status == "ready"` equals `composition_supported`, and assert that bundle-only records are not line-supported. Ensure the API child process is stopped and the fixture schema is dropped even on assertion failure.

- [ ] **Step 2: Run the targeted parity test to verify its unmet prerequisites fail clearly**

Run the ignored test directly without `WEB_API_BINARY`.

Expected: it is excluded from normal `cargo test` and reports a clear missing-binary error when explicitly invoked.

- [ ] **Step 3: Create `deploy/mcp-parity.py`**

Implement a standard-library Python runner that:
  - requires nonempty `TEST_DATABASE_URL` and `--web-repo`;
  - verifies the web migration directory exists;
  - builds `backend\cmd\server` into a temporary directory;
  - sets `WEB_POSTGRES_MIGRATIONS_DIR` and `WEB_API_BINARY`;
  - invokes only `postgres_readiness_matches_mcp_list_functions` with `--ignored`;
  - returns a nonzero exit code for build/test failure, removes temporary files, and never prints database credentials.

- [ ] **Step 4: Run the parity runner against an authorized local disposable database**

Run from the MCP repository:

```bash
export TEST_DATABASE_URL='<disposable test database URL>'
python deploy/mcp-parity.py --web-repo '<path to Web_Program_Functions_Database>'
```

The user runs this on the approved host. Expected: the runner applies the web migrations in an isolated schema and reports parity for every Function; intentionally changing a seeded readiness result makes the comparison fail.

- [ ] **Step 5: Document the one-command workflow in both repositories**

Add the runner command and prerequisites to both READMEs. State explicitly that the command is user-run on the approved host, requires a disposable test database, and must not use the production catalog.
