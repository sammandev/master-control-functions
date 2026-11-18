# 05: Orphan report + resolve

**What to build:** Every unresolved reference (link or Example pointing at a missing Function, Parameter definition, or Criteria definition) is flagged in the UI wherever it appears, and a dedicated orphan report lists them all grouped by kind. From any orphan row — including the report — the maintainer can create the missing target in one action, which resolves the reference. Nothing is ever dropped or quarantined (ADR-0003).

**Blocked by:** 03 (Workbook import), 04 (Function editor links).

**Status:** done

- [x] API-level orphan detection for all three link/Example kinds, exposed in list responses and a report endpoint (GET /api/orphans, groups: parameter-links/criteria-links/examples; EXISTS probes — a LEFT JOIN on (source, name) multiplied rows for multi-interface functions)
- [x] Orphan report page: grouped counts, filterable list, jump-to-row (link to the function detail when the function exists)
- [x] One-action resolution: create the missing Function or definition from the orphan row (prefilled name), which clears the flag (report page dialog + "Create definition" action on unresolved rows in function detail tabs)
- [x] Orphan badges visible in function detail tabs, dictionary usage lists, and examples views (examples views ship with ticket 06; meanwhile orphaned examples surface in the report)
- [x] Integration tests through the HTTP seam: seeded orphans from the real workbook fixture are reported, resolved in place, and flagged correctly after resolution (real-fixture union counts pinned: 155/64/16)
