# 06: Examples CRUD

**What to build:** A test engineer can manage Examples: the Examples tab on a Function detail page lists that Function's examples, and a standalone examples view offers full CRUD with device scope editing, script text, and meaning. Creating/editing validates function references and surfaces orphans per ticket 05's flagging.

**Blocked by:** 04 (Function editor — links).

**Status:** done

- [x] REST endpoints: list (filter by function, device scope; search in script/meaning), get, create, update, delete
- [x] UI: Examples tab on function detail + standalone examples page; create/edit dialog with device scope, script, meaning fields; delete with confirm
- [x] Script text renders readably (long pipe-delimited strings wrap or soft-wrap, no overflow)
- [x] Integration tests through the HTTP seam: CRUD round-trip, function filter, orphan example flagged
