# 05: MACRO family + import/orphan gate

**Status:** done

Two parts.

**Part 1 â€” MACRO family.** Device prefix `MACRO` (`General_DUT\Script.cs:2286-2289` â†’ `cRunMacro`, `GeneralFunction.cs:5048`): the function name is the macro label (e.g. MARCOT7) defined as a section in `Scripts\MACRO_LIST.ini`; params forwarded as `P1=...&P2=...`, criteria as `C1=...&c2=...`. Decide the representation with the maintainer before writing rows: (a) one generic `function_main` row `MACRO|RUN` with a `MACRO_NAME` parameter, or (b) one row per macro section. Do not pollute the dictionary with per-station macro names beyond one or two representative examples.

**Part 2 â€” Verification gate.** After all content tickets:
1. Re-import the workbook through the app (`POST /api/import/reset` against a seeded database) and confirm the import change log records only expected Normalization entries.
2. `GET /api/orphans` â€” total must be 0 (parameter-links / criteria-links / examples all clean).
3. Run the full test pass (`go -C backend test ./... ; pnpm check ; pnpm build ; pnpm test:e2e`) â€” the E2E suite drives the real workbook fixture.
4. Export round-trip (`GET` export endpoint) must regenerate the six-sheet workbook without data loss (ticket 08 contract).

**Blocked by:** 01, 02, 03, 04

**Acceptance:** decision recorded for Part 1; all four gate checks green.

## Comments

- Done 2026-10-02. Part 1 decision recorded: generic MACRO|RUN row (maintainer approved). Applied to the merged canonical workbook: 1 function_main row (BUILT-IN/MACRO/RUN, routing + per-station-macros note in description), 1 MACRO_NAME parameter definition, 1 param link, 1 criteria link (CRITERIA_ARGUMENTS), 2 representative examples. MACRO_LIST.ini sections (MARCOT1..7+) stay per-station data, not registered individually.
- Gate 1 (import): functions=260, parameter_definitions=176, parameter_links=505, criteria_definitions=52, criteria_links=180, examples=145; normalization log = the 7 expected entries only (4 pre-existing + 3 PTest->PTEST device-scope uppercases from ticket 04).
- Gate 2 (orphans): total 0 (all three kinds).
- Gate 3 (full test pass): go -C backend test ./... PASS; pnpm check 0 errors; pnpm build OK; pnpm test:e2e 6/6 PASS.
- Gate 4 (export round-trip): exported six-sheet workbook re-imported into a fresh DB -> identical counts (52/180/145/260/176/505), normalizations: null (export is clean), orphans 0. Ticket 08 contract holds.
- Pre-existing drift fixed en route (not caused by tickets 01-04): the E2E smoke suite still pinned the OLD fixture reality (76 functions, orphans 155/64/16) while the fixture + go-test pins had been updated by a previous session (fixture now: 179 functions, orphans 1/0/0). Updated the two smoke.spec.ts pins and aligned the orphan-report assertions with the actual UI contract (zero-count groups render toolbar buttons only, no section heading).
- Scope note: the E2E fixture (backend/internal/importer/testdata) still holds the pre-expansion workbook; syncing it with the merged canonical (260 functions) is a separate follow-up for the app repo, not part of this expansion effort.
