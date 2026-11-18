# Frontend

Svelte 5/SvelteKit browser application for the Web Program Functions Database. [vite.config.ts](vite.config.ts) uses `adapter-static` with an `index.html` SPA fallback. Production is a static directory served by Nginx; it does not need a Node application server.

For the paired Docker release, follow [the blue-green runbook](../deploy/README.md). That isolated builder uses Bun for installation and execution; production serves the static bundle through Go behind Nginx. The native-host deployment below remains an alternative.

## Requirements

- Node.js 24 or newer.
- pnpm 12.6.0, pinned in [package.json](package.json). Use pnpm for installation and keep [pnpm-lock.yaml](pnpm-lock.yaml) as the only lockfile. Include [pnpm-workspace.yaml](pnpm-workspace.yaml), which owns transitive security overrides.
- The [Go backend](../backend/README.md) on the approved Ubuntu host for live API workflows.
- Nginx on Ubuntu for deployment.

Commands in the local sections run from `frontend/`. Server deployment and database-backed tests are operator-run on Ubuntu. Do not install or connect to a database on the restricted Windows workstation.

## Environment

The browser calls same-origin `/api`. No public API URL or database environment variable is required for development/build/deployment. Never put database credentials in `PUBLIC_*` or `VITE_*` variables, which may be exposed in the client bundle.

[.env.example](.env.example) is only for database-backed Playwright tests on Ubuntu. Its `localhost:5432` URL refers to PostgreSQL on that Ubuntu host and selects the disposable `master_control_functions_e2e` database. It is not frontend runtime configuration. Leave the production database URL in the backend's private environment file, outside the static build. Do not copy `.env` into `build/`.

## Device Workspace

The app opens the Devices directory. Search by Device name or sort by its Function, Parameter, Criteria, and Example counts. Global definitions appear separately.

Each Device has a searchable Function list beside the selected Function's scoped details and examples. Selection and tabs are recorded in the URL for deep links and browser history. Parameters and Criteria use raw links and server-resolved definition IDs, preserving Device overrides and Global fallback. Global references remain a separate group. Scripts are selectable and copyable in both Function details and the Device Examples tab.

Device creation/deletion and scoped definition forms remain available. Manage function opens the existing Function editor. Import, sequences, and the script checker are unchanged.

Every non-Global Device card has a Delete action, including populated Devices. The dialog lists the permanent cascade and requires the exact text `DELETE <device_name>` before enabling Delete. Cancel changes nothing; failed deletion stays open with an inline error. Deletion removes device-owned content and related Sequence steps, while preserving shared data for other Devices and the Sequence containers. See the [API deletion contract](../backend/README.md#device--interface-contract) for details.

The production workspaces share the approved Device visual language: full-width layouts, teal accents, monospaced identifiers, and underlined section tabs. Functions, dictionaries, Examples, Sequences, Orphans, and Import retain their specialized workflows. Prototype Device families and descriptions are mock-only; production cards use the names and counts supplied by the API.

## Styling Contract

[src/app.css](src/app.css) owns production colors, including modal overlays, plus the radius and Cascadia mono stack. It uses shadcn-svelte token names so registry components can use the owned theme without per-component renaming. Migration aliases have been removed. `primary` is the brand accent; `accent` and `muted` are backgrounds; subdued text uses `muted-foreground`.

- Import interactive primitives through the owned wrappers in [src/lib/components/ui/](src/lib/components/ui/). Pages do not import Bits UI directly.
- Use semantic color tokens in production components and pair changes with light/dark verification. Fixed colors in Storybook samples and the development-only prototype are outside the production theme contract.
- Keep page sections unframed, with local scrolling for wide tables and scripts. Cards are for repeated items and framed tools. The workspace has no fixed content-width cap or density preference.
- Keep scripts in the Cascadia mono stack. Presentation changes must preserve API payloads and server-emitted text.
- Success toasts are polite feedback. Validation errors and blocking failures remain persistent inline alerts; shared banner, field, toolbar, and muted-text classes remain while they have live consumers.
- Preserve keyboard focus, dialog focus restoration, accessible route state, and sortable-header state. Quiet actions may remain inline; existing crowded action menus must support arrow keys and Escape.

## Local Development

```sh
pnpm install --frozen-lockfile
pnpm dev --host 127.0.0.1 --port 5173 --strictPort
```

Open `http://127.0.0.1:5173`. Vite proxies `/api` to `http://127.0.0.1:8080`, so UI workflows need an API there. For an authorized Ubuntu development backend running on loopback port `8080`, the operator can create an SSH tunnel in a separate workstation terminal:

```sh
ssh -N -o ExitOnForwardFailure=yes -L 127.0.0.1:8080:127.0.0.1:8080 <ssh-user>@<ubuntu-host>
```

Replace the SSH placeholders and keep the tunnel running while using Vite. Database connections remain on Ubuntu. If deliberately targeting the deployed backend, change only the tunnel's final port to `24993`; the workstation-side port stays `8080`. That sends real edits/imports to the deployed database. Use a dedicated development backend/database for development work. The agent does not access the server or create this tunnel.

## Dependency Maintenance

The November 2026 security-first update is followed by coordinated SvelteKit 3.1.1/adapter-static 4.0.1, Vitest 5.1.0, and Paraglide JS 3.2.1 migrations. Vite is at least 8.3.6. Vitest and its Playwright provider share version 5.1.0, with vitest-browser-svelte 3.1.0. Playwright and `@playwright/test` move together to 1.65.0; install their matching Chromium revision after updating:

```sh
pnpm install --frozen-lockfile
pnpm exec playwright install chromium
node ./node_modules/@inlang/paraglide-js/bin/run.js compile --project ./project.inlang --outdir ./src/lib/paraglide --emit-ts-declarations
pnpm check
pnpm test:unit --run
pnpm build
```

The narrowly version-scoped overrides in [pnpm-workspace.yaml](pnpm-workspace.yaml) fix transitive `postcss-selector-parser` and `minimatch` advisories that their parents' declared ranges cannot resolve. The `minimatch` override crosses major versions for development lint tooling, so validate ESLint as well as the application when changing it. Kit 3 includes a current `cookie`, so the former cookie override has been removed. The `runed@0.35.1` override selects 0.38, whose optional Kit integration supports Kit 3, for Bits UI and svelte-toolbelt; UI regression coverage validates that combination. Remove overrides when upstream ranges include the fixes. Do not ignore advisories merely to produce a clean audit.

The Docker builder copies this policy before `bun pm migrate`; Bun imports the overrides into its temporary build manifest. Keep pnpm as the repository lockfile owner.

After the Kit 3 migration, `pnpm audit` reports **1 low advisory, 0 moderate, 0 high, and 0 critical**, down from 14:

| Remaining advisory                                                                 | Status and next step                                                                                                                                                                                                                                                    |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [deepmerge prototype poisoning](https://github.com/advisories/GHSA-gcqx-w48m-w2h4) | No published patched version as of this check. It comes through `@sveltejs/vite-plugin-svelte` and is build tooling, not a production Node server. Track an upstream fix/replacement; do not substitute a different merge implementation without compatibility testing. |

Production serves static files through Go/Nginx, not a SvelteKit Node server. That limits the relevant runtime surface but is not a blanket exemption. Keep Vite and browser-test servers bound to loopback and use trusted build configuration.

### Why Not Update Everything To Latest?

`pnpm update --latest` is **not** a drop-in update for the remaining outdated packages:

| Upgrade                | Project impact                                                                                                                                                                                                                                       |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| TypeScript 6 -> 7      | Currently blocked by `typescript-eslint` 8.72.1's declared peer range, `>=4.8.4 <6.1.0`. Wait for compatible lint/parser tooling and verify `svelte-check`, generated declarations, and build configuration; do not force past this peer constraint. |
| `@types/node` 24 -> 26 | Keep types aligned with the documented Node 24 toolchain. Newer types can allow APIs that the actual build runtime does not provide; upgrading types alone does not upgrade Node or Bun.                                                             |

Keep these two upgrades deferred until the runtime and lint toolchain support them. A latest-version install alone does not demonstrate compatibility or eliminate the unpatched `deepmerge` advisory. Full database-backed E2E validation remains operator-run on Ubuntu.

### SvelteKit 3 Compatibility

The [Kit 3 migration](https://svelte.dev/docs/kit/migrating-to-sveltekit-3) uses explicit `#lib` subpath imports from [package.json](package.json), `$app/env`, the relocated hook types, and `$app/tsconfig`. TypeScript only checks source; Vite emits the bundle, so `allowImportingTsExtensions` and `noEmit` allow explicit `.ts` subpath imports without rewriting aliases during emit.

URL query editing copies readonly search parameters to strings. Navigation uses `reset: false` to preserve focus/scroll and `replace` for history replacement. Route IDs retain a leading slash; pathname inputs to `resolve` do not. Keep these distinctions when adding links or filters. The static `index.html` fallback and same-origin `/api` contract remain unchanged. Version polling stays disabled to preserve Kit 2 behavior.

### Vitest 5 Compatibility

The [Vitest 5 migration](https://vitest.dev/guide/migration) upgrades the runner and browser provider together. Existing `test.projects` and `browser.enabled` configuration remains supported. Mock history now clears before each test by default; tests must not depend on calls made by an earlier test or a `beforeAll` hook. The clipboard tests already isolate their mock history, and both Node and Chromium component projects pass with the new default.

### Paraglide 3 Compatibility

[Paraglide 3](https://github.com/opral/paraglide-js/blob/main/blog/v3/index.md) removes `additionalFiles`, the older SDK compatibility path, and `experimentalMiddlewareLocaleSplitting`; none are configured here. Keep custom application helpers outside the generated [src/lib/paraglide/](src/lib/paraglide/) directory. The Vite plugin may emit virtual message modules; do not copy those generated files without the plugin or hand-edit them.

The standalone compile command above also remains the Docker builder's entrypoint. Run it before type checking a fresh checkout, where the generated directory is absent. [Localization tests](tests/unit/paraglide.spec.ts) exercise all three configured locales, message interpolation, encoded Device paths, query/fragment preservation, and the actual reroute hook.

## Build And Preview

```sh
pnpm check
pnpm build
pnpm preview --host 127.0.0.1 --port 4173 --strictPort
```

The build is written to `build/`; preview opens at `http://127.0.0.1:4173`. Vite preview inherits the configured `/api` proxy to `127.0.0.1:8080`, so keep the same API tunnel running for live workflows. Do not use `pnpm dev` or `pnpm preview` as the Ubuntu production server.

## Ubuntu Deployment

Use `/opt/master-control-functions` as the repository location, matching the [backend deployment](../backend/README.md#ubuntu-deployment). The backend service must already be healthy on `127.0.0.1:24993`.

| Component      | Port    | Exposure                               |
| -------------- | ------- | -------------------------------------- |
| Frontend Nginx | `31627` | HTTP browser entry point               |
| Backend Nginx  | `39063` | Optional separate HTTP API entry point |
| Go API         | `24993` | Loopback only                          |

The random high-port defaults are fixed in the checked-in templates. The Ubuntu operator must check for conflicts before activation. These HTTP templates are for an access-controlled internal network. Add HTTPS and authentication before public exposure, and restrict firewall access to approved clients. Keep PostgreSQL and the Go listener private.

### Publish The Static Build

Build on Ubuntu with the pinned toolchain, or publish the complete `build/` produced by an approved build machine. For an Ubuntu build:

```bash
cd /opt/master-control-functions/frontend
pnpm install --frozen-lockfile
pnpm check
pnpm build
test -f build/index.html
sudo -u www-data test -r /opt/master-control-functions/frontend/build/index.html
```

Nginx's Ubuntu worker normally runs as `www-data`; adapt the permission check if the configured worker user differs. It needs directory traversal and read access to the complete build. Serve only `frontend/build`, never the repository or the frontend source directory. Preserve the previous build for rollback and publish a complete release so HTML and its assets stay together. Never grant world-write permissions to fix a `403`.

### Frontend Nginx

[nginx.conf](nginx.conf) listens on port `31627`, serves `/opt/master-control-functions/frontend/build`, and proxies `/api`, `/api/...` and `/health` directly to the loopback Go API. The proxy preserves the `/api` prefix, so no CORS configuration is needed. Missing `/_app/` assets return 404; other client-side routes fall back to `index.html`.

Back up existing application site configuration before replacing it. Install this template as a separate site inside the existing Nginx `http` context:

```bash
sudo install -m 644 nginx.conf /etc/nginx/sites-available/master-control-functions-frontend
```

For first activation only:

```bash
sudo ln -s /etc/nginx/sites-available/master-control-functions-frontend /etc/nginx/sites-enabled/master-control-functions-frontend
```

Validate all sites before reloading; start Nginx instead if it is not yet running:

```bash
sudo nginx -t
sudo systemctl reload nginx
curl --fail http://127.0.0.1:31627/health
curl --head --fail http://127.0.0.1:31627/import
```

Health should return `{"status":"ok","db":"ok"}` and `/import` should return HTML with HTTP 200. Open `http://<ubuntu-host>:31627` in a browser and verify navigation and API loading. If you change the deployment path or Go port, update this template and the backend service/configuration together. Both Nginx sites have a 4 MiB upload limit.

## Tests

Specs and fixtures live under [tests/](tests/). Database-free unit and browser component checks can run locally:

```sh
pnpm exec playwright install chromium
pnpm exec vitest run --project server --project client
```

The isolated UI config starts Vite and mocks API calls:

```sh
pnpm exec playwright test --config playwright.ui.config.ts
```

On Windows with HTTP proxy variables set, add `127.0.0.1,localhost` to the command's `NO_PROXY` environment before running Playwright. The mocked [UI suite](tests/e2e/polish.spec.ts) covers all production workspaces, including Import upload/reset/history, scoped overrides, Global references, retries, and delayed responses. Keyboard coverage includes browsing/filtering, edit and link dialogs, Composer, Sequence reorder/edit, Orphan repair, and reset confirmation. The suite does not start the backend or use a database.

Visual contracts check the prototype's card geometry, typography, section tabs, and light/dark accents. The final audit visits eleven production routes in both themes at desktop/mobile sizes, checking route highlighting, skip-link focus, overflow, and browser errors. Reset confirmation also checks destructive-button text contrast and keyboard focus wrapping. Screenshots include populated Import results and history, 32-Device catalogs, and long identifiers and scripts.

For static-bundle verification, run `pnpm build`, start `pnpm preview --host 127.0.0.1 --port 4174 --strictPort`, and run the UI suite with `CI` unset. If port `4174` is occupied, choose another and set `PLAYWRIGHT_UI_PORT` to match. Restart the preview after rebuilding. API responses are mocked for this run.

For the full E2E suite, work only on the approved Ubuntu host. Create `.env` from the example, protect it with `chmod 600 .env`, and configure a separate disposable database and test login. Playwright does not automatically load this file:

```bash
set -a
. ./.env
set +a
pnpm exec playwright install --with-deps chromium
pnpm build
pnpm test:e2e
```

E2E starts its own backend on `127.0.0.1:8123` and resets database contents. Never set `E2E_DATABASE_URL` to the production catalog. Do not run `pnpm test` locally: it includes this database-backed suite. Failed traces are retained under `test-results/`.

## Troubleshooting

- `/api` returns HTML: confirm Nginx loaded the supplied API locations and that `proxy_pass` has no trailing slash that would remove the API prefix.
- Nginx `502`: check backend health on `127.0.0.1:24993` and the backend service logs.
- Static `403` or missing scripts: check the Nginx build path, read/traverse permissions and whether the full build was published.
- Deep-link refresh fails: retain `try_files $uri $uri/ /index.html` and the generated fallback file.
- Browser calls fail during development/preview: check the API tunnel on workstation port `8080` or use the Nginx site.
