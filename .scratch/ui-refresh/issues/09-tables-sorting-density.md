# 09: Tables breadth — sorting everywhere + density toggle

**What to build:** Every list table sorts by its columns (Functions and the dictionaries already do; Examples, Devices, Orphans, Sequences gain it where missing), with aria-sort exposed on all sortable headers. A density toggle lands in the header: compact (default, the app's identity) vs comfortable, persisted across visits, swapping a root-level class that adjusts table cell padding app-wide. One shared mechanism, not per-page reimplementations.

**Blocked by:** 03, 04, 05, 06, 07, 08 (sorting lands once, on restyled tables).

**Status:** done

- [x] All list tables sortable; sort state exposed via aria-sort; sort state visible icon/affordance in both themes
- [x] Density toggle in header; choice persists across reloads; default compact
- [x] Density affects table spacing app-wide through one mechanism; no per-page hacks
- [x] e2e: new sorting assertions on newly sortable tables; density toggle + persistence; a11y aria-sort coverage extended; full gate green — see gate note below

**Gate note:** backend 105 tests ✓, svelte-check 0/0 ✓, build ✓. New e2e specs (tables.spec.ts: device sort flip + examples sort flip + density persistence; a11y.spec.ts: aria-sort across devices/examples/sequences) were written and type-check, but the full Playwright run was skipped at user direction (runner again produced no output before being aborted — the same startup hang seen in ticket 08; not diagnosed). Full-suite verification rides the ticket 10 gate.
