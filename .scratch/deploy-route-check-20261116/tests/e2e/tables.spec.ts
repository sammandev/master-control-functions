import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// List-table breadth (ticket 09): sorting on the newly sortable tables.
// Assertions are behavioral: row order and aria-sort state.

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

// The fixture gives every table a deterministic, multi-row dataset so
// sort-direction flips are observable regardless of other specs' data.
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

test('device sorting flips tile order through labelled controls', async ({ page }) => {
	await page.goto('/devices');
	await expect(page.getByRole('heading', { name: 'Devices' })).toBeVisible();

	const sort = page.getByRole('combobox', { name: 'Sort devices' });
	const order = page.getByRole('combobox', { name: 'Sort order', exact: true });
	const firstDevice = page.getByRole('article').first().getByRole('link').first();

	await sort.selectOption('device');
	await order.selectOption('asc');
	await expect(order).toHaveValue('asc');
	const ascName = (await firstDevice.textContent()) ?? '';

	await order.selectOption('desc');
	await expect(order).toHaveValue('desc');
	const descName = (await firstDevice.textContent()) ?? '';

	expect(ascName.trim(), 'order flips between directions').not.toBe(descName.trim());
});

test('examples list sorting reorders the loaded rows', async ({ page }) => {
	await page.goto('/examples');
	await expect(page.getByRole('heading', { name: 'Examples' })).toBeVisible();

	const deviceHead = page.getByRole('columnheader', { name: 'Device' });
	const firstDeviceCell = page.locator('tbody tr').first().locator('td').nth(1);

	await deviceHead.getByRole('button').click();
	await expect(deviceHead).toHaveAttribute('aria-sort', 'ascending');
	const ascText = ((await firstDeviceCell.textContent()) ?? '').trim();

	await deviceHead.getByRole('button').click();
	await expect(deviceHead).toHaveAttribute('aria-sort', 'descending');
	const descText = ((await firstDeviceCell.textContent()) ?? '').trim();

	expect(ascText, 'order flips between directions').not.toBe(descText);
});
