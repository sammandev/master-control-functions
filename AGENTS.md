# master-control-functions

CRUD web application over the script-definition knowledge workbook in `.dev-reference/`. PostgreSQL production database; SQLite is retained as migration input.

For database cutover or rollback, read `docs/postgresql-migration.md`. Database operations are user-run on the approved Ubuntu host; the agent must not access that host.

## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-role vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: `CONTEXT.md` + `docs/adr/` at repo root. See `docs/agents/domain.md`.
