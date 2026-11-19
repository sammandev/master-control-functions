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

function renderDialog(
	options: {
		onChanged?: () => void | Promise<void>;
		categories?: typeof categories;
	} = {}
) {
	const onOpenChange = vi.fn();
	const onChanged = vi.fn(options.onChanged ?? (() => {}));
	render(TroubleshootingRegistryDialog, {
		open: true,
		categories: options.categories ?? categories,
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

function emptyResponse() {
	return new Response(null, { status: 204 });
}

afterEach(() => vi.unstubAllGlobals());

describe('TroubleshootingRegistryDialog', () => {
	it('offers category and severity tabs while hiding delete and tone for categories', async () => {
		renderDialog();
		await expect
			.element(page.getByRole('dialog', { name: 'Manage troubleshooting lists' }))
			.toBeVisible();
		await expect.element(page.getByRole('tab', { name: 'Categories' })).toBeVisible();
		await expect.element(page.getByRole('tab', { name: 'Severities' })).toBeVisible();
		await expect
			.element(page.getByRole('button', { name: 'Delete DUT Connection' }))
			.not.toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: /^Tone:/ })).not.toBeInTheDocument();
		await expect.element(page.getByText('4 guides')).toBeVisible();
	});

	it('offers exactly the four supported severity tone options', async () => {
		renderDialog();
		await page.getByRole('tab', { name: 'Severities' }).click();
		await page.getByRole('button', { name: 'Tone: warn' }).click();

		for (const tone of ['destructive', 'warn', 'default', 'outline']) {
			await expect.element(page.getByRole('option', { name: tone })).toBeVisible();
		}
		await expect.element(page.getByRole('option').nth(4)).not.toBeInTheDocument();
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

	it('does not turn a successful save into a failure when refresh fails', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) =>
				jsonResponse({ ...categories[1], ...JSON.parse(String(init?.body)) })
			)
		);
		const { onChanged } = renderDialog({
			onChanged: async () => {
				throw new Error('Refresh failed');
			}
		});

		await page.getByRole('textbox', { name: 'Flow Execution label' }).fill('Script Execution');
		await page.getByRole('button', { name: 'Save Flow Execution' }).click();

		expect(onChanged).toHaveBeenCalledOnce();
		await expect.element(page.getByRole('alert')).not.toBeInTheDocument();
		await expect.element(page.getByRole('button', { name: 'Save Script Execution' })).toBeEnabled();
	});

	it('serializes reorder requests and persists only order while retaining drafts', async () => {
		const calls: Array<{ url: string; body: unknown }> = [];
		let releaseFirst: () => void = () => {};
		const firstResponse = new Promise<Response>((resolve) => {
			releaseFirst = () => resolve(jsonResponse({ ...categories[1], sort_order: 10 }));
		});
		vi.stubGlobal(
			'fetch',
			vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
				const body = JSON.parse(String(init?.body));
				calls.push({ url: String(input), body });
				if (calls.length === 1) return firstResponse;
				return jsonResponse({ ...categories[0], ...body });
			})
		);
		renderDialog();
		await page.getByRole('textbox', { name: 'Flow Execution label' }).fill('Draft Execution');
		await page
			.getByRole('textbox', { name: 'Flow Execution description' })
			.fill('Unsaved description');

		await page.getByRole('button', { name: 'Move Flow Execution up' }).click();
		await expect.poll(() => calls).toHaveLength(1);
		expect(calls[0]).toEqual({
			url: '/api/troubleshooting/categories/execution',
			body: { label: 'Flow Execution', description: 'Execution failures', sort_order: 10 }
		});
		releaseFirst();
		await expect.poll(() => calls).toHaveLength(2);

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
		await expect
			.element(page.getByRole('textbox', { name: 'Flow Execution label' }))
			.toHaveValue('Draft Execution');
		await expect
			.element(page.getByRole('textbox', { name: 'Flow Execution description' }))
			.toHaveValue('Unsaved description');
	});

	it('reorders severity with persisted tone instead of an unsaved tone draft', async () => {
		const calls: Array<{ url: string; body: unknown }> = [];
		vi.stubGlobal(
			'fetch',
			vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
				const body = JSON.parse(String(init?.body));
				calls.push({ url: String(input), body });
				return jsonResponse({
					...severities.find((item) => String(input).endsWith(item.key)),
					...body
				});
			})
		);
		renderDialog();
		await page.getByRole('tab', { name: 'Severities' }).click();
		await page.getByRole('button', { name: 'Tone: warn' }).click();
		await page.getByRole('option', { name: 'default' }).click();

		await page.getByRole('button', { name: 'Move Major up' }).click();

		expect(calls[0]).toEqual({
			url: '/api/troubleshooting/severities/major',
			body: { label: 'Major', description: 'Needs attention', sort_order: 10, tone: 'warn' }
		});
		await expect.element(page.getByRole('button', { name: 'Tone: default' })).toBeVisible();
	});

	it('requires a different replacement and keeps failed deletion open with visible error', async () => {
		const fetchMock = vi.fn<typeof fetch>(async () =>
			jsonResponse({ error: 'Reassignment failed' }, 500)
		);
		vi.stubGlobal('fetch', fetchMock);
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

		expect(fetchMock).toHaveBeenCalledOnce();
		expect(fetchMock.mock.calls[0][0]).toBe(
			'/api/troubleshooting/categories/execution?reassign_to=connection'
		);
		expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
		await expect.element(alert.getByRole('alert')).toHaveTextContent('Reassignment failed');
		await expect.element(alert).toBeVisible();
	});

	it('encodes reassignment keys and closes after DELETE despite refresh failure', async () => {
		const fetchMock = vi.fn<typeof fetch>(async () => emptyResponse());
		vi.stubGlobal('fetch', fetchMock);
		const { onChanged } = renderDialog({
			categories: [
				{
					key: 'connection & setup',
					label: 'DUT Connection',
					description: 'Connection failures',
					sort_order: 10,
					built_in: true,
					guide_count: 4
				},
				{
					key: 'execution/flow',
					label: 'Flow Execution',
					description: 'Execution failures',
					sort_order: 20,
					built_in: false,
					guide_count: 2
				}
			],
			onChanged: async () => {
				throw new Error('Refresh failed');
			}
		});

		await page.getByRole('button', { name: 'Delete Flow Execution' }).click();
		await page.getByRole('button', { name: 'Choose replacement category' }).click();
		await page.getByRole('option', { name: 'DUT Connection' }).click();
		await page.getByRole('button', { name: 'Delete category' }).click();

		expect(fetchMock).toHaveBeenCalledOnce();
		expect(fetchMock.mock.calls[0][0]).toBe(
			'/api/troubleshooting/categories/execution%2Fflow?reassign_to=connection%20%26%20setup'
		);
		expect(fetchMock.mock.calls[0][1]?.method).toBe('DELETE');
		expect(onChanged).toHaveBeenCalledOnce();
		await expect.element(page.getByRole('alertdialog')).not.toBeInTheDocument();
		await expect
			.element(page.getByRole('button', { name: 'Delete Flow Execution' }))
			.not.toBeInTheDocument();
	});
});
