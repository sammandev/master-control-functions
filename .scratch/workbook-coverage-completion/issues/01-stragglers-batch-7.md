# 01: Batch 7 — low-usage stragglers

**What to build:** Re-run the trunk diff against the workbook's function registry and append every remaining low-usage script function that production trunk scripts call but the workbook still lacks. Each new row carries evidence per the established hierarchy (usage line > implementation source > inference). After the append, run the established batch cycle: workbook backup → pre-flight assertions → append → additions-only verify → importer fixture sync → pinned counts in the four expectation files → full backend suite → commit only the fixture and expectation files.

**Blocked by:** None (can start immediately).

**Status:** done

- [x] Trunk scan re-run; every missing function name from the diff is either appended with evidence or explicitly listed as deferred
- [x] Pre-flight assertions pass: no duplicate natural keys, reused definitions exist, no pre-existing rows for target functions
- [x] Append verified additions-only (+N −0) against the dated backup
- [x] Importer fixture re-synced from the canonical workbook; fixture hash verified before the suite runs
- [x] Pinned import counts in the four expectation files updated; full backend suite green
- [x] Batch commit touches only the fixture copy and the four test/expectation files

## Completion notes (2026-10-02, commit aae837e)

**Appended (+6 functions / +9 parameter definitions / +1 criteria definition / +10 parameter links / +1 criteria link / +7 examples):**

- `DLL|Common` family (dispatch: `DLL*` prefix → DLLControlCmd verbatim forwarding; DLL source not in trunk, semantics from call sites): TURN_OFF_SERVICE, TURN_ON_SERVICE (all call sites commented out — noted in description), CHECK_FILE, COPY_ZIP_REPORT. Parameter reuse: QSPR_ROOT/QSPR_REPORT_DIR registered as `Common`-scoped rows per ADR-0009; SERVICE_NAME + three COMMON_COPY_* + COMMON_FTP_ENABLE new; criteria CHECK_FILE_EXTENSION.
- `BUILT-IN|CADB|CA_ACCESS` — 19 STB-spec uses; no dispatch case exists in current trunk sources (older engine generation), documented from call sites only.
- `BUILT-IN|CELLULAR|MRRESETINSTR` — 1 bkM7 use; same treatment.

**Deferred (recorded):**

- `DLLDLLSAMPLE|{OPEN,WRITE,TIMES,TEST,RANDOM,PARSE_XML}` — SDK sample DLL (DynamicSample), sample flows only; not production content.
- `MACRO|MARCOT7` — macro label, covered by the generic `MACRO|RUN` decision (ecosystem-expansion ticket 05).
- `DLLQSPR|XML_CHECH_ITEM_PASS`, `HTML_CHECH_ITEM_PASS`, `PARSE_XML` — variant spellings already documented inside the canonical rows' descriptions (ticket 04 precedent); no separate rows.
- `WL -I …` / `GET NAME|…` scan hits — C# source lines and device-enumeration strings, not script calls.

**Verification:** additions-only diff vs `.pre_batch7.xlsx` backup (every pre-existing row byte-identical); duplicate natural-key scan clean; 0 orphans under the app's join semantics (function existence by (source, name); link device_name is a device scope); fixture hash-verified against the master; `go vet` clean; full backend suite green (`api` 6.7s, `importer`, `db`).
