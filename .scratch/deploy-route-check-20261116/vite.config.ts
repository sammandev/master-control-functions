import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { paraglideVitePlugin } from '@inlang/paraglide-js';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [
		// Tailwind v4 processes src/app.css (single @import entry, ticket 01).
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// Static SPA bundle served by the Go/Fiber binary (ADR-0001).
			// fallback gives client-side routes a page on refresh/deep links.
			adapter: adapter({ fallback: 'index.html' }),
			version: { pollInterval: 0 }
		}),

		paraglideVitePlugin({
			project: './project.inlang',
			outdir: './src/lib/paraglide',
			emitTsDeclarations: true
		})
	],
	server: {
		// In dev, hand API calls to the Go server started with:
		//   go -C backend run ./cmd/server -addr 127.0.0.1:8080
		proxy: { '/api': 'http://127.0.0.1:8080' }
	},
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.ts',
				// Keep component dependencies on the test renderer's Svelte runtime.
				optimizeDeps: { exclude: ['@lucide/svelte', 'bits-ui', 'svelte-sonner'] },
				test: {
					name: 'client',
					browser: {
						enabled: true,
						provider: playwright(),
						instances: [{ browser: 'chromium', headless: true }]
					},
					include: ['tests/components/**/*.svelte.{test,spec}.{js,ts}']
				}
			},

			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['tests/unit/**/*.{test,spec}.{js,ts}'],
					exclude: ['tests/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
