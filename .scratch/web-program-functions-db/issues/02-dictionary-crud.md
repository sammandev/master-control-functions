# 02: Dictionary CRUD (parameter + criteria definitions)

**What to build:** A knowledge-base maintainer can manage the Parameter definition and Criteria definition dictionaries on dedicated pages: browse/search/paginate, create/edit/delete, with usage counts showing which Functions reference each entry. Edit forms enforce type, unit, default, allowed values (set syntax), cardinality (SINGLE/MULTIPLE), value pattern, and required flag per the definitions' own constraints.

**Blocked by:** 01 (Function CRUD end-to-end).

**Status:** done

- [x] Parameter definition + criteria definition schemas + migrations (same shape, separate tables)
- [x] REST endpoints for both dictionaries: list (search/sort/pagination), get, create, update, delete
- [x] Delete of a definition still referenced by links returns a warning/clear error per contract (no silent breakage)
- [x] UI: two dictionary pages with instant search, pagination, create/edit/delete, usage counts per entry
- [x] Field validation enforced server-side and surfaced as inline field errors; allowed-values set parsing covered
- [x] Integration tests through the HTTP seam cover CRUD, validation rejections, and usage-count correctness

## Comments

**2026-10-01 — implementation notes**

- Migration 0002 creates ALL remaining entity tables (parameter_definitions, criteria_definitions, parameter_links, criteria_links, examples) in one migration: the dictionaries' usage counts join real link tables, and tickets 03 (import) and 04 (links) become data-only work. Link tables carry no FKs (ADR-0003 orphans must be storable); uniqueness on (function_source, function_name, device_name, name).
- Validation is the definitions' own constraints, enforced server-side: type ∈ {int, double, string, list[int]}; cardinality ∈ {SINGLE, MULTIPLE}; allowed_values must parse as a `{v1,v2,…}` set with every member parseable per type; default_value must parse per type AND be a member when a set exists. Field-level 422s name the failing field and reason.
- Delete semantics (ADR-0003): DELETE → 409 `{error, usage_count}` when referenced; `?force=true` → 204, link rows survive as orphans for the report/resolution flow (ticket 05). Unreferenced deletes are unconditional 204. UI confirms with the reference count before force-deleting.
- API JSON uses `required: true/false` (bool) — stored as INTEGER 0/1. Live smoke caught the UI initially sending 0/1; fixed in the client before commit.
- Shared generic `dictionaryRoutes[T]` handler set (one implementation, two adapters) — parameter and criteria dictionaries stay behaviorally identical by construction.
- UI: one shared DictionaryPage component, two thin routes (`/parameter-definitions`, `/criteria-definitions`), nav links; search/sort/pagination/CRUD with inline field errors; usage column on parameters; type-change clears default/allowed inputs to avoid stale invalid values.
- Verification: 7 HTTP-seam tests (round-trip, 7-case validation table, duplicate 409, update/name-conflict 409, pagination+search, usage-count + delete-guard + force-orphan survival, criteria CRUD); `go vet` clean, `svelte-check` 0 errors, full suite green; live smoke: create/validation-rejects/duplicate/list/update+rename/delete/SPA all correct.
