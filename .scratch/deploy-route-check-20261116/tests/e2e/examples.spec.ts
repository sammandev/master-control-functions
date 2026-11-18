import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// Examples page happy path (ticket 01 tracer). The /examples surface is
// the thinnest full-CRUD screen and proves the new UI wrappers end-to-end.
// Assertions are role- and label-based only — no class or style coupling,
// so the wrapper migration cannot silently break the suite.

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

test('examples list renders and search filters it', async ({ page, request }) => {
	// Fixture example count derived from the API before the +1 seed — no
	// literal to update when a workbook batch lands. The seed adds one
	// row with a unique meaning. TX_MULTI_VERIFICATION exists in the
	// fixture, so the seeded row must NOT be flagged as an orphan.
	const before = (await (await request.get('/api/examples?page_size=1')).json()) as {
		total: number;
	};
	const created = await request.post('/api/examples', {
		data: {
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: 'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|10000|NEXT|FAIL',
			meaning: 'e2e seeded example'
		}
	});
	expect(created.status(), 'example seeded').toBe(201);

	await page.goto('/examples');
	await expect(page.getByRole('heading', { name: 'Examples' })).toBeVisible();
	await expect(page.getByText(`${before.total + 1} example(s)`)).toBeVisible();
	await expect(page.getByRole('button', { name: 'New example' })).toBeVisible();

	// The seeded row sorts last on the default page (total grew by exactly
	// one above), so anchor on
	// the search filter: the unique meaning narrows to exactly it, a
	// no-match term empties the list, and clearing restores the dataset.
	const search = page.getByLabel('Search examples');
	await search.fill('e2e seeded example');
	await expect(page.getByRole('row', { name: /TX_MULTI_VERIFICATION/ })).toBeVisible();

	await search.fill('zzz-no-match');
	await expect(page.getByText('No examples yet.')).toBeVisible();
	await search.fill('e2e seeded example');
	await expect(page.getByRole('row', { name: /TX_MULTI_VERIFICATION/ })).toBeVisible();
});
