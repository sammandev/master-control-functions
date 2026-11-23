import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

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

test('Function filter fetches only names and still filters example rows', async ({ page }) => {
	const exampleRequests: string[] = [];
	page.on('request', (request) => {
		const { pathname, search } = new URL(request.url());
		if (pathname === '/api/examples' || pathname === '/api/examples/function-names')
			exampleRequests.push(pathname + search);
	});

	await page.goto('/examples');
	const details = page.getByRole('region', { name: 'Example details' });
	await expect(details.getByRole('button', { name: /^Copy example \d+$/ })).toBeVisible();
	await page.getByRole('button', { name: 'Function name filter' }).click();
	const name = 'TX_MULTI_VERIFICATION';
	await expect(page.getByRole('option', { name, exact: true })).toBeVisible();
	expect(exampleRequests.sort()).toEqual([
		'/api/examples/function-names',
		'/api/examples?page=1&page_size=20'
	]);

	await page.getByRole('option', { name, exact: true }).click();
	await expect(page).toHaveURL(/function_name=TX_MULTI_VERIFICATION/);
	const rows = page.getByRole('button', { name: /^Select example of / });
	await expect(rows.first()).toBeVisible();
	await expect(rows.first()).toHaveAccessibleName(/^Select example of TX_MULTI_VERIFICATION /);
	await expect(
		page.getByRole('button', { name: /^Select example of (?!TX_MULTI_VERIFICATION )/ })
	).toHaveCount(0);
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

async function createWifiExample(page: Page, script: string, meaning: string) {
	await page.getByRole('button', { name: 'New example', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await expect(dialog).toContainText('Device: WIFI_TEST');
	await dialog
		.getByRole('textbox', { name: 'Function name', exact: true })
		.fill('TX_MULTI_VERIFICATION');
	await dialog.getByRole('textbox', { name: 'Script', exact: true }).fill(script);
	await dialog.getByRole('textbox', { name: 'Meaning', exact: true }).fill(meaning);
	const saved = page.waitForResponse(
		(response) => response.url().endsWith('/api/examples') && response.request().method() === 'POST'
	);
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	const response = await saved;
	expect(response.status()).toBe(201);
	const { id } = await response.json();
	await expect(dialog).toBeHidden();
	await expect(page.getByText('Example created', { exact: true })).toBeVisible();
	return id as number;
}

test('creating an example selects it even when it sorts onto a later page', async ({
	page,
	request
}) => {
	const firstPage = await request.get('/api/examples?device=WIFI_TEST&page_size=10');
	expect(firstPage.ok()).toBe(true);
	const { total } = await firstPage.json();
	expect(total).toBeGreaterThanOrEqual(10);

	await page.goto('/examples?device=WIFI_TEST&page_size=10');
	const details = page.getByRole('region', { name: 'Example details' });
	await expect(details.getByRole('button', { name: /^Copy example \d+$/ })).toBeVisible();

	const script = 'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|12345|NEXT|FAIL';
	const meaning = 'Verify that the newly saved WIFI_TEST script stays selected.';
	const id = await createWifiExample(page, script, meaning);
	await expect(
		details.getByRole('button', { name: `Copy example ${id}`, exact: true })
	).toBeVisible();

	await expect(details.getByText(script, { exact: true })).toBeVisible();
	await expect(details.getByText(meaning, { exact: true })).toBeVisible();
	await expect(
		page.getByRole('button', {
			name: 'Select example of TX_MULTI_VERIFICATION (WIFI_TEST)',
			pressed: true
		})
	).toHaveText(/Verify that the newly saved/);
	expect(Object.fromEntries(new URL(page.url()).searchParams)).toMatchObject({
		device: 'WIFI_TEST',
		page: String(Math.floor(total / 10) + 1),
		page_size: '10',
		example: String(id)
	});
	await page.reload();
	await expect(
		details.getByRole('button', { name: `Copy example ${id}`, exact: true })
	).toBeVisible();
});

test('creating an example reveals it when the active search excludes it', async ({ page }) => {
	await page.goto('/examples?device=WIFI_TEST&search=not-a-matching-script');
	await expect(page.getByText('No examples match.', { exact: false })).toBeVisible();

	const script = 'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|12346|NEXT|FAIL';
	const meaning = 'Verify that a search cannot hide the newly saved script.';
	const id = await createWifiExample(page, script, meaning);
	const details = page.getByRole('region', { name: 'Example details' });
	await expect(
		details.getByRole('button', { name: `Copy example ${id}`, exact: true })
	).toBeVisible();
	await expect(details.getByText(meaning, { exact: true })).toBeVisible();
	await expect(page.getByLabel('Search examples')).toHaveValue('');
	expect(new URL(page.url()).searchParams.get('device')).toBe('WIFI_TEST');
});
