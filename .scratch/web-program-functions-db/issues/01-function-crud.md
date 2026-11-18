# 01: Function CRUD end-to-end

**What to build:** A test engineer can browse, search, sort, paginate, create, edit, and delete Functions through the web app. The functions schema exists, list endpoints support search/sort/filter/pagination parameters, and the UI has a functions list page with instant (debounced) search plus a detail view. All behavior verified through the HTTP seam (integration tests against a temporary SQLite database), per the primary test seam.

**Blocked by:** 00 (Environment + app skeleton).

**Status:** done

- [x] Function schema + migrations: unique on (source, interface, name); columns for description, convention, category, execution_role, execution_phase
- [x] REST endpoints: list (search/sort/filter/pagination), get, create, update, delete — with field-level validation errors on bad input
- [x] Delete warns or fails cleanly per API contract; duplicate (source, interface, name) rejected with a clear message
- [x] UI: functions list with search-as-you-type, sort, pagination, create/edit/delete flows with inline error messages
- [x] Integration tests through the HTTP seam cover CRUD round-trip, uniqueness rejection, pagination contract, and search behavior

## Comments

**2026-10-01 — implementation notes**

- Migrations moved from code-defined strings to embedded `.sql` files under the db package (filename = `NNNN_name.sql`, version parsed from prefix, duplicate versions rejected at load). Reason: sqlc consumes plain SQL files as its schema source; embedded files keep the binary self-contained. Schema test updated: fresh DB is now version 1.
- sqlc 1.31.1 adopted per spec (installed from official GitHub release, SHA-256 verified, MIT). Two sqlc SQLite-grammar constraints shaped the queries: typed-parameter casts (`::text`) and `ESCAPE` clauses are rejected by its SQLite parser, and dynamic `ORDER BY` via CASE is not parseable. Consequences: search is plain substring (wildcard chars stripped from the term in Go), and sort+pagination happen in Go over filtered rows — deliberate at workbook-scale data volumes, documented in the handler; revisit with keyset pagination at the PostgreSQL migration.
- Contract details: `POST /api/functions` → 201 / 422 with field-level errors / 409 on natural-key duplicate; `GET /api/functions` → `{items,total,page,page_size}` with `search`, `source`, `sort`, `order`, `page`, `page_size` (page_size capped at 100); `PUT` is full-replace with key-change support (409 on collision); `DELETE` → 204/404. Sorting is server-side with a column whitelist + deterministic tiebreak.
- UI: functions list page (debounced search-as-you-type, sortable columns, source filter, pagination, create/edit dialog with inline field errors from the API, delete with confirm), root redirects to `/functions`, typed API client, dev-mode Vite proxy for `/api`.
- Verification: 9 new HTTP-seam integration tests (validation, round-trip, duplicate 409, pagination contract, search, filters, key-conflict), full suite green (`go vet` clean, `svelte-check` 0 errors), and live smoke: create 201 → list 200 → duplicate 409 → delete 204 → SPA deep link 200.
