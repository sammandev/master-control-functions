# 11: Repository layout — backend/ and frontend/

**What to build:** The mechanical move ADR-0006 records. Everything Go goes under `backend/` (`go.mod`, `go.sum`, `cmd/`, `internal/`, `sqlc.yaml`); `web/` becomes `frontend/`. Module path `webprogfndb` unchanged — no import rewrites. Repo-level docs stay at root.

- [x] `git mv` (history preserved): `go.mod`, `go.sum`, `cmd/`, `internal/`, `sqlc.yaml` → `backend/`; `web/` → `frontend/`
- [x] `sqlc.yaml`: schema/queries paths still `internal/...` relative to `backend/` (no change needed once sqlc runs from `backend/`); regenerate to confirm output paths land in `backend/internal/store/`
- [x] `cmd/server/main.go`: `-web` flag default `web/build` → `frontend/build`
- [x] `frontend/playwright.config.ts`: webServer cwd = `../backend` (module root — `go run`'s child inherits it), command `go run ./cmd/server -addr … -db … -web ../frontend/build` (`E2E_GO` override preserved)
- [x] `frontend/` npm scripts: any repo-root assumptions fixed (typecheck/build/test:e2e all work from `frontend/`)
- [x] `.gitignore`: `server.exe` → `backend/server.exe`, `web/build` → `frontend/build`, e2e db artifact path if present
- [x] README rewrite: development commands (`go -C backend …`), build/run (`backend\server.exe -db mydata.db` from root), one-command full pass (`go -C backend test ./… ; cd frontend ; npm run check ; npm run build ; npm run test:e2e ; cd ..`), E2E section, sqlc/migration notes (paths now relative to `backend/`)
- [x] Post-move reindex: `graphify update .` ran clean (graph.json/html + GRAPH_REPORT.md regenerated); codegraph resyncs on next query
- [x] Full verification (2026-10-02): `go -C backend vet ./…` ✓, `go -C backend test ./…` ✓, `npm run check` 0 errors ✓, `npm run build` ✓, E2E 6/6 ✓, manual `backend\server.exe` run from root → `/health` 200 `{"db":"ok"}` + `/functions` deep link 200 (SPA served)

**Blocked by:** None.

**Status:** done

Out of scope: no CI/Docker rewiring (none exists); no doc-content edits beyond path references (ADR-0001's stack text stays historical); `.dev-reference/` untouched.
