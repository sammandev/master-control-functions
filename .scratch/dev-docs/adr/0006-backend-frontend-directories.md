# Repository layout: backend/ and frontend/ directories

The app is one Go module (API + single binary) and one SvelteKit frontend, previously laid out with `go.mod`, `cmd/`, `internal/` and `sqlc.yaml` at the repo root and the frontend in `web/`. We decided to group them symmetrically: everything Go moves under `backend/` (module path unchanged, `sqlc.yaml` moves with it) and `web/` becomes `frontend/`. Repo-level docs stay at the root (`CONTEXT.md`, `docs/adr/`, `.scratch/`, `AGENTS.md`, `README.md`, `.dev-reference/`): this is a single-context product, and the docs govern both halves.

## Consequences

- All Go commands reroot: `go -C backend …` (or `cd backend`), `sqlc` runs from `backend/`.
- The server's `-web` flag default becomes `frontend/build`, and the server is run from the repo root (`backend\server.exe -db mydata.db`).
- E2E webServer launches from `backend/` (the module root) with `-web ../frontend/build` — note `go run`'s child process inherits that cwd, which is why the repo-root-relative default only works when the server itself is run from the repo root.
- Build artifacts relocate: `backend/server.exe`, `frontend/build` (`.gitignore` updated).
- The `webprogfndb` module path does not change; no import rewrites.
- This supersedes the implicit root layout ADR-0001 described; the stack decision itself is unchanged.
