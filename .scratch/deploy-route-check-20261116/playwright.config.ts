import { defineConfig } from '@playwright/test';

const databaseUrl = process.env.E2E_DATABASE_URL;
if (!databaseUrl) {
	throw new Error(
		'E2E_DATABASE_URL must select a disposable PostgreSQL database; E2E resets its data.'
	);
}

// E2E smoke suite (ticket 09). Drives the real UI served by the real Go
// server (single binary: API + built SPA) against a disposable PostgreSQL database that
// the suite seeds with the real knowledge workbook through /api/import
// (beforeAll in smoke.spec.ts), so stale runs never leak in.
//
// Prerequisites (see README "End-to-end tests"):
//   npm run build   (SPA bundle in frontend/build)
//   npx playwright install chromium
export default defineConfig({
	testDir: './tests/e2e',
	timeout: 30_000,
	fullyParallel: false,
	workers: 1,
	reporter: [['list']],
	use: {
		baseURL: process.env.E2E_BASE_URL ?? 'http://127.0.0.1:8123',
		trace: 'retain-on-failure'
	},
	webServer: {
		// Serves the built SPA and the API in one process — the app's real
		// production shape. cwd is backend/ (the module root), so the bundle
		// path is ../frontend/build; the production default (-web
		// frontend/build, run from the repo root) is unaffected.
		// E2E_GO overrides the go binary when `go` is not on PATH. The
		// E2E_DATABASE_URL must be disposable; seeding resets its data.
		command:
			`${process.env.E2E_GO ?? 'go'} run ./cmd/server ` +
			`-addr ${process.env.E2E_ADDR ?? '127.0.0.1:8123'} ` +
			`-web ../frontend/build`,
		env: { DATABASE_URL: databaseUrl },
		url: process.env.E2E_BASE_URL
			? `${process.env.E2E_BASE_URL}/health`
			: 'http://127.0.0.1:8123/health',
		cwd: '../backend',
		reuseExistingServer: false,
		timeout: 120_000
	}
});
