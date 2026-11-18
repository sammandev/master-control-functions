# 01: Test Program family

**Status:** done

Add the `TEST PROGRAM` built-in family to the workbook: every function dispatched by `General_DUT\Script.cs` `CreateCmd()` lines **1638-1865** (device string `TEST PROGRAM`). ~75 functions: DELAY, FAIL, LABEL_COMPARE, LABEL_CONTAIN, LABEL_SUBSTRING, LABEL_SUBSTRING_SP, LABEL_SUBSTRING_EX, LABEL_REMOVESUBSTRING, LABEL_REPLACESTRING, LABEL_TRANSFERTOASCII, LABEL_SPLIT, HEX_TO_DEC, HEX_TO_BIN, DEC_TO_HEX, DEC_TO_BIN, BIN_TO_HEX, BIN_TO_DEC, COMPLEMENT, VALUE_IN_RANGE, ARITHMETIC, SCAN_BARCODE, SCAN_QR_8825, NEXT_SCAN, MD5, CRC32, MD5_LABEL, CRC32_LABEL, CONFIRM_DIALOG, IMAGE_CONFIRM_DIALOG, IMAGE_CONFIRM_DIALOG2, LOWERCASE, UPPERCASE, STR_TO_INT, STR_TO_BASE64, COMMAND, GET_HTTP, PING, CMD_PING, DETECT_CONNECTION, EXPONENT_CONVERT, DISSOLUTION, AUTO_RUN, SHOW_READY, SHOW_PATH_LOSS, LABEL_TO_PATH_CSV, LABEL_TO_WIFIGOLDEN_CSV, COMPARE_PATH_LOSS, VALUE_COMPARE, WRITE_INI, READ_INI, WRITE_PUBLIC_INI, READ_PUBLIC_INI, WRITE_FILE, CHECK_CAL, UPDATE_RUN_DATE, CHECK_GOLDEN, CHECK_WIFI_GOLDEN_CSV, MATCH_GOLDEN, CHECK_WIRELESS_GOLDEN, MATCH_WIRELESS_GOLDEN, CHECK_PING, SET_ISN, LOCK, COPY_BRCMNVMTOSYSTEM32, DELETE_WLEXE, DELETE_VICREADER, RESULT_PASS, GET_SYSTEM_DATE, LINQ_TO_EXCEL_QUERY, TAKE_SN_FROM_POOL, REMOVE_FILE, SCRIPT_CHECKSUM, CHANGE_RESULT_PIC, CYCLE_TIMER, LOOP, GOTOITEM, RUNITEM, WRITE_VAR, ARITHMETICVAR, CSV_PATHLOSS_TO_LOG.

**Rules:**
- One `function_main` row per function (source=BUILT-IN, interface=`TEST PROGRAM`, convention=2); category from semantics (utility / label / hashing / communication / logging); execution_role + execution_phase inferred (e.g. SCAN_BARCODE → identification/setup, DELAY → timing/test_execution).
- Parameter definitions + links per function from the getParam/getCriteria calls (single PARAM string for most; criteria where the switch wires C).
- Criteria definitions for new criteria keys (e.g. PASS/Min/Max/Check Count patterns seen in MAIN_FLOW.ini aliases) — only add what the code reads.
- 2+ Examples per high-traffic function (DELAY, LOOP, GOTOITEM, SCAN_BARCODE, VALUE_IN_RANGE, WRITE_VAR, READ_INI, CMD_PING) derived from MAIN_FLOW.ini alias lines.

**Blocked by:** none

**Acceptance:** import into the app introduces zero new orphans; function names match the case labels exactly; (source, interface, name) unique.

## Comments

- Done 2026-10-02. Applied first-hand from General_DUT\Script.cs:1638-1865 (all 80 switch cases verified against the workbook; 0 missing).
- Appended to the canonical workbook: 64 new function_main rows (16 already registered, dedupe kept the richer existing rows), 11 parameter definitions, 87 parameter links, 1 criteria definition (CRITERIA_ARGUMENTS — the free-form criteria payload most TEST PROGRAM items evaluate), 45 criteria links, 16 examples (2 each for DELAY, LOOP, GOTOITEM, SCAN_BARCODE, VALUE_IN_RANGE, WRITE_VAR, READ_INI, CMD_PING).
- Scoping call: opaque single-string params got one honest generic definition (COMMAND_ARGUMENTS / CRITERIA_ARGUMENTS) instead of 60 invented grammars; specific definitions only where evidence exists (DURATION, LOOP_COUNT, TARGET_ITEM, INI_FILE/KEY/VALUE, BARCODE_PROMPT, HTTP_URL, FILE_PATH, LOCK_NAME, SOURCE_FILE, SAVE_VARIABLE).
- Gate via the real app (POST /api/import/reset): functions=177, parameter_definitions=150, parameter_links=390, criteria_definitions=51, criteria_links=150, examples=105; orphan total 0; normalization log shows only the 4 pre-existing expected fixes.
- Flag for ticket 03: pre-existing row BUILT-IN/TEST PROGRAM/SFIS_TO_VAR is NOT a TEST PROGRAM switch case — it belongs to the SFIS family; retarget interface when ticket 03 lands.
- Note: the workbook itself is gitignored (.dev-reference/*); the tracked test-fixture copy under backend/internal/importer/testdata was updated in a previous session and its pinned counts are out of scope here.
