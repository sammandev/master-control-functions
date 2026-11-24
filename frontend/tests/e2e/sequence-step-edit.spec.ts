import { expect, test } from '@playwright/test';
import type { SequenceDetail } from '../../src/lib/sequences';

const step = (function_name: string, parameters: Record<string, string> = {}) => ({
	function_source: 'DLL',
	function_interface: 'WIFI_TEST',
	function_name,
	parameters,
	timeout_ms: '5000',
	flow: 'NEXT',
	on_fail: 'FAIL'
});

test('an open step edit follows its step when the step is moved', async ({
	page,
	request
}, testInfo) => {
	const created = await request.post('/api/sequences', {
		data: {
			name: `E2E step edit reorder ${Date.now()}`,
			steps: [
				step('DUT_INITIALIZE', { BAND: '6G', PROJECT: 'AST1', WL_INTERFACE: 'wl0' }),
				step('DUT_REMOVE')
			]
		}
	});
	expect(created.status()).toBe(201);
	const seq = (await created.json()) as SequenceDetail;
	const path = `/sequences/${seq.id}`;
	const emitArrived = Promise.withResolvers<void>();
	const releaseEmit = Promise.withResolvers<void>();
	try {
		await page.goto(path);
		await expect(page.locator('.emitted-preview')).toBeVisible();
		const steps = page.locator('.steps > li');
		await steps.nth(1).getByRole('button', { name: 'Edit', exact: true }).click();
		await page.getByRole('textbox', { name: 'Note' }).fill('Remove the DUT first');

		// Answer the post-reorder emit from the server, then hold it in the browser.
		let emits = 0;
		await page.route(`**/api/sequences/${seq.id}/emit`, async (route) => {
			const response = await route.fetch();
			if (emits++ === 0) {
				emitArrived.resolve();
				await releaseEmit.promise;
			}
			await route.fulfill({ response });
		});
		await steps.nth(1).getByRole('button', { name: 'Move step up', exact: true }).click();
		await emitArrived.promise;

		const editedStep = steps.filter({ has: page.locator('form.step-form') });
		await page.screenshot({ path: testInfo.outputPath('step-edit-during-emit.png') });
		await expect.soft(editedStep.locator('.step-identity')).toHaveText('DLLWIFI_TEST · DUT_REMOVE');

		await editedStep.getByRole('button', { name: 'Save step', exact: true }).click();
		await expect(page.getByText('Step saved', { exact: true })).toBeVisible();
		releaseEmit.resolve();

		const persisted = await request.get(`/api${path}`);
		expect(persisted.status()).toBe(200);
		const { steps: saved } = (await persisted.json()) as SequenceDetail;
		expect(
			saved.map(({ function_name, parameters, note }) => ({ function_name, parameters, note }))
		).toEqual([
			{ function_name: 'DUT_REMOVE', parameters: {}, note: 'Remove the DUT first' },
			{
				function_name: 'DUT_INITIALIZE',
				parameters: { BAND: '6G', PROJECT: 'AST1', WL_INTERFACE: 'wl0' },
				note: ''
			}
		]);
	} finally {
		releaseEmit.resolve();
		await request.delete(`/api${path}`);
	}
});

test('step changes wait for a pending step write', async ({ page, request }) => {
	const created = await request.post('/api/sequences', {
		data: {
			name: `E2E step write gate ${Date.now()}`,
			steps: [
				step('DUT_INITIALIZE', { BAND: '6G', PROJECT: 'AST1', WL_INTERFACE: 'wl0' }),
				step('DUT_REMOVE')
			]
		}
	});
	expect(created.status()).toBe(201);
	const seq = (await created.json()) as SequenceDetail;
	const path = `/sequences/${seq.id}`;
	const writeArrived = Promise.withResolvers<void>();
	const releaseWrite = Promise.withResolvers<void>();
	try {
		await page.goto(path);
		await expect(page.locator('.emitted-preview')).toBeVisible();
		const steps = page.locator('.steps > li');
		await steps.nth(0).getByRole('button', { name: 'Edit', exact: true }).click();

		await page.route(`**/api/sequences/${seq.id}/steps`, async (route) => {
			const response = await route.fetch();
			writeArrived.resolve();
			await releaseWrite.promise;
			await route.fulfill({ response });
		});
		await steps.nth(1).getByRole('button', { name: 'Move step up', exact: true }).click();
		await writeArrived.promise;

		const stepChanges = [
			page.getByRole('button', { name: 'Save step', exact: true }),
			...['Move step up', 'Move step down', 'Remove'].map((name) =>
				page.locator('.steps').getByRole('button', { name, exact: true }).first()
			)
		];
		for (const button of stepChanges) await expect(button).toBeDisabled();

		releaseWrite.resolve();
		await expect(steps.nth(1).locator('form.step-form')).toBeVisible();
		await expect(steps.nth(1).locator('.step-identity')).toHaveText(
			'DLLWIFI_TEST · DUT_INITIALIZE'
		);
		await expect(page.getByRole('button', { name: 'Save step', exact: true })).toBeEnabled();
	} finally {
		releaseWrite.resolve();
		await request.delete(`/api${path}`);
	}
});
