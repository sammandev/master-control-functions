# Device-scoped parameter and criteria definitions

The knowledge workbook now carries `device_name` on `parameter_main` and `criteria_main` rows, and the database dictionaries key definitions by `(name, device_name)` with a blank scope meaning "global/shared definition". This amends ADR-0002, which scoped device to link rows only.

## Context

OSP and WIFI_TEST both define parameters named `ANT` and `FREQ` with incompatible semantics (bitmask integers vs `ANT_<n>` strings; comma-list MHz vs `2412MHZ` strings). The dictionaries assumed name-global uniqueness, so a second project's definition overwrote the first on import (`ON CONFLICT (name) DO UPDATE`, last row wins) and every link join (`d.name = l.name`) served one definition's semantics to both devices' links.

## Decision

- `parameter_definitions` and `criteria_definitions` gain a `device_name TEXT NOT NULL DEFAULT ''` column and a `UNIQUE (name, device_name)` key (replacing `UNIQUE (name)`).
- Blank `device_name` = **global/shared definition**, valid for every device. This is the default for the pre-existing catalog: most parameters (e.g. `CSV_TAG`, `POWER`) genuinely have one shared definition.
- Import upserts definitions by `(name, device_name)`; `device_name` casing follows the ADR-0002 uppercase normalization, recorded in the change log.
- Link-to-definition resolution is **exact scope first, global fallback**: a link `(name, device)` resolves to the definition with the same device scope when one exists, else to the blank-scope definition, else the link is an Orphan (ADR-0003 — kept and flagged, never dropped).
- The workbook's `parameter_main`/`criteria_main` sheets gain a `device_name` column directly after `name`; Export emits it and the round-trip invariant (export → import, no data loss) holds.

## Considered Options

- `function_name` on definitions — rejected: `function_parameter`/`function_criteria` already are the name×device×function join tables; adding function columns would duplicate the join, force one definition row per function for shared parameters (`ANT` serves ~15 functions), and create two sources of truth for the same fact.
- Renaming the script keys (`OSP_ANT`) — rejected: `name` is the literal script-language token; renaming breaks script fidelity and the Composer's emitted strings.
- Keep name-global definitions and disambiguate in prose — rejected: prose is unqueryable and the overwrite corruption remains.
- A `devices` entity table — rejected again (same reasoning as ADR-0002): free text, casing variants, no source taxonomy.

## Consequences

- Definitions are still Function-independent (CONTEXT.md unchanged on that point); they are now Device-scope-aware instead of implicitly global.
- All read paths that join definitions to links (function links, compose, orphans) must use the exact-scope-then-global resolution; a join on `name` alone is now ambiguous by construction.
- The Composer and Orphan report inherit the resolution rule automatically through the shared join.
- Exported workbooks and re-imports preserve the scope column; blank stays blank.
