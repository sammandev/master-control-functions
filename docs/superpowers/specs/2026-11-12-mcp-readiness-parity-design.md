# MCP Readiness Parity Design

## Intent

Keep the web application's MCP readiness verdict aligned with the actual
`script-audit-mcp` catalog behavior, and ensure the MCP PostgreSQL contract
tests exercise the web application's real `mcp_v1_*` views.

## Goals

- Add a repeatable cross-repository parity check using the same disposable
  PostgreSQL schema for the web API and MCP.
- Build that schema from the web application's PostgreSQL migrations rather
  than a hand-maintained imitation of the schema and views.
- Compare each Function's web line readiness with the MCP
  `list_functions.composition_supported` result. Bundle-only readiness remains
  distinct: a web `bundle_ready` Function is not line-composable.
- Add MCP PostgreSQL test coverage for the actual `mcp_v1_*` projections,
  including interface inference for links and examples.

## Non-goals

- Replace the web readiness implementation with a runtime call to the MCP.
- Change readiness rules, catalog schema, tool behavior, or migration policy.
- Connect to or modify the approved Ubuntu/production database.

## Constraints

- The web application owns PostgreSQL migration sources; the MCP is maintained
  in a separate repository.
- Integration checks must require an explicit `TEST_DATABASE_URL` and isolate
  their objects in a temporary schema. They must not default to a deployment or
  production URL. Database-backed test runs are user-run on the approved host;
  the agent does not access that host.
- Preserve the existing fast unit tests and their synthetic fixtures.

## Approaches

1. **Cross-repository parity harness (selected).** Create a repeatable
   integration path that applies the web migrations to an isolated local
   schema, starts or invokes the web API and MCP against that schema, and
   compares Function identity plus web line readiness against MCP composition
   support. Reuse the same real migrations to exercise the MCP's view contract.
   This checks actual runtime behavior while keeping runtime architecture
   unchanged.
2. **Make the MCP the web application's readiness source of truth.** Remove
   the Go rule mirror and query the MCP at runtime. This avoids rule duplication
   but couples the web API's availability to a running MCP and changes
   production architecture; it is not needed to add regression protection.

## Design calls

- Keep the Go and Rust implementations; detect behavioral drift with a local
  integration check rather than introducing a runtime service dependency.
- Use the exact PostgreSQL migrations from the web repository as the schema
  source. Do not copy the MCP views into test-only SQL or replace existing
  fast unit fixtures.
- Compare `ready` / `line.Ready` with MCP `composition_supported`, keyed by
  source, interface, and name. Do not equate `bundle_ready` with line support.
- Make the PostgreSQL checks opt-in through `TEST_DATABASE_URL`; document the
  command and required prerequisites in both repositories for user-run
  execution on the approved host.

## Builder decisions

The implementation may choose the smallest maintainable runner and test
ownership split across the two repositories, provided it can locate the web
migrations explicitly, keeps temporary PostgreSQL objects isolated and
cleaned up, invokes the actual MCP `list_functions` behavior, and fails with
clear diagnostics when prerequisites or parity are missing.
