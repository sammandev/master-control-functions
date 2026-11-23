import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// Creating a Function or dictionary definition whose name sorts off the
// current page must open the page that holds it and select it, not fall
// back to the first row of the page the user was on.

const FIXTURE = join(
	dirname(fileURLToPath(import.meta.url)),
	'..',
	'..',
	'..',
	'backend',
	'tests',
	'testdata',
	'Master_Control_V2_Script_Definition_Knowledge.xlsx'
);

test.beforeAll(async ({ request }: { request: APIRequestContext }) => {
	const res = await request.post('/api/import/reset', {
		multipart: {
			file: {
				name: 'Master_Control_V2_Script_Definition_Knowledge.xlsx',
				mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
				buffer: readFileSync(FIXTURE)
			}
		}
	});
	expect(res.status(), 'fixture import').toBe(200);
});

async function saveAndReadId(page: Page, endpoint: string) {
	const saved = page.waitForResponse(
		(response) => response.url().endsWith(endpoint) && response.request().method() === 'POST'
	);
	await page.getByRole('dialog').getByRole('button', { name: 'Save', exact: true }).click();
	const response = await saved;
	expect(response.status()).toBe(201);
	return ((await response.json()) as { id: number }).id;
}

test('creating a function that sorts onto a later page selects it', async ({ page, request }) => {
	const { total } = await (await request.get('/api/functions?page_size=1')).json();
	expect(total).toBeGreaterThan(20);

	await page.goto('/functions');
	const details = page.getByRole('region', { name: 'Function details' });
	await expect(details.getByRole('heading', { level: 2 })).toBeVisible();

	await page.getByRole('button', { name: 'New function', exact: true }).click();
	const name = 'ZZZ_E2E_CREATED_FUNCTION';
	await page.getByRole('dialog').getByLabel('Name', { exact: true }).fill(name);
	const id = await saveAndReadId(page, '/api/functions');

	await expect(details.getByRole('heading', { name, level: 2 })).toBeVisible();
	await expect(page.getByRole('button', { name: `Select ${name}`, pressed: true })).toBeVisible();
	expect(Object.fromEntries(new URL(page.url()).searchParams)).toMatchObject({
		page: String(Math.floor(total / 20) + 1),
		function: String(id)
	});
});

for (const [kind, title] of [
	['parameter-definitions', 'Parameter Definitions'],
	['criteria-definitions', 'Criteria Definitions']
] as const) {
	test(`creating a ${title} entry that sorts onto a later page selects it`, async ({
		page,
		request
	}) => {
		const { total } = await (await request.get(`/api/${kind}?page_size=1`)).json();
		expect(total).toBeGreaterThan(20);

		await page.goto(`/${kind}`);
		await expect(page.getByRole('heading', { name: title, level: 1 })).toBeVisible();
		const details = page.getByRole('region', { name: 'Definition details' });
		await expect(details.getByRole('heading', { level: 2 })).toBeVisible();

		await page.getByRole('button', { name: 'New', exact: true }).click();
		const name = 'zzz_e2e_created_definition';
		await page.getByRole('dialog').getByLabel('Name', { exact: true }).fill(name);
		const id = await saveAndReadId(page, `/api/${kind}`);

		await expect(details.getByRole('heading', { name, level: 2 })).toBeVisible();
		await expect(
			page.getByRole('button', { name: `Select ${name} (Global)`, pressed: true })
		).toBeVisible();
		expect(Object.fromEntries(new URL(page.url()).searchParams)).toMatchObject({
			page: String(Math.floor(total / 20) + 1),
			definition: String(id)
		});
	});
}
