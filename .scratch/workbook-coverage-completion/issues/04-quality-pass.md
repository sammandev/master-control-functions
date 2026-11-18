# 04: Examples and criteria quality pass

**What to build:** Close the quality pass over the completed registry. Every documented Function carries at least one real example harvested from trunk scripts (deduplicated by shape, meaning paraphrased from the same line); parameter definitions reused across functions where semantics match instead of duplicated; new definitions carry a `value_pattern` derived from real usage; criteria reuse (e.g. `EXPECTED_RESPONSE` for every function whose second segment is a response expectation) applied uniformly. Sparse name-only parameter rows are left alone unless evidence clarifies them — no speculation replacing honest gaps. Batch cycle and fixture sync as usual for any rows that change.

**Blocked by:** 01 (stragglers), 02 (IOT decision) — the row set must be final before deepening it.

**Status:** done

- [x] Every Function in the registry has at least one real example, or its absence is an honest, evidence-backed gap
- [x] No duplicate parameter definitions where semantics match; reuse applied per the established cross-family reuse set
- [x] New/updated definitions carry value patterns grounded in real usage
- [x] Criteria reuse audited: response-expectation segments uniformly use the shared criteria definitions
- [x] Sparse name-only parameter rows untouched unless evidence justified a change
- [x] Fixture + pinned counts synced (if rows changed); full backend suite green; batch commit scoped to fixture + expectation files

## Completion notes (2026-10-02, commit 8c26abd)

**Example coverage:** full-registry audit: 123 of 254 distinct `(source, name)` functions lacked examples. Trunk re-scan found verbatim usage lines for 23 of them; **34 examples appended** (185 → 219), the 11 high-traffic ones (≥14 call sites: ARITHMETIC, AUTO_RUN, CONFIRM_DIALOG, IMAGE_CONFIRM_DIALOG, LABEL_COMPARE, LABEL_SUBSTRING, SET_ISN, TSP_CHECKROUTE, TSP_GETVERSION, TSP_REPAIR, WRITE_DATA) carrying 2 distinct shapes each. Every script string is verbatim from `STB_spec.ini` / `MAIN_FLOW.ini` / alias lines; meanings paraphrase the line's description segment and flow targets. Example sheet now covers 154/254 functions.

**Honest gaps (100 functions, no trunk usage line found — left undocumented rather than invented):**

- `TEST PROGRAM` (58): ARITHMETICVAR, BIN_TO_DEC, BIN_TO_HEX, CHECK_CAL, CHECK_GOLDEN, CHECK_PING, CHECK_WIFI_GOLDEN_CSV, CHECK_WIRELESS_GOLDEN, COMMAND, COMPARE_PATH_LOSS, COMPLEMENT, COPY_BRCMNVMTOSYSTEM32, CYCLE_TIMER, DEC_TO_BIN, DELETE_VICREADER, DELETE_WLEXE, DETECT_CONNECTION, DISSOLUTION, FAIL, GET_HTTP, GET_SYSTEM_DATE, HEX_TO_BIN, IMAGE_CONFIRM_DIALOG2, LABEL_CONTAIN, LABEL_REMOVESUBSTRING, LABEL_REPLACESTRING, LABEL_SPLIT, LABEL_SUBSTRING_EX, LABEL_SUBSTRING_SP, LABEL_TO_PATH_CSV, LABEL_TO_WIFIGOLDEN_CSV, LABEL_TRANSFERTOASCII, LINQ_TO_EXCEL_QUERY, LOCK, MATCH_GOLDEN, MATCH_WIRELESS_GOLDEN, NEXT_SCAN, PING, READ_PUBLIC_INI, REMOVE_FILE, RESULT_PASS, SCAN_QR_8825, SCRIPT_CHECKSUM, SFIS_TO_VAR, SHOW_PATH_LOSS, SHOW_READY, STR_TO_BASE64, STR_TO_INT, TAKE_SN_FROM_POOL, UPDATE_RUN_DATE, UPPERCASE, VALUE_COMPARE, WRITE_FILE, WRITE_PUBLIC_INI
- `HTML` (17): BUTTONCLICK, CHECKBOX, CLEARSESSION, CLICK, COMBOBOX, JAVASCRIPT, PF_URL, POPUP, RADIOBUTTON, READANDCOMPARE, READINI, REFRESH, WAITURL, WRITE, XPATH, XPATHANDCOMPARE, XPATHANDCOMPARE_DYNAMIC
- `GOLDENKV` (9): COMPARE_GOLDEN_CAL_DATE, COMPARE_GOLDEN_CAL_PATHLOSS_MD5, COMPARE_GOLDEN_ISN, COMPARE_GOLDEN_VEL_DATE, COMPARE_PATH_LOSS, CONVERT_BINDATA_TO_DIC, RECORD_TESTRESULT, UPDATE_CAL_INFO, UPDATE_VEL_INFO
- `SFIS` (3): TSP_GETIMAC, TSP_GETIMACNEW, TSP_SSD_INPUTDATA, WRITE_DATA_TO_SFIS
- `DMM` (6): INIT, MATRIX_CLOSE, MATRIX_OPEN, MATRIX_OPENALL, OPEN, VOL_MEASURE
- `FTP` (3): COPY_PATHLOSS, COPY_TEST_RESULTS_ALL, COPY_TEST_RESULTS_FOLDER, MARS_EAP
- `AEROFLEX_PPM_16` (2): CONNECTION, MEASURE_PULSE_POWER, MEASURE_SETTING
- Others: NI6008|MEASURE, RS232|SENDSTR_AND_WAITSTR_NOLINE, IPLAS|WRITE_TIME

(The registry documents them from dispatch/implementation evidence; example absence is recorded here as the honest gap per the evidence hierarchy. Ticket 05's audit artifact re-lists these.)

**Definition audit:** 31 same-type-same-pattern signature collisions inspected — all are semantically distinct parameters sharing a loose pattern (LEFT_OPERAND vs RIGHT_OPERAND, COMMON_COPY_SN vs SERVICE_NAME, QSPR_CSV_ITEM vs QSPR_ERR_ITEM, per-metric WIFI_TEST limits, …); no merges. The ADR-0009 per-scope rows (`IOT`/`WIFI_TEST`) are by-design duplicates. `CMD`, `DEV_NAME`, `SETTING` are fully sparse (every column empty) — left untouched per the sparse-row rule.

**Criteria audit:** uniform already. Session/console families link the shared device-scoped `EXPECTED_RESPONSE` (CMD_CONSOLE, CMD_CONSOLEV2, CMD_SINGLE, DUT, RS232, SSH, TELNET); HTML/SFIS use `EXPECTED_VALUE`; argument-style criteria segments (TEST PROGRAM, QSPR, PTest, MACRO, AEROFLEX, IPLAS) use `CRITERIA_ARGUMENTS`; WIFI_TEST carries its 42 per-metric definitions by design. The one apparent outlier (`COMMAND` criteria on TELNET|OPEN) is correct — TELNET OPEN's criteria segment carries the login prompt/command echo. No changes.

**Verification:** additions-only diff vs `.pre_quality_pass.xlsx` backup (examples +34, all other sheets byte-identical); pre-flight duplicate-example check; 0 orphans under app semantics; fixture hash-verified (`60CBBBDD…`); `go vet` clean; full backend suite green.
