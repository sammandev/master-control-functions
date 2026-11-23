import { expect, test } from '@playwright/test';

test('function detail uses the loaded function name in the browser tab', async ({
	page,
	request
}) => {
	const name = `E2E_TAB_TITLE_${Date.now()}`;
	const created = await request.post('/api/functions', {
		data: { name, source: 'BUILT-IN', interface: '' }
	});
	expect(created.status()).toBe(201);
	const { id } = (await created.json()) as { id: number };

	try {
		await page.goto(`/functions/${id}`);
		await expect(page).toHaveTitle(`${name} | Functions | MasterControl Functions Database`);
	} finally {
		await request.delete(`/api/functions/${id}`);
	}
});

test('Examples uses its own browser tab title', async ({ page }) => {
	await page.goto('/examples');
	await expect(page).toHaveTitle('Examples | MasterControl Functions Database');
});
