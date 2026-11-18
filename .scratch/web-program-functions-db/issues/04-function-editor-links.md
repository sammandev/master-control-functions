# 04: Function editor — parameter + criteria links

**What to build:** On a Function's detail page, a test engineer manages its Parameter links and Criteria links as tabs: rows ordered by position with a required flag, added by picking an existing definition from a searchable dropdown, or created inline as a new definition and linked in one flow. Detaching a link never deletes the definition. Delete-with-warning protects definitions still referenced. Device scope is an editable free-text attribute on each link row (ADR-0002).

**Blocked by:** 02 (Dictionary CRUD), 03 (Workbook import).

**Status:** done

- [x] Link schemas + migrations: (function_source, function_name, name, device scope) natural keys, position, required
- [x] REST endpoints: list links for a function, attach, update (position/required/device scope), detach
- [x] Inline creation flow: new definition created and linked in one dialog without leaving the page
- [x] UI: Parameters and Criteria tabs on the function detail page; position ordering; required toggle; searchable dropdown; detach action; delete-with-warning for referenced definitions
- [x] Orphan links (already imported) render in the tabs with an unresolved warning state, editable in place
- [x] Integration tests through the HTTP seam: attach/update/detach round-trips, inline creation, warning on referenced-definition delete, orphan link editing

Notes: link tables shipped with migration 0002/0003 (this ticket added the queries + endpoints, not new schema). Inline creation is one atomic `POST .../links` call (`definition` object in the body) so the HTTP seam can test the rollback; the dialog drives it from the frontend.
