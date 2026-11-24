import { expect, test } from '@playwright/test';

test('workbook export explains why a long description cannot be downloaded', async ({
	page,
	request
}) => {
	const name = `E2E_LONG_DESCRIPTION_${Date.now()}`;
	const description = 'x'.repeat(32768);
	const created = await request.post('/api/functions', {
		data: { name, source: 'BUILT-IN', interface: '', description }
	});
	expect(created.status()).toBe(201);
	const { id } = (await created.json()) as { id: number };

	try {
		await page.goto('/import');
		const exportAction = page.getByText('Export workbook', { exact: true });
		await exportAction.click();
		await expect(page.getByRole('alert')).toContainText(name);
		await expect(page.getByRole('alert')).toContainText('32767');
		const saved = await request.get(`/api/functions/${id}`);
		expect((await saved.json()).description).toBe(description);
	} finally {
		await request.delete(`/api/functions/${id}`);
	}
});

test('workbook export still downloads a valid workbook', async ({ page }) => {
	await page.goto('/import');
	const exportAction = page.getByText('Export workbook', { exact: true });
	const download = page.waitForEvent('download');
	await exportAction.click();
	const file = await download;
	expect(file.suggestedFilename()).toBe('knowledge_workbook.xlsx');
	expect(await file.failure()).toBeNull();
});
