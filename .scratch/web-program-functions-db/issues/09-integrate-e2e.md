# 09: Integrate — E2E smoke, filter polish, performance pass

**What to build:** The whole app hangs together. A thin Playwright smoke suite drives the real UI over the happy paths (search a function, edit a parameter link, compose-and-save a script, view the orphan report) against a seeded database. Enum filter bars (source, interface, category, execution role, execution phase) work on the functions list; pages stay fast with the full real dataset; any wiring bugs the HTTP tests could not catch are fixed.

**Blocked by:** 05 (Orphan report), 06 (Examples CRUD), 07 (Script composer), 08 (Export round-trip).

**Status:** done

- [x] Playwright smoke suite: the four happy paths pass against a seeded database
- [x] Enum filter bar on functions list composes with search/sort/pagination (server-side, tested through the HTTP seam)
- [x] Full-dataset sanity: every list page loads in under a second locally; no N+1 query hotspots visible in timing
- [x] All integration tests + smoke suite green from one command; dev README updated

Annotations:
- Backend enum filters already existed (parseListQuery + SQL narg filters); this ticket added the facets endpoint, the UI bar, and the HTTP-seam tests.
- Smoke suite has 5 tests: the four happy paths plus one driving a filter select UI→API→render (expected total read from the API first). Edit-path assertion uses a sentinel device scope the fixture row provably lacks, so it cannot pass on a no-op save.
- Perf pin is server-latency (Go test, <1s per read endpoint, single-digit ms observed); timings are logged for N+1 inspection.
