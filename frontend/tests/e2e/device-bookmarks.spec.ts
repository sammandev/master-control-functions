import { readFileSync } from 'node:fs';
import { test, expect } from '@playwright/test';

test.beforeAll(async ({ request }) => {
	const workbook = readFileSync(
		new URL('../../../.dev-reference/Master_Control_V2_Script_Definition_Knowledge.xlsx', import.meta.url)
	);
	const response = await request.post('/api/import/reset', {
		multipart: {
			file: {
				name: 'Master_Control_V2_Script_Definition_Knowledge.xlsx',
				mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
				buffer: workbook
			}
		}
	});
	expect(response.status(), 'fixture import').toBe(200);
});

for (const tab of ['parameters', 'criteria'] as const) {
	test(`devices: ${tab} restores a Global definition bookmark on reload and history`, async ({
		page
	}) => {
		const name = tab === 'parameters' ? 'CHAINTOCALIBRATE' : 'MASK_MARGIN_UPPER_2';
		const url = `/devices/_global?tab=${tab}`;
		const panel = page.getByRole('tabpanel', { name: new RegExp(`^${tab}`, 'i') });
		const firstPage = panel.getByText(/^Page 1 of /);
		const secondPage = panel.getByText(/^Page 2 of /);
		const next = page.getByRole('button', { name: 'Next page' });
		const previous = page.getByRole('button', { name: 'Previous page' });
		const details = page.getByRole('region', { name: 'Definition details' });
		const selected = details.getByRole('heading', { name, exact: true });

		await page.goto(url);
		await expect(firstPage).toBeVisible();
		await expect(details).toBeVisible();
		await expect(page).toHaveURL(url);
		await next.click();
		await page.getByRole('button', { name: `Inspect ${name} (Global)` }).click();
		await expect(selected).toBeVisible();
		await expect(page).toHaveURL(/[?&]definition=\d+$/);
		const bookmark = page.url();

		await page.reload();
		await expect(page).toHaveURL(bookmark);
		await expect(secondPage).toBeVisible();
		await expect(selected).toBeVisible();

		await previous.click();
		await expect(firstPage).toBeVisible();
		await expect(panel.getByText(/^Select a .* definition\.$/)).toBeVisible();
		await expect(page).toHaveURL(bookmark);
		await next.click();
		await expect(selected).toBeVisible();
		await previous.click();
		await expect(panel.getByText(/^Select a .* definition\.$/)).toBeVisible();

		await page.getByRole('tab', { name: /^Functions/ }).click();
		await expect(page).toHaveURL(/\?tab=functions$/);
		await page.goBack();
		await expect(page).toHaveURL(bookmark);
		await expect(secondPage).toBeVisible();
		await expect(selected).toBeVisible();
	});

	test(`devices: ${tab} leaves unavailable bookmarks on page one`, async ({ page }) => {
		const url = `/devices/_global?tab=${tab}&definition=2147483647`;
		const panel = page.getByRole('tabpanel', { name: new RegExp(`^${tab}`, 'i') });
		await page.goto(url);
		await expect(panel.getByText(/^Page 1 of [2-9]/)).toBeVisible();
		await expect(panel.getByText(/^Select a .* definition\.$/)).toBeVisible();
		await expect(page).toHaveURL(url);
		await page.getByRole('button', { name: 'Next page' }).click();
		await expect(panel.getByText(/^Page 2 of /)).toBeVisible();
		await page.getByRole('button', { name: 'Previous page' }).click();
		await expect(panel.getByText(/^Page 1 of /)).toBeVisible();
	});
}
