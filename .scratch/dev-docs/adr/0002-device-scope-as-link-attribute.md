# Device scope is an attribute on link rows, not an entity

The workbook carries `device_name` on `function_parameter`, `function_criteria`, and `example` rows. It is NOT the Function's `interface` (a 3-part join on source+interface+name loses rows; `IOT` is not an interface), and it is not modeled as its own device table. We decided: link rows join Functions by `(function_source, function_name)` and carry `device_name` as a free-text scope attribute meaning "these parameters/criteria apply when the Function runs against this device".

## Considered Options

- Own `devices` table with FK — rejected: source data is free-text with casing variants (GOLDENKV/GoldenKV) and values that overlap but don't match interfaces; an entity would force a taxonomy the workbook doesn't have.
- Merge away as redundant — rejected: it would drop information.
- Treat as the Interface — rejected: disproved by the join analysis.

## Consequences

- `device_name` casing is normalized on import (recorded in the change log) but otherwise unvalidated free text.
- Editing a link row means editing a link attribute, not creating/choosing a Device record.
