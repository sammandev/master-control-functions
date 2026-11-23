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

test('downloads templates, previews bulk imports, and exports selected or matching guides', async ({
	page
}) => {
	await mockTroubleshooting(page);
	const exports: Array<{ scope: string; format: string; ids?: number[]; search: string | null }> =
		[];
	let commits = 0;
	await page.route('**/api/troubleshooting/templates/json', (route) =>
		route.fulfill({
			json: { guides: [{ title: 'Example' }] },
			headers: {
				'Content-Disposition': 'attachment; filename="troubleshooting-guide-template.json"'
			}
		})
	);
	await page.route('**/api/troubleshooting/templates/xlsx', (route) =>
		route.fulfill({
			body: 'template workbook',
			headers: {
				'Content-Disposition': 'attachment; filename="troubleshooting-guide-template.xlsx"'
			}
		})
	);
	await page.route('**/api/troubleshooting/import/preview', (route) => {
		const invalid = route.request().postData()?.includes('invalid.json');
		return route.fulfill({
			json: {
				valid: !invalid,
				rows: [
					{ index: 1, title: 'Imported guide', action: 'create' },
					invalid
						? { index: 2, title: 'Rejected guide', action: 'error', reasons: ['unknown category'] }
						: { index: 2, id: 2, title: 'Updated guide', action: 'update' }
				]
			}
		});
	});
	await page.route('**/api/troubleshooting/import/commit', (route) => {
		commits++;
		return route.fulfill({ json: { created: 1, updated: 1 } });
	});
	await page.route(/\/api\/troubleshooting\/export/, (route) => {
		const req = route.request();
		exports.push({ ...req.postDataJSON(), search: new URL(req.url()).searchParams.get('search') });
		return route.fulfill({ body: 'export file' });
	});

	await page.goto('/troubleshooting');
	const sheet = process.env.TROUBLESHOOTING_SCREENSHOT_DIR;
	async function capture(name: string) {
		if (sheet) await page.screenshot({ path: `${sheet}/troubleshooting-${name}.png` });
	}
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	await capture('template-download');
	const template = page.waitForEvent('download');
	await page.getByRole('menuitem', { name: 'Download JSON template' }).click();
	expect((await template).suggestedFilename()).toBe('troubleshooting-guide-template.json');
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	const workbook = page.waitForEvent('download');
	await page.getByRole('menuitem', { name: 'Download XLSX template' }).click();
	expect((await workbook).suggestedFilename()).toBe('troubleshooting-guide-template.xlsx');
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	await page.getByRole('menuitem', { name: 'Bulk import' }).click();
	const bulk = page.getByRole('dialog', { name: 'Bulk Import Guides' });
	const upload = bulk.getByLabel('Guide file');
	const confirm = bulk.getByRole('button', { name: 'Confirm import' });
	await upload.setInputFiles({
		name: 'invalid.json',
		mimeType: 'application/json',
		buffer: Buffer.from('{}')
	});
	await bulk.getByRole('button', { name: 'Preview import' }).click();
	await expect(bulk.getByText('unknown category')).toBeVisible();
	await expect(confirm).toBeDisabled();
	await capture('preview-errors');
	await upload.setInputFiles({
		name: 'clean.json',
		mimeType: 'application/json',
		buffer: Buffer.from('{}')
	});
	await bulk.getByRole('button', { name: 'Preview import' }).click();
	await expect(bulk.getByText('Updated guide')).toBeVisible();
	await expect(confirm).toBeEnabled();
	expect(commits).toBe(0);
	await capture('preview-clean');
	await confirm.click();
	await expect(bulk).toBeHidden();
	expect(commits).toBe(1);

	await page.getByRole('checkbox', { name: 'Select First guide for export' }).check();
	await page.getByRole('button', { name: 'Export', exact: true }).click();
	const exportDialog = page.getByRole('dialog', { name: 'Export Guides' });
	await expect(exportDialog.getByText('1 selected guide')).toBeVisible();
	await exportDialog.getByRole('radio', { name: /Both/ }).check();
	await capture('export-dialog');
	const selected = page.waitForEvent('download');
	await exportDialog.getByRole('button', { name: 'Download export' }).click();
	expect((await selected).suggestedFilename()).toBe('troubleshooting-guides.zip');
	expect(exports[0]).toEqual({ scope: 'selected', format: 'both', ids: [1], search: null });
	await page.getByRole('checkbox', { name: 'Select page' }).check();
	await page.getByRole('button', { name: 'Export', exact: true }).click();
	await expect(exportDialog.getByText('2 selected guides')).toBeVisible();
	await exportDialog.getByRole('radio', { name: 'XLSX', exact: true }).check();
	await exportDialog.getByRole('button', { name: 'Download export' }).click();
	await expect(exportDialog).toBeHidden();
	expect(exports[1]).toEqual({ scope: 'selected', format: 'xlsx', ids: [1, 2], search: null });
	await page.getByRole('searchbox', { name: 'Search troubleshooting guides' }).fill('prompt');
	await page.getByRole('button', { name: 'Select all matching filters' }).click();
	await expect(
		exportDialog.getByRole('radio', { name: 'All matching current filters' })
	).toBeChecked();
	await exportDialog.getByRole('button', { name: 'Download export' }).click();
	await expect(exportDialog).toBeHidden();
	await expect
		.poll(() => exports[2])
		.toEqual({ scope: 'matching', format: 'json', search: 'prompt' });
	await page.getByRole('button', { name: 'Export', exact: true }).click();
	await exportDialog.getByRole('radio', { name: 'All guides', exact: true }).check();
	await exportDialog.getByRole('button', { name: 'Download export' }).click();
	await expect(exportDialog).toBeHidden();
	expect(exports[3]).toEqual({ scope: 'all', format: 'json', search: null });
});

test('single import prefills an existing edit without saving or changing its target', async ({
	page
}) => {
	await mockTroubleshooting(page);
	await page.route('**/api/troubleshooting/import/single', (route) =>
		route.fulfill({
			json: { ...seedGuides[0], id: 2, title: 'Imported edit', symptom: '- One\n- Two' }
		})
	);
	await page.goto('/troubleshooting');
	await page.getByRole('button', { name: 'Edit', exact: true }).click();
	const edit = page.getByRole('dialog', { name: 'Edit Troubleshooting Guide' });
	await edit.getByLabel('Import from file').setInputFiles({
		name: 'guide.xlsx',
		mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
		buffer: Buffer.from('workbook')
	});
	await expect(edit.getByRole('textbox', { name: /Solution \/ Manifestation Title/ })).toHaveValue(
		'Imported edit'
	);
	await expect(edit.getByRole('textbox', { name: /Observed Symptom/ })).toHaveValue('- One\n- Two');
	await expect(page.getByRole('heading', { name: 'First guide' })).toBeVisible();
	const saved = page.waitForRequest((req) => req.method() === 'PUT');
	await edit.getByRole('button', { name: 'Save Changes' }).click();
	expect(new URL((await saved).url()).pathname).toBe('/api/troubleshooting/1');
});

async function assertDialogChromeIsSeparated(
	dialog: ReturnType<Page['getByRole']>,
	stickyTabs = false
) {
	const heading = dialog.locator('.dialog-heading');
	const body = dialog.locator('.dialog-body');
	await expect(body).toBeVisible();
	const headingBox = await heading.boundingBox();
	const bodyBox = await body.boundingBox();
	expect(headingBox).not.toBeNull();
	expect(bodyBox).not.toBeNull();
	expect(bodyBox!.y).toBeGreaterThanOrEqual(headingBox!.y + headingBox!.height - 0.5);

	if (stickyTabs) {
		const tabs = dialog.getByRole('tablist', { name: 'Troubleshooting lists' });
		await body.evaluate((element) => {
			element.scrollTop = element.scrollHeight;
		});
		await expect.poll(() => body.evaluate((element) => element.scrollTop)).toBeGreaterThan(0);
		await expect(tabs).toBeVisible();
		await expect(tabs).toHaveCSS('position', 'sticky');
		const tabsBox = await tabs.boundingBox();
		expect(tabsBox).not.toBeNull();
		expect(Math.abs(tabsBox!.y - bodyBox!.y)).toBeLessThan(2);
	}

	const overlap = await dialog.evaluate((element) => {
		const heading = element.querySelector('.dialog-heading')!.getBoundingClientRect();
		return (
			element.ownerDocument
				.elementFromPoint(heading.left + 12, heading.top - 1)
				?.closest('[role="dialog"] [role="tabpanel"], [role="dialog"] form') != null
		);
	});
	expect(overlap).toBe(false);
}

async function mockTroubleshooting(page: Page, options: { failOneCategoryRefresh?: boolean } = {}) {
	const guides = structuredClone(seedGuides);
	categories.splice(
		0,
		categories.length,
		{
			key: 'measurement',
			label: 'Measurement & Criteria',
			description: '',
			sort_order: 1,
			built_in: true,
			guide_count: 2
		},
		...Array.from({ length: 10 }, (_, index) => ({
			key: `category-${index + 2}`,
			label: `Category ${index + 2}`,
			description: '',
			sort_order: index + 2,
			built_in: true,
			guide_count: 0
		}))
	);
	severities.splice(0, severities.length, {
		key: 'critical',
		label: 'Critical',
		description: '',
		sort_order: 1,
		built_in: true,
		guide_count: 2,
		tone: 'destructive'
	});
	refreshes.list = 0;
	refreshes.detail = 0;
	let categoryLoads = 0;
	let categoryRefreshFailed = false;
	await page.route('**/api/**', async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		const path = url.pathname;
		if (path === '/api/troubleshooting/categories' && request.method() === 'GET') {
			categoryLoads++;
			if (options.failOneCategoryRefresh && categoryLoads > 1 && !categoryRefreshFailed) {
				categoryRefreshFailed = true;
				return route.fulfill({ status: 500, json: { error: 'Metadata refresh failed' } });
			}
			return route.fulfill({ json: { items: categories, total: categories.length } });
		}
		if (path === '/api/troubleshooting/categories/measurement' && request.method() === 'PUT') {
			if (!request.headers()['content-type']?.startsWith('application/json')) {
				return route.fulfill({ status: 400, json: { error: 'invalid JSON body' } });
			}
			categories[0].label = 'Renamed Measurement';
			return route.fulfill({ json: categories[0] });
		}
		if (path === '/api/troubleshooting/categories' && request.method() === 'POST') {
			if (!request.headers()['content-type']?.startsWith('application/json')) {
				return route.fulfill({ status: 400, json: { error: 'invalid JSON body' } });
			}
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
			if (!request.headers()['content-type']?.startsWith('application/json')) {
				return route.fulfill({ status: 400, json: { error: 'invalid JSON body' } });
			}
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
			if (request.method() === 'POST') {
				if (!request.headers()['content-type']?.startsWith('application/json')) {
					return route.fulfill({ status: 400, json: { error: 'invalid JSON body' } });
				}
				const body = request.postDataJSON();
				const created = {
					...body,
					id: Math.max(0, ...guides.map((guide) => guide.id)) + 1,
					steps: body.steps.map((step: object, index: number) => ({
						...step,
						id: index + 100,
						position: index + 1
					})),
					created_at: '2026-11-21T00:00:00Z',
					updated_at: '2026-11-21T00:00:00Z'
				};
				guides.push(created);
				return route.fulfill({ status: 201, json: created });
			}
			refreshes.list++;
			return route.fulfill({
				json: {
					items: guides.map(({ steps: _steps, ...guide }) => ({
						...guide,
						step_count: _steps.length,
						related_count: Math.max(0, guides.length - 1)
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
			const id = Number(match[1]);
			const index = guides.findIndex((guide) => guide.id === id);
			if (request.method() === 'PUT') {
				if (!request.headers()['content-type']?.startsWith('application/json')) {
					return route.fulfill({ status: 400, json: { error: 'invalid JSON body' } });
				}
				if (index < 0) return route.fulfill({ status: 404, json: { error: 'not found' } });
				const body = request.postDataJSON();
				guides[index] = {
					...guides[index],
					...body,
					steps: body.steps.map((step: object, stepIndex: number) => ({
						...step,
						id: stepIndex + 200,
						position: stepIndex + 1
					})),
					updated_at: '2026-11-21T01:00:00Z'
				};
				return route.fulfill({ json: guides[index] });
			}
			if (request.method() === 'DELETE') {
				if (index < 0) return route.fulfill({ status: 404, json: { error: 'not found' } });
				guides.splice(index, 1);
				return route.fulfill({ status: 204 });
			}
			refreshes.detail++;
			return route.fulfill({ json: guides.find((guide) => guide.id === id) });
		}
		return route.fulfill({ status: 404, json: { error: 'not mocked' } });
	});
}

const seedGuides = [
	{
		id: 1,
		title: 'First guide',
		category: 'measurement',
		severity: 'critical',
		device_name: 'WIFI_TEST',
		function_name: 'TX_VERIFY',
		error_code: 'ERRCODE:1',
		symptom: '1. First symptom\n2. Confirm prompt state',
		root_cause: '- First cause\n- Verify configuration',
		steps: [
			{
				id: 1,
				position: 1,
				title: 'Inspect first',
				instruction: '- Inspect it\n- Record the prompt',
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

test('troubleshooting shows compact guides and supports registry controls', async ({ page }) => {
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
	const firstGuide = page.getByRole('button', { name: 'Select guide First guide' });
	await expect(firstGuide.getByText('Critical', { exact: true })).toHaveCount(0);
	await expect(firstGuide).not.toContainText('First symptom');
	await expect(firstGuide.locator('.guide-labels > span')).toHaveText([
		'ERRCODE:1',
		'WIFI_TEST',
		'TX_VERIFY'
	]);
	const detailSeverity = page.locator('.detail-head .severity-badge');
	await expect(detailSeverity).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
	await expect(page.locator('.diagnostic-grid ol li')).toHaveText([
		'First symptom',
		'Confirm prompt state'
	]);
	await expect(page.locator('.diagnostic-grid ul li')).toHaveText([
		'First cause',
		'Verify configuration'
	]);
	await expect(page.locator('.step-card ul li')).toHaveText(['Inspect it', 'Record the prompt']);
	const alternativeSection = page.getByRole('region', { name: 'Alternative Solutions' });
	await expect(alternativeSection.getByText(/One error code or symptom/)).toContainText(
		'One error code or symptom can have multiple causes. Choose another solution below if needed.'
	);
	const titleBox = await alternativeSection.locator('.alternative-solutions-title').boundingBox();
	const subtitleBox = await alternativeSection
		.locator('.alternative-solutions-title')
		.locator('xpath=../following-sibling::p')
		.boundingBox();
	expect(titleBox).not.toBeNull();
	expect(subtitleBox).not.toBeNull();
	expect(subtitleBox!.y).toBeGreaterThan(titleBox!.y + titleBox!.height);
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

	const resolution = page.getByRole('region', { name: 'Step-by-Step Resolution Guide' });
	await expect(resolution.getByRole('checkbox')).toHaveCount(0);
	await page.getByRole('button', { name: 'Select guide Second guide' }).click();
	await page.getByRole('button', { name: 'Select guide First guide' }).click();
	await expect(resolution.getByRole('heading', { name: 'Inspect first' })).toBeVisible();

	const listRefreshes = refreshes.list;
	const detailRefreshes = refreshes.detail;
	await page.getByRole('button', { name: 'Manage Lists' }).click();
	const manager = page.getByRole('dialog', { name: 'Manage Troubleshooting Lists' });
	await expect(manager).toBeVisible();
	await assertDialogChromeIsSeparated(manager, true);
	await manager.screenshot({
		path: test.info().outputPath('troubleshooting-manager.png')
	});
	await manager
		.getByRole('textbox', { name: 'Measurement & Criteria label' })
		.fill('Renamed Measurement');
	await manager.getByRole('button', { name: 'Save Measurement & Criteria' }).click();
	await expect(page.getByText('Renamed Measurement', { exact: true }).first()).toBeVisible();
	await expect.poll(() => refreshes.list).toBeGreaterThan(listRefreshes);
	await expect.poll(() => refreshes.detail).toBeGreaterThan(detailRefreshes);
	await manager.getByRole('button', { name: 'Close dialog' }).click();

	await page.getByRole('button', { name: 'Add Alternate Solution', exact: true }).click();
	const alternate = page.getByRole('dialog', {
		name: 'Add Alternate Solution for Existing Problem / Error Code'
	});
	await expect(alternate).toBeVisible();
	await assertDialogChromeIsSeparated(alternate);
	await alternate.getByRole('button', { name: 'Close dialog' }).click();

	await page.getByRole('button', { name: 'New Troubleshooting Guide' }).click();
	const editor = page.getByRole('dialog', { name: 'New Troubleshooting Guide' });
	await expect(editor.getByText('Testing Stage / Category', { exact: true })).toBeVisible();
	await expect(editor.getByText('Severity', { exact: true })).toBeVisible();
	await expect(editor.getByText('Device / Interface Scope', { exact: true })).toBeVisible();
	await editor.getByRole('textbox', { name: /Observed Symptom/ }).scrollIntoViewIfNeeded();
	await editor.screenshot({ path: test.info().outputPath('troubleshooting-new-guide.png') });
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

test('creates, edits, and deletes a troubleshooting guide', async ({ page }) => {
	await mockTroubleshooting(page);
	await page.goto('/troubleshooting');

	await page.getByRole('button', { name: 'New Troubleshooting Guide' }).click();
	const createDialog = page.getByRole('dialog', { name: 'New Troubleshooting Guide' });
	await createDialog
		.getByRole('textbox', { name: /Solution \/ Manifestation Title/ })
		.fill('CRUD fixture');
	const symptom = createDialog.getByRole('textbox', { name: /Observed Symptom/ });
	await symptom.pressSequentially('1. Fixture symptom');
	await symptom.press('Enter');
	await symptom.pressSequentially('Fixture log');
	await expect(symptom).toHaveValue('1. Fixture symptom\n2. Fixture log');
	await createDialog.getByRole('textbox', { name: /Action Headline/ }).fill('Inspect fixture');
	await createDialog
		.getByRole('textbox', { name: /Detailed Instruction/ })
		.fill('Run fixture check');
	const verification = createDialog.getByRole('textbox', { name: /Expected Verification Result/ });
	await verification.pressSequentially('- Passes');
	await verification.press('Enter');
	await verification.pressSequentially('Is stable');
	await createDialog.getByRole('button', { name: 'Create Guide' }).click();
	await expect(page.getByRole('heading', { name: 'CRUD fixture' })).toBeVisible();
	await expect(page.locator('.diagnostic-grid ol li')).toHaveText([
		'Fixture symptom',
		'Fixture log'
	]);
	await expect(page.locator('.step-card ul li')).toHaveText(['Passes', 'Is stable']);

	await page.getByRole('button', { name: 'Edit', exact: true }).click();
	const editDialog = page.getByRole('dialog', { name: 'Edit Troubleshooting Guide' });
	await editDialog
		.getByRole('textbox', { name: /Solution \/ Manifestation Title/ })
		.fill('CRUD fixture updated');
	await editDialog.getByRole('button', { name: 'Save Changes' }).click();
	await expect(page.getByRole('heading', { name: 'CRUD fixture updated' })).toBeVisible();

	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	const deleteDialog = page.getByRole('alertdialog', { name: 'Delete Troubleshooting Guide?' });
	await deleteDialog.getByRole('button', { name: 'Delete Guide', exact: true }).click();
	await expect(page.getByRole('button', { name: 'Select guide CRUD fixture updated' })).toHaveCount(
		0
	);
});

test('metadata refresh failure preserves the current guide list and detail', async ({ page }) => {
	await mockTroubleshooting(page, { failOneCategoryRefresh: true });
	await page.goto('/troubleshooting');
	await expect(page.getByRole('button', { name: 'Select guide First guide' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'First guide' })).toBeVisible();

	const listRefreshes = refreshes.list;
	const detailRefreshes = refreshes.detail;
	await page.getByRole('button', { name: 'Manage Lists' }).click();
	const manager = page.getByRole('dialog', { name: 'Manage Troubleshooting Lists' });
	await manager
		.getByRole('textbox', { name: 'Measurement & Criteria label' })
		.fill('Refresh Draft');
	await manager.getByRole('button', { name: 'Save Measurement & Criteria' }).click();

	await expect(page.getByRole('alert')).toContainText('Metadata refresh failed');
	await expect(page.getByRole('button', { name: 'Select guide First guide' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'First guide' })).toBeVisible();
	expect(refreshes.list).toBe(listRefreshes);
	expect(refreshes.detail).toBe(detailRefreshes);

	await page.getByRole('button', { name: 'Close' }).click();
	await page.getByRole('button', { name: 'Retry metadata' }).click();
	await expect.poll(() => refreshes.list).toBeGreaterThan(listRefreshes);
	await expect.poll(() => refreshes.detail).toBeGreaterThan(detailRefreshes);
});
