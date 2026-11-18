# 02: Batch 8 — IOT interface decision

**What to build:** Settle the 18 pre-existing `DLL|IOT` orphan link rows. First hunt for implementation evidence of the IOT interface functions (device-prefix dispatch candidates and versioned-component listings in the trunk sources). If real evidence exists, register the IOT functions properly and the orphan links resolve. If evidence does not exist in reachable sources, record IOT as an accepted known gap — an explicit decision entry, not silence — so the orphan report stays trustworthy. Either way, sync the fixture and pinned counts if row counts change, and reconcile the orphan report.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] Evidence hunt performed against the trunk sources; outcome (evidence found / accepted gap) recorded
- [x] If evidence found: IOT functions registered per the established row conventions (source taxonomy, description style) and the 18 orphan links resolve through the normal resolution flow — **N/A, see outcome**
- [x] If no evidence: accepted-gap decision written into project docs and the orphan report's known-gap accounting distinguishes these from batch-introduced orphans
- [x] Workbook append (if any) verified additions-only against a dated backup, pre-flight assertions passing — **no rows changed, nothing to append**
- [x] Importer fixture + pinned counts synced and full backend suite green (only if rows changed) — **rows unchanged, suite already green**
- [x] Orphan report total reconciles with the decision: zero unaccounted orphans

## Outcome (2026-10-02): IOT was never an orphan gap — it is a device scope

**Evidence hunt results:**

1. **No IOT dispatch exists.** `General_DUT\Script.cs` (the engine's `CreateCmd()` switch) has no `IOT` case; no `DLLIOT` device string occurs anywhere in the trunk sources, so the `DLL*` → `DLLControlCmd` path cannot dispatch it either.
2. **No IOT component exists.** `RD_SETTING.ini` `[Dll_Version]` (all script-folder variants) lists the versioned DLL components: QSPR, Common, DLLSample, DLLStruct, DLLTestProgram, WriteCSV, OSP, TESTPROGRAM — no IOT.
3. **The links are well-formed scope rows.** The 15 `device_name='IOT'` parameter links reference TESTER_REGISTER / TESTER_CONNECTION / DUT_INSERT, which all exist as `DLL|WIFI_TEST` functions; every referenced parameter definition exists in BOTH `IOT` and `WIFI_TEST` scopes (APT, DUT_IP, IP_ADDRESS, LOGIN, REQUEST_TIMEOUT, RF_PORT, RF_ROUTE, SRCNIC_IP, TESTER_CONTROL_METHOD, TESTER_TYPE).

**Decision (corroborates ADR-0002, "device scope is an attribute on link rows, not an entity"):** the spec's premise — "18 pre-existing `DLL|IOT` orphan links" — was written under a 3-part join reading of the data. Under the app's actual orphan semantics (function existence by `(function_source, function_name)`; definition existence by `(name, device_scope)` with global blank-scope fallback; ADR-0002), every IOT link resolves: `IOT` is the device scope a station used when running the WIFI_TEST-family DUT-control functions, with IOT-scoped parameter overrides. Registering phantom `DLL|IOT` function rows would misrepresent the source taxonomy and invent an interface the engine never dispatched — rejected per the spec's "do not invent".

**Reconciliation:** `GET /api/orphans` pins 0/0/0 for the current fixture (`TestOrphanReportFromRealWorkbook` green, re-confirmed 2026-10-02); no batch-introduced orphans exist to distinguish. Nothing left unaccounted.

**Carry-over for ticket 05 (coverage audit):** the audit's diff must use the app's join semantics — function existence by `(source, name)`, definition existence by `(name, scope)` or global — not a naive `(source, interface, name)` 3-tuple, or it will re-report these 15 links as missing.
