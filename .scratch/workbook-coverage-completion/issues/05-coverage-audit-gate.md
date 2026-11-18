# 05: Batch 9 — coverage audit + final gate

**What to build:** Prove completion. Re-run the trunk scan end to end, diff against the final registry, and produce the audit artifact: every trunk function name that still has no workbook row. Fill any last gaps the audit surfaces as a small final batch (same procedure), then run the full gate through the real app: import/reset, import change log shows only expected Normalization entries, export round-trip, orphan report reconciled with ticket 02's accepted-gap accounting, and the complete test pass (backend suite, typecheck, build, E2E) green. The audit list is the closing deliverable: completion asserted by evidence, not by claim.

**Blocked by:** 01 (stragglers), 02 (IOT decision), 04 (quality pass).

**Status:** done

- [x] Trunk scan re-run against the final registry; audit artifact lists every remaining undocumented trunk call
- [x] Gaps found by the audit appended (evidence-backed) or explicitly deferred with reasons
- [x] Import gate through the real app: import/reset change log clean, export round-trip idempotent
- [x] Orphan report reconciles with ticket 02's outcome; total matches the accepted-gap record
- [x] Full test pass green: backend suite, typecheck, build, E2E
- [x] Fixture + pinned counts final; closing commit scoped to fixture + expectation files (plus the audit artifact)

## Completion notes (2026-10-02, commit 3afaf78)

**Audit:** full trunk walk re-run with the app's join semantics (function existence by `(source, name)`; DLL prefix + `#QUICK`/`#PROXY` mode suffixes stripped; ADR-0002 scope rules). Every usage resolves to a registry row or to an explicitly recorded class: 8 SDK sample-DLL calls (`DLLSample`, incl. the `DLLDLLSAMPLE`/`#QUICK` variants — deferred ticket 01), 1 macro label (`MACRO|RUN` coverage), 3 QSPR variant spellings (documented in canonical rows), 7 scan-noise hits (C# command arrays / device-enumeration strings). **Genuinely-new undocumented production calls: 0.** Artifact: `.scratch/workbook-coverage-completion/audit.md`.

**No workbook rows changed** — no fixture sync or backend count updates needed this batch (fixture stays hash-verified at `60CBBBDD…`).

**Gate results:**
- Import/reset + change-log + export round-trip + re-import idempotency + orphan report 0/0/0 (reconciles with ticket 02): all green through the API suite.
- `go vet` clean; `go -C backend test ./...` green.
- `pnpm check` 0 errors/0 warnings; `pnpm build` green.
- `pnpm test:e2e` **25/25 green** — required re-syncing two stale E2E count pins (examples spec asserted `fixture+1` = 220 after seeding; smoke asserted `296 functions`) — included in the closing commit.

The trunk-usage coverage axis is complete; spec and tickets 01–05 closed.
