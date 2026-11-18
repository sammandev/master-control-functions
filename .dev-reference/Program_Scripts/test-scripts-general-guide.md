# General Test-Scripts Guide

This document describes the common structure and purpose of test-script folders that use the same INI-based pattern as this workspace.

It is intentionally station-agnostic. The exact section names, values, file names, and station labels may change from one test station to another, but the file roles and execution pattern are usually the same.

## Overall Purpose

The test scripts define how an automated test station should:

- Load shared settings and environment values.
- Select and execute a main test flow.
- Call reusable test macros.
- Compare measured results against specification values.
- Read and write runtime data such as login prompts, file names, and device settings.

The scripts are usually consumed by a host test program or automation engine. That engine reads the INI files, resolves variable references, and executes the function rows in the expected order.

## High-Level File Roles

### `MAIN_FLOW.ini` checks

This is the orchestration file.

It usually contains the top-level test flow definitions and the detailed step-by-step execution blocks for each test item.

Typical responsibilities:

- Define the ordered test flow for a station or SKU.
- Map a visible test item name to a detail section.
- Select which macros or sub-items run for each test item.
- Control init, test, final, save, and cleanup phases.

For audit and parser purposes, only `[TEST FLOW:...#SKU:...]` sections count as
the ordered test-item list. `[FINAL FLOW:...]` sections are cleanup/handoff and
`[INIT FLOW:...]` sections are setup, so they should not be counted as test
items.

This file is normally the main consumer of all other script files.

### `MACRO_LIST.ini` checks

This is the reusable macro library.

It usually contains the executable function rows that the main flow calls indirectly.

Typical responsibilities:

- Hold reusable test procedures.
- Encapsulate common DUT actions, telnet steps, delays, checks, retries, and cleanup.
- Avoid repeating the same command sequence in many different flows.

This file is one of the main places where executable function-style rows are
expected to appear. In bundles such as HH5K `Wireless-1`, `MAIN_FLOW.ini`
item sections also contain executable function rows directly (item
implementations live in `MAIN_FLOW.ini`), while `MACRO_LIST.ini` holds the
reusable macros those items call with `MACRO|NAME|...` rows.

### `FACTORY_SETTING.ini`

This file usually stores factory or site-level values.

Typical responsibilities:

- Store station identity values.
- Store DUT-related network values such as IP addresses.
- Store tester metadata, labels, and production-side information.
- Store shared values that rarely change during a run.

These values are commonly used by both the main flow and the macro library.

### `PARAMETER_LIST.ini`

This file usually stores general runtime parameters for the test engine.

Typical responsibilities:

- Hold global timing values.
- Hold analysis defaults.
- Hold capture, averaging, or measurement behavior settings.
- Hold tuning values that affect how the test software behaves.

These are usually definition-only values and not executable steps.

### `RD_SETTING.ini`

This file usually stores runtime device, script, and environment settings.

Typical responsibilities:

- Store login credentials and prompt strings.
- Store interface aliases such as radio names or port names.
- Store script bundle file names.
- Store path-loss file names or other support file names.
- Store version pins, calibration timing, and auxiliary test settings.

This file is commonly used by telnet-based flows, file-load steps, and device-setup routines.

### `TEST_SPEC.ini` checks

This file usually stores limits and specification values.

Typical responsibilities:

- Store pass/fail thresholds.
- Store power, frequency, PER, RSSI, EVM, or other measurement limits.
- Store mode-specific targets for different bands, standards, or data rates.
- Store calibration or golden-reference spec values.

The main flow and macros use these values when comparing measured results against expected limits.

## Common Script Pattern

Most folders follow the same pattern:

1. `MAIN_FLOW.ini` defines the station-level test path.
2. A visible test item name is mapped to a detail section.
3. The detail section contains sub-items that are executed in order.
4. Those sub-items may call macros from `MACRO_LIST.ini`.
5. The macros read settings from the other INI files and from runtime inputs.
6. The test result is compared against values from `TEST_SPEC.ini` or other spec sections.

### Typical Section Patterns

Common section names include:

- `[TEST FLOW:...]`
- `[FINAL FLOW:...]`
- `[INIT FLOW:...]`
- `[ITEM_NAME#ErrCode:...]`
- `[MACRO_NAME]`
- `[SETTING_NAME]`

The exact text inside the brackets changes by station, but the structure usually stays the same.

### Typical Value Patterns

Common value patterns include:

- Plain scalar settings such as IP addresses, timeouts, and filenames.
- Section-key references using `*V:Section-Key#`.
- Whole-section references using `*V:Section#`, which expand to that section's key-value assignments.
- Nested references such as `*V:*V:Macro-Criteria_Criteria##`, where the inner reference resolves to the section name used by the outer reference.
- Host/runtime values that are supplied by the automation engine rather than stored directly in the repo.

## Function Row Template

Executable rows usually follow this format:

```text
DEVICE|FUNCTION|PARAM|CRITERIA|WAIT_TIME|PASS_PATH|FAIL_PATH|ERRCODE:Index
```

The command template has seven core fields plus one optional eighth field:

```text
DEVICE|FUNCTION|PARAM|CRITERIA|WAIT_TIME|PASS_PATH|FAIL_PATH|ERROR_CODE
```

The first seven fields are the normal command contract. The eighth field is
optional but, when present, must use `ERRCODE:<index>` or `ErrCode:<index>`
syntax. Rows with seven fields remain valid. Fields after the eighth are not
part of the documented template and should be reported as extra fields.

General meaning of the fields:

- `DEVICE`: the module, device, or handler that provides the action.
- `FUNCTION`: the function or command name to run.
- `PARAM`: runtime arguments or key-value inputs.
- `CRITERIA`: the comparison rule or validation data.
- `WAIT_TIME`: maximum time allowed for the action, normally in milliseconds.
- `PASS_PATH`: next step when the action succeeds.
- `FAIL_PATH`: next step when the action fails.
- `ERROR_CODE`: optional error-code expression used for failure reporting.

In many flows, the `MAIN_FLOW.ini` entry calls a macro name, and the macro body contains rows in this exact format.

### Function-row control flow

The last two fields are result-dependent branches. Follow `PassPath` after a
successful action and `FailPath` after a failed action. A standalone uppercase
`QUIT` path is terminal for the current script segment: execution stops when
that branch is taken. Do not treat a label such as `Delay_QUIT` as terminal just
because its name contains the text `QUIT`; resolve that label and continue until
the row whose path is the standalone `QUIT` token is reached.

`NEXT` is the documented fall-through path token, not a label. The C# runtime
compares this token exactly as uppercase `NEXT`; do not silently normalize
`Next` or other casing when validating runtime compatibility. After the action
result is known, move to the next active script row in source order within the
current item section. Skip blank and commented-out rows when identifying the
next active row. A checker may report mixed-case `Next` as a portability or
runtime-compatibility warning rather than treating it as a valid runtime token.

If there is no later active row in the current item section, `Next` continues to
the next test item in the selected `[TEST FLOW:...#SKU:...]` order. Use that
top-level flow order for the handoff; do not scan into the next bracketed item
section by physical file position and treat its implementation rows as part of
the current item. For example, when `GET_SERIAL_NUMBER` is followed by
`TSP_CheckRoute` in the selected TEST FLOW and its final active row uses
`PassPath=Next` or `FailPath=Next`, execution enters `TSP_CheckRoute` after the
current item branch. If the selected TEST FLOW has no next item, do not invent a
label or an additional step; record that the fall-through reaches the end of the
selected flow and follow the host's end-of-flow behavior.

The following syntax has two different retry/loop meanings:

- A row or state-item suffix such as `#Retry:30` gives that action its retry
  budget. It does not create a new top-level test item and it does not replace
  a loop controller.
- A row whose function field is `Loop`, commonly written as
  `Test Program or TEST PROGRAM or TESTPROGRAM|Loop|Test_Count,3|Parameters|TimeOut|PassPath|FailPath`, is a
  loop controller. The count in `Test_Count,3` is the maximum number of loop
  iterations. Other rows in the same item may point to this controller through
  `PassPath` or `FailPath`, which is how a retrying action re-enters the loop.

Keep the two budgets separate when explaining a script: `Test_Count,3` allows
up to three loop passes, while `#Retry:30` supplies the retry budget for the
action row reached during each pass. When the loop budget is exhausted, follow
the loop controller's configured terminal branch. In the common pattern where
that branch is `FAIL`, the enclosing test item fails; according to the seven
field template this is the `FailPath` field when `FAIL` is the seventh value.

Only active rows participate in this control-flow graph. Lines whose first
non-whitespace characters are `;` or `//` are comments and are not executable,
including lines beginning with `; //`. A commented-out alternative row must
not be counted as an executed step, available branch, retry, or loop.

## Macro Call Pattern

The main flow often uses a wrapper like this:

```text
MACRO|FunctionName|...
```

That means:

- The flow is calling a reusable macro.
- `FunctionName` should exist as a section in `MACRO_LIST.ini`.
- The macro section contains the actual executable function rows.

The runtime usually resolves the macro name first, then executes the function rows that belong to that macro section.

### Macro PassPath and FailPath control flow

The rows inside a `MACRO_LIST.ini` section form their own result-driven graph.
Resolve each macro row's `PassPath` and `FailPath` against labels in that macro
section, just as for an active `MAIN_FLOW.ini` section. A macro section is not a
separate selected TEST FLOW item and must not fall through to the next physical
section in `MACRO_LIST.ini`.

When a macro row reaches a standalone `QUIT`, the branch that reached it is the
macro result: `PassPath=QUIT` returns success and `FailPath=QUIT` returns
failure. A `FAIL` terminal also returns failure. The result then returns to the
`MAIN_FLOW.ini` row that called the macro, and the caller's own `PassPath` or
`FailPath` is followed. For example, a macro call with
`PassPath=AFTER_PASS|FailPath=AFTER_FAIL` reaches `AFTER_PASS` after a successful
macro return and `AFTER_FAIL` after a failed macro return.

The documented runtime does not support nested macros: `cRunMacro.execute`
rejects a macro call while a macro is already executing. A checker should report
a macro call from inside `MACRO_LIST.ini` as unsupported runtime nesting. Static
graph code may still expand such rows for analysis, but that graph must not be
presented as executable production behavior. A macro-local `NEXT` advances only
to the next active row in that macro section. `NEXT` at the end of a macro
without an explicit return is incomplete control flow and should be reported
rather than treated as a new TEST FLOW item.

The execution graph should inspect both MAIN_FLOW and expanded MACRO_LIST
branches for reachable cycles. Report a potential infinite loop only when the
reachable cycle contains an active `Test Program|Loop|Test_Count,N` controller
and that Loop row's `FailPath` targets another active row inside the same cycle.
That re-entry means the loop can be entered again after its configured count is
exhausted. Do not report ordinary cycles without a Loop controller, or a Loop
whose `FailPath` terminates outside the cycle. Keep this separate from
row-level `#Retry:N` and item-header `Retry:N` budgets.

Macro call arguments also create scoped variable names while the macro body is
being expanded. In the seven-core-field row, caller `Criteria` assignments become
`Macro-Param_*` variables, and caller `Parameters` assignments become
`Macro-Criteria_*` variables. For example, a caller row containing
`TX_TARGET_POWER=*V:6G_TX_TARGET-11AX_BW160_MCS11_POWER#` in the `Criteria`
field and `Criteria=6G_TX_11AX_BW160_MCS11_SPEC` in the `Parameters` field can
be consumed by a macro body as `*V:Macro-Param_TX_TARGET_POWER#` and
`*V:*V:Macro-Criteria_Criteria##`.

## How MCV2 Renders Script Items Into the Log

Every log element is the rendered trace of a script element. The mapping is
documented in full in
[`mcv2-log-format-guide.md`](./mcv2-log-format-guide.md); the script-side
facts that matter when joining the two are:

- State-Log block headers are **function-row labels** (the LHS of
  `LABEL = "DEVICE|FUNCTION|..."`), not item names. Item names appear only in
  the SFIS/iPLAS summary rows and the `<-------Result------->` trailer.
- A `MACRO|NAME|...` row logs one block under the caller row label, containing
  a `Clean Run Flow.` / `Macro add : <MACRO_NAME>` preamble input plus one
  input per executed macro sub-row; macro sub-row labels are not printed.
- A failed `#Retry:N` row logs one block per attempt, with the attempt number
  appended to the label (`[PING_DUT#Retry:30:2]`, `:3`, `:4`). Attempts are
  one logical action, not duplicate items.
- `ANT=` in DLL parameters is a chain bitmask (`1,2,4,8` = chains 1-4; `7` =
  chains 1+2+3; `15` = all four), not a chain count. A row with `ANT=4`
  produces chain-3 (`TX3`/`RX3`) measurement names.
- Pipe characters inside telnet PARAM strings are escaped as `\x7C` so they
  survive the seven-field row format; the state log echoes `\x7C` literally
  while the DUT receives a real `|`. Similarly `\x0A`, `\x0D`, `\x03`.
- Each INI ends with a `checkSumed:<md5>~<author>~<created>~<expires>` trailer;
  its `.bin` form is listed in the log header DUT CONSOLE block. Multiple
  trailing `checkSumed:` lines are repackaging history. `*.bak` siblings are
  not part of the active bundle.

## Variable Reference Pattern

The most common variable substitution pattern is:

```text
*V:Section-Key#
```

This means:

- Find the section named `Section`.
- Read the key named `Key` in that section.
- Replace the token with the stored value.

This style is used to share settings across files without copying values into many places.

At runtime, resolved values are stored and looked up through the test program's
`Global.sdVarMap`. Preflight should therefore distinguish a literal value from
an unresolved `*V:` reference and report the source file, section, key, and
field when a reference cannot be resolved. Do not infer a missing variable from
a similarly named key in another scope.

The audit resolver also supports the related whole-section and nested patterns:

```text
*V:Section#
*V:*V:Macro-Criteria_Criteria##
```

`*V:Section#` expands the active section into `KEY=VALUE&KEY2=VALUE2` text.
Nested references resolve from the inside out. In the common macro pattern,
`*V:Macro-Criteria_Criteria#` first resolves to a section name such as
`BT_TX_SPEC`, and then the outer `*V:...#` expands that section's limits.

For Test Spec Review, resolve variables against the selected script bundle's
`MAIN_FLOW.ini`, `MACRO_LIST.ini`, `TEST_SPEC.ini`, `RD_SETTING.ini`, and
`PARAMETER_LIST.ini`. Preserve unresolved references as evidence instead of
guessing missing values.

## Test Spec Review Limit Derivation

The rendered `Test Spec Review` table compares the relevant measurement or TP
requirement against `Scripts limit`, `TP/MTP limit`, and `Log Actual`.
The `Scripts limit` may come from `TEST_SPEC.ini`, script Parameters, or resolved
macro context. The `TP/MTP limit` must come only from the selected TP/MTP
reference; show `-` when that TP/MTP row has no explicit limit. The three
evidence columns should be rendered as de-duplicated bullet lists, not as long
semicolon- or pipe-separated strings.

For SZ Wi-Fi rows, the numeric criteria may be printed in a separate Test Spec
or result table from the channel/mode requirement list. Normalize those
criteria onto the matching row by band, frequency, mode, and bandwidth, then
render metric-specific TP/MTP values such as POWER, EVM, FREQ, MASK,
LO_LEAKAGE_DB, PER, and RSSI. A `-` is appropriate only when the TP row is
qualitative or has no explicit numeric criterion.

For QCA F5984 PowerPoint references, the same metric values come from the
numbered `Pass/Fail Criteria` rows. Map each row's 2.4G/5G mode, bandwidth, and
frequency to its generated TP requirement row before rendering `TP/MTP limit`.

Render script/log metric bullets in this order: POWER/POW, EVM,
FREQ/FREQ_ERROR, MASK, then LO/LO_LEAKAGE. For BT TX evidence, show
`TX1_POW_*_BT` before `TX1_FREQ_KHZ_*_BT`. The Test Spec Review source summary
should read `Scripts: <path-to-scripts-folder>` because the limit can come from
any selected script-bundle file, not only `TEST_SPEC.ini`.

For power-style checks, combine the target value from `Criteria` with the
tolerance from `Parameters` or the resolved spec section. For example:

```text
TX_TARGET_POWER=20
POWER=-3~1
```

produces the effective limit:

```text
POWER=17 to 21
```

A single POWER value is a symmetric tolerance around the target. If the target
is `20`, then `POWER=1.5` or `POWER=1.5dB` means:

```text
POWER=18.5 to 21.5
```

Direct criteria such as `PER=10%`, `RSSI=-6dB~6dB`, or
`FREQ_ERROR_KHZ=-150~150` should be carried through as resolved limits without
target arithmetic.

### Why This Matters

The same test station may reuse the same flow structure, but the values underneath can change by:

- Station
- SKU
- DUT type
- Band or radio
- Calibration target
- Customer or project variant

The `*V:` pattern is what allows the same flow template to work across multiple stations.

## Execution Order

The normal runtime behavior is:

1. Read the top-level `TEST FLOW` from `MAIN_FLOW.ini`.
2. Resolve the target test item name.
3. Open the matching detail section for that test item.
4. Execute the sub-items in sequence.
5. Resolve any `MACRO|...` call into `MACRO_LIST.ini`.
6. Substitute any direct, whole-section, or nested `*V:` values from the appropriate settings or spec file.
7. Continue through pass and fail paths until the flow completes.

## Comment Rules

Lines that begin with `;` or `//` are comments.

They may document historical behavior, alternative flows, disabled test steps, or notes for engineers, but they are not executed by the test program.

The runtime `TINI` reader trims the values it returns and skips lines whose
first non-whitespace characters are `;` or `//`. Source-level preflight should
still inspect raw lines first, because trimming before validation would hide
leading/trailing whitespace defects in TEST FLOW item values and function rows.

## Verified Flow Syntax and Lint Rules

The test scripts bundle confirms that flow syntax needs a raw-line validation
pass before normal parsing. The parser can otherwise discard malformed lines
and make a broken script look shorter instead of reporting the syntax defect.

### TEST FLOW, INIT FLOW, and FINAL FLOW pairing

`TEST FLOW` is the required execution entry point. `INIT FLOW` and `FINAL FLOW`
are optional. A script may have a TEST FLOW without either lifecycle section,
although a matching FINAL FLOW is strongly recommended when the test reserves,
initializes, or removes hardware, tester connections, DUT state, or temporary
files:

```text
[TEST FLOW:Wireless_Test_6G#SKU:HH5K]
...

[FINAL FLOW:Wireless_Test_6G]
...
```

The `#SKU:...` suffix belongs to the TEST FLOW selection header. The FINAL
FLOW header normally uses the same name without the SKU suffix. Pairing is by
the exact normalized name after `TEST FLOW:` or `FINAL FLOW:`; do not pair
flows by a partial prefix. A missing FINAL FLOW is a warning/recommendation,
not a syntax failure. A missing INIT FLOW is allowed unless the station profile
requires setup.

Whitespace after the flow colon is allowed, and station/SKU names may contain
spaces:

```text
[TEST FLOW:Wireless_Test_6G#SKU:HH5K]
[TEST FLOW: Wireless_Test_6G#SKU:HH5K]
[TEST FLOW: NFT Test#SKU:F3897 LG]
[INIT FLOW: Wireless_Test_6G]
[FINAL FLOW: Wireless_Test_6G]
```

The checker should trim only boundary whitespace around the kind, flow name,
and SKU. It must preserve spaces inside the actual station and SKU names. A
tolerant header shape is equivalent to:

```text
^\[\s*(TEST|INIT|FINAL)\s+FLOW\s*:\s*([^\]#]+?)(?:\s*#SKU\s*:\s*([^\]]+?))?\s*\]\s*$
```

The test scripts bundle follows this pattern for the production flow:

```text
[TEST FLOW:Wireless_Test_2_5G#SKU:HH5K]
...

[FINAL FLOW:Wireless_Test_2_5G]
DUT_REMOVE = "DUT_REMOVE"
TESTER_DISCONNECTION = "IQ_DISCONNECTION"
FTP_CopyFile = "FTP_CopyFile"
AUTO_RUN = "AUTO_RUN"
```

The sampling flow is a separate lifecycle:
`Wireless_Test_2_5G_Sampling` has its own TEST FLOW and FINAL FLOW and must
not be merged into the production flow. An audit should report:

- a TEST FLOW with no matching FINAL FLOW as a warning/recommendation;
- an orphan FINAL FLOW with no matching TEST FLOW;
- duplicate TEST or FINAL headers for the same name;
- a selected TEST FLOW whose cleanup flow has a different name.

Only TEST FLOW rows are test-item order. FINAL FLOW rows are cleanup and
handoff operations, and INIT FLOW rows are setup operations.

### Exact active TEST FLOW row grammar

Inside a TEST FLOW, an active row has an unquoted identifier on the left and a
quoted test-item name on the right:

```text
Variable_A = "TEST_ITEM1"
SET_IPLAS_INFO_WIRELESS_1 = "SET_IPLAS_INFO_WIRELESS_1"
DELAY = "Wait_WIFI_Stable"
```

The workspace parser's exact identifier grammar is equivalent to:

```text
^[A-Za-z0-9_]+\s*=\s*"[^"]+"$
```

Therefore:

```text
Variable_A = "TEST_ITEM1"       # valid
"Variable_A" = "TEST_ITEM1"   # invalid: quoted left-hand key
Variable_A = " TEST_ITEM1"     # invalid: leading space inside item value
Variable_A = "TEST_ITEM1 "     # invalid: trailing space inside item value
```

Use underscores in identifiers. A literal key containing a space, such as
`Variable A`, is not valid for this INI grammar even if the name is readable to
humans.

The left-hand key and right-hand item may differ. For example,
`DELAY = "Wait_WIFI_Stable"` is valid; the right-hand value is the item name
that appears in the log and is the value used for sequence and coverage
matching.

The raw validator should reject leading or trailing whitespace inside the
quoted item before the parser calls `.strip()`. The current parser strips the
RHS value after matching, so a malformed line such as
`Variable_A = " TEST_ITEM1"` can otherwise be normalized silently. That is a
linting gap, not evidence that the source syntax is valid.

### Comment tokens and invalid slash usage

Only these prefixes disable a row, after optional indentation:

```text
;Variable_A = "TEST_ITEM1"
; Variable_A = "TEST_ITEM1"
//Variable_A = "TEST_ITEM1"
// Variable_A = "TEST_ITEM1"
```

The following are also comments because their first non-whitespace character
is `;` or the two-character token `//`:

```text
;;Variable_A = "TEST_ITEM1"
;// Variable_A = "TEST_ITEM1"
;--[TEST FLOW:Disabled#SKU:HH5K]
```

A single leading slash is not a comment marker:

```text
/Variable_A = "TEST_ITEM1"
```

It must be reported as an invalid active/configuration line. It must not be
silently treated as a disabled test item. The same rule applies to malformed
comment prefixes in TEST FLOW, FINAL FLOW, INIT FLOW, item sections, and macro
sections.

### Section boundary rules

The bracketed header must be recognized before rows are classified:

```text
[TEST FLOW:NAME#SKU:SKU]
[FINAL FLOW:NAME]
[INIT FLOW:NAME]
[ITEM_NAME#ErrCode:ALIAS]
[MACRO_NAME]
```

An unknown or malformed bracket header ends the current flow context for audit
purposes. Rows after it must not be attached to the previous TEST FLOW or
macro section merely because they occur nearby in the file.

## Test Scripts Findings and Reusable Insights

The supplied test scripts bundle shows the following
stable design patterns:

- `MAIN_FLOW.ini` owns flow selection, production order, sampling order, and
  cleanup flow declarations.
- `MACRO_LIST.ini` contains reusable state machines such as
  `TX_MULTI_VERIFICATION`, `FIXTURE_VERIFICATION`,
  `FIXTURE_VERIFICATION_Composite`, `FREQ_VERIFICATION`, `RX_VERIFY_PER`,
  `RX_VERIFY_RSSI_CAL`, `RX_VERIFY_PER_Collect_RSSI`, and
  `RX_11BE_VERIFY_PER`.
- A MAIN_FLOW row with `MACRO|FunctionName|...` passes caller arguments into a
  macro. Caller `Criteria` values become `Macro-Param_*`; caller `Parameters`
  values become `Macro-Criteria_*` during expansion.
- `CSV_TAG` is runtime measurement metadata, not a fixed enumeration. The
  bundle uses `_FIXTURE`, `_Check_FREQ`, `_Before_RSSI_CAL`, and
  `_Collect_RSSI_Value`. New tag names should be discovered from resolved
  macro parameters and matched case-insensitively by measurement identity.
- `RX_PACKAGE=rxdtocast` is a DLL parameter, not a measurement-name tag. It
  must remain separate from `CSV_TAG` handling.
- `#Retry:N` on a row label is a retry budget. It is not a duplicate TEST FLOW
  item and is not equivalent to `Test Program|Loop|Test_Count,N`.
- `#ErrCode:ALIAS` belongs to an item section header and should be checked
  against the selected ErrorCodeMap independently of row syntax.
- Fixture rows use both item-name markers such as `_FIXTURE` and emitted
  measurement names such as `FIXTURE_OR_DUT_PROBLEM`. They should be excluded
  from ordinary TP requirement coverage but retained as explicit evidence in
  Test Spec Review when present in the log.
- `TEST_SPEC.ini` separates target values, fixture limits, PA trimming offsets,
  per-mode TX limits, RX power/frame-count settings, calibration limits, and
  collection-only variants. Do not flatten these sections into one global
  limit map.
- PA trimming uses `*_PA_TRIMMING_POWER` (`11B_TARGET_POWER`, `BW20_TARGET_POWER`,
  ...) as `TX_TARGET_POWER` and `*_PA_TRIMMING_SPEC` (`PA_UPDATE_LIMIT_UPPER`,
  `PA_UPDATE_LIMIT_LOWER`) as the update window. DUT `WiFi_PA*_POW_OLD_*`
  bounds are target +/- UPPER. DUT `WiFi_PA*_POW_DIF_ABS_*` is
  `|target - POW_OLD|`.
- Beamforming calibration uses `[2G_IBF_TEST]` / `[5G_IBF_TEST]` `IBF_SPEC`
  ranges for `IBF_2G_TEST` / `IBF_5G_TEST` PASS items.
- `RD_SETTING.ini` owns runtime identity and environment values, including
  radio aliases, login prompts, shell script names, path-loss names, and
  firmware metadata. `PARAMETER_LIST.ini` owns global tester/measurement
  behavior and should not be treated as executable flow.
- 5G PA calibration for BW40/BW80/BW160 uses `BRCM_PA_PREDICTION`, fed by the
  prediction engine that `WIFI_PRED_INITIALIZATION` sets up
  (`PRED_SETTING`/`PRED_INITIALIZE`/`LOAD_STATE`/`LOAD_MODEL`, persisted by
  `SAVE_PREDICTION_STATE`/`SAVE_STATE`); `BRCM_PA_TRIMMING` remains for 2G
  and 5G BW20. Prediction tuning lives in `[BANDWIDTH_CHANNEL_LIST_5G]`,
  `[DRIFT_PARAMETERS]`, `[COUNTER_LIMITS]` in `RD_SETTING.ini`.
- `RX_VERIFY_RSSI_CAL` / `RX_VERIFY_RSSI_CAL_Multi` macros are check-only
  (`FailPath=QUIT`, `CSV_TAG=_Before_RSSI_CAL`); the recovery ladder
  (re-initialize -> default RX gain -> `BRCM_RSSI_CALIBRATION` -> retest)
  lives in the calling MAIN_FLOW item, not in the macro.
- `NVRAM_MD5_Check` FAILS the item when the NVRAM MD5 is UNCHANGED — the
  calibration flow must have written data. Its forced-fail row is
  `LABEL_COMPARE|PASS,0|FAIL,-1` (0 vs -1 never equal -> FailPath `FAIL`).
  Outcome labels describe results, not compare polarity.
- Item-header metadata order is not fixed: `#ErrCode:X`, `#ErrCode:X,Retry:N`,
  and `#Retry:2,ErrCode:X` all occur; the ErrCode alias may contain spaces
  (`Other in program error`). Row retry suffix casing varies and is
  significant for State-Log block matching: `#Retry:3` and `#RETRY:2`
  (`DUT_INSERT_WIFI#RETRY:2`).
- Eighth-field error codes appear in both casings:
  `|FAIL|ERRCODE:DUT_BOOT_UP_Ping_PEGA_FW_Fail` and
  `|FAIL|ErrCode:READ_2G_PAVARS`.
- Six-field rows (missing `WAIT_TIME`) exist in active sections
  (`Check_ZB_Bootloader`, `Check_DUT_Version` in
  `Check_ZIGBEE_SGC_FW`/`Check_DUT_SGC_FW`); a checker should warn instead of
  silently normalizing them, and runtime tolerance must be verified.
- `FACTORY_SETTING.ini` per-slot sections `[Setting:1]`..`[Setting:4]`
  (`DUT_IP`, `SRCNIC_IP`, `RF_PORT`, `RF_ROUTE`) are referenced as
  `*V:Setting#` / `*V:Setting-DUT_IP#` resolved per DUT slot; the slot suffix
  is `Setting:<DUT_Number>`. `[DeviceID_List]` maps
  `<flow1>&<flow2>#<DUTn>` keys to tester device IDs and a separate
  `[TestItemMonitor]`/`[MonitorItems]` block defines station-side fixture
  spot checks (`NAME=count,GT,value`).
- Function-name spellings that look like typos are load-bearing and must not
  be corrected on one side only: `COMPARE_GOLDEN_VEL_DATE` /
  `UPDATE_VEL_INFO` (VEL) in `DLLGoldenKV` calls, and the
  `*_WIFI_DOWN_ROMPT` keys in `RD_SETTING.ini`.

The test scripts header also contains a malformed metadata line:

```text
PROGRAM_VERSION = "2.69.9685.21265
```

It is missing the closing quote. A robust Scripts Correctness check should report this
as malformed metadata even when the TEST FLOW rows themselves are valid.

## Sample Test Script Deep Dive: Items, Functions, Parameters, Criteria, Control Flow

Reference bundle: `data-source/BRCM_Chip_HH5K/Scripts/Wireless-1/`
(`MAIN_FLOW.ini` 1446 lines, plus `MACRO_LIST.ini`, `TEST_SPEC.ini`,
`RD_SETTING.ini`, `PARAMETER_LIST.ini`, `FACTORY_SETTING.ini`). Every claim
below was read directly from those files. Log-side confirmation for this
bundle lives in the HH5K sections of `mcv2-log-format-guide.md`.

### Bundle header, versions, packaging

```text
[TEST_PROGRAM_VER]
SCRIPT_VERSION  = "HH5K_Wireless_1"
PROGRAM_VERSION	= "2.69.9685.21265        <- closing quote missing (malformed)

[Script_Info]
Ver				= "HH5K_Wireless_1_260714"
```

- `[Script_Info] Ver` appears in both `MAIN_FLOW.ini` and `RD_SETTING.ini`
  with the same value; the DUT log header `[Script_Ver]` echoes exactly this
  value (not `SCRIPT_VERSION`). `FACTORY_SETTING.ini` `[Path] TP_Dir`
  carries the TP packaging folder (`...20260713_V122\...`). Three lineage
  layers exist: bundle version, declared program version, TP packaging path.
- Every INI ends with
  `checkSumed:<md5>~Henry_Chuang~2026/07/13-16:12:25~2036/07/13-16:12:25`;
  the DUT log `[DUT CONSOLE]` block lists the same MD5s with `.bin` script
  paths (compiled copies; no FACTORY_SETTING line).

### Flows in the bundle

| TEST FLOW | SKU | Lifecycle | Purpose |
|---|---|---|---|
| `Wireless_Test_2_5G` | HH5K | TEST + FINAL | Production 2.4G+5G RF test (67 items) |
| `Wireless_Test_2_5G_Sampling` | HH5K | TEST + FINAL | Sampling variant |
| `WIFI_25G_GOLDEN_VERIFICATION` | HH5K | INIT + TEST + FINAL | Golden re-verification |
| `WIFI_25G_CHECK_IOT_VER` | HH5K | TEST + FINAL | Zigbee/Z-Wave FW check only |
| `WIFI_25G_MAKE_GOLDEN` | HH5K_RD_USE | commented out | disabled lab golden build |
| `WIFI_25G_GOLDEN_CALIBRATION` | HH5K | commented out (`;--[TEST FLOW:...]`) | disabled golden calibration |

`WIFI_25G_GOLDEN_VERIFICATION` is the one active flow with an `INIT FLOW`
(`KILL_PowerResults_FOLDER`:
`Test Program|COMMAND|cmd.exe,RD /S /Q Temp\*V:Label-DUT_Number#\GenerateResultTemp`).
The disabled `WIFI_25G_GOLDEN_CALIBRATION` header shows the `;--`
decorated-comment form. FINAL items for the production flow: `DUT_REMOVE`,
`TESTER_DISCONNECTION = "IQ_DISCONNECTION"` (LHS alias maps to the
`[IQ_DISCONNECTION]` item section), `FTP_CopyFile`, `AUTO_RUN`.

Production vs Sampling deltas (sequence matchers must honor these):

- Sampling drops `Check_DUT_SGC_FW` and `Default_WiFi_5G_NVRAM`.
- `TX_2G_11B_BW20_CCK11` (production, single point 2462) becomes
  `TX_2G_11B_BW20_CCK11_Sampling` (2412 -> 2462 chain) and
  `TX_2G_11AC_BW20_MCS8_Sampling` (2462) is added.
- `RX_2G_11AX_BW20_MCS0_RSSI_CAL` becomes `2G_RX_RSSI_CAL_Sampling`
  (11N MCS7 + 11AX MCS0 chain, each with the RSSI-CAL recovery ladder).
- 2G RX order swaps: `RX_2G_11AC_BW20_MCS0` runs before
  `RX_2G_11AG_BW20_OFDM6` in Sampling.
- Production uses the composite 5G fixture item
  (`TX_5G_11AC_BW20_MCS8_FIXTURE_Composite`, ANT=15); Sampling uses the
  enumerated-chain `TX_5G_11AC_BW20_MCS8_FIXTURE`.

### Ordered item list: `Wireless_Test_2_5G` (active rows only)

```text
SET_IPLAS_INFO_WIRELESS_1, WIFI_PRED_INITIALIZATION, CHECK_CAL_PATHLOSS_MD5,
COMPARE_GOLDEN_VER_DATE, COMPARE_CAL_DATE, DUT_BOOT_UP_PING, Wait_WIFI_Stable,
DUT_Login_TELNET, GET_SERIAL_NUMBER, TSP_CheckRoute, Check_DUT_SGC_FW,
CHECK_WiFi_6G_FW_VER, CHECK_WiFi_5G_FW_VER, CHECK_WiFi_2G_FW_VER,
Check_ZIGBEE_SGC_FW, Check_ZWAVE_SGC_FW, Default_WiFi_5G_NVRAM,
NVRAM_MD5_Before_Test, LOAD_SHELL_SCRIPT_ALL, IQ_GLOBAL_SETTING,
IQ_CONNECTION, IQ_LOAD_PATHLOSS, DUT_INSERT, DUT_INITIALIZE_5G,
TX_5G_11AC_BW20_MCS8_FIXTURE_Composite, DUT_INITIALIZE_2G,
TX_2G_11N_BW20_MCS0_FIXTURE, DUT_BCM6726_GET_2G_PAVARS, WIFI_2G_PA_TRIMMING,
DUT_BCM6726_SET_2G_PAVARS, TX_2G_11B_BW20_CCK11, TX_2G_11N_BW20_MCS0,
TX_2G_11N_BW40_MCS7, TX_2G_11AX_BW40_MCS11, TX_2G_11BE_BW40_MCS0,
TX_2G_11BE_BW40_MCS13, DUT_INITIALIZE_2G_RX, RX_2G_11AX_BW20_MCS0_RSSI_CAL,
RX_2G_11AG_BW20_OFDM6, RX_2G_11AC_BW20_MCS0, RX_2G_11BE_BW40_MCS0,
DUT_INITIALIZE_5G_RX, DUT_BCM6726_GET_5G_PAVARS, WIFI_5G_PA_TRIMMING,
DUT_BCM6726_SET_5G_PAVARS, TX_5G_11N_BW40_MCS7, TX_5G_11AX_BW20_MCS0,
TX_5G_11AX_BW80_MCS0, TX_5G_11AX_BW80_MCS11, TX_5G_11BE_BW80_MCS0,
TX_5G_11BE_BW160_MCS13, RX_5G_11AX_BW80_MCS11_RSSI_CAL, RX_5G_11AX_BW80_MCS0,
RX_5G_11BE_BW160_MCS7, DUT_5G_IBF_INITIALIZE, IBF_5G_TEST,
DUT_2G_IBF_INITIALIZE, IBF_2G_TEST, SAVE_2G_PA_BRCM, SAVE_5G_PA_BRCM,
SAVE_2G_IBF_DATA, SAVE_5G_IBF_DATA, SAVE_2G_RSSI_CAL, SAVE_5G_RSSI_CAL,
SAVE_PREDICTION_STATE, SAVE_NVRAM_FILE, NVRAM_MD5_Check
```

### Test-item anatomy as used here

Three verified item shapes:

1. **Single DLL row** — `[IQ_GLOBAL_SETTING#ErrCode:GLOBAL_SETTING]` holds
   one row `IQ_GLOBAL_SETTING#Retry:3 = "DLLWIFI_TEST|GLOBAL_SETTING|*V:WIRELESS_GLOBAL_SETTING#|NULL|10000|QUIT|FAIL"`.
2. **Labeled-row state machine** — `[GET_SERIAL_NUMBER#ErrCode:Get_SN_fail,Retry:2]`
   holds ~13 labeled rows forming pass/fail graphs with a telnet re-login
   recovery loop.
3. **Single MACRO call** — every TX/RX measurement item is one row
   `MACRO|TX_MULTI_VERIFICATION|BAND=..&STANDARD=..&FREQ=..&BW=..&DATA_RATE=..&TX_TARGET_POWER=*V:<BAND>_TX_TARGET-KEY#|Criteria=<SPEC_SECTION>|120000|QUIT|FAIL`;
   the `MACRO_LIST.ini` macro performs measure -> recover -> re-measure.

Item-header metadata order is not fixed; all three forms occur:
`#ErrCode:X`, `#ErrCode:X,Retry:N`, `#Retry:2,ErrCode:X`. The ErrCode alias
may contain spaces (`Other in program error`, `Other in program error.`).
Row-level retry suffix casing varies and is logged verbatim: `#Retry:3`,
`#RETRY:2` (`DUT_INSERT_WIFI#RETRY:2`). The optional eighth field appears as
`|FAIL|ERRCODE:DUT_BOOT_UP_Ping_PEGA_FW_Fail` (uppercase) and
`|FAIL|ErrCode:READ_2G_PAVARS` (mixed case).

Multi-frequency items chain rows by PassPath, e.g.
`TX_2G_11BE_BW40_MCS13`: `2422_11BE_BW40_MCS13` -> `2452_11BE_BW40_MCS13` ->
`QUIT`; `RX_5G_11AX_BW80_MCS11_RSSI_CAL`: `5210_...` -> `5530_...` ->
`5775_...` -> `QUIT`, with each measurement row's FailPath ->
`DUT_INITIALIZE_5G`.

### Device catalog (row field 1) and functions (row field 2)

All first fields seen in this bundle (rows mix `Test Program` and
`TEST PROGRAM`; runtime matching is case-insensitive, keep tokens verbatim
for log joins):

| DEVICE | Functions verified in this bundle |
|---|---|
| `Test Program` / `TEST PROGRAM` | `Delay`, `CMD_PING`, `LABEL_COMPARE`, `LABEL_CONTAIN`, `LABEL_SUBSTRING_EX`, `VALUE_IN_RANGE`, `SET_ISN`, `LOCK`, `NEXT_SCAN`, `SCAN_BARCODE`, `Loop`, `COMMAND`, `Confirm_DIALOG`, `CHANGE_RESULT_PIC`, `AUTO_RUN`, `SFIS_TO_VAR` |
| `TELNET` | `OPEN`, `SENDSTR`, `SENDSTR_AND_WAITSTR`, `RECV_ONE_REG`, `SENDSTR_AND_RECV_ONE_REG` |
| `DLLWIFI_TEST` | `GLOBAL_SETTING`, `TESTER_REGISTER`, `TESTER_CONNECTION`, `TESTER_DISCONNECTION`, `LOAD_PATH_LOSS`, `DUT_INSERT`, `DUT_INITIALIZE`, `DUT_REMOVE`, `TX_MULTI_VERIFICATION`, `RX_VERIFY_PER`, `BRCM_PA_TRIMMING`, `BRCM_PA_PREDICTION`, `BRCM_GET_SROM`, `BRCM_SET_SROM_TO_PAVARS`, `BRCM_RSSI_CALIBRATION`, `BRCM_BEAMFORMING_TEST`, `PRED_SETTING`, `PRED_INITIALIZE`, `LOAD_STATE`, `LOAD_MODEL`, `SAVE_STATE` |
| `DLLGoldenKV` | `COMPARE_GOLDEN_ISN`, `COMPARE_GOLDEN_CAL_PATHLOSS_MD5`, `COMPARE_GOLDEN_VEL_DATE`, `COMPARE_GOLDEN_CAL_DATE`, `UPDATE_CAL_INFO`, `UPDATE_VEL_INFO`, `RECORD_TESTRESULT`, `WRITE_DATA_TO_BIN`, `CONVERT_BINDATA_TO_DIC`, `GOLDEN_CALIBRATION`, `COMPARE_PATH_LOSS_DATA`, `GOLDEN_VERIFY` |
| `DLLPOWER_SWITCH` | `REPOWERON` |
| `SFIS` | `TSP_CHECKROUTE`, `TSP_REPAIR`, `TSP_GETVERSION`, `WRITE_DATA_TO_SFIS` |
| `IPLAS` | `WRITE_DATA` |
| `FTP` | `COPY_TEST_RESULTS_ALLCSV` |
| `CMD_CONSOLEV2` | `OPEN`, `WAITSTR` |
| `MACRO` | call into a `MACRO_LIST.ini` section |

Function notes worth knowing during audits:

- `BRCM_PA_PREDICTION` is the prediction-driven PA trim used for 5G
  BW40/BW80/BW160; `BRCM_PA_TRIMMING` remains for 2G and 5G BW20. The
  prediction engine is initialized by `WIFI_PRED_INITIALIZATION`
  (`PRED_SETTING` with
  `*V:BANDWIDTH_CHANNEL_LIST_5G#&*V:DRIFT_PARAMETERS#&*V:COUNTER_LIMITS#`,
  `PRED_INITIALIZE DUT_NUMBER&BAND=5G`, `LOAD_STATE`,
  `LOAD_MODEL FILE_DIR=Support`) and persisted by `SAVE_PREDICTION_STATE`
  (`SAVE_STATE`).
- `DUT_INITIALIZE` params vary by phase: plain
  (`PROJECT=HH5K&BAND=2G&WL_INTERFACE=wl0`), RX
  (`TEST_CASE=RX&STANDARD=11AC`), IBF (`DYNAMIC_SHELL_SCRIPT=IBF`).
- `Loop|Test_Count,N` appears both as a controller row
  (`Loop_LOAD_SHELL = "Test Program|Loop|Test_Count,3|NULL|10000|Delay_For_TELNET|FAIL"`)
  and as a retry wrapper around PC-console openings (`OPEN_TEST_PC_CMD`
  macro: Loop -> `CMD_CONSOLEV2|OPEN TEST_PC@CMDProxy.exe` ->
  `WAITSTR TEST_PC@Microsoft` -> `Delay_Open` -> `QUIT`).
- `AUTO_RUN|TRUE/FALSE` sets the station's auto-next-unit flag;
  `CHANGE_RESULT_PIC` switches the operator screen picture. The AUTO_RUN
  item also carries the operator dialog
  `Confirm_DIALOG "Tolong Keluarkan Board dari Shielding Box!"`
  (Indonesian: please remove the board from the shielding box).
- `LOCK|DMM,1/0` and `LOCK|Manual,1,Manual*V:Setting-DUT_Number#` guard
  station resources during barcode scans and route dialogs; `NEXT_SCAN`
  releases the scanner.
- `IPLAS|WRITE_DATA` params: `Project=HH5K&TSP=Wireless_Test_2_5G`
  (production), `Project=HH5K_GOLDEN&Model=Calibration|Verification`, or
  `Type|Offline` (offline fallback in route handling). These values surface
  as iPLAS metadata rows in the DUT log.

### Criteria field (row field 4) semantics by function

| Function | Criteria meaning | Verified example |
|---|---|---|
| `CMD_PING` | `TRUE` / `FALSE` (inverted: expect no reply) / `PassN=k&Connect=TRUE` / `PassN=k&TTL=64&Connect=TRUE` | `PassN=4&TTL=64&Connect=TRUE`; plain `FALSE` when waiting for DUT power-off |
| `LABEL_COMPARE` | two `name,expr` pairs; equal -> PassPath | `iPLAS Flag,*V:TestInfo-iPLAS_Enable#` vs `iPLAS default,1`; `NVRAM_FILE_MD5,*V:Label-NVRAM_MD5_BEFORE_TEST#` vs `Calibrated MD5,*V:Label-NVRAM_MD5_AFTER_TEST#` |
| `LABEL_CONTAIN` | label + substring | `P_Msg,*T:Label-P_Msg#` vs `Check String,PLS CHANGE DEVICE` |
| `LABEL_SUBSTRING_EX` | extract substring into a label | `*T:Label-P_Msg#,[,]` -> `Station` |
| `VALUE_IN_RANGE` | label + numeric window | `Counts of Failure: ,*T:Label-LF#,times` vs `2,0` |
| `SCAN_BARCODE` | label to store the scan | `SCAN ISN BarCode,15` with criteria `ISN` |
| `TELNET OPEN` | `login,prompt` pair | `*V:DUT_LOGIN_SETTING-LOGIN#,*V:DUT_LOGIN_SETTING-LOGIN_PROMPT#` |
| `SENDSTR_AND_WAITSTR` | wait string | `#`, `DONE`, `done`, `has been frozen`, `HH5K`, `pega_write`, `messages` |
| `RECV_ONE_REG` | regex with `(?<Value>...)` captured into the label named after `,` in PARAM | `SERIAL_NUMBER:+\s*(?<Value>[\w\d]{15}),ISN` (criteria `NULL`) |
| `SENDSTR_AND_RECV_ONE_REG` | PARAM = command + regex; Criteria = label to store | `LOAD_SH@md5sum /opt/conf/.kernel_nvram.setting_CA,setting_CA+\s*(?<Value>[\w\d]{32})` with criteria `NVRAM_MD5_BEFORE_TEST` |
| DLL measurement rows | spec expression: whole-section ref, direct k=v pairs, or nested `*V:*V:Macro-Criteria_Criteria##` | `*V:2G_PA_TRIMMING_SPEC#`, `PER=*V:2G_RSSI_CAL_SPEC-PER#&RSSI=*V:2G_RSSI_CAL_SPEC-RSSI#` |
| `DLLGoldenKV` | key filters and limits | `CONTAINKEY=POW&LIMIT=1.5`, `ANT_RANGE=0.5&FRQ_RANGE=20`, `LIMIT=<6000,10~30` |
| `FTP COPY_TEST_RESULTS_ALLCSV` | `folder,enable_flag` | `*V:SELECTED STATION-NAME#\*V:Script_Info-Ver#,*V:TestInfo-FTP_Enable#` |
| `Confirm_DIALOG` | message text; criteria `NULL` | `Please Enable TFTP Server and Select right Folder` |

Note the label-location asymmetry: `RECV_ONE_REG` carries the capture label
inside PARAM (`,ISN` suffix, criteria `NULL`) while
`SENDSTR_AND_RECV_ONE_REG` carries it in the Criteria field. Parsers must
handle criteria semantics per function.

### Where criteria and target values live (inline vs `*V:`)

Measurement targets and spec parameters may be written **inline in the
`MAIN_FLOW.ini` row itself** or **referenced from the support files** via the
`*V:` variable pattern; the spec-review pipeline must resolve both forms:

- Inline direct values live in the row's PARAM field
  (`DLLIOT|TX_MULTI_VERIFY|ANT=1&FREQ=2402&TX_TARGET_POWER=10|*V:BT_TX_SPEC#|...`
  in the 6G bundle's `[BT_TX_MULTI_VERIFY]`, and
  `RX_TARGET_POWER=-92&FRAME_COUNT=200` in `[BT_RX_VERIFY_PER]`). A direct
  numeric target is authoritative on its own.
- Referenced values read other bundle files through `*V:<SECTION>-<KEY>#`,
  e.g.
  `TX_TARGET_POWER=*V:2G_TX_TARGET-11B_BW20_CCK11_POWER#|Criteria=2G_TX_11B_CCK11_SPEC`
  pulls the target from `TEST_SPEC.ini [2G_TX_TARGET]`, and whole-section
  criteria like `*V:BT_TX_SPEC#` / `*V:2G_PA_TRIMMING_SPEC#` expand a
  TEST_SPEC section into the row's criteria. Support files that may be
  consulted this way: `TEST_SPEC.ini`, `RD_SETTING.ini`,
  `FACTORY_SETTING.ini`, `PARAMETER_LIST.ini`, and `MACRO_LIST.ini`
  (macro rows add `*V:Macro-...` scopes).
- Offset-form spec windows (`POWER=-4~1` in `[BT_TX_SPEC]`,
  `POWER=-3~5dB` in fixture specs) are relative to the row's target —
  inline (`TX_TARGET_POWER=10`) or referenced — and the displayed Script
  limit is `target+low to target+high` (e.g. 6 to 11dBm). Absolute windows
  with positive bounds display as-is.
- Catch-all spec sections may key values by band
  (`NFT_TX_VERIFY_Spec`: `2G_Tx_Power` / `5G_Tx_Power` / `6G_Tx_Power` /
  `IOT_Tx_Power`); a requirement naming one band resolves only the matching
  key.

### PARAM field idioms (row field 3)

- `k=v&k2=v2` chaining everywhere; whole-section references inject entire
  sections into params (`*V:Setting#&*V:DUT_INSERT_SET#` into `DUT_INSERT`;
  `*V:WIRELESS_GLOBAL_SETTING#` into `GLOBAL_SETTING`;
  `*V:2G_IBF_TEST#&AUTO_CHANGE_VER_FREQ=...` into `BRCM_BEAMFORMING_TEST`).
- Telnet session prefix before `@` names the log session: `LOAD_SH`,
  `SAVE_DATA`, `RX_GAIN`, `RX_GAIN_5G`, `READ_PAVARS*V:Label-DUT_Number#`,
  `TEST_PC` (CMD_CONSOLEV2). Addressing form:
  `OPEN|alias@*V:Setting-SRCNIC_IP#:0,*V:Setting-DUT_IP#:23`.
- Escapes inside quoted values: `\x7C` for `|` (shell pipes in
  `pavars \x7C grep pa2g`), `\x0D` CR in Zigbee/Z-Wave command lines.
- Embedded double quotes are legal inside the quoted value:
  `LOAD_SH@scos-iot-send-cmd -d /dev/ttyS1 -c "bootload"` — the value runs
  from the first `"` to the last `"` on the line.

### PassPath / FailPath idioms found here

1. **Flag gating without failure.** `SET_IPLAS_INFO_WIRELESS_1`:
   `Check_Flag = "Test Program|LABEL_COMPARE|iPLAS Flag,*V:TestInfo-iPLAS_Enable#|iPLAS default,1|2000|SET_IPLAS_INFO|Delay_QUIT"`
   — iPLAS enabled -> write `Project/TSP`; disabled -> `Delay_QUIT`
   (delay then `QUIT`) and the item still passes. FailPath is "next step on
   mismatch", not "fail the item".
2. **Recovery ladder inside macros.** TX/RX macros on measurement failure:
   `DUT_REMOVE` -> `Delay` -> `CMD_PING` (`#Retry:3`) -> `DUT_INSERT#RETRY:2`
   -> `DUT_INITIALIZE` -> re-measure `#Retry:3` -> `QUIT`. The re-measure
   row's FailPath is plain `FAIL` (no second recovery round).
3. **Loop re-entry through a dialog.** `LOAD_SHELL_SCRIPT_ALL`:
   `Loop_LOAD_SHELL|Test_Count,3` wraps tftp/tar setup; any step failure ->
   `CHECK_DUT_STATUS#Retry:5` (ping, eighth-field ERRCODE) ->
   `Check_TFTP_Status` `Confirm_DIALOG` ("Please Enable TFTP Server and
   Select right Folder") -> `Loop_LOAD_SHELL`. Bounded by the loop count;
   after exhaustion the Loop row's FailPath (`FAIL`) ends the item. This is
   the canonical bounded loop-retry shape; only cycles where an active Loop
   row's FailPath re-enters the cycle are infinite-loop findings.
4. **Inverted-name outcome labels.** `NVRAM_MD5_Check`: MD5 unchanged ->
   `Compare_NVRAM` PassPath -> `NVRAM_FAIL` -> forced-fail row
   `LABEL_COMPARE|PASS,0|FAIL,-1|NULL|Delay_QUIT|FAIL` (0 vs -1 never equal
   -> FailPath `FAIL`). MD5 changed (calibration wrote data) -> FailPath ->
   `NVRAM_PASS` -> delay -> `QUIT`. Label names describe outcomes, not
   compare polarity; do not rename or "simplify" these rows.
5. **Same-target paths.** `[Check_ZWAVE_SGC_FW] Check_CMD` uses
   `|5000|ZW_GPIO54_Cmd|ZW_GPIO54_Cmd` — pass and fail converge; the row
   exists for its side effect (sending a space to flush the console).
6. **Segmented retry chains.** `IBF_5G_TEST` splits four 5G sub-bands:
   `IBF_5G_TEST_BAND1..4#Retry:5` chained by PassPath; band-1 failure ->
   `DUT_5G_IBF_INITIALIZE` -> `IBF_5G_TEST_BAND1_Re#Retry:5`; bands 2-4
   FailPath is plain `FAIL`. `IBF_2G_TEST` is the single-band version with
   `AUTO_CHANGE_CAL_FREQ=2437,2412,2462`.
7. **Diagnostic detour on calibration failure.** `WIFI_5G_PA_TRIMMING` BW20
   cal failure -> `Check_5G_FREQ` (`MACRO|FREQ_VERIFICATION` at 5300,
   `CSV_TAG=_Check_FREQ`) -> pass: `5G_TX_BW20_CAL_MCS7_Retest`; fail:
   `FAIL`. The CSV_TAG distinguishes fixture-vs-frequency problems in
   measurement exports.
8. **Multi-budget rows.** `PING_DUT#Retry:30` appears inside items whose
   other rows carry `#Retry:2/#Retry:3`; each budget is independent, and the
   `#RETRY:` casing variant is equivalent.

### Effective limits from TEST_SPEC.ini (Sample: HH5K values)

- TX verify = target +/- `POWER` tolerance. Targets (`[2G_TX_TARGET]`):
  `11B_BW20_CCK11=23`, `11N_BW20_MCS0=21`, `11N_BW40_MCS7=20`,
  `11AC_BW20_MCS8=20.5`, `11AX_BW40_MCS11=18.5`, `11BE_BW40_MCS0=21`,
  `11BE_BW40_MCS13=18` dBm. (`[5G_TX_TARGET]`): `11N_BW40_MCS7=17.5`,
  `11AC_BW20_MCS8=17`, `11AX_BW20_MCS0=19`, `11AX_BW80_MCS0=19`,
  `11AX_BW80_MCS11=16`, `11BE_BW80_MCS0=19`, `11BE_BW160_MCS13=17` dBm, plus
  CAL targets `11AC20MCS0/7=19`, `11AC40MCS7=17.5`, `11AC80MCS7=19`,
  `11AC160MCS7=17`, `11AX80MCS7=19`. Tolerance `POWER=1.5dB` everywhere.
- Fixture checks use relative windows `POWER=-3~5dB`
  (`2G_TX_FIXTURE_SPEC`, `5G_TX_FIXTURE_SPEC`, `_Composite` variants) and
  `FREQ_ERROR=20PPM` — e.g. 2G MCS0 target 21 dBm renders bounds 18-26.
- Per-mode EVM (dB, upper bounds): 2G `11B=-9.1, 11N20MCS0=-5,
  11N40MCS7=-27, 11AC20MCS8=-30, 11AX40MCS11=-35, 11BE40MCS0=-5,
  11BE40MCS13=-38`; 5G mirrors with `11AX80MCS0=-5`. All TX rows also carry
  `MASK=0`, `FREQ_ERROR=20PPM`, `LO_LEAKAGE=-90~-20PPM`.
- PA trimming targets: 2G `11B=23/BW20=21/BW40=20dBm`; 5G
  `BW20=19/BW40=17.5/BW80=19/BW160=17dBm`; window
  `PA_UPDATE_LIMIT_UPPER=5dB`, `PA_UPDATE_LIMIT_LOWER=0.1dB`; the 2G
  SPEC section adds `PA_OFFSET_MODE=PA_OFFSET_ABSOLUTE`,
  `PA_OFFSET_ABSOLUTE=0.3dB`. POW_OLD bounds = target +/- UPPER (2G CCK:
  18-28); POW_DIF_ABS below LOWER (0.1dB) leaves `SROM_NEW == SROM_OLD`.
- RSSI calibration: 2G `POWER_LEVEL=-82dBm`, 5G `-50dBm`, `FRAME_COUNT=500`;
  criteria `PER=10%`, `RSSI=-10dB~10dB`.
- RX sensitivity (`[2G_RX_POWER_LEVEL]`): OFDM6 -92 / 11N MCS7 -75 / 11AC
  MCS0 -92 / 11AX MCS0 -89 / 11BE40 MCS0 -90 / 11BE40 MCS7 -61 dBm;
  (`[5G_RX_POWER_LEVEL]`): 11AX80 MCS0 -85 / MCS11 -56 / 11BE160 MCS7 -64
  dBm; `FRAME_COUNT=500`, criteria `PER=10%` (renders `<10,0>`).
- FRAME_COUNT semantics: script RX params carry the NOMINAL capture count
  (e.g. `FRAME_COUNT=*V:2G_RX_FRAME_COUNT-11B_BW20_CCK11#` -> 200, or a
  flat `FRAME_COUNT=500`). TP/MTP documents usually express the frame
  window as nominal +/-10% (the PER acceptance window): `180 ~ 220` for a
  nominal `200` (200 + (10% x 200) = 220, 200 - (10% x 200) = 180).
  Comparisons must therefore use containment (nominal inside the window),
  not endpoint equality — endpoint equality would compare a nominal against
  a window edge and report a false mismatch.
- IBF: `[2G_IBF_TEST]` / `[5G_IBF_TEST]` `IBF_SPEC=4.5dB~12dB` (renders
  `<12,4.5>`).
- GOLDEN specs: `[2G_GOLDEN_SPEC]`/`[5G_GOLDEN_SPEC]` `POWER=1dB`;
  `[GOLDEN_CAL_SPEC]` `POWER=5dB`; `[GOLDEN_VER_SPEC]` `POWER=1dB`.

### Bundle-specific lint expectations

- Six-field rows (missing `WAIT_TIME`): `Check_ZB_Bootloader`,
  `Check_DUT_Version`. Treat as warning-level compatibility findings and
  verify runtime tolerance; do not silently normalize.
- Trailing whitespace after the closing quote on the `Unzip_Table_File` row
  (`LOAD_SHELL_SCRIPT_ALL`); strict `"..."$` grammars reject it, tolerant
  ones accept it — decide deliberately.
- Duplicate section names across files: `[TSP_CheckRoute]` exists both as a
  MAIN_FLOW item (active, used by the flow) and as a `MACRO_LIST.ini` macro
  (alternative called via `MACRO|TSP_CheckRoute` in a commented row);
  `[Script_Info] Ver` exists in MAIN_FLOW and RD_SETTING. Resolve by file.
- Item DUT_BOOT_UP_PING (MAIN_FLOW item) vs macro DUT_BOOT_UP_Ping
  (MACRO_LIST) differ only in case; the production flow resolves to the
  MAIN_FLOW item section.
- `IQ_LOAD_PATHLOSS` hardcodes `PATH_LOSS_NAME=pathloss_WiFi.csv` while
  `RD_SETTING.ini` says `Pathloss_WiFi.csv` — same file, case-insensitive
  runtime; do not flag as mismatch.
- Commented rows contain known copy-paste defects (e.g. a commented
  `RX_2G_11BE_BW40_MCS0` row references the `FRAME_COUNT` MCS7 key in a
  MCS0 context); commented rows are never evidence.

## Recommended Scripts Correctness

Before sequence, coverage, or macro analysis, run a source-level preflight in
this order:

1. Read physical lines without stripping meaningful spaces inside quoted values.
2. Classify comments using only `;` and `//` after indentation.
3. Validate TEST/FINAL/INIT headers and pair TEST/FINAL names.
4. Validate active flow rows against the exact unquoted-key/quoted-value grammar.
5. Reject leading/trailing whitespace inside a quoted test-item value.
6. Report single-slash lines and other malformed active lines instead of
   silently skipping them.
7. Build the ordered TEST FLOW list from RHS item values only.
8. Parse item sections and macro sections separately, then validate every
   `MACRO|FunctionName` target.
9. Resolve `*V:` references and validate the resulting seven-core/eight-field function
   rows, including PassPath and FailPath targets.
10. Validate limits and runtime settings against the selected station and
    TP/MTP revision.

This ordering prevents a syntax error from being misdiagnosed later as a
missing test item, a sequence mismatch, an unresolved macro, or a TP/MTP
coverage defect.

## Six-File Correctness Model

The script bundle is a small program spread across six INI files. A correctness
checker should validate each file according to its role instead of applying one
generic key/value rule to everything.

### `MAIN_FLOW.ini`

Check:

- TEST FLOW headers and their optional INIT/FINAL companions.
- Exact normalized TEST/FINAL names, duplicate flow names, and SKU identity.
- Active flow rows using unquoted keys and quoted RHS item names.
- Leading/trailing whitespace inside quoted item names.
- Comment prefixes and invalid single-slash lines.
- RHS item sections and `#ErrCode:` metadata.
- Separation of production, sampling, golden, and calibration flows.

The RHS item is the execution identity. The LHS key is an alias used inside the
INI file and may differ, as shown by `DELAY = "Wait_WIFI_Stable"`.

### `MACRO_LIST.ini`

Check:

- Unique macro sections and active seven-core/eight-field rows.
- Macro call targets from MAIN_FLOW.
- PassPath/FailPath targets, terminal tokens, and `Next` fall-through.
- Nested macro calls and recursive macro cycles.
- `#Retry:N` budgets separately from `Loop|Test_Count,N` controllers.
- CSV_TAG and other runtime parameters after variable resolution.

Wireless-1 demonstrates distinct macro state machines for normal TX, fixture
TX, frequency verification, normal RX, RSSI calibration, collection-only RSSI,
and 11BE RX. These must not be collapsed into one generic macro behavior.

### `TEST_SPEC.ini`

Check:

- Section/key syntax and duplicate keys.
- Range direction (`low <= high`) and unit compatibility.
- Separate target, fixture, PA trimming, RX power/frame-count, calibration,
  and collection-only sections.
- References from resolved macro Criteria/Parameters.
- Whether a script limit is numeric, qualitative, or intentionally commented.

For Wireless-1, examples include `2G_TX_TARGET`, `5G_TX_TARGET`, fixture spec
sections, PA update offsets, RX power levels, RX frame counts, and RSSI/PER
criteria. A global flattened limit map would incorrectly mix these roles.

### `RD_SETTING.ini` checks

Check:

- Runtime login/prompt fields as a complete group.
- Radio mappings such as `2G=wl0`, `5G=wl1`, and `6G=wl2`.
- Referenced shell/support filenames and their existence when local files are
  available.
- Firmware/version and path-loss settings.
- Missing or duplicate runtime keys without exposing credential values in the
  report.

### `PARAMETER_LIST.ini` checks

Check:

- Section/key syntax and duplicate keys.
- Numeric ranges and known enumerations such as analysis modes.
- References expected by macro parameters.
- Separation from executable flow rows: this file supplies runtime behavior,
  timing, averaging, and analysis defaults; it does not define test order.

### `FACTORY_SETTING.ini` checks

Check:

- Site/station identity and required environment sections.
- IP, port, boolean, and enumerated value formats where the station profile
  defines them.
- Duplicate device/radio identifiers.
- Separation of factory/site defaults from RD runtime settings.

### Cross-folder discovery

The workspace uses both `scripts` and `Scripts`, lowercase and mixed-case
station directories, flat SZ chip bundles, nested QCA bundles, and BRCM bundles
with one subdirectory per station. Discovery must locate files by role name,
not by one hard-coded casing or directory depth. On Windows, path lookup is
case-insensitive, but reports should preserve the original path and station
spelling.

## Correctness Checker Implementation

The checker is implemented as
`../scripts/scripts_correctness_checker.py` (module docstring and tests in
`../scripts/test_*.py` describe the current diagnostics). It calls the existing
`scripts_flow_parser.py`, `scripts_flow_graph.py`, `scripts_spec_parser.py`, and `scripts_context.py`
parsers, then adds source-line diagnostics and cross-file validation. Run it
once during `run_audit.py` setup, before log workers are submitted.

Use three severities:

- **Error**: invalid active flow row, malformed header, broken seven-core/eight-field row,
  unresolved selected-flow branch, or invalid required reference.
- **Warning**: missing optional FINAL/INIT flow, orphan helper section,
  duplicate non-selected helper section, suspicious range/unit, or missing
  optional support file.
- **Info**: unused helper macro, commented-out alternative, or recognized
  station-specific convention.

The preflight result should be stored as structured evidence with file, line,
code, message, and original source text. It should remain separate from DUT
measurement anomalies and Test Spec Review so a script syntax defect is not
misreported as a log or TP/MTP failure.

## General Use Across Stations

This INI layout is reusable because it separates the responsibilities cleanly:

- `MAIN_FLOW.ini` handles orchestration.
- `MACRO_LIST.ini` handles reusable executable logic.
- `FACTORY_SETTING.ini` and `RD_SETTING.ini` handle environment and runtime settings.
- `PARAMETER_LIST.ini` handles analysis defaults and global tuning values.
- `TEST_SPEC.ini` handles limits and expected results.

That separation makes it easier to:

- Port the same software pattern to another station.
- Update a single setting without rewriting flow logic.
- Reuse macros across many items.
- Keep the test program behavior predictable.

## Summary Guidance Context

When reading a future station’s test scripts, infer these roles first:

- `MAIN_FLOW.ini` is the entry point and orchestration layer.
- `MACRO_LIST.ini` contains the executable macro implementations.
- The other INI files are usually supporting value banks.
- Comments are useful for context but should not be treated as live execution.
- The same section and substitution patterns usually apply even when the actual names change.

If a new station uses the same structure, the safest first step is to identify the flow entry points, the macro sections, the variable banks, and the specification bank before analyzing individual test steps.

## False-Positive and Interpretation Rules

When judging whether an engine finding is a real defect, apply these rules:

- A measurement out-of-bounds FAIL is not a script bug by itself; verify the
  limit source and the script's criteria mapping before calling it a script defect.
- A naming alias is not a script bug when the family, channel, and limits still
  match the intended measurement.
- A retry loop is not a duplicate item when it is explicitly labeled as a retry
  branch (`#Retry:N` or an equivalent retry branch).
- Commented-out loop rows and alternative branches are never executable retries,
  loops, or test steps; do not count them as evidence.
- When explaining a script failure, trace the active result paths until a
  standalone `QUIT`, `FAIL`, or another terminal outcome is reached; do not call
  a downstream item missing until the active branch makes it unreachable.
- Treat a `Next` handoff at an item boundary as execution of the next selected
  test-flow item, not as an unexplained missing row in the current item.
- A blank `TP/MTP limit` is legitimate only when the selected TP row has no
  explicit numeric criterion (for example, a qualitative `PA trimming PASS`
  statement). A blank caused by an unparsed result table is not evidence that
  the TP has no limit.
- PASS/FAIL-only `TX_*` / `RX_*` execution-status items (for example
  `TX_6G_11AG_BW20_OFDM6`) are the DUT log envelope for the same TX/RX tests
  that already have measurement rows; do not treat them as extra findings.
