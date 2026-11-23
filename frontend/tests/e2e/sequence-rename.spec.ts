import { expect, test } from '@playwright/test';
import type { SequenceDetail } from '../../src/lib/sequences';

test('renaming a sequence in a stale tab preserves a saved step edit', async ({
	page,
	request
}) => {
	const name = `E2E stale rename ${Date.now()}`;
	const created = await request.post('/api/sequences', {
		data: {
			name,
			steps: [{ function_source: 'DLL', function_name: 'E2E_RENAME_STEP', note: 'Original note' }]
		}
	});
	expect(created.status()).toBe(201);
	const { id } = (await created.json()) as SequenceDetail;
	const path = `/sequences/${id}`;
	const renameTab = await page.context().newPage();

	try {
		await Promise.all([page.goto(path), renameTab.goto(path)]);
		await expect(renameTab.getByText('Original note', { exact: true })).toBeVisible();

		await page.getByRole('button', { name: 'Edit', exact: true }).click();
		await page.getByLabel('Note', { exact: true }).fill('Saved in the other tab');
		await page.getByRole('button', { name: 'Save step', exact: true }).click();
		await expect(page.getByText('Step saved', { exact: true })).toBeVisible();
		const saved = await request.get(`/api${path}`);
		expect(saved.status()).toBe(200);
		const { steps } = (await saved.json()) as SequenceDetail;
		expect(steps[0].note).toBe('Saved in the other tab');
		await expect(renameTab.getByText('Original note', { exact: true })).toBeVisible();

		await renameTab.getByRole('button', { name: 'Rename', exact: true }).click();
		await renameTab.getByLabel('Sequence name', { exact: true }).fill(`${name} renamed`);
		await renameTab.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(
			renameTab.getByRole('heading', { name: `${name} renamed`, exact: true })
		).toBeVisible();

		const renamed = await request.get(`/api${path}`);
		expect(renamed.status()).toBe(200);
		expect(await renamed.json()).toMatchObject({ name: `${name} renamed`, steps });
	} finally {
		await request.delete(`/api${path}`);
	}
});
