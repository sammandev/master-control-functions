import { readFileSync } from 'node:fs';
import { expect, test } from '@playwright/test';

const workbookName = 'Master_Control_V2_Script_Definition_Knowledge.xlsx';

test('Function Examples keeps the latest search results when an older response arrives', async ({
	page,
	request
}, testInfo) => {
	const seed = await request.post('/api/import/reset', {
		multipart: {
			file: {
				name: workbookName,
				mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
				buffer: readFileSync(new URL(`../../../.dev-reference/${workbookName}`, import.meta.url))
			}
		}
	});
	expect(seed.status(), 'workbook import').toBe(200);

	const functions = await request.get('/api/functions?source=BUILT-IN&search=DELAY');
	expect(functions.ok()).toBe(true);
	const { items } = (await functions.json()) as { items: { id: number; name: string }[] };
	const fn = items.find(({ name }) => name === 'DELAY');
	expect(fn).toBeDefined();

	const oldTerm = '1000 ms';
	const latestTerm = '500 ms';
	const held = Promise.withResolvers<boolean>();
	const release = Promise.withResolvers<void>();
	const oldSearch = (url: URL) =>
		url.pathname === '/api/examples' && url.searchParams.get('search') === oldTerm;

	await page.route(oldSearch, async (route) => {
		const response = await route.fetch();
		held.resolve(response.ok());
		await release.promise;
		await route.fulfill({ response });
	});

	try {
		await page.goto(`/functions/${fn!.id}`);
		await page.getByRole('tab', { name: 'Examples', exact: true }).click();
		const panel = page.getByRole('tabpanel', { name: 'Examples', exact: true });
		const search = panel.getByRole('textbox', { name: 'Search examples' });
		const oldResult = panel.getByRole('cell', { name: /^Waits 1000 ms/ });
		const latestResult = panel.getByRole('cell', { name: /^Waits 500 ms/ });
		await expect(oldResult).toBeVisible();
		await expect(latestResult).toBeVisible();

		await search.fill(oldTerm);
		expect(await held.promise).toBe(true);
		await search.fill(latestTerm);
		await expect(latestResult).toBeVisible();
		await expect(oldResult).toBeHidden();

		const delayedResponse = page.waitForResponse((response) => oldSearch(new URL(response.url())));
		release.resolve();
		await (await delayedResponse).finished();
		await page.screenshot({ path: testInfo.outputPath('function-examples-search.png') });

		await expect(search).toHaveValue(latestTerm);
		await expect(latestResult).toBeVisible();
		await expect(oldResult).toBeHidden();
	} finally {
		release.resolve();
	}
});
