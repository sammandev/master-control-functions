import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';
import type { FunctionRow, ListResult } from '../../src/lib/api';

// Function detail page chrome (ticket 05): tab semantics with arrow-key
// switching, keyboard-reachable tooltips on orphan hints, and breadcrumb
// navigation. Assertions are role- and label-based only.

const E2E_DIR = dirname(fileURLToPath(import.meta.url));
const FIXTURE = join(
	E2E_DIR,
	'..',
	'..',
	'..',
	'backend',
	'tests',
	'testdata',
	'Master_Control_V2_Script_Definition_Knowledge.xlsx'
);

async function seed(request: APIRequestContext) {
	const workbook = readFileSync(FIXTURE);
	const res = await request.post('/api/import/reset', {
		multipart: {
			file: {
				name: 'Master_Control_V2_Script_Definition_Knowledge.xlsx',
				mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
				buffer: workbook
			}
		}
	});
	expect(res.status(), 'fixture import').toBe(200);
}

async function firstFunctionId(request: APIRequestContext): Promise<number> {
	const res = await request.get('/api/functions?page_size=1');
	const body = (await res.json()) as { items: Array<{ id: number }> };
	return body.items[0].id;
}

test.beforeAll(async ({ request }) => {
	await seed(request);
});

test('opening Function details fetches each link list once', async ({ page }) => {
	const fnId = await firstFunctionId(page.request);
	const paths = [`/api/functions/${fnId}/parameter-links`, `/api/functions/${fnId}/criteria-links`];
	const requests: string[] = [];
	page.on('request', (request) => {
		const { pathname } = new URL(request.url());
		if (request.method() === 'GET' && paths.includes(pathname)) requests.push(pathname);
	});

	await page.goto(`/functions/${fnId}`);
	const tabs = page.getByRole('tablist', { name: 'Function sections' });
	await expect(tabs.getByRole('tab', { name: 'Parameters' }).locator('.tab-count')).toHaveText(
		/^\d+$/
	);
	await expect(tabs.getByRole('tab', { name: 'Criteria' }).locator('.tab-count')).toHaveText(/^\d+$/);
	await expect(page.getByRole('tabpanel').getByText(/^\d+ linked definitions$/)).toBeVisible();
	expect(requests.sort()).toEqual(paths.sort());
});

test('Function switcher fetches compact options on first open and preserves catalog order', async ({
	page,
	request
}) => {
	const query = '/api/functions?sort=name&order=asc&page_size=100&page=';
	const catalog: FunctionRow[] = [];
	for (let pageNumber = 1; ; pageNumber++) {
		const response = await request.get(query + pageNumber);
		expect(response.ok()).toBe(true);
		const { items, total }: ListResult = await response.json();
		expect(items.length).toBeGreaterThan(0);
		catalog.push(...items);
		if (catalog.length >= total) break;
	}
	expect(catalog.length).toBeGreaterThan(100);
	const options = catalog.map(({ id, name, source, interface: scope }) => ({
		id,
		name,
		source,
		interface: scope
	}));

	const catalogRequests: string[] = [];
	page.on('request', (request) => {
		const path = new URL(request.url()).pathname;
		if (path === '/api/functions' || path === '/api/functions/options') {
			catalogRequests.push(path);
		}
	});
	await page.goto(`/functions/${options[1].id}`);
	const trigger = page.getByRole('button', { name: 'Switch function' });
	await expect(trigger).toHaveText(options[1].name);
	await page.waitForLoadState('networkidle');
	expect(catalogRequests).toEqual([]);

	const responsePromise = page.waitForResponse(
		(response) => new URL(response.url()).pathname === '/api/functions/options'
	);
	await trigger.focus();
	await trigger.press('Enter');
	const response = await responsePromise;
	expect(response.ok()).toBe(true);
	expect(await response.json()).toEqual({ items: options });

	const popup = page.getByRole('dialog', { name: 'Choose Switch function' });
	const labels = options.map((row) => `${row.name} (${row.source} / ${row.interface || 'Global'})`);
	await expect(popup.getByRole('option')).toHaveText(labels);
	const search = popup.getByRole('combobox', { name: 'Search options' });
	await expect(search).toBeFocused();
	await search.press('Escape');
	await expect(popup).toBeHidden();
	await trigger.click();
	await expect(popup.getByRole('option')).toHaveText(labels);

	const target = options.at(-1)!;
	await search.fill(target.name);
	await popup.getByRole('option', { name: labels.at(-1), exact: true }).click();
	await expect(page).toHaveURL(new RegExp(`/functions/${target.id}$`));
	await expect(trigger).toHaveText(target.name);
	expect(catalogRequests).toEqual(['/api/functions/options']);
});

test('tabs switch with arrow keys and announce the active panel', async ({ page }) => {
	const fnId = await firstFunctionId(page.request);
	await page.goto(`/functions/${fnId}`);

	// The tablist is labelled so screen readers announce what it groups.
	const tablist = page.getByRole('tablist', { name: 'Function sections' });
	await expect(tablist).toBeVisible();

	const parameters = tablist.getByRole('tab', { name: 'Parameters' });
	const criteria = tablist.getByRole('tab', { name: 'Criteria' });
	const examples = tablist.getByRole('tab', { name: 'Examples' });

	await expect(parameters).toHaveAttribute('aria-selected', 'true');
	// The active panel is a real tabpanel carrying the links panel content.
	await expect(page.getByRole('tabpanel')).toBeVisible();

	// Arrow keys move focus AND activate (bits-ui automatic activation).
	await parameters.focus();
	await page.keyboard.press('ArrowRight');
	await expect(criteria).toHaveAttribute('aria-selected', 'true');
	await expect(parameters).toHaveAttribute('aria-selected', 'false');
	await page.keyboard.press('ArrowRight');
	await expect(examples).toHaveAttribute('aria-selected', 'true');
	await page.keyboard.press('ArrowLeft');
	await expect(criteria).toHaveAttribute('aria-selected', 'true');
	await expect(page.getByRole('tabpanel')).toBeVisible();
});

test('orphan hint tooltip opens on keyboard focus', async ({ page }) => {
	// A link whose definition does not exist renders the unresolved badge —
	// the one hint that used to ride on a title attribute. The API refuses
	// links to unknown names (422), so seed the chain: definition, link,
	// then force-delete the definition to leave the link unresolved.
	const created = await page.request.post('/api/functions', {
		data: {
			source: 'DLL',
			name: `E2E_TT_FN_${Date.now()}`,
			interface: 'IF',
			description: 'tooltip test target'
		}
	});
	expect(created.status(), 'function seeded').toBe(201);
	const fn = (await created.json()) as { id: number };

	const defRes = await page.request.post('/api/parameter-definitions', {
		data: { name: `E2E_TT_PARAM_${Date.now()}`, type: 'int', cardinality: 'SINGLE' }
	});
	expect(defRes.status(), 'definition seeded').toBe(201);
	const def = (await defRes.json()) as { id: number; name: string };

	const attached = await page.request.post(`/api/functions/${fn.id}/parameter-links`, {
		data: { name: def.name, device_name: 'TT_DEV', position: 1, required: false }
	});
	expect(attached.status(), 'link seeded').toBe(201);

	const deleted = await page.request.delete(`/api/parameter-definitions/${def.id}?force=true`);
	expect(deleted.status(), 'definition force-deleted (link becomes unresolved)').toBe(204);

	await page.goto(`/functions/${fn.id}`);

	const badge = page.getByText('unresolved', { exact: true });
	await expect(badge).toBeVisible();

	// Keyboard reachability: focusing the badge announces the hint. A tooltip
	// that only opens on hover fails this.
	await badge.focus();
	await expect(page.getByRole('tooltip')).toBeVisible();
	await expect(page.getByRole('tooltip')).toContainText(
		'No parameter definition with this name exists'
	);
});

test('breadcrumbs label the trail back to the list', async ({ page }) => {
	const fnId = await firstFunctionId(page.request);
	const fn = (await (await page.request.get(`/api/functions/${fnId}`)).json()) as {
		name: string;
	};
	await page.goto(`/functions/${fnId}`);

	const crumbs = page.getByRole('navigation', { name: 'Breadcrumb' });
	await expect(crumbs.getByRole('link', { name: 'Functions' })).toBeVisible();
	await expect(crumbs.locator('[aria-current="page"]')).toHaveText(fn.name);
});
