# Frontend package management: pnpm canonical, bun as optional script runner

The frontend under `frontend/` uses **pnpm 12** as its single canonical package manager, pinned via the `packageManager` field in `package.json` (corepack resolves it). Every `package.json` script can alternatively be executed with **bun** (`bun run dev` / `build` / `check`), which shells the script out through the system shell so vite/svelte-kit still run under Node. Bun as an *installer* is intentionally not supported.

## Considered Options

- Status quo (npm only) — rejected; the goal was pnpm + bun usability, and npm's flat `node_modules` + `package-lock.json` were the thing being replaced.
- Bun as canonical installer — rejected as default: Vite 8 + SvelteKit's native plugin chain on Windows is an unproven compatibility surface, and it would make pnpm entirely redundant.
- Both pnpm and bun as supported installers — rejected: each writes its own lockfile (`pnpm-lock.yaml` vs `bun.lock`) and its own `node_modules` layout (symlinked store vs hoisted). Dual-lockfile repos drift, and switching installers between machines leaves stale layouts that break resolution.

## Consequences

- `frontend/pnpm-lock.yaml` is the single source of truth for dependency versions; no `package-lock.json` and no `bun.lock` is committed.
- Enforcement is soft: the `packageManager` field + corepack pin pnpm for npm/pnpm users, but bun ignores corepack, so nothing hard-blocks a `bun install`. The README states the convention instead of adding a `preinstall` user-agent guard — this repo has no CI to betray and a single-operator audience.
- `bun run <script>` is safe by construction: it executes package.json scripts via the shell (not the bun runtime), so the Node-based toolchain (vite, svelte-check, Playwright) is unaffected. Verified minimum: `dev`, `build`, `check` under both runners.
- Removing bun later costs nothing (it is a runner, not an owner of state); removing pnpm means regenerating a lockfile and rewriting docs — the reason this decision is recorded.
