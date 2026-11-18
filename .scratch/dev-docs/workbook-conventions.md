# Workbook conventions

How the script-definition knowledge workbook is grown and kept trustworthy. Written for the next person (or agent) appending batches, so the discipline does not have to be reverse-engineered from the data. Domain vocabulary lives in [CONTEXT.md](../CONTEXT.md); architectural decisions in [docs/adr/](adr/). This document describes what the committed batches actually did (batches 1–7 plus the Script.cs-axis expansion tracked in `.scratch/workbook-ecosystem-expansion/`).

## Files involved

- The **canonical workbook** is `Master_Control_V2_Script_Definition_Knowledge_merged.xlsx` in `.dev-reference/` (gitignored). It is the single write target (ADR-0004); the SQLite database is never edited directly.
- The **importer fixture** `backend/tests/testdata/Master_Control_V2_Script_Definition_Knowledge.xlsx` is the only tracked copy. After every batch it is copied from the canonical workbook and hash-compared before the suite runs.
- The **expectation files** take their row-count expectations from the fixture itself through the fixture oracle (`backend/tests/support/workbook/oracle.go`, which parses the testdata workbook with the same `ParseWorkbook` the importer uses): `backend/tests/importer/workbook_test.go` (parser counts), `backend/tests/api/import_test.go` (import/reset counts), `backend/tests/api/export_test.go` (re-import idempotency counts). The exception is `backend/tests/api/orphans_test.go`, which pins a true invariant rather than a workbook count: the orphan report must stay `0/0/0`.

## Script-line layouts

Every documented call is a pipe-delimited string quoted inside a script line (`KEY = "DEVICE|FUNC|…"`). Two layouts are in production use:

**Convention 1 — 5-segment operations.** `DEVICE|FUNC|PARAM|CRITERIA|WAIT_TIME`. Used by the plain session operations (OPEN, CLOSE, SENDSTR, SENDSTR_AND_WAITSTR, WAITSTR on TELNET / SSH / RS232 / CMD_CONSOLE), by single-shot SFIS writes, and by a handful of TEST PROGRAM and Common calls — 41 functions in the current registry. Example: `SSH|OPEN|DUT@192.168.1.101,192.168.1.1:22,pegacell,pega#1234|root@OpenWrt:~#|5000`. Session payloads `@`-split `PARAM` into `<session_name>@<payload or regex>`; `CRITERIA` is the expected prompt or `NULL`.

**Convention 2 — flow-controlled operations.** `DEVICE|FUNC|PARAM|CRITERIA|WAIT_TIME|PASS_PATH|FAIL_PATH|ERROR_CODE|DESCRIPTION`. The canonical layout for everything flow-controlled: the rest of the session/console families (RECV_*, CMD_SINGLE, CMD_CONSOLEV2), comm/DLL families (QSPR, DLLCommon checks, SFIS queries, IPLAS, FTP, HTML), TEST PROGRAM flow functions (LOOP, GOTOITEM, SCAN_BARCODE, …), and the instrument families (GPIB_DEVICE, DAQ, VIDEOMASTER, SPECTRUM, CISCO, RS_SFE100, NI6008). Call sites may appear truncated at 7 fields (no error code or description) — still convention 2. Beware the boundary: `DLLCommon|TURN_OFF_SERVICE|ServiceName=WwanSvc|NULL|3000` is 5 fields and registered convention 1, while a 7-field `DLLQSPR|PARSE_XML|…|20000|CHECH_CRITERIA_TXPower_2dBm|FAIL|E00VU000003|PARSE_XML` is convention 2.

**Which applies:** a function row's `convention` column reflects the segment count of its call sites — count the pipes in the usage evidence, not the family.

## Source taxonomy

`function_main.source` is exactly one of `BUILT-IN` or `DLL` (CONTEXT.md, **Source**):

- **`BUILT-IN`** — comm/console and TEST PROGRAM families implemented in the engine's dispatch switch (`General_DUT\Script.cs` `CreateCmd()`, plus the MasterControl-local families `AEROFLEX_PPM_16`, `ECHANNEL`).
- **`DLL`** — device drivers and instruments reached through the `DLL*` prefix (`General_DUT\DLLControl.cs`): the prefix is stripped and `Func|Param|Criteria|WaitTime` is forwarded verbatim to the dynamically loaded DLL (`Global.dLLControl.dyn(...)`). `#QUICK` loads in-process, `#PROXY` routes through the named-pipe host (`MasterControl\Proxy\ProxyControl.cs`). Registered interface names drop the `DLL` prefix and the mode suffix: device string `DLLQSPR#QUICK` → interface `QSPR`; `DLLCommon` → `Common`; `DLLPTest#PROXY` → `PTest`.
- **Usage-only rows** (no dispatch case and no DLL source in the trunk — e.g. `CADB|CA_ACCESS`, `CELLULAR|MRRESETINSTR` from the older engine generation) are registered under `BUILT-IN` with an explicit "documented from call sites only" sentence in the description. Never invent an interface the engine does not dispatch — `IOT`, for example, is a device scope, not an interface (ADR-0002; see `.scratch/workbook-coverage-completion/issues/02-iot-interface-batch-8.md` for the full evidence hunt).

`interface` carries the family's device string exactly as dispatched (`TEST PROGRAM` keeps its space). Function identity is `(source, interface, name)`; same-name functions on different interfaces are legitimate and all registered (e.g. `OPEN` on TELNET, RS232, SSH, CMD_CONSOLE).

## Evidence hierarchy

When documenting a function, prefer sources in this order and say which you used:

1. **Usage line** — the verbatim pipe-delimited call from a trunk script (`STB_spec.ini`, `MAIN_FLOW.ini`, `MACRO_LIST.ini`). This is the top evidence; examples are harvested from it.
2. **Implementation source** — the C# case label or DLL call site (`Script.cs` line ranges, `DLLControl.cs`, `ProxyControl.cs`). Grounds semantics when the usage line is opaque.
3. **Inference** — only where 1 and 2 are silent, and it must be named as such in the description (the QSPR family's closing sentence "Source of QSPR.dll is not in this repository, so semantics are documented from call sites only" is the template).

Descriptions follow the established verbose two-to-four sentence style; categories reuse the existing vocabulary (`communication`, `utility`, `instrument_control`, `logging`, `data_processing`, `system`, … — see `function_main`).

## Parameter and criteria definitions

- **Reuse first.** Before creating a definition, look for an existing one with the same semantics (the batch-3/6 cross-family reuse set: `EXPECTED_RESPONSE`, `VARIABLE`, `GPIB_COMMAND`, `EXPECTED_VALUE`, `QSPR_ROOT`, `QSPR_REPORT_DIR`). The dictionary stays lean; duplicate definitions with matching semantics are a quality bug.
- **Device-scoped rows** (ADR-0009): a definition is keyed by `(name, device_name)`; a blank `device_name` is the global definition and a device-scoped row overrides it for that device's links. The same logical parameter may legitimately have one row per device scope (e.g. `APT`, `DUT_IP`, `RF_PORT` each exist scoped `WIFI_TEST` and `IOT`).
- **Spec-style multi-sub-field segments:** when one parameter segment packs several sub-fields (comma- or `&`-joined), either (a) map each well-named key to its own link, as the QSPR family does (`Root=…&ReportDir=…` → links `QSPR_ROOT` position 1, `QSPR_REPORT_DIR` position 2), or (b) when the sub-fields are opaque or their grammar is uncertain, use **one loose `MULTIPLE` definition** with a `<a>,<b>[,...]` pattern instead of inventing a grammar without evidence (e.g. `DAQ_MEASURE_SPEC` `<device>[,<parameter>...]`, `SPECTRUM_SEARCH_SPEC` `<board>,<device>,<parameter>...`, `CADB_ACCESS_SPEC`).
- **`value_pattern`** is derived from real usage, in the established style (`<path>`, `<limit>dB`, `<lower_limit>~<upper_limit>PPM`).
- **Sparse name-only parameter rows are left alone** unless evidence clarifies them — speculation does not replace honest gaps.
- **Criteria reuse:** every function whose second segment is a response expectation uses the shared `EXPECTED_RESPONSE` definition (device-scoped per family) rather than new per-function criteria.

## Example rows

- Harvested **verbatim** from trunk `.ini` files, deduplicated by shape (the same call with different labels is one example).
- The `meaning` column paraphrases the flow targets, beeps, and verdicts from the same line.
- **DLL example-source rule (batch-6 correction):** example rows for DLL families carry `DLL` as `function_source` and the interface name in `device_name` — the raw device string (with `DLL` prefix and mode suffix) appears only inside the `script` text. E.g. the `QSPR|XML_PARSE` example is stored as `('DLL', 'QSPR', 'XML_PARSE', 'DLLQSPR#QUICK|XML_PARSE|…', …)`.

## Orphan semantics (reconciliation)

The orphan report (`GET /api/orphans`, ADR-0003) checks **function existence by `(function_source, function_name)`** and **definition existence by `(name, device_scope)` with blank-scope fallback**. A link row's `device_name` is a device scope, not the interface — a naive `(source, interface, name)` 3-tuple join reports phantom orphans (the 15 `IOT`-scoped links resolve via the `WIFI_TEST`-interface functions). Any coverage audit or count reconciliation must use the app's join, or it re-reports resolved rows.

## The standard batch procedure

Every content batch runs the same cycle; the batch commit touches **only the fixture copy** (count expectations re-derive from it automatically — see step 6):

1. **Backup** — dated copy of the canonical workbook (e.g. `*.pre_batch7.xlsx`) before touching anything.
2. **Pre-flight assertions** — no duplicate natural keys among new rows, every reused definition exists, no pre-existing rows for the target functions.
3. **Append** — see the append-script pattern below.
4. **Verify additions-only** — reload backup + new workbook; every pre-existing row byte-identical, only appends; duplicate natural-key scan; join-integrity check (0 new orphans under the app's semantics).
5. **Fixture sync** — copy canonical workbook → importer fixture; SHA-256 compare the two before running anything.
6. **Count expectations** — nothing to edit: the expectation files assert against the fixture oracle, so a batch requires zero test edits. Record the new totals in the registry growth log below.
7. **Test cycle** — `go vet` + full backend suite (`go -C backend test ./...`) once at the end; single-package probe tests during development are deleted before commit.
8. **Commit** — message convention `chore: sync fixture + expectations to merged master (batch N - <families>)`; only the fixture (plus this document's growth log) in the diff.

### Append-script pattern

- **Temp-file Python, never inline shell commands** — inline one-liners mangle quoting and drop formatting.
- **Per-column style copying** — load the workbook with `openpyxl` (not `read_only`), append after the last row, and copy `font`/`border`/`fill`/`alignment`/`number_format` from a reference row (the current last row) column by column, so the new rows look like the rest of the sheet.
- **Bool-type matching** — `required` columns must be written as real Python booleans (`True`/`False`), not strings, or the importer's expectations drift.
- Assert sheet headers before appending (`function_main`, `parameter_main`, `criteria_main`, `function_parameter`, `function_criteria`, `example`) so a workbook layout drift fails loudly instead of silently misplacing columns.

## Registry growth log

| Batch | Families / change | Registry after |
|---|---|---|
| 1 | SSH session family | — |
| 2 | TEST PROGRAM data operations | — |
| 3 | TEST PROGRAM stragglers + NI6008 (+ team expansion) | — |
| 4 | DUT console (OPEN_PORT … RECEIVE_RSP) | — |
| 5 | GPIB_DEVICE + DAQ audio | — |
| 6 | VIDEOMASTER, SPECTRUM, CISCO, RS_SFE100 | 289 functions / 199 params / 53 criteria / 542 links / 184 criteria links / 174 examples |
| (ADR-0009 split) | device-scoped definition rows expanded one-per-scope | 290 / 261 / 70 / 562 / 192 / 178 |
| 7 | DLLCommon family + CADB/CELLULAR usage-only rows; DLLSAMPLE, MACRO|MARCOT7, QSPR typo-variants explicitly deferred | 296 / 270 / 71 / 572 / 193 / 185 |

Pre-batch-7 committed counts are pinned in the expectation files' git history (`a41afac` … `0212579`).
