# Stack: Go (Fiber) + Svelte 5, SQLite with a PostgreSQL exit path

The CRUD web app is a single Go binary (Fiber HTTP + static file serving) with a Svelte 5 frontend built via SvelteKit + adapter-static, and SQLite as the starter database. SQLite is explicitly temporary: the data layer must keep a clean path to PostgreSQL later.

## Considered Options

- FastAPI + React / HTMX — rejected; stack preference was Go + Svelte
- mattn/go-sqlite3 (cgo) — rejected; requires a C toolchain on Windows
- GORM — rejected; ORM hinders the planned database swap

## Consequences

- Driver: `modernc.org/sqlite` (pure Go). Queries via `sqlc`.
- **Portability discipline**: all SQL stays ANSI-compatible — no SQLite-only functions, no `AUTOINCREMENT` (use `INTEGER PRIMARY KEY`), ISO-8601 timestamps, no `PRAGMA`-dependent logic outside the connection setup. Migrating to PostgreSQL + `pgx` later means swapping the driver, changing sqlc's engine config, and adjusting the few dialect-specific queries — not a rewrite.
- Deployment shape: one executable + one SQLite file next to it; SPA assets embedded or served from disk.
