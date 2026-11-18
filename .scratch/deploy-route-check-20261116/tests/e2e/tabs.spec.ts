import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

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
