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

test('renaming after another tab reorders steps cannot move a draft to a different step', async ({
	page,
	request
}) => {
	const name = `E2E reordered rename ${Date.now()}`;
	const created = await request.post('/api/sequences', {
		data: {
			name,
			steps: [
				{ function_source: 'DLL', function_name: 'E2E_FIRST_STEP' },
				{ function_source: 'DLL', function_name: 'E2E_SECOND_STEP' }
			]
		}
	});
	expect(created.status()).toBe(201);
	const { id } = (await created.json()) as SequenceDetail;
	const path = `/sequences/${id}`;
	const otherTab = await page.context().newPage();

	try {
		await Promise.all([page.goto(path), otherTab.goto(path)]);
		const steps = page.locator('.steps');
		await steps.getByRole('button', { name: 'Edit', exact: true }).first().click();
		await page.getByLabel('Note', { exact: true }).fill('Unsaved draft');
		await page.getByRole('button', { name: 'Rename', exact: true }).click();
		await page.getByLabel('Sequence name', { exact: true }).fill(`${name} renamed`);
		await expect(page.getByLabel('Note', { exact: true })).toHaveValue('Unsaved draft');

		const reorder = otherTab.waitForResponse(
			(response) => response.url().endsWith(`/api${path}`) && response.request().method() === 'PUT'
		);
		await otherTab.getByRole('button', { name: 'Move step down', exact: true }).first().click();
		const reordered = await reorder;
		expect(reordered.status()).toBe(200);
		const saved = (await reordered.json()) as SequenceDetail;
		expect(saved.steps.map((step) => step.function_name)).toEqual([
			'E2E_SECOND_STEP',
			'E2E_FIRST_STEP'
		]);

		await page.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(page.getByRole('heading', { name: `${name} renamed`, exact: true })).toBeVisible();
		await expect(page.getByRole('button', { name: 'Save step', exact: true })).not.toBeVisible();
		await expect(
			page.getByText('Steps changed. Reopen the step to continue editing.')
		).toBeVisible();
		const renamed = await request.get(`/api${path}`);
		expect(await renamed.json()).toMatchObject({ name: `${name} renamed`, steps: saved.steps });
	} finally {
		await request.delete(`/api${path}`);
	}
});
