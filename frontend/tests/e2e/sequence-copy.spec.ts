import { expect, test, type APIRequestContext, type Page } from '@playwright/test';
import type { EmitResult, SequenceDetail } from '../../src/lib/sequences';

const step = (function_name: string, parameters: Record<string, string> = {}) => ({
	function_source: 'DLL',
	function_interface: 'WIFI_TEST',
	function_name,
	parameters,
	timeout_ms: '5000',
	flow: 'NEXT',
	on_fail: 'FAIL'
});

async function createSequence(request: APIRequestContext, name: string) {
	const created = await request.post('/api/sequences', {
		data: {
			name,
			steps: [
				step('DUT_INITIALIZE', { BAND: '6G', PROJECT: 'AST1', WL_INTERFACE: 'wl0' }),
				step('DUT_REMOVE'),
				step('DUT_INITIALIZE', { BAND: '6G', PROJECT: 'AST2', WL_INTERFACE: 'wl1' })
			]
		}
	});
	expect(created.status()).toBe(201);
	return (await created.json()) as SequenceDetail;
}

async function emittedText(request: APIRequestContext, id: number) {
	const emitted = await request.post(`/api/sequences/${id}/emit`);
	expect(emitted.status()).toBe(200);
	const { text } = (await emitted.json()) as EmitResult;
	expect(text.split('\n')).toHaveLength(3);
	return text;
}

// Each emit is answered by the server when it is sent, then held until released,
// so a held response still describes the step order it was requested for.
function holdEmits(page: Page, id: number) {
	const arrivals = [Promise.withResolvers<void>(), Promise.withResolvers<void>()];
	const releases = [Promise.withResolvers<void>(), Promise.withResolvers<void>()];
	let count = 0;
	return {
		arrived: (n: number) => arrivals[n].promise,
		release: (n: number) => releases[n].resolve(),
		install: () =>
			page.route(`**/api/sequences/${id}/emit`, async (route) => {
				const n = count++;
				const response = await route.fetch();
				arrivals[n].resolve();
				await releases[n].promise;
				await route.fulfill({ response });
			})
	};
}

test('Copy all waits for the reordered steps to revalidate', async ({ page, context, request }) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	const seq = await createSequence(request, `E2E copy reorder ${Date.now()}`);
	const path = `/sequences/${seq.id}`;
	try {
		await page.goto(path);
		const preview = page.locator('.emitted-preview');
		await expect(preview).toHaveText(await emittedText(request, seq.id));
		const copyAll = page.getByRole('button', { name: 'Copy all', exact: true });

		const emits = holdEmits(page, seq.id);
		await emits.install();
		await page
			.locator('.steps')
			.getByRole('button', { name: 'Move step down', exact: true })
			.first()
			.click();
		await emits.arrived(0);
		await expect(page.getByText('Validating steps…', { exact: true })).toBeVisible();
		await expect(copyAll).toBeDisabled();

		emits.release(0);
		const reordered = await emittedText(request, seq.id);
		await expect(preview).toHaveText(reordered);
		await copyAll.click();
		await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
		expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(reordered);
	} finally {
		await request.delete(`/api${path}`);
	}
});

test('Copy all ignores an emit response for an older step order', async ({
	page,
	context,
	request
}) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	const seq = await createSequence(request, `E2E copy stale emit ${Date.now()}`);
	const path = `/sequences/${seq.id}`;
	try {
		await page.goto(path);
		const preview = page.locator('.emitted-preview');
		await expect(preview).toHaveText(await emittedText(request, seq.id));
		const moveDown = page.locator('.steps').getByRole('button', { name: 'Move step down' });

		const emits = holdEmits(page, seq.id);
		await emits.install();
		await moveDown.nth(0).click();
		await emits.arrived(0);
		await moveDown.nth(1).click();
		await emits.arrived(1);
		await expect(page.locator('.step-identity')).toHaveText([
			'DLLWIFI_TEST · DUT_REMOVE',
			'DLLWIFI_TEST · DUT_INITIALIZE',
			'DLLWIFI_TEST · DUT_INITIALIZE'
		]);
		const latest = await emittedText(request, seq.id);

		emits.release(1);
		await expect(preview).toHaveText(latest);
		emits.release(0);
		await page.waitForLoadState('networkidle');
		await expect(preview).toHaveText(latest);

		await page.getByRole('button', { name: 'Copy all', exact: true }).click();
		await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
		expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(latest);
	} finally {
		await request.delete(`/api${path}`);
	}
});
