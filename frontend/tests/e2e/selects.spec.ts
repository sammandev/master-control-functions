import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// Select-driven filter tests (ticket 03). Every dropdown is a bits-ui
// Select wrapper: a button trigger with aria-haspopup="listbox" and a
// portaled listbox of role=option items, keyboard operable (Enter opens,
// arrows highlight, Enter commits — bits-ui handles focus). A select
// drives UI -> API -> render, so the rendered count is cross-checked
// against the API's own total for the chosen facet. Assertions are
// role/label-based only; portal-rendered listboxes are found by role,
// never by DOM location.

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

test.beforeAll(async ({ request }) => {
	await seed(request);
});

test('source facet filter drives UI -> API -> render', async ({ page, request }) => {
	await page.goto('/functions');

	const facets = await (await request.get('/api/functions/facets')).json();
	const sources: string[] = facets.sources ?? [];
	expect(sources.length, 'fixture has at least one source facet').toBeGreaterThan(0);
	const pick = sources[0];
	const apiList = await (
		await request.get(`/api/functions?source=${encodeURIComponent(pick)}`)
	).json();

	const trigger = page.getByRole('button', { name: 'Filter by source' });
	await trigger.click();
	await page.getByRole('option', { name: pick, exact: true }).click();

	await expect(page.getByText(`${apiList.total} functions`)).toBeVisible();
	await expect(page.locator('tbody tr td:nth-child(2)')).toHaveText(
		Array.from({ length: await page.locator('tbody tr').count() }, () => pick)
	);
});

test('page-size dropdown is keyboard operable', async ({ page }) => {
	await page.goto('/functions');

	const trigger = page.getByRole('button', { name: 'Page size' });
	await trigger.focus();
	// Keyboard events go to the trigger (aria-activedescendant pattern —
	// focus never leaves the button). Opening highlights the selected
	// item ("20 / page"), so ArrowUp moves to "10 / page".
	await page.keyboard.press('Enter'); // open the listbox
	await page.keyboard.press('ArrowUp'); // highlight "10 / page"
	await page.keyboard.press('Enter'); // commit the selection

	await expect(trigger).toHaveText(/10 \/ page/);
	// The trigger text updates at commit; the refetched page arrives a beat
	// later, so the row count must be retried rather than read once.
	await expect
		.poll(() => page.locator('tbody tr').count(), {
			message: 'at most 10 rows per page after page-size change'
		})
		.toBeLessThanOrEqual(10);
	expect(await page.locator('tbody tr').count()).toBeGreaterThan(0);
});

test('dictionary type select works inside the create dialog', async ({ page }) => {
	await page.goto('/criteria-definitions');
	await page.getByRole('button', { name: 'New', exact: true }).click();

	const dialog = page.getByRole('dialog');
	const typeTrigger = dialog.getByRole('button', { name: 'Type' });
	await typeTrigger.click();
	// The listbox portals to the document body — a sibling of the portaled
	// dialog, not a descendant — so options are looked up at page level.
	await page.getByRole('option', { name: 'double', exact: true }).click();
	await expect(typeTrigger).toHaveText('double');
});
