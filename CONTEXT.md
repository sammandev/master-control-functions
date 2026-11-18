# Web Program Functions Database

A CRUD web application over the script-definition knowledge workbook (`.dev-reference/Master_Control_V2_Script_Definition_Knowledge.xlsx`). It manages the definitions of test-script functions, their parameters and criteria, and usage examples, with the workbook as a repeatable upsert source.

## Language

### Core entities

**Function**:
A single test-script command, uniquely identified by `(source, interface, name)`. Its `interface` is its Device name; blank means Global. Parameter and Criteria links may also reference it from other Devices.
_Avoid_: command, method, API call

**Source**:
Where a Function is implemented. One of `BUILT-IN` or `DLL`.

**Interface**:
An alias for Device. Function records retain the field name `interface` for API and workbook compatibility; it does not identify a separate catalog.

**Device**:
A named execution context, such as TELNET, GPIB or WIFI_TEST. Device, Device Name and Interface refer to the same concept. A Device can be **registered** before it has content, or **derived** from Function interfaces, links, definitions or Examples. Blank means Global. Names are trimmed and uppercased when saved. A rename preserves structured references; saved script text is unchanged.
_Avoid_: scope (as a user-facing term)

**Device scope**:
The Device a link, definition or Example applies to, stored as `device_name`. It uses the same Device names as a Function's `interface`. Blank means Global/shared. Forms offer the Device catalog and a blank choice.

**Parameter definition**:
A dictionary entry in `parameter_main`: the name, type, unit, default, allowed values, and pattern of a Parameter, independent of any Function. Identified by `(name, device_name)`; a blank `device_name` is the global definition, and a device-scoped one overrides it for that device's links (ADR-0009).

**Parameter link**:
A row in `function_parameter`: binds a Parameter definition to a Function at a `position` with a `required` flag and a Device scope.

**Criteria definition**:
A dictionary entry in `criteria_main`: the verification threshold definition (same shape and same `(name, device_name)` keying as a Parameter definition, ADR-0009).

**Criteria link**:
A row in `function_criteria`: binds a Criteria definition to a Function.

**Example**:
A usage record: one pipe-delimited script string plus a human "meaning", attached to a Function with a Device scope. Examples demonstrating Criteria remain Function usages, not separate records owned by a Criteria definition.

### Data-quality terms

**Orphan**:
A link or Example row whose target Function, Parameter definition, or Criteria definition does not exist yet. Orphans are kept and flagged in the UI, never dropped.

**Normalization**:
Import-time fixes applied deterministically to source data (e.g. `GoldenKV` → `GOLDENKV`, `text_execution` → `test_execution`), recorded in a change log.

### App-specific terms

**Composer**:
The form-driven tool on a Function page that emits that Function's script string from parameter/criteria values; its output can be saved as an Example. Emission is server-owned.
_Avoid_: script builder

**Export**:
Regenerating the complete six-sheet workbook from the application database in the original column layout; the result re-imports without data loss.
_Avoid_: dump

### Sequences

**Sequence**:
A named, ordered set of Sequence steps stored in the app. Web-native: it never enters the workbook's six sheets. Identified by a unique name.
_Avoid_: program, macro, script (a script is a single Function's string)

**Sequence step**:
One ordered row of a Sequence: a Function reference (by source, interface, name) plus parameter/criteria values, per-step timeout/flow/on-fail, device scope, and an optional note. Steps are the source of truth; the sequence's emitted text is always generated from them, never parsed back.
_Avoid_: line, command
