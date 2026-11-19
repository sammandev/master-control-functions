import { page } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import TroubleshootingRegistryDialog from '../../src/lib/components/TroubleshootingRegistryDialog.svelte';

const categories = [
	{
		key: 'connection',
		label: 'DUT Connection',
		description: 'Connection failures',
		sort_order: 10,
		built_in: true,
		guide_count: 4
	},
	{
		key: 'execution',
		label: 'Flow Execution',
		description: 'Execution failures',
		sort_order: 20,
		built_in: false,
		guide_count: 2
	}
];

const severities = [
	{
		key: 'critical',
		label: 'Critical',
		description: 'Blocks the flow',
		sort_order: 10,
		built_in: true,
		guide_count: 3,
		tone: 'destructive'
	},
	{
		key: 'major',
		label: 'Major',
		description: 'Needs attention',
		sort_order: 20,
		built_in: false,
		guide_count: 1,
		tone: 'warn'
	}
];

function renderDialog(options: { onChanged?: () => void | Promise<void> } = {}) {
	const onOpenChange = vi.fn();
	const onChanged = vi.fn(options.onChanged ?? (() => {}));
	render(TroubleshootingRegistryDialog, {
		open: true,
		categories,
		severities,
		onOpenChange,
		onChanged
	});
	return { onChanged, onOpenChange };
}

function jsonResponse(body: unknown, status = 200) {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'content-type': 'application/json' }
	});
}

afterEach(() => vi.unstubAllGlobals());

describe('TroubleshootingRegistryDialog', () => {
	it('offers category and severity tabs while hiding delete for built-in rows', async () => {
		renderDialog();
		await expect
			.element(page.getByRole('dialog', { name: 'Manage troubleshooting lists' }))
			.toBeVisible();
		await expect.element(page.getByRole('tab', { name: 'Categories' })).toBeVisible();
		await expect.element(page.getByRole('tab', { name: 'Severities' })).toBeVisible();
		await expect
			.element(page.getByRole('button', { name: 'Delete DUT Connection' }))
			.not.toBeInTheDocument();
		await expect.element(page.getByText('4 guides')).toBeVisible();
	});

	it('saves edited metadata with the preserved stable key and supported severity tone', async () => {
		const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
			const body = JSON.parse(String(init?.body));
			return jsonResponse({ ...severities[1], ...body });
		});
		vi.stubGlobal('fetch', fetchMock);
		const { onChanged } = renderDialog();

		await page.getByRole('tab', { name: 'Severities' }).click();
		await page.getByRole('textbox', { name: 'Major label' }).fill('High Priority');
		await page.getByRole('textbox', { name: 'Major description' }).fill('Escalate quickly');
		await page.getByRole('button', { name: 'Tone: warn' }).click();
		await page.getByRole('option', { name: 'default' }).click();
		await page.getByRole('button', { name: 'Save Major' }).click();

		expect(fetchMock).toHaveBeenCalledOnce();
		const [url, init] = fetchMock.mock.calls[0];
		expect(String(url)).toBe('/api/troubleshooting/severities/major');
		expect(init?.method).toBe('PUT');
		expect(JSON.parse(String(init?.body))).toEqual({
			label: 'High Priority',
			description: 'Escalate quickly',
			sort_order: 20,
			tone: 'default'
		});
		expect(onChanged).toHaveBeenCalledOnce();
	});

	it('moves rows by serially swapping sort order through update requests', async () => {
		const calls: Array<{ url: string; body: unknown }> = [];
		vi.stubGlobal(
			'fetch',
			vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
				const body = JSON.parse(String(init?.body));
				calls.push({ url: String(input), body });
				return jsonResponse({
					...categories.find((item) => item.key === String(input).split('/').at(-1)),
					...body
				});
			})
		);
		renderDialog();

		await page.getByRole('button', { name: 'Move Flow Execution up' }).click();

		expect(calls).toEqual([
			{
				url: '/api/troubleshooting/categories/execution',
				body: { label: 'Flow Execution', description: 'Execution failures', sort_order: 10 }
			},
			{
				url: '/api/troubleshooting/categories/connection',
				body: { label: 'DUT Connection', description: 'Connection failures', sort_order: 20 }
			}
		]);
	});

	it('requires a different replacement and keeps failed deletion open with visible error', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(async () => jsonResponse({ error: 'Reassignment failed' }, 500))
		);
		renderDialog();

		await page.getByRole('button', { name: 'Delete Flow Execution' }).click();
		const alert = page.getByRole('alertdialog');
		await expect
			.element(alert.getByRole('heading', { name: 'Delete Flow Execution?' }))
			.toBeVisible();
		await expect.element(alert.getByText(/2 guides use this category/)).toBeVisible();
		await page.getByRole('button', { name: 'Delete category' }).click();
		await expect.element(alert.getByRole('alert')).toHaveTextContent('Choose a different category');

		await page.getByRole('button', { name: 'Choose replacement category' }).click();
		await expect.element(page.getByRole('option', { name: 'DUT Connection' })).toBeVisible();
		await expect
			.element(page.getByRole('option', { name: 'Flow Execution' }))
			.not.toBeInTheDocument();
		await page.getByRole('option', { name: 'DUT Connection' }).click();
		await page.getByRole('button', { name: 'Delete category' }).click();

		await expect.element(alert.getByRole('alert')).toHaveTextContent('Reassignment failed');
		await expect.element(alert).toBeVisible();
	});
});
