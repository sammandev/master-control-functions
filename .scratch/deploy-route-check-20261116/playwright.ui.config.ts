import { defineConfig } from '@playwright/test';

const port = process.env.PLAYWRIGHT_UI_PORT ?? '4174';
const baseURL = `http://127.0.0.1:${port}`;

export default defineConfig({
	testDir: './tests/e2e',
	testMatch: 'polish.spec.ts',
	timeout: 30_000,
	workers: 1,
	reporter: [['list']],
	use: {
		baseURL,
		trace: 'retain-on-failure'
	},
	webServer: {
		command: `pnpm exec vite dev --host 127.0.0.1 --port ${port} --strictPort`,
		url: baseURL,
		reuseExistingServer: !process.env.CI,
		timeout: 60_000
	}
});
