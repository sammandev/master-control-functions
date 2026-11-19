import { expect, test, type Page } from '@playwright/test';

const categories = [
	{
		key: 'measurement',
		label: 'Measurement & Criteria',
		description: '',
		sort_order: 1,
		built_in: true,
		guide_count: 2
	}
];
const severities = [
	{
		key: 'critical',
		label: 'Critical',
		description: '',
		sort_order: 1,
		built_in: true,
		guide_count: 2,
		tone: 'destructive'
	}
];

const refreshes = { list: 0, detail: 0 };

async function mockTroubleshooting(page: Page) {
	refreshes.list = 0;
	refreshes.detail = 0;
	await page.route('**/api/**', async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		const path = url.pathname;
		if (path === '/api/troubleshooting/categories' && request.method() === 'GET') {
			return route.fulfill({ json: { items: categories, total: categories.length } });
		}
		if (path === '/api/troubleshooting/categories/measurement' && request.method() === 'PUT') {
			categories[0].label = 'Renamed Measurement';
			return route.fulfill({ json: categories[0] });
		}
		if (path === '/api/troubleshooting/categories' && request.method() === 'POST') {
			const item = {
				key: 'advanced-measurement',
				label: 'Advanced Measurement',
				description: '',
				sort_order: 2,
				built_in: false,
				guide_count: 0
			};
			categories.push(item);
			return route.fulfill({ status: 201, json: item });
		}
		if (path === '/api/troubleshooting/severities' && request.method() === 'GET') {
			return route.fulfill({ json: { items: severities, total: severities.length } });
		}
		if (path === '/api/troubleshooting/severities' && request.method() === 'POST') {
			const item = {
				key: 'needs-review',
				label: 'Needs Review',
				description: '',
				sort_order: 2,
				built_in: false,
				guide_count: 0,
				tone: 'outline'
			};
			severities.push(item);
			return route.fulfill({ status: 201, json: item });
		}
		if (path === '/api/devices') return route.fulfill({ json: { items: [], total: 0 } });
		if (path === '/api/troubleshooting') {
			refreshes.list++;
			return route.fulfill({
				json: {
					items: guides.map(({ steps: _steps, ...guide }) => ({
						...guide,
						step_count: _steps.length,
						related_count: 1
					})),
					total: guides.length,
					page: 1,
					page_size: 20
				}
			});
		}
		if (/^\/api\/troubleshooting\/\d+\/related$/.test(path)) {
			const id = Number(path.match(/\d+/)?.[0]);
			return route.fulfill({
				json: {
					items: guides
						.filter((guide) => guide.id !== id)
						.map(({ steps, ...guide }) => ({
							...guide,
							step_count: steps.length,
							related_count: 1
						})),
					total: 1
				}
			});
		}
		const match = path.match(/^\/api\/troubleshooting\/(\d+)$/);
		if (match) {
			refreshes.detail++;
			return route.fulfill({ json: guides.find((guide) => guide.id === Number(match[1])) });
		}
		return route.fulfill({ status: 404, json: { error: 'not mocked' } });
	});
}

const guides = [
	{
		id: 1,
		title: 'First guide',
		category: 'measurement',
		severity: 'critical',
		device_name: 'WIFI_TEST',
		function_name: 'TX_VERIFY',
		error_code: 'ERRCODE:1',
		symptom: 'First symptom',
		root_cause: 'First cause',
		steps: [
			{
				id: 1,
				position: 1,
				title: 'Inspect first',
				instruction: 'Inspect it',
				script_snippet: '',
				verification: 'Verified'
			}
		],
		created_at: '2026-11-19T00:00:00Z',
		updated_at: '2026-11-19T00:00:00Z'
	},
	{
		id: 2,
		title: 'Second guide',
		category: 'measurement',
		severity: 'critical',
		device_name: 'WIFI_TEST',
		function_name: 'TX_VERIFY',
		error_code: 'ERRCODE:1',
		symptom: 'Second symptom',
		root_cause: 'Second cause',
		steps: [
			{
				id: 2,
				position: 1,
				title: 'Inspect second',
				instruction: 'Inspect it too',
				script_snippet: '',
				verification: ''
			}
		],
		created_at: '2026-11-19T00:00:00Z',
		updated_at: '2026-11-19T00:00:00Z'
	}
];

test('troubleshooting workflow preserves checklist and supports registry controls', async ({
	page
}) => {
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	page.on('console', (message) => {
		if (message.type() === 'error') errors.push(message.text());
	});
	await mockTroubleshooting(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/troubleshooting');

	await expect(page.getByText('Testing Stage / Category', { exact: true })).toBeVisible();
	await expect(page.getByText('Device / Interface', { exact: true })).toBeVisible();
	const categoryFilter = page.getByRole('button', { name: 'Filter by category' });
	const deviceFilter = page.getByRole('button', { name: /Device \/ Interface:/ });
	await expect(categoryFilter).toHaveCSS('height', '40px');
	await expect(deviceFilter).toHaveCSS('height', '40px');
	await page.screenshot({
		path: test.info().outputPath('troubleshooting-desktop.png'),
		fullPage: true
	});
	await expect(
		page.getByRole('button', { name: 'Add Alternate Solution', exact: true })
	).toHaveText('Add Alternate Solution');

	const firstStep = page.getByRole('checkbox', { name: /Mark step 1: Inspect first/ });
	await firstStep.check();
	await page.getByRole('button', { name: 'Select guide Second guide' }).click();
	await page.getByRole('button', { name: 'Select guide First guide' }).click();
	await expect(page.getByRole('checkbox', { name: /Mark step 1: Inspect first/ })).toBeChecked();
	await page.getByRole('button', { name: 'Reset progress' }).click();
	await expect(
		page.getByRole('checkbox', { name: /Mark step 1: Inspect first/ })
	).not.toBeChecked();

	const listRefreshes = refreshes.list;
	const detailRefreshes = refreshes.detail;
	await page.getByRole('button', { name: 'Manage Lists' }).click();
	const manager = page.getByRole('dialog', { name: 'Manage troubleshooting lists' });
	await expect(manager).toBeVisible();
	await manager
		.getByRole('textbox', { name: 'Measurement & Criteria label' })
		.fill('Renamed Measurement');
	await manager.getByRole('button', { name: 'Save Measurement & Criteria' }).click();
	await expect(page.getByText('Renamed Measurement', { exact: true }).first()).toBeVisible();
	await expect.poll(() => refreshes.list).toBeGreaterThan(listRefreshes);
	await expect.poll(() => refreshes.detail).toBeGreaterThan(detailRefreshes);
	await page.getByRole('button', { name: 'Close' }).click();

	await page.getByRole('button', { name: 'New Troubleshooting Guide' }).click();
	const editor = page.getByRole('dialog', { name: 'New Troubleshooting Guide' });
	await expect(editor.getByText('Testing Stage / Category', { exact: true })).toBeVisible();
	await expect(editor.getByText('Severity', { exact: true })).toBeVisible();
	await expect(editor.getByText('Device / Interface Scope', { exact: true })).toBeVisible();
	await page
		.getByRole('button', { name: /Category:/ })
		.last()
		.click();
	const categorySearch = page.getByRole('combobox', { name: 'Search categories' });
	await categorySearch.fill('Advanced Measurement');
	await page.getByRole('button', { name: 'Create “Advanced Measurement”' }).click();
	await expect(page.getByRole('button', { name: 'Category: Advanced Measurement' })).toBeVisible();
	await page
		.getByRole('button', { name: /Severity:/ })
		.last()
		.click();
	const severitySearch = page.getByRole('combobox', { name: 'Search severities' });
	await severitySearch.fill('Needs Review');
	await page.getByRole('button', { name: 'Create “Needs Review”' }).click();
	await expect(page.getByRole('button', { name: 'Severity: Needs Review' })).toBeVisible();
	await page.getByRole('button', { name: 'Close' }).last().click();

	await page.setViewportSize({ width: 390, height: 844 });
	await page.screenshot({
		path: test.info().outputPath('troubleshooting-mobile.png'),
		fullPage: true
	});
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth <= document.documentElement.clientWidth
		)
	).toBe(true);
	expect(errors).toEqual([]);
});
