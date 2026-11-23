import { test, expect, type Locator, type Page } from '@playwright/test';
import type { DefinitionKind, DefinitionRow } from '../../src/lib/definitions';
import type { DeviceSummary } from '../../src/lib/devices';
import type { OrphanExampleRow, OrphanLinkRow, OrphanReport } from '../../src/lib/orphans';
import type { ImportResult, ImportRun, ImportNormalization } from '../../src/lib/imports';
import type { Readiness } from '../../src/lib/mcp';
import type {
	SequenceDetail,
	SequenceInput,
	SequenceStep,
	SequenceStepInput
} from '../../src/lib/sequences';

test('all production routes share the reference navigation and theme', async ({ page }) => {
	test.setTimeout(120_000);
	await page.route('**/api/**', (route) => route.fulfill({ status: 503, json: {} }));
	const paths = [
		'/devices',
		'/functions',
		'/parameter-definitions',
		'/criteria-definitions',
		'/examples',
		'/sequences',
		'/orphans',
		'/import'
	];
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: 900 });
		for (const path of paths) {
			await page.goto(path);
			await expect(page.locator('header.app-chrome'), `${path} at ${width}px`).toHaveClass(
				/reference-chrome/,
				{ timeout: 15_000 }
			);
			await expect(page.locator('main')).toHaveAttribute('data-reference-theme', '');
			await expect(page.locator('main')).toHaveCSS('max-width', '1504px');
			if (width < 1200) {
				await expect(page.getByRole('button', { name: 'Open navigation menu' })).toBeVisible();
			} else {
				await expect(page.getByRole('navigation', { name: 'Main navigation' })).toBeVisible();
				await expect(page.getByRole('heading', { level: 1 })).toHaveCSS('font-size', '30px');
			}
		}
	}
});

test('shared shell aligns desktop navigation and keeps its brand visible on mobile', async ({
	page
}) => {
	await page.route('**/api/**', (route) => route.fulfill({ status: 503, json: {} }));
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices');
	const centers = await page.evaluate(() => {
		const brand = document.querySelector('.brand')!.getBoundingClientRect();
		const nav = document.querySelector('.reference-chrome nav')!;
		const navBox = nav.getBoundingClientRect();
		const activeLink = nav.querySelector('[aria-current="page"]')!.getBoundingClientRect();
		return {
			brandCenter: brand.top + brand.height / 2,
			navCenter: navBox.top + navBox.height / 2,
			underlineGap: activeLink.bottom - brand.bottom
		};
	});
	expect(Math.abs(centers.brandCenter - centers.navCenter)).toBeLessThan(2);
	expect(centers.underlineGap).toBeLessThan(3);

	await page.setViewportSize({ width: 390, height: 844 });
	await expect(page.locator('.brand-name')).toBeVisible();
	await expect(page.locator('.app-footer .footer-brand')).toBeVisible();
	expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
});

test('Devices and Functions follow their selected prototype catalog layouts', async ({ page }) => {
	test.setTimeout(60_000);
	await page.route('**/api/**', (route) => route.fulfill({ status: 503, json: {} }));
	await page.goto('/devices');
	await expect(page.locator('.directory-heading .eyebrow')).toHaveCount(0);
	await expect(page.getByRole('heading', { name: 'Devices', exact: true })).toBeVisible();
	await page.goto('/functions');
	await expect(page.locator('.directory-heading .eyebrow')).toHaveCount(0);
	await expect(
		page.getByRole('heading', { level: 1, name: 'Functions', exact: true })
	).toBeVisible();
	await expect(page.locator('.filters .filter-field > span')).toHaveText([
		'Source',
		'Device / Interface',
		'Category',
		'Execution role',
		'Execution phase'
	]);
	await expect(page.locator('.functions-workspace')).toHaveCSS('border-radius', '6px');
});

test('Device directory search aligns with the sorting controls', async ({ page }) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices');
	const search = page.getByRole('textbox', { name: 'Search devices' });
	const sort = page.getByRole('button', { name: 'Sort devices' });
	const order = page.getByRole('button', { name: 'Sort order' });
	await expect(search).toBeVisible();
	const boxes = await Promise.all([search.boundingBox(), sort.boundingBox(), order.boundingBox()]);
	expect(boxes.every((box) => box !== null)).toBe(true);
	for (const box of boxes) expect(box!.height).toBe(40);
	for (const box of boxes.slice(1)) {
		expect(Math.abs(box!.y + box!.height - (boxes[0]!.y + boxes[0]!.height))).toBeLessThan(1);
	}
});

test('Device directory shows four cards at desktop width and keeps the shared shell full-width', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1104, height: 900 });
	await page.goto('/devices');

	const grid = page.locator('.device-grid').first();
	await expect(grid).toBeVisible();
	expect(
		await grid.evaluate(
			(element) => getComputedStyle(element).gridTemplateColumns.split(' ').length
		)
	).toBe(4);
	await expect(page.getByRole('region', { name: 'Global definitions' })).toBeVisible();
	await expect(page.getByRole('contentinfo')).toBeVisible();
	await expect(page.locator('header.app-chrome')).toHaveCSS('position', 'sticky');
	const clientWidth = await page.locator('html').evaluate((element) => element.clientWidth);
	await expect(page.locator('header.app-chrome')).toHaveCSS('width', `${clientWidth}px`);
	await expect(page.locator('main')).toHaveCSS('width', `${clientWidth}px`);
});

test('Device workspace exposes section counts and adapts its list-detail layout', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices/WIFI_TEST');

	for (const [section, count] of [
		['Functions', '2'],
		['Parameters', '1'],
		['Criteria', '1'],
		['Examples', '1']
	]) {
		const sectionTab = page.getByRole('tab', { name: section, exact: true });
		await expect(sectionTab).toBeVisible();
		await expect(sectionTab.locator('.section-count')).toHaveText(count);
	}

	const browser = page.locator('.function-browser');
	expect(
		await browser.evaluate(
			(element) => getComputedStyle(element).gridTemplateColumns.split(' ').length
		)
	).toBe(2);
	const headerHeight = await page
		.locator('header.app-chrome')
		.evaluate((element) => element.getBoundingClientRect().height);
	const railTop = Number.parseFloat(
		await page
			.getByRole('complementary', { name: 'Device functions' })
			.evaluate((element) => getComputedStyle(element).top)
	);
	expect(railTop).toBeGreaterThan(headerHeight);

	await page.setViewportSize({ width: 390, height: 844 });
	expect(
		await browser.evaluate(
			(element) => getComputedStyle(element).gridTemplateColumns.split(' ').length
		)
	).toBe(1);
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth <= document.documentElement.clientWidth
		)
	).toBe(true);
});

test('standalone Function header uses a full-width searchable switcher', async ({ page }) => {
	await mockFunctionCatalog(page);
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/functions/1');
	const detail = page.getByRole('region', { name: 'Function details' });
	await expect(detail).toBeVisible();
	await expect(detail.locator('.ownership')).toHaveCount(0);
	await expect(detail).toHaveCSS('padding-left', '0px');
	await page.screenshot({
		path: test.info().outputPath('function-heading-desktop.png'),
		fullPage: true
	});
	await page.setViewportSize({ width: 390, height: 844 });
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth > document.documentElement.clientWidth
		)
	).toBe(false);
	await page.screenshot({
		path: test.info().outputPath('function-heading-mobile.png'),
		fullPage: true
	});
	await detail.getByRole('button', { name: 'Switch function' }).click();
	const popup = page.getByRole('dialog', { name: 'Choose Switch function' });
	await popup.getByRole('combobox', { name: 'Search options' }).fill('RX_PER');
	await popup.getByRole('option', { name: /RX_PER/ }).click();
	await expect(page).toHaveURL(/\/functions\/2$/);
	await expect(page.getByRole('heading', { level: 1 })).toHaveText('RX_PER');
	expect(errors).toEqual([]);
});

test('Device field links open and highlight definitions beyond the first page', async ({
	page
}) => {
	await mockDeviceHub(page);
	const rows = Array.from({ length: 30 }, (_, index) => ({
		id: index === 29 ? 11 : index + 100,
		name: index === 29 ? 'ANT' : `FIELD_${index}`,
		device_name: 'WIFI_TEST',
		type: 'integer',
		unit: '',
		description: '',
		default_value: '1',
		allowed_values: '1,2',
		cardinality: 'single',
		value_pattern: '',
		required: true
	}));
	await page.route('**/api/devices/WIFI_TEST/parameters?*', (route) => {
		const query = new URL(route.request().url()).searchParams;
		const number = Number(query.get('page') ?? 1);
		const size = Number(query.get('page_size') ?? 20);
		return route.fulfill({
			json: {
				items: rows.slice((number - 1) * size, number * size),
				total: rows.length,
				global_items: [],
				global_total: 0,
				page: number,
				page_size: size
			}
		});
	});
	await page.goto('/devices/WIFI_TEST');
	const fields = page.getByRole('region', { name: 'WIFI_TEST parameters', exact: true });
	await fields.getByRole('link', { name: 'ANT', exact: true }).click();
	await expect(page).toHaveURL(/tab=parameters/);
	await expect(page).toHaveURL(/definition=11/);
	await expect(
		page
			.getByRole('region', { name: 'Definition details' })
			.getByRole('heading', { name: 'ANT', exact: true })
	).toBeVisible();
	const panel = page.getByRole('tabpanel', { name: /^Parameters/ });
	await expect(panel.locator('tr[data-selected="true"]')).toContainText('ANT');
	await expect(panel.locator('[data-pagination-root]')).toContainText('Page 2 of 2');
	await page.goBack();
	await expect(page.getByRole('tab', { name: /^Functions/ })).toHaveAttribute(
		'aria-selected',
		'true'
	);
});

test('Device field table headers sort and criteria links retain Device context', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST');
	const fields = page.getByRole('region', { name: 'WIFI_TEST parameters', exact: true });
	for (const label of ['Name', 'Type', 'Default', 'Allowed values', 'Required']) {
		const header = fields.getByRole('columnheader', { name: new RegExp(`^${label}`) });
		await header.getByRole('button').click();
		await expect(header).toHaveAttribute('aria-sort', 'ascending');
		await header.getByRole('button').click();
		await expect(header).toHaveAttribute('aria-sort', 'descending');
	}
	await fields.getByRole('columnheader', { name: /^Name/ }).getByRole('button').click();
	await fields.getByRole('columnheader', { name: /^Name/ }).getByRole('button').click();
	await expect(fields.locator('tbody tr').first()).toContainText('UNRESOLVED_FIELD');
	await fields.getByRole('columnheader', { name: /^Name/ }).getByRole('button').click();
	await expect(fields.getByRole('columnheader', { name: /^Name/ })).not.toHaveAttribute(
		'aria-sort'
	);
	await expect(fields.locator('tbody tr').first()).toContainText('ANT');
	await page
		.getByRole('region', { name: 'WIFI_TEST criteria', exact: true })
		.getByRole('link', { name: 'POWER', exact: true })
		.click();
	await expect(page).toHaveURL(/tab=criteria/);
	await expect(page).toHaveURL(/definition=23/);
	await expect(
		page.getByRole('tabpanel', { name: /^Criteria/ }).locator('tr[data-selected="true"]')
	).toContainText('POWER');
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Device tab toolbars align and their pagination stays pinned', async ({ page }) => {
	await mockDeviceHub(page);
	for (const kind of ['parameters', 'criteria']) {
		await page.route(`**/api/devices/WIFI_TEST/${kind}?*`, (route) =>
			route.fulfill({
				json: {
					items: Array.from({ length: 20 }, (_, index) => ({
						id: index === 0 ? (kind === 'parameters' ? 11 : 23) : index + 100,
						name: `FIELD_${index}`,
						device_name: 'WIFI_TEST',
						type: 'integer',
						unit: '',
						description: '',
						default_value: '1',
						allowed_values: '1,2',
						cardinality: 'single',
						value_pattern: '',
						required: true
					})),
					total: 40,
					global_items: [],
					global_total: 0,
					page: 1,
					page_size: 20
				}
			})
		);
	}
	await page.route('**/api/examples?*', (route) =>
		route.fulfill({
			json: {
				items: Array.from({ length: 20 }, (_, index) => ({
					id: index + 50,
					function_source: 'DLL',
					function_name: `EXAMPLE_${index}`,
					device_name: 'WIFI_TEST',
					script: 'DLL|TEST|ANT=1|POWER=10|1000',
					meaning: `Example ${index}`,
					function_exists: true,
					created_at: '',
					updated_at: ''
				})),
				total: 40,
				page: 1,
				page_size: 20
			}
		})
	);
	await page.setViewportSize({ width: 1440, height: 900 });
	for (const kind of ['parameters', 'criteria', 'examples']) {
		await page.goto(`/devices/WIFI_TEST?tab=${kind}`);
		const panel = page.getByRole('tabpanel', { name: new RegExp(`^${kind}`, 'i') });
		const search = panel.locator('.section-search input');
		const add = panel.locator('.section-actions > button');
		await expect(search).toBeVisible();
		const searchBox = (await search.boundingBox())!;
		const addBox = (await add.boundingBox())!;
		expect(Math.abs(searchBox.y - addBox.y)).toBeLessThan(2);
		expect(searchBox.height).toBe(addBox.height);
		const pager = panel.locator('.device-pagination');
		await expect(pager).toBeVisible();
		const before = (await pager.boundingBox())!;
		await panel
			.locator(kind === 'examples' ? '.examples-table-scroll' : '.definition-tables')
			.evaluate((element) => {
				element.scrollTop = element.scrollHeight;
			});
		const after = (await pager.boundingBox())!;
		expect(Math.abs(before.y - after.y)).toBeLessThan(2);
		await expect(page.locator('.banner-error')).toHaveCount(0);
		await page.screenshot({
			path: test.info().outputPath(`device-${kind}-desktop.png`),
			fullPage: true
		});
		await page.setViewportSize({ width: 390, height: 844 });
		expect(
			await page.evaluate(
				() => document.documentElement.scrollWidth > document.documentElement.clientWidth
			)
		).toBe(false);
		await page.screenshot({
			path: test.info().outputPath(`device-${kind}-mobile.png`),
			fullPage: true
		});
		await page.setViewportSize({ width: 1440, height: 900 });
	}
});

test('MCP readiness data headers cycle through ascending descending and default', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.route(/\/api\/functions\/1\/mcp-readiness(?:\?.*)?$/, (route) =>
		route.fulfill({
			json: {
				function_id: 1,
				device: 'WIFI_TEST',
				status: 'not_ready',
				grammar: null,
				line: { ready: false, issues: [] },
				bundle: { ready: false, issues: [] },
				warnings: [],
				criteria: [],
				parameters: ['ZETA', 'ALPHA'].map((name, index) => ({
					name,
					device_name: 'WIFI_TEST',
					required: true,
					has_definition: index === 1,
					script_key_scope: index === 1 ? 'bundle' : ''
				}))
			}
		})
	);
	await page.goto('/functions/1');
	await page.getByRole('tab', { name: 'MCP readiness', exact: true }).click();
	const table = page.getByRole('region', { name: 'MCP readiness', exact: true }).locator('table');
	await expect(table.locator('tbody tr')).toHaveCount(2);
	for (const label of ['Key', 'Definition', 'Script key audit']) {
		const header = table.getByRole('columnheader', { name: new RegExp(`^${label}`) });
		await header.getByRole('button').click();
		await expect(header).toHaveAttribute('aria-sort', 'ascending');
		if (label === 'Key') await expect(table.locator('tbody tr').first()).toContainText('ALPHA');
		await header.getByRole('button').click();
		await expect(header).toHaveAttribute('aria-sort', 'descending');
		await header.getByRole('button').click();
		await expect(header).not.toHaveAttribute('aria-sort');
		await expect(table.locator('tbody tr').first()).toContainText('ZETA');
	}
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('pagination page-size popup preserves document scroll and focus', async ({ page }) => {
	await mockFunctionCatalog(page);
	await page.setViewportSize({ width: 1440, height: 500 });
	await page.goto('/functions');
	const trigger = page
		.getByRole('complementary', { name: 'Functions list' })
		.getByRole('button', { name: 'Page size', exact: true });
	await trigger.scrollIntoViewIfNeeded();
	const before = await page.evaluate(() => window.scrollY);
	expect(before).toBeGreaterThan(0);
	await trigger.click();
	const popup = page.getByRole('dialog', { name: 'Choose Page size', exact: true });
	await expect(popup).toBeVisible();
	await expect(popup.getByRole('combobox', { name: 'Search options' })).toBeFocused();
	expect(await page.evaluate(() => window.scrollY)).toBe(before);
	await popup.getByRole('option', { name: '50 / page', exact: true }).click();
	await expect(trigger).toHaveText('50 / page');
	expect(await page.evaluate(() => window.scrollY)).toBe(before);
});

test('dictionary page-size selection preserves document position', async ({ page }) => {
	await mockDictionary(page, 'criteria-definitions');
	await page.setViewportSize({ width: 1440, height: 500 });
	await page.goto('/criteria-definitions');
	const trigger = page
		.getByRole('complementary', { name: 'Definitions list' })
		.getByRole('button', { name: 'Page size', exact: true });
	await trigger.scrollIntoViewIfNeeded();
	const before = await page.evaluate(() => window.scrollY);
	await trigger.click();
	expect(await page.evaluate(() => window.scrollY)).toBe(before);
	await page
		.getByRole('dialog', { name: 'Choose Page size', exact: true })
		.getByRole('option', { name: '50 / page', exact: true })
		.click();
	await expect(trigger).toHaveText('50 / page');
	expect(await page.evaluate(() => window.scrollY)).toBe(before);
});

test('Examples main search and function dropdown cover complete example identities', async ({
	page
}) => {
	await mockDeviceHub(page);
	const rows = Array.from({ length: 101 }, (_, index) => ({
		id: index + 1,
		function_source: index === 100 ? 'BUILT-IN' : 'DLL',
		function_name: index === 100 ? 'LATE_ORPHAN' : 'TX_MATCH',
		device_name: index === 100 ? 'GPIB' : 'WIFI_TEST',
		script: 'COMMAND|VALUE|1',
		meaning: `Example ${index}`,
		function_exists: index !== 100,
		created_at: '2026-10-07',
		updated_at: '2026-10-07'
	}));
	await page.route('**/api/examples/function-names', (route) =>
		route.fulfill({ json: { items: ['TX_MATCH', 'LATE_ORPHAN'] } })
	);
	const requests: URLSearchParams[] = [];
	await page.route('**/api/examples?*', (route) => {
		const query = new URL(route.request().url()).searchParams;
		requests.push(query);
		const matches = rows.filter(
			(row) =>
				(!query.get('function_name') || row.function_name === query.get('function_name')) &&
				`${row.function_source} ${row.function_name} ${row.device_name} ${row.script} ${row.meaning}`
					.toLowerCase()
					.includes((query.get('search') ?? '').toLowerCase())
		);
		const number = Number(query.get('page') ?? 1),
			size = Number(query.get('page_size') ?? 20);
		return route.fulfill({
			json: {
				items: matches.slice((number - 1) * size, number * size),
				total: matches.length,
				page: number,
				page_size: size
			}
		});
	});
	await page.goto('/examples?source=DLL');
	await expect(
		page.getByRole('textbox', { name: 'Function source filter', exact: true })
	).toHaveCount(0);
	const search = page.getByRole('textbox', { name: 'Search examples', exact: true });
	await expect(search).toHaveCSS('height', '40px');
	await search.fill('GPIB');
	await expect(page).toHaveURL(/search=GPIB/);
	await expect(page).not.toHaveURL(/source=/);
	await expect(
		page.getByRole('button', { name: 'Select example of LATE_ORPHAN (GPIB)', exact: true })
	).toBeVisible();
	await search.fill('');
	await expect(page).not.toHaveURL(/search=/);
	const filter = page.getByRole('button', { name: 'Function name filter', exact: true });
	await expect(filter).toBeEnabled();
	await expect(filter).toHaveCSS('height', '40px');
	await filter.click();
	const popup = page.getByRole('dialog', { name: 'Choose Function name filter', exact: true });
	await popup.getByRole('combobox', { name: 'Search options' }).fill('LATE_ORPHAN');
	await popup.getByRole('option', { name: 'LATE_ORPHAN', exact: true }).click();
	await expect(page).toHaveURL(/function_name=LATE_ORPHAN/);
	expect(requests.every((query) => !query.has('function_source'))).toBe(true);
	await page.setViewportSize({ width: 1440, height: 500 });
	const pager = page
		.getByRole('complementary', { name: 'Examples list' })
		.getByRole('button', { name: 'Page size', exact: true });
	await pager.scrollIntoViewIfNeeded();
	const scrollBefore = await page.evaluate(() => window.scrollY);
	await pager.click();
	await page
		.getByRole('dialog', { name: 'Choose Page size', exact: true })
		.getByRole('option', { name: '50 / page', exact: true })
		.click();
	await expect(pager).toHaveText('50 / page');
	expect(await page.evaluate(() => window.scrollY)).toBe(scrollBefore);
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: 844 });
		expect(
			await page.evaluate(
				() => document.documentElement.scrollWidth > document.documentElement.clientWidth
			)
		).toBe(false);
		await page.screenshot({
			path: test.info().outputPath(`examples-filters-${width}.png`),
			fullPage: true
		});
	}
});

test('CRUD dialogs provide multiline fields and preserve saved line breaks', async ({ page }) => {
	const description = 'First line of the description.\nSecond line with additional context.';
	const { mutations } = await mockFunctionCatalog(page);
	await page.setViewportSize({ width: 390, height: 844 });
	await page.goto('/functions');
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	let dialog = page.getByRole('dialog', { name: 'New Function', exact: true });
	await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('MULTILINE_FUNCTION');
	await expect(dialog.getByRole('textbox', { name: 'Description', exact: true })).toHaveJSProperty(
		'tagName',
		'TEXTAREA'
	);
	await dialog.getByRole('textbox', { name: 'Description', exact: true }).fill(description);
	expect(await dialog.evaluate((element) => element.scrollWidth > element.clientWidth)).toBe(false);
	await page.screenshot({
		path: test.info().outputPath('function-dialog-mobile.png'),
		fullPage: true
	});
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	expect(mutations.at(-1)?.body?.description).toBe(description);
	for (const kind of ['parameter-definitions', 'criteria-definitions'] as const) {
		const store = await mockDictionary(page, kind);
		await page.goto(`/${kind}`);
		await page.getByRole('button', { name: 'New', exact: true }).click();
		dialog = page.getByRole('dialog');
		await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('MULTILINE_FIELD');
		await expect(
			dialog.getByRole('textbox', { name: 'Description', exact: true })
		).toHaveJSProperty('tagName', 'TEXTAREA');
		await dialog.getByRole('textbox', { name: 'Description', exact: true }).fill(description);
		await expect(
			dialog.getByRole('textbox', { name: 'Value pattern', exact: true })
		).toHaveJSProperty('tagName', 'TEXTAREA');
		await page.screenshot({
			path: test.info().outputPath(`${kind}-dialog-mobile.png`),
			fullPage: true
		});
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(dialog).toHaveCount(0);
		expect(store.mutations.at(-1)?.body?.description).toBe(description);
	}
});

test('catalog workspace columns and pagination meet the bottom edge', async ({ page }) => {
	await mockFunctionCatalog(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/functions');
	await expect(page.getByRole('region', { name: 'Function details' })).toBeVisible();
	await expect(page.locator('.banner-error')).toHaveCount(0);
	const geometry = await page.locator('.functions-workspace').evaluate((workspace) => {
		const rail = workspace.querySelector('aside')!;
		const pager = rail.querySelector('[data-pagination-root]')!;
		const frame = workspace.getBoundingClientRect();
		return {
			railGap: frame.bottom - rail.getBoundingClientRect().bottom,
			pagerGap: frame.bottom - pager.getBoundingClientRect().bottom
		};
	});
	expect(Math.abs(geometry.railGap)).toBeLessThan(2);
	expect(Math.abs(geometry.pagerGap)).toBeLessThan(2);
});

test('catalog tab toolbars align search and add controls responsively', async ({ page }) => {
	await mockFunctionCatalog(page);
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: 900 });
		await page.goto('/functions');
		for (const [tab, searchName, actionName] of [
			['Parameters', 'Search parameter links', 'Attach parameter'],
			['Criteria', 'Search criteria links', 'Attach criteria'],
			['Examples', 'Search examples', 'New example']
		]) {
			await page.getByRole('tab', { name: tab, exact: true }).click();
			const panel = page.getByRole('tabpanel', { name: tab, exact: true });
			const search = panel.getByRole('textbox', { name: searchName, exact: true });
			const action = panel.getByRole('button', { name: actionName, exact: true });
			await expect(search).toBeVisible();
			await expect(action).toBeVisible();
			const searchBox = (await search.boundingBox())!;
			const actionBox = (await action.boundingBox())!;
			expect(actionBox.height).toBe(searchBox.height);
			if (width > 850) {
				expect(Math.abs(searchBox.y - actionBox.y)).toBeLessThan(2);
				expect(actionBox.x).toBeGreaterThanOrEqual(searchBox.x + searchBox.width);
			}
			expect(
				await page.evaluate(
					() => document.documentElement.scrollWidth > document.documentElement.clientWidth
				)
			).toBe(false);
			await expect(page.locator('.banner-error')).toHaveCount(0);
			await page.screenshot({
				path: test.info().outputPath(`toolbar-${tab}-${width}.png`),
				fullPage: true
			});
		}
	}
});

test('Functions catalog controls and pagination stay visible while its rows scroll', async ({
	page
}) => {
	test.setTimeout(60_000);
	const { functions } = await mockFunctionCatalog(page);
	functions.push(
		...Array.from({ length: 18 }, (_, index) => ({
			...functions[0],
			id: index + 3,
			name: `SCROLL_FUNCTION_${index}`
		}))
	);
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: 900 });
		await page.goto('/functions');
		const rail = page.getByRole('complementary', { name: 'Functions list' });
		const search = rail.getByRole('textbox', { name: 'Search functions' });
		const sorting = rail.getByRole('button', { name: 'Sort functions' });
		const pagination = rail.getByRole('button', { name: 'Page size' });
		await expect(rail.getByRole('button', { name: 'Select SCROLL_FUNCTION_17' })).toBeAttached();
		const before = await Promise.all([
			search.boundingBox(),
			sorting.boundingBox(),
			pagination.boundingBox()
		]);
		await rail.evaluate((element) => {
			element.scrollTop = element.scrollHeight;
		});
		const after = await Promise.all([
			search.boundingBox(),
			sorting.boundingBox(),
			pagination.boundingBox()
		]);
		expect(await rail.evaluate((element) => element.scrollTop)).toBeGreaterThan(0);
		for (let index = 0; index < before.length; index++) {
			expect(Math.abs(before[index]!.y - after[index]!.y)).toBeLessThan(2);
		}
		const railBox = await rail.boundingBox();
		for (const box of after) {
			expect(box!.y).toBeGreaterThanOrEqual(railBox!.y);
			expect(box!.y + box!.height).toBeLessThanOrEqual(railBox!.y + railBox!.height);
		}
		await pagination.click();
		await expect(page.getByRole('option', { name: '50 / page' })).toBeVisible();
		await page.keyboard.press('Escape');
	}
});

test('Functions detail uses the prototype reference hierarchy with live actions', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/functions?function=1');
	const detail = page.getByRole('region', { name: 'Function details' });
	await expect(detail.getByRole('heading', { name: 'TX_MULTI_VERIFICATION' })).toBeVisible();
	await expect(detail.locator('.ownership')).toHaveText('Function reference / WIFI_TEST');
	await expect(detail.getByRole('heading', { name: 'TX_MULTI_VERIFICATION' })).toHaveCSS(
		'font-size',
		'20px'
	);
	await expect(detail.locator('.description')).toHaveCSS('font-size', '13px');
	await expect(detail.getByRole('button', { name: 'Edit function' })).toBeVisible();
	await expect(detail.getByRole('button', { name: 'Compose script' })).toBeVisible();
	await expect(detail.getByRole('link', { name: 'Manage function' })).toHaveAttribute(
		'href',
		'/functions/1'
	);
	await expect(detail.getByRole('tab', { name: 'Parameters' }).locator('.tab-count')).toHaveText(
		'7'
	);
	await expect(detail.getByRole('tab', { name: 'Criteria' }).locator('.tab-count')).toHaveText('2');
	await expect(detail.getByText('Parameter links', { exact: true })).toBeVisible();
	await expect(detail.getByText('7 linked definitions', { exact: true })).toBeVisible();
	await detail.getByRole('tab', { name: 'Examples' }).click();
	const examples = detail.getByRole('tabpanel', { name: 'Examples' });
	await expect(examples.getByText('Examples', { exact: true })).toBeVisible();
	await expect(examples.getByText('1 example', { exact: true })).toBeVisible();
	await expect(examples.getByRole('textbox', { name: 'Search examples' })).toBeVisible();
	await expect(examples.getByRole('button', { name: 'New example' })).toBeVisible();
});

test('Functions detail refreshes the tab count after detaching a link', async ({ page }) => {
	await mockFunctionCatalog(page);
	let attached = true;
	const row = {
		id: 101,
		name: 'ANT',
		device_name: 'WIFI_TEST',
		position: 1,
		required: true,
		function_source: 'DLL',
		function_name: 'TX_MULTI_VERIFICATION',
		source_file: '',
		resolved: true,
		definition_id: 11,
		created_at: '',
		updated_at: ''
	};
	await page.route('**/api/functions/1/parameter-links', (route) =>
		route.fulfill({ json: { items: attached ? [row] : [] } })
	);
	await page.route('**/api/functions/1/parameter-links/101', (route) => {
		attached = false;
		return route.fulfill({ status: 204 });
	});
	await page.goto('/functions?function=1');
	const detail = page.getByRole('region', { name: 'Function details' });
	const count = detail.getByRole('tab', { name: 'Parameters' }).locator('.tab-count');
	await expect(count).toHaveText('1');
	await detail.getByRole('button', { name: 'Actions for ANT' }).click();
	await page.getByRole('menuitem', { name: 'Detach' }).click();
	await page
		.getByRole('alertdialog', { name: 'Detach Link' })
		.getByRole('button', { name: 'Detach' })
		.click();
	await expect(count).toHaveText('0');
	await page.getByRole('button', { name: 'Select RX_PER' }).click();
	await expect(
		page
			.getByRole('region', { name: 'Function details' })
			.getByRole('tab', { name: 'Parameters' })
			.locator('.tab-count')
	).toHaveText('7');
	await page.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION' }).click();
	await expect(
		page
			.getByRole('region', { name: 'Function details' })
			.getByRole('tab', { name: 'Parameters' })
			.locator('.tab-count')
	).toHaveText('0');
});

async function mockDeviceHub(page: Page, functionName = 'TX_MULTI_VERIFICATION') {
	await page.route('**/api/**', async (route) => {
		const url = new URL(route.request().url());
		const functionRow = {
			id: 1,
			name: functionName,
			source: 'DLL',
			interface: 'WIFI_TEST',
			description: 'Verify transmit power.',
			category: 'Transmit',
			convention: 1,
			execution_role: '',
			execution_phase: '',
			updated_at: '2026-10-06'
		};
		const receive = {
			...functionRow,
			id: 2,
			name: 'RX_PER',
			description: 'Verify packet error rate.'
		};
		const scope = url.pathname.includes('GOLDENKV') ? 'GOLDENKV' : 'WIFI_TEST';
		const definition = {
			id: scope === 'GOLDENKV' ? 13 : 11,
			name: 'ANT',
			device_name: scope,
			description: '',
			type: 'integer',
			unit: '',
			default_value: scope === 'GOLDENKV' ? '2' : '1',
			allowed_values: '1,2',
			cardinality: 'single',
			value_pattern: '',
			required: true
		};
		const globalDefinition = {
			...definition,
			id: 12,
			name: 'SETTLE_MS',
			device_name: '',
			default_value: '200'
		};
		const globalAnt = { ...definition, id: 15, device_name: '', default_value: '0' };
		const globalTimeout = { ...globalDefinition, id: 16, name: 'TIMEOUT', default_value: '5000' };
		const criterion = {
			...definition,
			id: scope === 'GOLDENKV' ? 24 : 23,
			name: 'POWER',
			default_value: scope === 'GOLDENKV' ? '20' : '10',
			allowed_values: '10,20'
		};
		const link = {
			function_source: 'DLL',
			function_name: functionName,
			position: 1,
			required: true,
			source_file: '',
			resolved: true,
			created_at: '',
			updated_at: ''
		};
		let json: unknown;
		if (url.pathname === '/api/devices')
			json = {
				items: [
					{
						device_name: 'WIFI_TEST',
						source: 'DLL',
						sources: ['DLL'],
						functions: 2,
						parameters: 1,
						criteria: 1,
						examples: 1
					},
					{
						device_name: 'GOLDENKV',
						source: 'BUILT-IN',
						sources: ['BUILT-IN'],
						functions: 2,
						parameters: 1,
						criteria: 1,
						examples: 1
					},
					{ device_name: '', functions: 2, parameters: 3, criteria: 0, examples: 1 },
					{
						device_name: 'GPIB',
						source: 'BUILT-IN',
						sources: ['BUILT-IN'],
						functions: 0,
						parameters: 0,
						criteria: 0,
						examples: 0
					},
					{
						device_name: 'OSP',
						source: null,
						sources: ['BUILT-IN', 'DLL'],
						functions: 1,
						parameters: 1,
						criteria: 1,
						examples: 1
					}
				],
				total: 5
			};
		else if (
			/^\/api\/(?:parameter|criteria)-definitions\/\d+$/.test(url.pathname) &&
			route.request().method() === 'GET'
		) {
			const canonical = [
				definition,
				{ ...definition, id: 13, device_name: 'GOLDENKV', default_value: '2' },
				globalDefinition,
				globalAnt,
				globalTimeout,
				criterion,
				{ ...criterion, id: 24, device_name: 'GOLDENKV', default_value: '20' }
			].find((row) => row.id === Number(url.pathname.split('/').pop()));
			if (!canonical)
				return route.fulfill({ status: 404, json: { error: 'Definition not found' } });
			json = {
				...canonical,
				usage_count: 2,
				orphan_usage_count: 0,
				created_at: '2026-10-06',
				updated_at: '2026-10-06'
			};
		} else if (url.pathname.endsWith('/functions'))
			json = {
				items: url.pathname.includes('GPIB') ? [] : [functionRow, receive],
				total: url.pathname.includes('GPIB') ? 0 : 2,
				page: 1,
				page_size: 100
			};
		else if (/^\/api\/functions\/\d+\/mcp-readiness$/.test(url.pathname)) {
			json = {
				function_id: Number(url.pathname.split('/')[3]),
				device: url.searchParams.get('device') ?? '',
				status: 'not_ready',
				grammar: null,
				line: { ready: false, issues: [] },
				bundle: { ready: false, issues: [] },
				parameters: [],
				criteria: [],
				warnings: []
			} satisfies Readiness;
		} else if (url.pathname.endsWith('/parameters'))
			json = {
				items: url.pathname.includes('/_global/')
					? [globalDefinition, globalAnt, globalTimeout]
					: [definition],
				total: url.pathname.includes('/_global/') ? 3 : 1,
				global_items: url.pathname.includes('/_global/') ? [] : [globalDefinition],
				global_total: url.pathname.includes('/_global/') ? 0 : 1,
				page: 1,
				page_size: 100
			};
		else if (url.pathname.endsWith('/criteria'))
			json = {
				items: url.pathname.includes('/_global/') ? [] : [criterion],
				total: url.pathname.includes('/_global/') ? 0 : 1,
				global_items: [],
				global_total: 0,
				page: 1,
				page_size: 100
			};
		else if (url.pathname.endsWith('/parameter-links'))
			json = {
				items: [
					{ ...link, id: 101, name: 'ANT', device_name: 'WIFI_TEST', definition_id: 11 },
					{ ...link, id: 102, name: 'ANT', device_name: 'GOLDENKV', definition_id: 13 },
					{
						...link,
						id: 103,
						name: 'SETTLE_MS',
						device_name: 'WIFI_TEST',
						definition_id: 12,
						position: 2
					},
					{ ...link, id: 104, name: 'ANT', device_name: 'OTHER_DEVICE', definition_id: 14 },
					{ ...link, id: 105, name: 'ANT', device_name: '', definition_id: 15 },
					{ ...link, id: 106, name: 'TIMEOUT', device_name: '', definition_id: 16, position: 3 },
					{
						...link,
						id: 107,
						name: 'UNRESOLVED_FIELD',
						device_name: 'WIFI_TEST',
						definition_id: null,
						resolved: false,
						position: 4
					}
				]
			};
		else if (url.pathname.endsWith('/criteria-links'))
			json = {
				items: [
					{ ...link, id: 201, name: 'POWER', device_name: 'WIFI_TEST', definition_id: 23 },
					{ ...link, id: 202, name: 'POWER', device_name: 'GOLDENKV', definition_id: 24 }
				]
			};
		else if (url.pathname.endsWith('/compose/fields'))
			json = {
				function: functionRow,
				parameters: [
					{ ...definition, device_name: 'WIFI_TEST', position: 1, resolved: true },
					{ ...globalDefinition, device_name: 'WIFI_TEST', position: 2, resolved: true },
					{ ...globalTimeout, position: 3, resolved: true }
				],
				criteria: [
					{
						...criterion,
						device_name: 'WIFI_TEST',
						default_value: '10',
						position: 1,
						resolved: true
					}
				]
			};
		else if (url.pathname === '/api/examples/function-names') json = { items: [functionName] };
		else if (url.pathname === '/api/examples') {
			const device = url.searchParams.get('device');
			json = {
				items: [
					{
						id: device ? 31 : 32,
						function_source: 'DLL',
						function_name: functionName,
						device_name: device ?? 'OTHER_DEVICE',
						script: device ? `TX|ANT=${device === 'GOLDENKV' ? '2' : '1'}` : 'GLOBAL_REFERENCE',
						meaning: device ? 'Device transmit example' : 'Shared transmit example',
						function_exists: true,
						created_at: '',
						updated_at: ''
					}
				],
				total: 1,
				page: 1,
				page_size: 20
			};
		} else {
			await route.fulfill({
				status: 500,
				json: { message: `Unexpected mocked endpoint: ${url.pathname}` }
			});
			return;
		}
		await route.fulfill({ json });
	});
}

async function mockFunctionCatalog(page: Page) {
	await mockDeviceHub(page);
	const functions = [
		{
			id: 1,
			name: 'TX_MULTI_VERIFICATION',
			source: 'DLL',
			interface: 'WIFI_TEST',
			description: 'Verify transmit power.',
			category: 'Transmit',
			convention: 1,
			execution_role: 'TEST',
			execution_phase: 'RUN',
			created_at: '2026-10-06',
			updated_at: '2026-10-06'
		},
		{
			id: 2,
			name: 'RX_PER',
			source: 'BUILT-IN',
			interface: 'GPIB',
			description: 'Verify packet error rate.',
			category: 'Receive',
			convention: null,
			execution_role: 'CHECK',
			execution_phase: 'VERIFY',
			created_at: '2026-10-06',
			updated_at: '2026-10-06'
		}
	];
	const requests: URLSearchParams[] = [];
	const mutations: { method: string; path: string; body: Record<string, unknown> | null }[] = [];
	await page.route('**/api/functions/facets', (route) =>
		route.fulfill({
			json: {
				sources: ['DLL', 'BUILT-IN'],
				interfaces: ['WIFI_TEST', 'GPIB'],
				categories: ['Transmit', 'Receive'],
				execution_roles: ['TEST', 'CHECK'],
				execution_phases: ['RUN', 'VERIFY']
			}
		})
	);
	await page.route(/\/api\/functions(?:\?.*)?$/, (route) => {
		if (route.request().method() === 'POST') {
			const body = route.request().postDataJSON();
			if (!body.name)
				return route.fulfill({
					status: 422,
					json: { error: 'Invalid function', fields: { name: 'Name is required.' } }
				});
			mutations.push({ method: 'POST', path: '/api/functions', body });
			const row = {
				...functions[0],
				...body,
				id: Math.max(...functions.map((fn) => fn.id), 2) + 1
			};
			functions.push(row);
			return route.fulfill({ status: 201, json: row });
		}
		const query = new URL(route.request().url()).searchParams;
		requests.push(query);
		const matches = functions.filter(
			(fn) =>
				(!query.get('search') ||
					`${fn.name} ${fn.description}`
						.toLowerCase()
						.includes(query.get('search')!.toLowerCase())) &&
				['source', 'interface', 'category', 'execution_role', 'execution_phase'].every(
					(key) => !query.get(key) || fn[key as keyof typeof fn] === query.get(key)
				)
		);
		const number = Number(query.get('page') ?? 1);
		const size = Number(query.get('page_size') ?? 20);
		return route.fulfill({
			json: {
				items: matches.slice((number - 1) * size, number * size),
				total: matches.length,
				page: number,
				page_size: size
			}
		});
	});
	await page.route(/\/api\/functions\/\d+$/, (route) => {
		const path = new URL(route.request().url()).pathname;
		const id = Number(path.split('/').pop());
		const index = functions.findIndex((fn) => fn.id === id);
		if (index < 0) return route.fulfill({ status: 404, json: { error: 'Function not found' } });
		const method = route.request().method();
		if (method === 'DELETE') {
			mutations.push({ method, path, body: null });
			functions.splice(index, 1);
			return route.fulfill({ status: 204 });
		}
		if (method === 'PUT') {
			const body = route.request().postDataJSON();
			mutations.push({ method, path, body });
			Object.assign(functions[index], body);
		}
		return route.fulfill({ json: functions[index] });
	});
	await page.route(/\/api\/devices\/(?:WIFI_TEST|GOLDENKV)\/functions\?/, (route) => {
		const device = new URL(route.request().url()).pathname.split('/')[3];
		const matches = functions.filter((fn) => fn.id <= 2 || fn.interface === device);
		return route.fulfill({
			json: {
				items: matches,
				total: matches.length,
				page: 1,
				page_size: 100
			}
		});
	});
	return { functions, requests, mutations };
}

async function mockDictionary(page: Page, kind: DefinitionKind) {
	await mockDeviceHub(page);
	const rows: DefinitionRow[] = ['', 'WIFI_TEST', 'GOLDENKV'].map((device_name, index) => ({
		id: index + 1,
		name: 'ANT',
		device_name,
		description: 'Antenna selection.',
		type: 'int',
		unit: 'ms',
		default_value: String(index),
		allowed_values: '{0,1,2}',
		cardinality: 'SINGLE',
		value_pattern: '^[0-9]+$',
		required: true,
		usage_count: 2,
		orphan_usage_count: 1,
		created_at: '2026-10-06',
		updated_at: '2026-10-06'
	}));
	const requests: URLSearchParams[] = [];
	const mutations: { method: string; path: string; body: Record<string, unknown> | null }[] = [];
	await page.route(new RegExp(`/api/${kind}(?:\\?.*)?$`), (route) => {
		if (route.request().method() === 'POST') {
			const body = route.request().postDataJSON();
			if (!body.name)
				return route.fulfill({
					status: 422,
					json: { error: 'Invalid definition', fields: { name: 'Name is required.' } }
				});
			if (body.unit === 'INVALID')
				return route.fulfill({
					status: 422,
					json: { error: 'Invalid definition', fields: { unit: 'Unit is invalid.' } }
				});
			const row = {
				...rows[0],
				...body,
				id: Math.max(...rows.map((row) => row.id), 0) + 1,
				usage_count: 0,
				orphan_usage_count: 0
			};
			rows.push(row);
			mutations.push({ method: 'POST', path: `/api/${kind}`, body });
			return route.fulfill({ status: 201, json: row });
		}
		if (route.request().method() !== 'GET')
			return route.fulfill({ status: 500, json: { error: 'Unexpected dictionary mutation' } });
		const query = new URL(route.request().url()).searchParams;
		requests.push(query);
		const matches = rows.filter((row) =>
			`${row.name} ${row.description}`
				.toLowerCase()
				.includes((query.get('search') ?? '').toLowerCase())
		);
		const sort =
			query.get('sort') === 'type'
				? 'type'
				: query.get('sort') === 'updated_at'
					? 'updated_at'
					: 'name';
		matches.sort(
			(left, right) =>
				left[sort].localeCompare(right[sort]) * (query.get('order') === 'desc' ? -1 : 1)
		);
		const number = Number(query.get('page') ?? 1);
		const size = Number(query.get('page_size') ?? 20);
		return route.fulfill({
			json: {
				items: matches.slice((number - 1) * size, number * size),
				total: matches.length,
				page: number,
				page_size: size
			}
		});
	});
	await page.route(new RegExp(`/api/${kind}/[0-9]+(?:\\?.*)?$`), (route) => {
		const url = new URL(route.request().url());
		const id = Number(url.pathname.split('/').pop());
		const index = rows.findIndex((row) => row.id === id);
		if (index < 0) return route.fulfill({ status: 404, json: { error: 'Definition not found' } });
		const method = route.request().method();
		if (method === 'PUT') {
			const body = route.request().postDataJSON();
			mutations.push({ method, path: url.pathname, body });
			Object.assign(rows[index], body);
			return route.fulfill({ json: rows[index] });
		}
		if (method === 'DELETE') {
			if (rows[index].usage_count > 0 && url.searchParams.get('force') !== 'true')
				return route.fulfill({ status: 409, json: { error: 'Definition is in use.' } });
			mutations.push({ method, path: url.pathname + url.search, body: null });
			rows.splice(index, 1);
			return route.fulfill({ status: 204 });
		}
		if (method === 'GET') return route.fulfill({ json: rows[index] });
		return route.fulfill({ status: 500, json: { error: 'Unexpected definition request' } });
	});
	return { rows, requests, mutations };
}

async function switchDevice(page: Page, device: string) {
	await page.getByRole('button', { name: /^Device: / }).click();
	const popup = page.getByRole('dialog', { name: 'Choose Device', exact: true });
	await popup.getByRole('combobox', { name: 'Search devices', exact: true }).fill(device);
	await popup.getByRole('option', { name: device || 'Global definitions', exact: true }).click();
}

async function chooseFormDevice(page: Page, device: string) {
	await page
		.getByRole('dialog')
		.getByRole('button', { name: /^Device \/ Interface:/ })
		.click();
	const popup = page.getByRole('dialog', { name: 'Choose Device / Interface', exact: true });
	await popup.getByRole('combobox', { name: 'Search devices', exact: true }).fill(device);
	await popup
		.getByRole('option', { name: device || 'Global (no Device / Interface)', exact: true })
		.click();
}

test('app uses the supplied ICO for the tab and navigation brand', async ({ page }) => {
	await mockDeviceHub(page);
	await page.goto('/devices');
	const icon = page.locator('.brand img');
	await expect(icon).toBeVisible();
	await expect
		.poll(() => icon.evaluate((image: HTMLImageElement) => image.naturalWidth))
		.toBeGreaterThan(0);
	const source = await icon.getAttribute('src');
	await expect(page.locator('link[rel="icon"]')).toHaveAttribute('href', source!);
	await expect(page.locator('link[rel="icon"]')).toHaveAttribute('type', 'image/x-icon');
});

test('Function Device / Interface searches the catalog and accepts blank on edit', async ({
	page
}) => {
	const { mutations } = await mockFunctionCatalog(page);
	await page.goto('/functions');
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'New Function', exact: true });
	await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('CATALOG_PICKER_FN');
	await dialog.getByRole('button', { name: /^Device \/ Interface:/ }).click();
	const popup = page.getByRole('dialog', { name: 'Choose Device / Interface', exact: true });
	await popup.getByRole('combobox', { name: 'Search devices' }).fill('gpi');
	await expect(popup.getByRole('option', { name: 'WIFI_TEST', exact: true })).toHaveCount(0);
	await popup.getByRole('option', { name: 'GPIB', exact: true }).click();
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	expect(mutations.at(-1)?.body?.interface).toBe('GPIB');
	await page.getByRole('button', { name: 'Edit function', exact: true }).click();
	await chooseFormDevice(page, '');
	await page
		.getByRole('dialog', { name: 'Edit Function #3', exact: true })
		.getByRole('button', { name: 'Save', exact: true })
		.click();
	await expect(page.getByRole('dialog')).toHaveCount(0);
	expect(mutations.at(-1)?.body?.interface).toBe('');
});

test('Function picker retries catalog failures without losing the blank choice', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.goto('/functions');
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	let unavailable = true;
	await page.route('**/api/devices', (route) =>
		unavailable
			? route.fulfill({ status: 503, json: { error: 'Catalog unavailable' } })
			: route.fallback()
	);
	await page
		.getByRole('dialog')
		.getByRole('button', { name: /^Device \/ Interface:/ })
		.click();
	const popup = page.getByRole('dialog', { name: 'Choose Device / Interface', exact: true });
	await expect(popup.getByRole('alert')).toContainText('Catalog unavailable');
	await expect(
		popup.getByRole('option', { name: 'Global (no Device / Interface)', exact: true })
	).toBeVisible();
	unavailable = false;
	await popup.getByRole('button', { name: 'Retry', exact: true }).click();
	await expect(popup.getByRole('option', { name: 'WIFI_TEST', exact: true })).toBeVisible();
	await popup.getByRole('option', { name: 'Global (no Device / Interface)', exact: true }).click();
	await page
		.getByRole('dialog', { name: 'New Function', exact: true })
		.getByRole('button', { name: 'Cancel' })
		.click();
});

for (const kind of ['parameter-definitions', 'criteria-definitions'] as const) {
	test(`${kind} picker permits catalog-only devices and blank scope`, async ({ page }) => {
		const { mutations } = await mockDictionary(page, kind);
		await page.goto(`/${kind}`);
		await page.getByRole('button', { name: 'New', exact: true }).click();
		const dialog = page.getByRole('dialog');
		await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('BLANK_SCOPE');
		await chooseFormDevice(page, 'GPIB');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
			'GPIB'
		);
		await chooseFormDevice(page, '');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(dialog).toHaveCount(0);
		expect(mutations.at(-1)?.body?.device_name).toBe('');
	});
	test(`${kind} dictionary separates same-name ownership and browses constraints before editing`, async ({
		page
	}) => {
		await mockDictionary(page, kind);
		await page.goto(`/${kind}?definition=2`);
		const list = page.getByRole('complementary', { name: 'Definitions list', exact: true });
		const detail = page.getByRole('region', { name: 'Definition details', exact: true });
		await expect(list).toBeVisible();
		await expect(detail.getByRole('heading', { name: 'ANT', exact: true })).toBeVisible();
		await expect(detail.getByText('WIFI_TEST', { exact: true })).toBeVisible();
		await expect(detail).toContainText('^[0-9]+$');
		await expect(detail).toContainText('{0,1,2}');
		await expect(
			detail.getByRole('button', { name: 'Edit definition', exact: true })
		).toBeVisible();
		await expect(
			detail.getByRole('button', { name: 'Delete definition', exact: true })
		).toBeVisible();
		await list.getByRole('button', { name: 'Select ANT (Global)', exact: true }).click();
		await expect(page).toHaveURL(/definition=1/);
		await expect(detail.getByText('Global - shared across Devices', { exact: true })).toBeVisible();
		await expect(detail.getByText('1 orphan link', { exact: true })).toBeVisible();
		await expect(page.getByRole('dialog')).toHaveCount(0);
		await page.goBack();
		await expect(detail.getByText('WIFI_TEST', { exact: true })).toBeVisible();
	});
	test(`${kind} dictionary filters complete Device results and keeps Global distinct`, async ({
		page
	}) => {
		const { rows, requests } = await mockDictionary(page, kind);
		rows.push(
			...Array.from({ length: 100 }, (_, index) => ({
				...rows[0],
				id: index + 10,
				name: `EARLY_${index}`
			}))
		);
		rows.push({ ...rows[1], id: 200, name: 'LATE_DEVICE_DEFINITION' });
		await page.goto(`/${kind}?device=WIFI_TEST`);
		await expect(page.getByText('2 entries', { exact: true })).toBeVisible();
		const list = page.getByRole('complementary', { name: 'Definitions list', exact: true });
		await list
			.getByRole('button', { name: 'Select LATE_DEVICE_DEFINITION (WIFI_TEST)', exact: true })
			.click();
		await expect(
			page
				.getByRole('region', { name: 'Definition details', exact: true })
				.getByRole('heading', { name: 'LATE_DEVICE_DEFINITION', exact: true })
		).toBeVisible();
		expect(
			requests.some((query) => query.get('page') === '2' && query.get('page_size') === '100')
		).toBe(true);
		await page.getByRole('button', { name: /^Filter definitions by Device:/ }).click();
		await page.getByRole('option', { name: 'Global definitions', exact: true }).click();
		await expect(page).toHaveURL(/device=_global/);
		await expect(page.getByText('101 entries', { exact: true })).toBeVisible();
		await expect(
			list.getByRole('button', { name: 'Select ANT (WIFI_TEST)', exact: true })
		).toHaveCount(0);
		await page
			.getByRole('textbox', { name: /^Search .* definitions$/ })
			.fill('LATE_DEVICE_DEFINITION');
		await expect(page.getByText('0 entries', { exact: true })).toBeVisible();
		await expect(
			list.getByText('No entries match. Adjust the search or create one.', { exact: true })
		).toBeVisible();
	});
	test(`${kind} dictionary CRUD retains validation context and usage-protected deletion`, async ({
		page
	}) => {
		const { mutations } = await mockDictionary(page, kind);
		await page.goto(`/${kind}?device=WIFI_TEST&search=ANT&definition=2`);
		await page.getByRole('button', { name: 'New', exact: true }).click();
		let dialog = page.getByRole('dialog');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
			'WIFI_TEST'
		);
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(dialog.getByText('Name is required.', { exact: true })).toBeVisible();
		await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('ANT_NEW');
		await dialog.getByRole('textbox', { name: 'Unit', exact: true }).fill('INVALID');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(dialog.getByText('Unit is invalid.', { exact: true })).toBeVisible();
		await dialog.getByRole('textbox', { name: 'Unit', exact: true }).fill('ms');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		const details = page.getByRole('region', { name: 'Definition details', exact: true });
		await expect(details.getByRole('heading', { name: 'ANT_NEW', exact: true })).toBeVisible();
		await expect(page).toHaveURL(/device=WIFI_TEST&search=ANT&definition=4/);
		await details.getByRole('button', { name: 'Edit definition', exact: true }).click();
		dialog = page.getByRole('dialog');
		await dialog.getByRole('textbox', { name: 'Description', exact: true }).fill('Discarded');
		await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
		expect(mutations.filter((call) => call.method === 'PUT')).toHaveLength(0);
		await details.getByRole('button', { name: 'Edit definition', exact: true }).click();
		await dialog
			.getByRole('textbox', { name: 'Description', exact: true })
			.fill('Updated definition');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(details.getByText('Updated definition', { exact: true })).toBeVisible();
		expect(mutations.find((call) => call.method === 'PUT')?.body?.device_name).toBe('WIFI_TEST');
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(details.getByRole('heading', { name: 'ANT', exact: true })).toBeVisible();
		await expect(page).toHaveURL(/definition=2/);
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		const confirmation = page.getByRole('alertdialog', { name: 'Force Delete ANT?', exact: true });
		await expect(confirmation).toContainText(
			'Links are retained and may resolve to a Global fallback.'
		);
		await expect(confirmation).toContainText('Without a fallback, they become unresolved.');
		await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click();
		expect(mutations.filter((call) => call.method === 'DELETE')).toHaveLength(1);
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
		await expect(page.getByText('0 entries', { exact: true })).toBeVisible();
		await expect(page).not.toHaveURL(/definition=/);
		expect(mutations.at(-1)?.path).toBe(`/api/${kind}/2?force=true`);
		await page.goto(`/${kind}?device=_global&definition=1`);
		await page
			.getByRole('region', { name: 'Definition details', exact: true })
			.getByRole('button', { name: 'Delete Global definition', exact: true })
			.click();
		await expect(confirmation).toContainText(
			'Links using this Global definition are retained but become unresolved.'
		);
		await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click();
		expect(mutations.at(-1)?.path).toBe(`/api/${kind}/2?force=true`);
	});
	test(`${kind} dictionary Device actions protect ownership and distinguish Global overrides`, async ({
		page
	}) => {
		const { rows, mutations } = await mockDictionary(page, kind);
		const section = kind === 'parameter-definitions' ? 'parameters' : 'criteria';
		rows.push({
			...rows[0],
			id: 5,
			name: 'GLOBAL_LIMIT',
			default_value: '77',
			allowed_values: '{77,88}'
		});
		await page.route(`**/api/devices/WIFI_TEST/${section}?*`, (route) => {
			const scoped = rows.filter((row) => row.device_name === 'WIFI_TEST');
			const globals = rows.filter(
				(row) =>
					!row.device_name &&
					['ANT', 'GLOBAL_LIMIT'].includes(row.name) &&
					!scoped.some((item) => item.name === row.name)
			);
			return route.fulfill({
				json: {
					items: scoped,
					total: scoped.length,
					global_items: globals,
					global_total: globals.length,
					page: 1,
					page_size: 20
				}
			});
		});
		await page.goto(`/devices/WIFI_TEST?tab=${section}`);
		await page
			.getByRole('button', {
				name:
					kind === 'parameter-definitions' ? 'Add parameter definition' : 'Add criteria definition',
				exact: true
			})
			.click();
		let dialog = page.getByRole('dialog');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
			'WIFI_TEST'
		);
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toBeDisabled();
		await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('DEVICE_ONLY');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		const details = page.getByRole('region', { name: 'Definition details', exact: true });
		await expect(details.getByRole('heading', { name: 'DEVICE_ONLY', exact: true })).toBeVisible();
		await page.getByRole('button', { name: 'Inspect ANT (WIFI_TEST)', exact: true }).click();
		await expect(details).toContainText('2 Function links');
		await details.getByRole('button', { name: 'Edit definition', exact: true }).click();
		dialog = page.getByRole('dialog');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toBeDisabled();
		await dialog
			.getByRole('textbox', { name: 'Description', exact: true })
			.fill('Device only edit');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(details.getByText('Device only edit', { exact: true })).toBeVisible();
		expect(mutations.at(-1)?.body?.device_name).toBe('WIFI_TEST');
		await page.getByRole('button', { name: 'Inspect GLOBAL_LIMIT (Global)', exact: true }).click();
		await expect(
			details.getByText('Global - shared across Devices', { exact: true })
		).toBeVisible();
		await details.getByRole('button', { name: 'Edit Global definition', exact: true }).click();
		dialog = page.getByRole('dialog');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
			'Global'
		);
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toBeDisabled();
		await expect(dialog.getByText('Global - shared across Devices', { exact: true })).toBeVisible();
		await dialog.getByRole('textbox', { name: 'Default value', exact: true }).fill('88');
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		expect(mutations.at(-1)?.path).toBe(`/api/${kind}/5`);
		expect(mutations.at(-1)?.body?.device_name).toBe('');
		await details.getByRole('button', { name: 'Create Device override', exact: true }).click();
		dialog = page.getByRole('dialog');
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
			'WIFI_TEST'
		);
		await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toBeDisabled();
		await expect(dialog.getByRole('textbox', { name: 'Name', exact: true })).toHaveValue(
			'GLOBAL_LIMIT'
		);
		await expect(dialog.getByRole('textbox', { name: 'Default value', exact: true })).toHaveValue(
			'88'
		);
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(details.getByText('WIFI_TEST', { exact: true })).toBeVisible();
		expect(mutations.at(-1)?.method).toBe('POST');
		expect(mutations.at(-1)?.body).toMatchObject({
			name: 'GLOBAL_LIMIT',
			device_name: 'WIFI_TEST',
			default_value: '88',
			allowed_values: '{77,88}'
		});
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(page).toHaveURL(new RegExp(`/devices/WIFI_TEST\\?tab=${section}`));
		await page.getByRole('button', { name: 'Inspect GLOBAL_LIMIT (Global)', exact: true }).click();
		await expect(
			details.getByText('Global - shared across Devices', { exact: true })
		).toBeVisible();
		await expect(details.getByText('88', { exact: true })).toBeVisible();
		expect(mutations.at(-1)?.path).toBe(`/api/${kind}/7`);
	});
	test(`${kind} dictionary retries failures and requires new force confirmation for stale usage`, async ({
		page
	}) => {
		const { rows, mutations } = await mockDictionary(page, kind);
		rows[1].usage_count = 0;
		let fail = true;
		await page.route(new RegExp(`/api/${kind}(?:\\?.*)?$`), (route) => {
			if (fail) {
				fail = false;
				return route.fulfill({
					status: 503,
					json: { error: 'Dictionary temporarily unavailable' }
				});
			}
			return route.fallback();
		});
		await page.goto(`/${kind}?definition=2`);
		await expect(page.getByRole('alert')).toContainText('Dictionary temporarily unavailable');
		await expect(page.getByText('Count unavailable', { exact: true })).toBeVisible();
		await page.getByRole('button', { name: 'Retry definitions', exact: true }).click();
		const details = page.getByRole('region', { name: 'Definition details', exact: true });
		await expect(details).toContainText('0 Function links');
		rows[1].usage_count = 2;
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(page.getByRole('alertdialog')).toContainText('Definition is in use.');
		await expect(
			page.getByRole('alertdialog', { name: 'Force Delete ANT?', exact: true })
		).toBeVisible();
		expect(mutations).toHaveLength(0);
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Cancel', exact: true })
			.click();
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(
			page.getByRole('alertdialog', { name: 'Force Delete ANT?', exact: true })
		).toBeVisible();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(page.getByRole('alertdialog')).not.toBeVisible();
		expect(mutations.at(-1)?.path).toBe(`/api/${kind}/2?force=true`);
	});
	test(`${kind} dictionary ignores a delayed response after leaving the workspace`, async ({
		page
	}) => {
		const { rows } = await mockDictionary(page, kind);
		let release!: () => void;
		const pending = new Promise<void>((resolve) => {
			release = resolve;
		});
		await page.route(new RegExp(`/api/${kind}(?:\\?.*)?$`), async (route) => {
			await pending;
			return route.fulfill({ json: { items: [rows[0]], total: 1, page: 1, page_size: 20 } });
		});
		await page.goto(`/${kind}?definition=999`);
		await expect(page.getByText('Loading definitions...', { exact: true })).toBeVisible();
		await page.locator('a[href="/devices"]').first().click();
		await expect(page).toHaveURL('/devices');
		const response = page.waitForResponse(new RegExp(`/api/${kind}(?:\\?.*)?$`));
		release();
		await (await response).finished();
		await page.waitForLoadState('networkidle');
		await expect(page).toHaveURL('/devices');
	});
	test(`${kind} dictionary long content stays readable across layouts and keyboard selection`, async ({
		page
	}, testInfo) => {
		const { rows } = await mockDictionary(page, kind);
		rows[1].name = `LONG_DEFINITION_${'NAME_'.repeat(24)}`;
		rows[1].device_name = `DEVICE_${'SCOPE_'.repeat(14)}`;
		rows[1].value_pattern = `^${'LONG_PATTERN_'.repeat(24)}$`;
		await page.goto(`/${kind}?definition=2`);
		for (const width of [1440, 960, 390]) {
			await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
			await expect(
				page.getByRole('region', { name: 'Definition details', exact: true })
			).toBeVisible();
			await page.screenshot({ path: testInfo.outputPath(`${kind}-${width}.png`), fullPage: true });
			const sizes = await page.evaluate(() => ({
				width: document.documentElement.clientWidth,
				scroll: document.documentElement.scrollWidth
			}));
			expect(sizes.scroll).toBeLessThanOrEqual(sizes.width);
		}
		await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
		await expect(page.locator('html')).toHaveClass(/dark/);
		await page.screenshot({ path: testInfo.outputPath(`${kind}-390-dark.png`), fullPage: true });
		const choice = page.getByRole('button', { name: 'Select ANT (Global)', exact: true });
		await choice.focus();
		await page.keyboard.press('Enter');
		await expect(
			page.getByRole('region', { name: 'Definition details', exact: true })
		).toContainText('Global - shared across Devices');
		await expect(choice).toBeFocused();
		const section = kind === 'parameter-definitions' ? 'parameters' : 'criteria';
		rows[1].device_name = 'WIFI_TEST';
		await page.route(`**/api/devices/WIFI_TEST/${section}?*`, (route) =>
			route.fulfill({
				json: {
					items: [rows[1]],
					total: 1,
					global_items: [rows[0]],
					global_total: 1,
					page: 1,
					page_size: 20
				}
			})
		);
		await page.goto(`/devices/WIFI_TEST?tab=${section}&definition=2`);
		await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
		await expect(page.locator('html')).not.toHaveClass(/dark/);
		for (const width of [1440, 390]) {
			await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
			await expect(
				page
					.getByRole('region', { name: 'Definition details', exact: true })
					.getByRole('heading', { name: rows[1].name, exact: true })
			).toBeVisible();
			const sizes = await page.evaluate(() => ({
				width: document.documentElement.clientWidth,
				scroll: document.documentElement.scrollWidth
			}));
			expect(sizes.scroll).toBeLessThanOrEqual(sizes.width);
			await page.screenshot({
				path: testInfo.outputPath(`${kind}-device-${width}.png`),
				fullPage: true
			});
		}
		await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
		await expect(page.locator('html')).toHaveClass(/dark/);
		await page.screenshot({
			path: testInfo.outputPath(`${kind}-device-390-dark.png`),
			fullPage: true
		});
	});
	test(`${kind} dictionary paging and sorting retain URL history and separate Device tabs`, async ({
		page
	}) => {
		const { rows, requests } = await mockDictionary(page, kind);
		for (let index = 10; index < 40; index++)
			rows.push({ ...rows[0], id: index, name: `VALUE_${index}` });
		await page.goto(`/${kind}`);
		await page.getByRole('button', { name: 'Next page', exact: true }).click();
		await expect(page.getByText('Page 2 of 2', { exact: true })).toBeVisible();
		await page.getByRole('button', { name: 'Select VALUE_39 (Global)', exact: true }).click();
		await expect(
			page
				.getByRole('region', { name: 'Definition details', exact: true })
				.getByRole('heading', { name: 'VALUE_39', exact: true })
		).toBeVisible();
		await page.getByRole('button', { name: 'Sort direction', exact: true }).click();
		await page.getByRole('option', { name: 'Descending', exact: true }).click();
		await expect(page.getByText('Page 1 of 2', { exact: true })).toBeVisible();
		await expect(
			page
				.getByRole('complementary', { name: 'Definitions list', exact: true })
				.getByRole('button')
				.first()
		).toHaveAccessibleName('Select VALUE_39 (Global)');
		expect(requests.at(-1)?.get('order')).toBe('desc');
		await page.goBack();
		await expect(page.getByText('Page 2 of 2', { exact: true })).toBeVisible();
		await expect(page).toHaveURL(new RegExp('page=2&definition=39$'));
		const section = kind === 'parameter-definitions' ? 'parameters' : 'criteria';
		const other = section === 'parameters' ? 'Criteria' : 'Parameters';
		const otherName = section === 'parameters' ? 'POWER' : 'ANT';
		await page.goto(
			`/devices/WIFI_TEST?tab=${section}&definition=${section === 'parameters' ? 11 : 23}`
		);
		await page.getByRole('tab', { name: other, exact: true }).click();
		await expect(
			page
				.getByRole('region', { name: 'Definition details', exact: true })
				.getByRole('heading', { name: otherName, exact: true })
		).toBeVisible();
		await expect(page).not.toHaveURL(/definition=/);
	});
	test(`${kind} dictionary Device deletion clamps a emptied final page`, async ({ page }) => {
		const { rows } = await mockDictionary(page, kind);
		const section = kind === 'parameter-definitions' ? 'parameters' : 'criteria';
		const seed = rows[1];
		rows.splice(
			0,
			rows.length,
			...Array.from({ length: 21 }, (_, index) => ({
				...seed,
				id: 100 + index,
				name: `FIELD_${index}`,
				usage_count: 0,
				orphan_usage_count: 0
			}))
		);
		await page.route(`**/api/devices/WIFI_TEST/${section}?*`, (route) => {
			const query = new URL(route.request().url()).searchParams;
			const number = Number(query.get('page') ?? 1);
			const size = Number(query.get('page_size') ?? 20);
			return route.fulfill({
				json: {
					items: rows.slice((number - 1) * size, number * size),
					total: rows.length,
					global_items: [],
					global_total: 0,
					page: number,
					page_size: size
				}
			});
		});
		await page.goto(`/devices/WIFI_TEST?tab=${section}`);
		await page.getByRole('button', { name: 'Next page', exact: true }).click();
		await page.getByRole('button', { name: 'Inspect FIELD_20 (WIFI_TEST)', exact: true }).click();
		const details = page.getByRole('region', { name: 'Definition details', exact: true });
		await details.getByRole('button', { name: 'Delete definition', exact: true }).click();
		await page
			.getByRole('alertdialog')
			.getByRole('button', { name: 'Delete', exact: true })
			.click();
		await expect(
			page.getByRole('button', { name: 'Inspect FIELD_0 (WIFI_TEST)', exact: true })
		).toBeVisible();
		await expect(details.getByRole('heading', { name: 'FIELD_0', exact: true })).toBeVisible();
		await expect(page.getByText('Page 2 of 1', { exact: true })).not.toBeVisible();
	});
}

test.beforeEach(async ({ page }) => {
	await page.route('**/api/**', async (route) => {
		await route.fulfill({ json: { items: [], total: 0 } });
	});
});

const importWorkbookName = `Master_Control_${'LONG_DEVICE_NAME_'.repeat(10)}Knowledge.xlsx`;
const importResult: ImportResult = {
	run_id: 42,
	file_name: importWorkbookName,
	reset: false,
	counts: { function_main: 12, parameter_main: 34, criteria_main: 0 },
	skipped: [{ table: 'example', row_key: 'DLL/WIFI_TEST/VERIFY', reason: 'Missing script' }],
	normalizations: [
		{
			table: 'function_main',
			row_key: 'VERIFY',
			column: 'interface',
			source: 'GoldenKV',
			normalized: 'GOLDENKV',
			rule: 'uppercase_device'
		}
	]
};
const importHistoryRun: ImportRun = {
	id: 41,
	file_name: importWorkbookName,
	reset: true,
	started_at: '2026-10-15T09:00:00Z',
	finished_at: '2026-10-15T09:00:02Z',
	counts: JSON.stringify({ function_main: 9, parameter_main: 20 }),
	skipped: JSON.stringify(importResult.skipped)
};
const importNormalization: ImportNormalization = {
	id: 1,
	run_id: 41,
	table_name: 'function_main',
	row_key: `DLL/${'LONG_DEVICE_NAME_'.repeat(12)}/VERIFY`,
	column_name: 'interface',
	source_value: 'GoldenKV',
	normalized_value: 'GOLDENKV',
	rule: 'uppercase_device'
};
const selectedWorkbook = {
	name: importWorkbookName,
	mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
	buffer: Buffer.from('mock workbook bytes')
};

test('Import upload preserves multipart semantics and shows pending success and API results', async ({
	page
}) => {
	const submissions: { path: string; contentType: string; body: string }[] = [];
	let completeImport: (() => void) | undefined;
	const pendingImport = new Promise<void>((resolve) => {
		completeImport = resolve;
	});
	let completed = false;
	await page.route('**/api/import**', async (route) => {
		const request = route.request();
		if (request.method() === 'POST') {
			submissions.push({
				path: new URL(request.url()).pathname,
				contentType: request.headers()['content-type'],
				body: request.postDataBuffer()!.toString()
			});
			await pendingImport;
			completed = true;
			return route.fulfill({ json: importResult });
		}
		return route.fulfill({
			json: { items: completed ? [{ ...importHistoryRun, id: 42, reset: false }] : [] }
		});
	});
	await page.goto('/import');
	await page.getByLabel('Workbook file', { exact: true }).setInputFiles(selectedWorkbook);
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	await expect(page.getByRole('status').filter({ hasText: 'Importing workbook' })).toBeVisible();
	await expect(page.getByRole('button', { name: 'Importing', exact: true })).toBeDisabled();
	await expect(page.getByRole('button', { name: 'Reset & import', exact: true })).toBeDisabled();
	await expect(page.getByLabel('Workbook file', { exact: true })).toBeDisabled();
	expect(submissions).toHaveLength(1);
	expect(submissions[0].path).toBe('/api/import');
	expect(submissions[0].contentType).toContain('multipart/form-data; boundary=');
	expect(submissions[0].body).toContain('name="file"');
	expect(submissions[0].body).toContain(importWorkbookName);
	expect(submissions[0].body).toContain('mock workbook bytes');
	completeImport!();
	const results = page.getByRole('region', { name: 'Import results', exact: true });
	await expect(results).toContainText('run #42');
	await expect(results).toContainText('function_main');
	await expect(results).toContainText('12');
	await expect(results).toContainText('34');
	await expect(results).toContainText('0');
	await expect(results).toContainText('Missing script');
	await results.getByRole('button', { name: 'Normalization details' }).click();
	await expect(results.getByRole('cell', { name: 'uppercase_device', exact: true })).toBeVisible();
	await expect(page.getByRole('button', { name: /#42/ })).toBeVisible();
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Import reset requires confirmation and cancellation restores keyboard focus without requests', async ({
	page
}) => {
	const submissions: string[] = [];
	await page.route('**/api/import**', async (route) => {
		if (route.request().method() === 'POST') {
			submissions.push(new URL(route.request().url()).pathname);
			return route.fulfill({ json: { ...importResult, reset: true } });
		}
		return route.fulfill({ json: { items: [] } });
	});
	await page.goto('/import');
	const reset = page.getByRole('button', { name: 'Reset & import', exact: true });
	await reset.click();
	await expect(page.getByRole('alert')).toContainText('Choose a workbook file first.');
	await expect(page.getByRole('alertdialog')).toHaveCount(0);
	await page.getByLabel('Workbook file', { exact: true }).setInputFiles(selectedWorkbook);
	await reset.focus();
	await page.keyboard.press('Enter');
	const dialog = page.getByRole('alertdialog');
	await expect(dialog).toContainText('Rows edited only in this app are lost.');
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(reset).toBeFocused();
	expect(submissions).toEqual([]);
	await page.keyboard.press('Enter');
	await dialog.getByRole('button', { name: 'Reset & import', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	await expect(page.getByRole('region', { name: 'Import results', exact: true })).toContainText(
		'Reset import'
	);
	await expect(reset).toBeFocused();
	expect(submissions).toEqual(['/api/import/reset']);
});

for (const theme of ['light', 'dark'] as const) {
	test(`Final audit: Import reset confirmation is readable and keyboard operable in ${theme} mode`, async ({
		page
	}) => {
		const submissions: string[] = [];
		await page.route('**/api/import**', (route) => {
			if (route.request().method() === 'POST') submissions.push(route.request().url());
			return route.fulfill({ json: { items: [] } });
		});
		await page.emulateMedia({ colorScheme: theme });
		await page.goto('/import');
		await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /^(?!.*dark)/);
		await page.getByLabel('Workbook file', { exact: true }).setInputFiles(selectedWorkbook);
		const reset = page.getByRole('button', { name: 'Reset & import', exact: true });
		await reset.focus();
		await page.keyboard.press('Enter');
		const dialog = page.getByRole('alertdialog');
		const cancel = dialog.getByRole('button', { name: 'Cancel', exact: true });
		const confirm = dialog.getByRole('button', { name: 'Reset & import', exact: true });
		await expect(dialog).toBeFocused();
		await page.keyboard.press('Tab');
		await expect(cancel).toBeFocused();
		await page.keyboard.press('Tab');
		await expect(confirm).toBeFocused();
		await page.screenshot({ path: test.info().outputPath(`reset-confirmation-${theme}.png`) });
		const contrast = await confirm.evaluate((element) => {
			const luminance = (color: string) => {
				const channels = color
					.match(/[\d.]+/g)!
					.slice(0, 3)
					.map(Number)
					.map((channel) => {
						const value = channel / 255;
						return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
					});
				return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
			};
			const style = getComputedStyle(element);
			const values = [luminance(style.color), luminance(style.backgroundColor)].sort(
				(first, second) => second - first
			);
			return (values[0] + 0.05) / (values[1] + 0.05);
		});
		expect(contrast).toBeGreaterThanOrEqual(4.5);
		expect(await confirm.evaluate((element) => getComputedStyle(element).outlineWidth)).not.toBe(
			'0px'
		);
		await page.keyboard.press('Tab');
		await expect(cancel).toBeFocused();
		await page.keyboard.press('Enter');
		await expect(dialog).toHaveCount(0);
		await expect(reset).toBeFocused();
		expect(submissions).toEqual([]);
	});
}

test('Import upload failure retains the workbook for retry and accepts clean null results', async ({
	page
}) => {
	let attempts = 0;
	await page.route('**/api/import**', async (route) => {
		if (route.request().method() === 'POST') {
			attempts += 1;
			return route.fulfill(
				attempts === 1
					? { status: 422, json: { error: 'Invalid workbook sheet' } }
					: { json: { ...importResult, normalizations: null, skipped: null } }
			);
		}
		return route.fulfill({ json: { items: [] } });
	});
	await page.goto('/import');
	await page.getByLabel('Workbook file', { exact: true }).setInputFiles(selectedWorkbook);
	const upload = page.getByRole('button', { name: 'Import', exact: true });
	await upload.click();
	await expect(page.getByRole('alert')).toContainText('Invalid workbook sheet');
	await expect(upload).toBeEnabled();
	await upload.click();
	await expect(page.getByRole('region', { name: 'Import results', exact: true })).toContainText(
		'No normalizations needed.'
	);
	await expect(page.locator('.banner-error')).toHaveCount(0);
	expect(attempts).toBe(2);
});

for (const width of [1440, 390]) {
	test(`Import history and normalization details fit ${width}px in light and dark themes`, async ({
		page
	}, testInfo) => {
		await page.setViewportSize({ width, height: 900 });
		await page.emulateMedia({ colorScheme: 'light' });
		let finishHistory: (() => void) | undefined;
		const pendingHistory = new Promise<void>((resolve) => {
			finishHistory = resolve;
		});
		let finishDetails: (() => void) | undefined;
		const pendingDetails = new Promise<void>((resolve) => {
			finishDetails = resolve;
		});
		let detailAttempts = 0;
		await page.route('**/api/import/runs', async (route) => {
			await pendingHistory;
			return route.fulfill({
				json: {
					items: [
						importHistoryRun,
						{ ...importHistoryRun, id: 40, file_name: 'clean.xlsx', skipped: 'null' }
					]
				}
			});
		});
		await page.route('**/api/import/runs/*', async (route) => {
			if (route.request().url().endsWith('/40'))
				return route.fulfill({ json: { run: importHistoryRun, normalizations: [] } });
			detailAttempts += 1;
			if (detailAttempts === 1)
				return route.fulfill({ status: 503, json: { error: 'Details unavailable' } });
			await pendingDetails;
			return route.fulfill({
				json: { run: importHistoryRun, normalizations: [importNormalization] }
			});
		});
		await page.goto('/import');
		await expect(
			page.getByRole('status').filter({ hasText: 'Loading import history' })
		).toBeVisible();
		finishHistory!();
		await page.getByLabel('Workbook file', { exact: true }).setInputFiles(selectedWorkbook);
		const run = page.getByRole('button', { name: /#41/ });
		await run.focus();
		await page.keyboard.press('Enter');
		await expect(page.getByRole('alert')).toContainText('Details unavailable');
		await page.getByRole('button', { name: 'Retry details' }).click();
		await expect(
			page.getByRole('status').filter({ hasText: 'Loading normalization details' })
		).toBeVisible();
		finishDetails!();
		await expect(
			page.getByRole('cell', { name: importNormalization.row_key, exact: true })
		).toBeVisible();
		await expect(page.getByText('Missing script', { exact: false })).toBeVisible();
		const table = page.getByRole('region', { name: 'Normalization records', exact: true });
		for (const label of ['Table', 'Row', 'Column', 'Source', 'Normalized', 'Rule']) {
			const header = table.getByRole('columnheader', { name: new RegExp(`^${label}`) });
			for (const direction of ['ascending', 'descending', null]) {
				await header.getByRole('button').click();
				if (direction) await expect(header).toHaveAttribute('aria-sort', direction);
				else await expect(header).not.toHaveAttribute('aria-sort');
			}
		}
		for (const theme of ['light', 'dark']) {
			if (theme === 'dark')
				await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
			await expect(page.locator('html')).toHaveClass(theme === 'dark' ? /dark/ : /^(?!.*dark)/);
			await expect(page.locator('.banner-error')).toHaveCount(0);
			const geometry = await page.evaluate(() => ({
				root: document.documentElement.clientWidth,
				scroll: document.documentElement.scrollWidth,
				local: [...document.querySelectorAll<HTMLElement>('[aria-label="Normalization records"]')]
					.filter((region) => region.clientWidth > 0)
					.map((region) => ({
						width: region.clientWidth,
						scroll: region.scrollWidth,
						overflow: getComputedStyle(region).overflowX
					}))
			}));
			expect(geometry.scroll).toBeLessThanOrEqual(geometry.root);
			expect(geometry.local).toHaveLength(1);
			expect(geometry.local[0].overflow).toBe('auto');
			if (width === 390) {
				expect(geometry.local[0].scroll).toBeGreaterThan(geometry.local[0].width);
				const records = page.getByRole('region', { name: 'Normalization records', exact: true });
				await records.focus();
				await expect(records).toBeFocused();
				await page.keyboard.press('ArrowRight');
				await expect
					.poll(() => records.evaluate((element) => element.scrollLeft))
					.toBeGreaterThan(0);
				await records.evaluate((element) => {
					element.scrollLeft = 0;
				});
			}
			await page.screenshot({
				path: testInfo.outputPath(`import-${width}-${theme}.png`),
				fullPage: true
			});
		}
		await run.click();
		await expect(
			page.getByRole('cell', { name: importNormalization.row_key, exact: true })
		).toHaveCount(0);
		await page.getByRole('button', { name: /#40/ }).click();
		await expect(
			page.getByRole('paragraph').filter({ hasText: 'No normalizations in this run.' })
		).toBeVisible();
	});
}

test('Import history retries a failed request without submitting a workbook', async ({ page }) => {
	let historyRequests = 0;
	const submissions: string[] = [];
	await page.route('**/api/import**', async (route) => {
		if (route.request().method() === 'POST') submissions.push(route.request().url());
		historyRequests += 1;
		await route.fulfill(
			historyRequests === 1
				? { status: 503, json: { error: 'History unavailable' } }
				: { json: { items: [] } }
		);
	});
	await page.goto('/import');
	await expect(page.getByRole('alert')).toContainText('History unavailable');
	await page.getByRole('button', { name: 'Retry history' }).click();
	await expect(page.getByText('No imports yet.', { exact: true })).toBeVisible();
	await expect(page.getByRole('alert')).toHaveCount(0);
	expect(historyRequests).toBe(2);
	expect(submissions).toEqual([]);
});

test('Functions prototype A uses searchable catalog controls in both themes', async ({ page }) => {
	await page.route('**/api/**', (route) => route.abort());
	await page.goto('/demo/functions-prototype?variant=A');
	const catalog = page.locator('aside.catalog');
	await expect(catalog.getByRole('textbox', { name: 'Search functions' })).toBeVisible();
	await expect(catalog.getByRole('button', { name: 'Sort functions' })).toBeVisible();
	for (const theme of ['light', 'dark']) {
		if (theme === 'dark') await page.getByRole('button', { name: 'Switch to dark mode' }).click();
		await page.getByRole('button', { name: 'Filter by source' }).click();
		const source = page.getByRole('dialog', { name: 'Choose Filter by source' });
		await source.getByRole('combobox', { name: 'Search options' }).fill('dll');
		await source.getByRole('option', { name: 'DLL', exact: true }).click();
		await expect(catalog.locator('.function-list > button')).toHaveCount(3);
		await page.getByRole('button', { name: 'Clear filters' }).click();
		await expect(catalog.locator('.function-list > button')).toHaveCount(6);
		await catalog.getByRole('textbox', { name: 'Search functions' }).fill('AUTO_RUN');
		await expect(catalog.locator('.function-list > button')).toHaveCount(1);
		await page.getByRole('button', { name: 'Clear filters' }).click();
		await catalog.getByRole('button', { name: 'Sort direction' }).click();
		await page
			.getByRole('dialog', { name: 'Choose Sort direction' })
			.getByRole('option', { name: 'Descending' })
			.click();
		await expect(catalog.locator('.function-list > button').first()).toContainText(
			'TX_MULTI_VERIFICATION'
		);
		await catalog.getByRole('button', { name: 'Sort direction' }).click();
		await page
			.getByRole('dialog', { name: 'Choose Sort direction' })
			.getByRole('option', { name: 'Ascending' })
			.click();
		const checkbox = page.getByRole('checkbox', { name: 'Required: LEFT_OPERAND' });
		await expect(checkbox).toBeChecked();
		const colors = await checkbox.evaluate((element) => ({
			background: getComputedStyle(element).backgroundColor,
			mark: getComputedStyle(element).color,
			svg: Boolean(element.querySelector('svg'))
		}));
		expect(colors.svg).toBe(true);
		expect(colors.background).not.toBe(colors.mark);
		await checkbox.click();
		await expect(checkbox).not.toBeChecked();
		await checkbox.click();
		await page.getByRole('button', { name: 'New function' }).click();
		const dialog = page.getByRole('dialog', { name: 'New Function' });
		await dialog.getByRole('button', { name: 'Source', exact: true }).click();
		const options = page.getByRole('dialog', { name: 'Choose Source' });
		await options.getByRole('combobox', { name: 'Search options' }).fill('built');
		await options.getByRole('option', { name: 'BUILT-IN', exact: true }).click();
		await expect(dialog.getByRole('button', { name: 'Source', exact: true })).toContainText(
			'BUILT-IN'
		);
		await dialog.getByRole('button', { name: 'Cancel' }).click();
		await page.getByRole('button', { name: 'Attach parameter' }).click();
		const linkDialog = page.getByRole('dialog', { name: 'Attach Parameter' });
		await expect(linkDialog.getByRole('checkbox', { name: 'Required' })).toBeChecked();
		await linkDialog.getByRole('checkbox', { name: 'Required' }).click();
		await expect(linkDialog.getByRole('checkbox', { name: 'Required' })).not.toBeChecked();
		await linkDialog.getByRole('button', { name: 'Cancel' }).click();
	}
	await page.setViewportSize({ width: 390, height: 844 });
	await expect(catalog.getByRole('textbox', { name: 'Search functions' })).toBeVisible();
	const geometry = await page.evaluate(() => ({
		client: document.documentElement.clientWidth,
		scroll: document.documentElement.scrollWidth
	}));
	expect(geometry.scroll).toBeLessThanOrEqual(geometry.client);
});

test('Functions prototype A fills the filter row and sorts detail tables', async ({ page }) => {
	await page.route('**/api/**', (route) => route.abort());
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/demo/functions-prototype?variant=A');
	const filters = page.locator('.filters');
	const filterBounds = await filters.boundingBox();
	const clearBounds = await filters.getByRole('button', { name: 'Clear filters' }).boundingBox();
	expect(
		Math.abs(filterBounds!.x + filterBounds!.width - clearBounds!.x - clearBounds!.width)
	).toBeLessThan(2);
	const panel = page.getByRole('tabpanel', { name: /Parameters/ });
	const nameHeader = panel.getByRole('columnheader', { name: /Name/i });
	await nameHeader.getByRole('button').click();
	await expect(nameHeader).toHaveAttribute('aria-sort', 'ascending');
	await nameHeader.getByRole('button').click();
	await expect(nameHeader).toHaveAttribute('aria-sort', 'descending');
	const firstRow = panel.getByRole('row').nth(1);
	await expect(firstRow).toContainText('RIGHT_OPERAND');
	await expect(firstRow.getByRole('cell').first()).toHaveText('3');
	await firstRow.getByRole('button', { name: 'Edit RIGHT_OPERAND' }).click();
	await expect(
		page.getByRole('dialog', { name: 'Edit Parameter' }).getByLabel('Definition')
	).toHaveValue('RIGHT_OPERAND');
	await page.getByRole('button', { name: 'Cancel' }).click();
	await page.getByRole('button', { name: 'Switch to dark mode' }).click();
	await page.getByRole('tab', { name: /Criteria/ }).click();
	const criteria = page.getByRole('tabpanel', { name: /Criteria/ });
	const criteriaHeader = criteria.getByRole('columnheader', { name: /Name/i });
	await criteriaHeader.getByRole('button').click();
	await expect(criteriaHeader).toHaveAttribute('aria-sort', 'ascending');
	await expect(criteria.getByRole('row').nth(1)).toContainText('RESULT');
	await page.getByRole('tab', { name: 'Examples' }).click();
	const examples = page.getByRole('tabpanel', { name: 'Examples' });
	await examples
		.getByRole('columnheader', { name: /Script/i })
		.getByRole('button')
		.click();
	await expect(examples.getByRole('columnheader', { name: /Script/i })).toHaveAttribute(
		'aria-sort',
		'ascending'
	);
	for (const width of [1024, 390]) {
		await page.setViewportSize({ width, height: 900 });
		const geometry = await filters.evaluate((element) => {
			const right = element.getBoundingClientRect().right;
			const children = [...element.children].map((child) => child.getBoundingClientRect());
			const rowEnds = children.filter(
				(rect) =>
					!children.some(
						(other) => Math.abs(other.bottom - rect.bottom) < 1 && other.left > rect.left
					)
			);
			return {
				right,
				ends: rowEnds.map((rect) => rect.right),
				client: document.documentElement.clientWidth,
				scroll: document.documentElement.scrollWidth
			};
		});
		for (const right of geometry.ends) expect(Math.abs(right - geometry.right)).toBeLessThan(2);
		expect(geometry.scroll).toBeLessThanOrEqual(geometry.client);
	}
});

test('Functions catalog selects shared details beside the list and retains URLs', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/functions?function=1');
	const list = page.getByRole('complementary', { name: 'Functions list', exact: true });
	const details = page.getByRole('region', { name: 'Function details', exact: true });
	await expect(list).toBeVisible();
	await expect(list.getByRole('textbox', { name: 'Search functions' })).toBeVisible();
	await expect(list.getByRole('button', { name: 'Sort functions' })).toBeVisible();
	await expect(list.getByRole('button', { name: 'Sort direction' })).toBeVisible();
	await expect(
		details.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await expect(details.getByText('Function reference / WIFI_TEST', { exact: true })).toBeVisible();
	await expect(details.getByRole('button', { name: 'Edit function', exact: true })).toBeVisible();
	await expect(details.getByRole('button', { name: 'Compose script', exact: true })).toBeVisible();
	const listBounds = await list.boundingBox();
	const detailBounds = await details.boundingBox();
	expect(detailBounds!.x).toBeGreaterThanOrEqual(listBounds!.x + listBounds!.width);
	await list.getByRole('button', { name: 'Select RX_PER', exact: true }).click();
	await expect(page).toHaveURL(/function=2/);
	await expect(details.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	await page.goBack();
	await expect(
		details.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await page.goto('/functions/1');
	await expect(page.getByRole('heading', { name: /^TX_MULTI_VERIFICATION/ })).toBeVisible();
	await expect(page.getByRole('button', { name: 'Compose script', exact: true })).toBeVisible();
});

test('Functions filters use server parameters and restore history', async ({ page }) => {
	const { requests } = await mockFunctionCatalog(page);
	await page.goto(
		'/functions?search=transmit&source=DLL&category=Transmit&execution_role=TEST&execution_phase=RUN&function=1'
	);
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await expect
		.poll(() => Object.fromEntries(requests.at(-1) ?? []))
		.toMatchObject({
			search: 'transmit',
			source: 'DLL',
			category: 'Transmit',
			execution_role: 'TEST',
			execution_phase: 'RUN',
			page: '1',
			page_size: '20',
			sort: 'name',
			order: 'asc'
		});
	await page.getByRole('textbox', { name: 'Search functions', exact: true }).fill('unmatched');
	await expect(page).toHaveURL(/search=unmatched/);
	await expect(page.getByText('0 functions', { exact: true })).toBeVisible();
	await expect(page.getByRole('textbox', { name: 'Search functions', exact: true })).toBeFocused();
	await page.goBack();
	await expect(page.getByRole('textbox', { name: 'Search functions', exact: true })).toHaveValue(
		'transmit'
	);
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await page.getByRole('button', { name: 'Clear filters', exact: true }).click();
	await expect(page).not.toHaveURL(/search=|source=|category=|execution_role=|execution_phase=/);
});

test('Functions detail sorts links without changing their identity', async ({ page }) => {
	await mockFunctionCatalog(page);
	const links = ['ZETA', 'ALPHA'].map((name, index) => ({
		id: index + 1,
		name,
		position: index + 1,
		device_name: 'WIFI_TEST',
		required: true,
		function_source: 'DLL',
		function_name: 'TX_MULTI_VERIFICATION',
		resolved: true,
		definition_id: index + 1,
		source_file: '',
		created_at: '',
		updated_at: ''
	}));
	await page.route('**/api/functions/1/parameter-links', (route) =>
		route.fulfill({ json: { items: links } })
	);
	await page.goto('/functions?function=1');
	const panel = page.getByRole('tabpanel', { name: 'Parameters' });
	const nameHeader = panel.getByRole('columnheader', { name: /Name/i });
	await nameHeader.getByRole('button').click();
	await expect(nameHeader).toHaveAttribute('aria-sort', 'ascending');
	const firstRow = panel.getByRole('row').nth(1);
	await expect(firstRow).toContainText('ALPHA');
	await expect(firstRow.getByRole('cell').first()).toHaveText('2');
	await nameHeader.getByRole('button').click();
	await expect(nameHeader).toHaveAttribute('aria-sort', 'descending');
	await nameHeader.getByRole('button').click();
	await expect(nameHeader).not.toHaveAttribute('aria-sort');
	await expect(panel.getByRole('row').nth(1)).toContainText('ZETA');
	await nameHeader.getByRole('button').click();
	await page.getByRole('button', { name: 'Toggle dark mode' }).click();
	const checkbox = firstRow.getByRole('checkbox', { name: 'Toggle required for ALPHA' });
	await expect(checkbox).toBeChecked();
	const checkboxColors = await checkbox.evaluate((element) => ({
		background: getComputedStyle(element).backgroundColor,
		mark: getComputedStyle(element).color,
		icon: Boolean(element.querySelector('svg'))
	}));
	expect(checkboxColors.icon).toBe(true);
	expect(checkboxColors.background).not.toBe(checkboxColors.mark);
	await firstRow.getByRole('button', { name: 'Actions for ALPHA' }).click();
	await page.getByRole('menuitem', { name: 'Edit', exact: true }).click();
	await expect(panel.getByRole('spinbutton', { name: 'Position' })).toHaveValue('2');
});

test('Functions Device discovery filters complete pages and separates Global', async ({ page }) => {
	const { functions } = await mockFunctionCatalog(page);
	const rows = Array.from({ length: 100 }, (_, index) => ({
		...functions[0],
		id: index + 10,
		name: `EARLY_${index}`
	}));
	const pages: number[] = [];
	await page.route('**/api/devices/WIFI_TEST/functions?*', (route) => {
		const number = Number(new URL(route.request().url()).searchParams.get('page'));
		pages.push(number);
		return route.fulfill({
			json: {
				items: number === 1 ? rows : [functions[1]],
				total: 101,
				page: number,
				page_size: 100
			}
		});
	});
	await page.goto('/functions?device=WIFI_TEST&search=packet&category=Receive&function=2');
	await expect(page.getByText('1 functions', { exact: true })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	expect(pages).toEqual([1, 2]);
	await expect(
		page.getByRole('link', { name: 'Open Device workspace', exact: true })
	).toHaveAttribute('href', '/devices/WIFI_TEST?tab=functions');
	await page.getByRole('button', { name: /^Filter by Device:/ }).click();
	await page.getByRole('option', { name: 'Global definitions', exact: true }).click();
	await expect(page).toHaveURL(/device=_global/);
	await expect(
		page.getByRole('link', { name: 'Open Device workspace', exact: true })
	).toHaveAttribute('href', '/devices/_global?tab=functions');
});

test('Functions shared CRUD preserves context validation cancellation and confirmed deletion', async ({
	page
}) => {
	const { mutations } = await mockFunctionCatalog(page);
	await page.goto('/functions?source=DLL&function=1');
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	let dialog = page.getByRole('dialog', { name: 'New Function', exact: true });
	await expect(dialog.getByText('Global - shared across Devices', { exact: true })).toBeVisible();
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog.getByText('Name is required.', { exact: true })).toBeVisible();
	await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('NEW_SHARED_FN');
	await chooseFormDevice(page, 'WIFI_TEST');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	await expect(page).toHaveURL(/source=DLL/);
	await expect(
		page.getByRole('button', { name: 'Select NEW_SHARED_FN', exact: true })
	).toBeVisible();
	await page.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION', exact: true }).click();
	await page.getByRole('button', { name: 'Edit function', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'Edit Function #1', exact: true });
	await dialog.getByRole('textbox', { name: 'Description', exact: true }).fill('Cancelled edit');
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	expect(mutations.filter((call) => call.method === 'PUT')).toHaveLength(0);
	await page.getByRole('button', { name: 'Edit function', exact: true }).click();
	await dialog
		.getByRole('textbox', { name: 'Description', exact: true })
		.fill('Shared updated description');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(
		page
			.getByRole('region', { name: 'Function details', exact: true })
			.getByText('Shared updated description', { exact: true })
	).toBeVisible();
	expect(mutations.find((call) => call.method === 'PUT')?.path).toBe('/api/functions/1');
	await page.getByRole('button', { name: 'Delete function', exact: true }).click();
	const confirmation = page.getByRole('alertdialog', { name: 'Delete Function', exact: true });
	await expect(confirmation).toContainText('across all Devices');
	await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click();
	expect(mutations.filter((call) => call.method === 'DELETE')).toHaveLength(0);
	await page.getByRole('button', { name: 'Delete function', exact: true }).click();
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(
		page.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION', exact: true })
	).toHaveCount(0);
	await expect(page.getByRole('heading', { name: 'NEW_SHARED_FN', exact: true })).toBeVisible();
	await expect(page).not.toHaveURL(/function=1(?:&|$)/);
	expect(mutations.find((call) => call.method === 'DELETE')?.path).toBe('/api/functions/1');
});

test('Device Function actions edit shared identity and retain Composer and link management', async ({
	page
}) => {
	const { mutations } = await mockFunctionCatalog(page);
	await page.route('**/api/functions/1/compose', (route) =>
		route.fulfill({ json: { script: 'SERVER_EMITTED_SCRIPT' } })
	);
	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	const details = page.getByRole('region', { name: 'Function details', exact: true });
	await expect(
		details.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await details.getByRole('button', { name: 'Edit function', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Edit Function #1', exact: true });
	await dialog
		.getByRole('textbox', { name: 'Description', exact: true })
		.fill('Updated shared Function');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(details.getByText('Updated shared Function', { exact: true })).toBeVisible();
	await expect(page).toHaveURL(/WIFI_TEST\?tab=functions&function=1$/);
	expect(mutations.at(-1)?.path).toBe('/api/functions/1');
	expect(mutations.at(-1)?.body).not.toHaveProperty('device_name');
	await page.goto('/devices/GOLDENKV?tab=functions&function=1');
	await expect(details.getByText('Updated shared Function', { exact: true })).toBeVisible();
	await details.getByRole('button', { name: 'Compose script', exact: true }).click();
	await expect(page.getByText('SERVER_EMITTED_SCRIPT', { exact: true })).toBeVisible();
	await page.keyboard.press('Escape');
	await details.getByRole('link', { name: 'Manage function', exact: true }).click();
	await expect(page).toHaveURL(/\/functions\/1$/);
	await expect(page.getByRole('button', { name: 'Attach parameter', exact: true })).toBeVisible();
	await page.getByRole('tab', { name: 'Criteria', exact: true }).click();
	await expect(page.getByRole('button', { name: 'Attach criteria', exact: true })).toBeVisible();
	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	await details.getByRole('button', { name: 'Delete function', exact: true }).click();
	await page
		.getByRole('alertdialog', { name: 'Delete Function', exact: true })
		.getByRole('button', { name: 'Delete', exact: true })
		.click();
	await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	await expect(page).not.toHaveURL(/function=1(?:&|$)/);
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	const createDialog = page.getByRole('dialog', { name: 'New Function', exact: true });
	await createDialog
		.getByRole('textbox', { name: 'Name', exact: true })
		.fill('NEW_DEVICE_SHARED_FN');
	await expect(createDialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
		'WIFI_TEST'
	);
	await createDialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page).toHaveURL(/\/functions\/3$/);
	await expect(
		page.getByRole('heading', { name: 'NEW_DEVICE_SHARED_FN', exact: true })
	).toBeVisible();
	expect(mutations.find((call) => call.method === 'POST')?.body).not.toHaveProperty('device_name');
	await page.goBack();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=functions$/);
	await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	await expect(
		page.getByRole('button', { name: 'Select NEW_DEVICE_SHARED_FN', exact: true })
	).toBeVisible();
});

test('Functions catalog ignores a late response after changing filters', async ({ page }) => {
	const { functions } = await mockFunctionCatalog(page);
	let releaseOld!: () => void;
	let finishOld!: () => void;
	const blocked = new Promise<void>((resolve) => {
		releaseOld = resolve;
	});
	const completed = new Promise<void>((resolve) => {
		finishOld = resolve;
	});
	await page.route('**/api/functions?*', async (route) => {
		const isOld = !new URL(route.request().url()).searchParams.has('source');
		if (isOld) await blocked;
		await route.fulfill({
			json: { items: [functions[isOld ? 0 : 1]], total: 1, page: 1, page_size: 20 }
		});
		if (isOld) finishOld();
	});
	await page.goto('/functions');
	await page.getByRole('button', { name: 'Filter by source', exact: true }).click();
	await page.getByRole('option', { name: 'BUILT-IN', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	releaseOld();
	await completed;
	await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toHaveCount(0);
});

test('Functions catalog retries failed discovery without stale details', async ({ page }) => {
	const { functions } = await mockFunctionCatalog(page);
	let attempts = 0;
	await page.route('**/api/functions?*', (route) => {
		attempts++;
		return attempts === 1
			? route.fulfill({ status: 503, json: { error: 'Catalog unavailable' } })
			: route.fulfill({ json: { items: [functions[0]], total: 1, page: 1, page_size: 20 } });
	});
	await page.goto('/functions?function=1');
	const railCount = page
		.getByRole('complementary', { name: 'Functions list', exact: true })
		.locator('.rail-heading span');
	await expect(page.getByRole('alert')).toContainText('Catalog unavailable');
	await expect(page.getByText('Count unavailable', { exact: true })).toBeVisible();
	await expect(page.getByText('0 functions', { exact: true })).toHaveCount(0);
	await expect(railCount).toHaveCount(0);
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toHaveCount(0);
	await page.getByRole('button', { name: 'Retry functions', exact: true }).click();
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await expect(page.getByText('1 functions', { exact: true })).toBeVisible();
	await expect(railCount).toHaveText('1');
	expect(attempts).toBe(2);
});

test('Functions catalog pagination uses real totals and restores URLs', async ({ page }) => {
	const { functions, requests } = await mockFunctionCatalog(page);
	functions.push(
		...Array.from({ length: 30 }, (_, index) => ({
			...functions[0],
			id: index + 3,
			name: `PAGED_${index}`
		}))
	);
	await page.goto('/functions');
	await expect(page.getByText('32 functions', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Next page', exact: true }).click();
	await expect(page).toHaveURL(/page=2/);
	await expect(page.getByRole('heading', { name: 'PAGED_18', exact: true })).toBeVisible();
	expect(requests.at(-1)?.get('page')).toBe('2');
	await page.goBack();
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await page.getByRole('button', { name: 'Page size', exact: true }).click();
	await page.getByRole('option', { name: '50 / page', exact: true }).click();
	await expect(page).toHaveURL(/page_size=50/);
	await expect.poll(() => requests.at(-1)?.get('page_size')).toBe('50');
	await expect(
		page
			.getByRole('complementary', { name: 'Functions list', exact: true })
			.getByText('Page 1 of 1', { exact: true })
	).toBeVisible();
});

test('Functions catalog preserves literal encoded Device names', async ({ page }) => {
	const { functions } = await mockFunctionCatalog(page);
	const paths: string[] = [];
	await page.route('**/api/devices/*/functions?*', (route) => {
		paths.push(new URL(route.request().url()).pathname);
		return route.fulfill({ json: { items: [functions[0]], total: 1, page: 1, page_size: 100 } });
	});
	await page.goto('/functions?device=MC2+%2F+LAB%2520&function=1');
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	expect(paths).toEqual(['/api/devices/MC2%20%2F%20LAB%2520/functions']);
	await expect(
		page.getByRole('link', { name: 'Open Device workspace', exact: true })
	).toHaveAttribute('href', '/devices/MC2%20%2F%20LAB%2520?tab=functions');
});

test('Functions catalog layout handles long content desktop mobile and dark mode', async ({
	page
}, testInfo) => {
	const { functions } = await mockFunctionCatalog(page);
	functions[0].name = `TX_${'LONG_IDENTIFIER_'.repeat(14)}`;
	functions[0].interface = `WIFI_${'LONG_INTERFACE_'.repeat(12)}`;
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	for (const width of [1440, 960, 390]) {
		await page.setViewportSize({ width, height: 900 });
		await page.goto('/functions?function=1');
		await expect(page.getByRole('heading', { name: functions[0].name, exact: true })).toBeVisible();
		const geometry = await page.evaluate(() => ({
			viewport: window.innerWidth,
			client: document.documentElement.clientWidth,
			scroll: document.documentElement.scrollWidth,
			filterWidths: [...document.querySelectorAll('.filters .filter-field')].map((field) => ({
				field: field.getBoundingClientRect().width,
				control: field.querySelector('button')!.getBoundingClientRect().width
			}))
		}));
		expect(geometry.scroll).toBeLessThanOrEqual(geometry.viewport);
		for (const item of geometry.filterWidths)
			expect(Math.abs(item.field - item.control)).toBeLessThan(2);
		const list = await page
			.getByRole('complementary', { name: 'Functions list', exact: true })
			.boundingBox();
		const detail = await page
			.getByRole('region', { name: 'Function details', exact: true })
			.boundingBox();
		if (width > 850) expect(detail!.x).toBeGreaterThanOrEqual(list!.x + list!.width);
		else expect(detail!.y).toBeGreaterThanOrEqual(list!.y + list!.height);
		await page.screenshot({ path: testInfo.outputPath(`functions-${width}.png`), fullPage: true });
	}
	await page.getByRole('button', { name: 'Open navigation menu', exact: true }).click();
	const nav = page.getByRole('menu');
	for (const name of [
		'Devices',
		'Functions',
		'Parameters',
		'Criteria',
		'Examples',
		'Sequences',
		'Orphans',
		'Import'
	])
		await expect(nav.getByRole('menuitem', { name, exact: true })).toHaveCount(1);
	await page.keyboard.press('Escape');
	await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
	await expect(page.locator('html')).toHaveClass(/dark/);
	await expect(page.getByRole('button', { name: 'Toggle table density', exact: true })).toHaveCount(
		0
	);
	await page.screenshot({ path: testInfo.outputPath('functions-mobile-dark.png'), fullPage: true });
	await page.getByRole('button', { name: 'New function', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'New Function', exact: true });
	await expect(dialog).toBeVisible();
	await dialog.getByRole('button', { name: 'Save', exact: true }).scrollIntoViewIfNeeded();
	const bounds = await dialog.boundingBox();
	expect(bounds!.x).toBeGreaterThanOrEqual(0);
	expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(390);
	await page.screenshot({
		path: testInfo.outputPath('function-dialog-mobile-dark.png'),
		fullPage: true
	});
	await page.keyboard.press('Escape');
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.screenshot({
		path: testInfo.outputPath('functions-desktop-dark.png'),
		fullPage: true
	});
	expect(errors).toEqual([]);
});

test('Functions deletion keeps errors retryable and handles the last selected result', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	let attempts = 0;
	await page.route('**/api/functions/1', (route) => {
		if (route.request().method() === 'DELETE' && ++attempts === 1)
			return route.fulfill({ status: 503, json: { error: 'Delete unavailable' } });
		return route.fallback();
	});
	await page.goto('/functions?source=DLL&function=1');
	await page.getByRole('button', { name: 'Delete function', exact: true }).click();
	const confirmation = page.getByRole('alertdialog', { name: 'Delete Function', exact: true });
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(confirmation.getByRole('alert')).toContainText('Delete unavailable');
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(confirmation).toHaveCount(0);
	await expect(page.getByText('0 functions', { exact: true })).toBeVisible();
	await expect(page.getByText('No Function selected.', { exact: true })).toBeVisible();
	await expect(page).not.toHaveURL(/function=1(?:&|$)/);
	expect(attempts).toBe(2);
});

test('app opens all Devices rather than a last-used device', async ({ page }) => {
	await page.route('**/api/devices', (route) =>
		route.fulfill({
			json: {
				items: [
					{ device_name: '', functions: 0, parameters: 1, criteria: 0, examples: 0 },
					{ device_name: 'WIFI_TEST', functions: 2, parameters: 3, criteria: 2, examples: 4 },
					{ device_name: 'GPIB', functions: 0, parameters: 0, criteria: 0, examples: 0 }
				],
				total: 3
			}
		})
	);
	await page.goto('/');
	await expect(page).toHaveURL(/\/devices$/, { timeout: 20_000 });
	await expect(page.getByRole('heading', { name: 'Devices', exact: true })).toBeVisible();
	await expect(page.getByRole('link', { name: 'WIFI_TEST', exact: true })).toBeVisible();
	await expect(page.getByRole('link', { name: 'GPIB', exact: true })).toBeVisible();
	await expect(page.getByRole('region', { name: 'Global definitions' })).toBeVisible();
	await page.getByRole('textbox', { name: 'Search devices' }).fill('wifi');
	await expect(page.getByRole('article', { name: 'WIFI_TEST' })).toBeVisible();
	await expect(page.getByRole('article', { name: 'GPIB' })).toHaveCount(0);
	await page.getByRole('textbox', { name: 'Search devices' }).fill('unmatched');
	await expect(page.getByText('No devices match your search.')).toBeVisible();
});

test('Devices and Functions provide prototype navigation on mobile', async ({ page }) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 390, height: 844 });
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Open navigation menu' }).click();
	const functions = page.getByRole('menuitem', { name: 'Functions', exact: true });
	await expect(functions).toHaveAttribute('href', '/functions');
	await functions.click();
	await expect(page).toHaveURL(/\/functions$/);
	await expect
		.poll(() =>
			page.evaluate(() =>
				getComputedStyle(document.documentElement).getPropertyValue('--device-accent').trim()
			)
		)
		.toBe('#181a19');
	await page.getByRole('button', { name: 'Open navigation menu' }).click();
	await page.getByRole('menuitem', { name: 'Parameters', exact: true }).click();
	await expect(page).toHaveURL(/\/parameter-definitions$/);
	await expect
		.poll(() =>
			page.evaluate(() =>
				getComputedStyle(document.documentElement).getPropertyValue('--device-accent').trim()
			)
		)
		.toBe('#181a19');
});

test('Device directory follows Prototype E visual hierarchy', async ({ page }, testInfo) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices');
	const tile = page.getByRole('article', { name: 'WIFI_TEST', exact: true });
	await expect(tile).toBeVisible({ timeout: 20_000 });
	const name = tile.getByRole('heading', { name: 'WIFI_TEST', exact: true });
	await expect(name).toHaveCSS('font-size', '16px');
	await expect(tile.locator('.device-icon')).toBeVisible();
	const tiles = await page.getByRole('article').all();
	expect(tiles).toHaveLength(4);
	const rows = await Promise.all(tiles.map((item) => item.boundingBox()));
	expect(rows.filter((item) => item!.y === rows[0]!.y)).toHaveLength(3);
	expect(new Set(rows.map((item) => item!.y)).size).toBe(2);
	const main = await page.getByRole('main').boundingBox();
	expect(main!.width).toBeGreaterThan(1280);
	await page.screenshot({ path: testInfo.outputPath('device-cards-desktop.png'), fullPage: true });
	await page.setViewportSize({ width: 390, height: 844 });
	const geometry = await page.evaluate(() => ({
		viewport: window.innerWidth,
		scroll: document.documentElement.scrollWidth
	}));
	expect(geometry.scroll).toBeLessThanOrEqual(geometry.viewport);
	await page.screenshot({ path: testInfo.outputPath('device-cards-mobile.png'), fullPage: true });
});

test('Device directory preserves sort, count links and create/delete workflows', async ({
	page
}) => {
	await mockDeviceHub(page);
	let devices: DeviceSummary[] = [
		{ device_name: 'WIFI_TEST', functions: 2, parameters: 1, criteria: 1, examples: 1 },
		{ device_name: 'GOLDENKV', functions: 1, parameters: 1, criteria: 1, examples: 1 },
		{ device_name: '', functions: 2, parameters: 3, criteria: 0, examples: 1 }
	];
	await page.route('**/api/devices', async (route) => {
		if (route.request().method() === 'POST') {
			expect(route.request().postDataJSON()).toEqual({
				device_name: ' new_device ',
				source: 'DLL'
			});
			devices = [
				...devices,
				{
					device_name: 'NEW_DEVICE',
					source: 'DLL',
					sources: ['DLL'],
					functions: 0,
					parameters: 0,
					criteria: 0,
					examples: 0
				}
			];
			await route.fulfill({ status: 201, json: { device_name: 'NEW_DEVICE' } });
		} else {
			expect(route.request().method()).toBe('GET');
			await route.fulfill({ json: { items: devices, total: devices.length } });
		}
	});
	await page.route('**/api/devices/NEW_DEVICE', async (route) => {
		expect(route.request().method()).toBe('DELETE');
		expect(route.request().postDataJSON()).toEqual({ confirmation: 'DELETE NEW_DEVICE' });
		devices = devices.filter((item) => item.device_name !== 'NEW_DEVICE');
		await route.fulfill({ status: 204 });
	});
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Sort devices', exact: true }).click();
	await page.getByRole('option', { name: 'Functions', exact: true }).click();
	await page.getByRole('button', { name: 'Sort order', exact: true }).click();
	await page.getByRole('option', { name: 'Descending', exact: true }).click();
	await expect(page.getByRole('article').first()).toHaveAccessibleName('WIFI_TEST');
	await expect(
		page.getByRole('article', { name: 'WIFI_TEST' }).getByRole('button', { name: 'Edit WIFI_TEST' })
	).toBeVisible();
	await expect(
		page
			.getByRole('article', { name: 'WIFI_TEST' })
			.getByRole('button', { name: 'Delete WIFI_TEST' })
	).toBeVisible();
	await page.getByRole('textbox', { name: 'Search devices' }).fill('gold');
	await expect(page.getByRole('article')).toHaveCount(1);
	await expect(page.getByRole('region', { name: 'Global definitions' })).toBeVisible();
	await page.getByRole('article').getByRole('link', { name: '1 parameters' }).click();
	await expect(page).toHaveURL(/\/devices\/GOLDENKV\?tab=parameters$/);
	await expect(page.getByRole('tab', { name: 'Parameters', exact: true })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Create device', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Create Device', exact: true });
	await expect(dialog.getByRole('button', { name: 'Type', exact: true })).toHaveText('BUILT-IN');
	await dialog.getByRole('button', { name: 'Type', exact: true }).click();
	await expect(page.getByRole('listbox').getByRole('option')).toHaveCount(2);
	await page.getByRole('option', { name: 'DLL', exact: true }).click();
	await dialog.getByRole('textbox', { name: 'Device name' }).fill(' new_device ');
	await expect(dialog.getByText('Will be created as NEW_DEVICE')).toBeVisible();
	await dialog.getByRole('button', { name: 'Create', exact: true }).click();
	await expect(dialog).toBeHidden();
	const created = page.getByRole('article', { name: 'NEW_DEVICE', exact: true });
	await expect(created.getByRole('img', { name: 'DLL type', exact: true })).toBeVisible();
	await page.reload();
	await expect(created.getByText('DLL', { exact: true })).toBeVisible();
	await expect(created.getByRole('link', { name: /^0 / })).toHaveCount(4);
	await created.getByRole('button', { name: 'Delete NEW_DEVICE', exact: true }).click();
	const confirmation = page.getByRole('dialog', { name: 'Delete NEW_DEVICE?' });
	await expect(confirmation).toContainText('Delete NEW_DEVICE?');
	await expect(confirmation.getByRole('button', { name: 'Delete', exact: true })).toBeDisabled();
	await confirmation.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(created).toBeVisible();
	await created.getByRole('button', { name: 'Delete NEW_DEVICE', exact: true }).click();
	await confirmation.getByLabel('Type DELETE NEW_DEVICE to confirm').fill('DELETE NEW_DEVICE');
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(created).toHaveCount(0);
	await expect(page.getByRole('region', { name: 'Global definitions' })).toBeVisible();
});

test('Device Type labels and icons distinguish native, DLL and mixed Devices', async ({
	page
}, testInfo) => {
	await mockDeviceHub(page);
	await page.goto('/devices');
	const native = page.getByRole('article', { name: 'GOLDENKV', exact: true });
	const dll = page.getByRole('article', { name: 'WIFI_TEST', exact: true });
	await expect(native.getByText('BUILT-IN', { exact: true })).toBeVisible();
	await expect(dll.getByText('DLL', { exact: true })).toBeVisible();
	await expect(
		page
			.getByRole('article', { name: 'OSP', exact: true })
			.getByText('BUILT-IN / DLL', { exact: true })
	).toBeVisible();
	const nativeIcon = native.getByRole('img', { name: 'BUILT-IN type', exact: true });
	const dllIcon = dll.getByRole('img', { name: 'DLL type', exact: true });
	expect(await nativeIcon.innerHTML()).not.toBe(await dllIcon.innerHTML());
	await page.setViewportSize({ width: 390, height: 844 });
	await page.getByRole('button', { name: 'Create device', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Create Device', exact: true });
	await expect(dialog.getByRole('button', { name: 'Type', exact: true })).toBeVisible();
	await dialog.getByRole('button', { name: 'Type', exact: true }).click();
	await expect(page.getByRole('option', { name: 'BUILT-IN', exact: true })).toBeVisible();
	await expect(page.getByRole('option', { name: 'DLL', exact: true })).toBeVisible();
	const geometry = await page.evaluate(() => ({
		viewport: window.innerWidth,
		scroll: document.documentElement.scrollWidth
	}));
	expect(geometry.scroll).toBeLessThanOrEqual(geometry.viewport);
	await page.screenshot({ path: testInfo.outputPath('device-type-mobile.png'), fullPage: true });
});

test('device detail shows its source type without duplicate function metadata', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST');

	const typeLabel = page.locator('.device-identity .device-label');
	await expect(typeLabel).toHaveText('DLL');
	await expect(page.getByRole('img', { name: 'DLL type' })).toBeVisible();

	const functionDetail = page.locator('.function-detail');
	await expect(
		functionDetail.getByRole('heading', { name: 'TX_MULTI_VERIFICATION' })
	).toBeVisible();
	await expect(functionDetail.locator('p.metadata')).toHaveCount(0);

	await page.goto('/devices/OSP');
	await expect(typeLabel).toHaveText('BUILT-IN / DLL');
	await expect(page.getByRole('img', { name: 'BUILT-IN / DLL type' })).toBeVisible();

	await page.goto('/devices/GOLDENKV');
	await expect(page.getByRole('img', { name: 'BUILT-IN type' })).toBeVisible();
});

test('Device parameter and criteria rows select details across the full row', async ({ page }) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=parameters');

	const detail = page.locator('.definition-detail');
	await expect(detail.locator('.ownership')).toHaveText('Parameter definition');
	await expect(detail.getByRole('heading', { name: 'ANT', exact: true })).toBeVisible();

	const globalParameter = page.getByRole('row').filter({ hasText: 'SETTLE_MS' }).last();
	await globalParameter.getByRole('cell').nth(1).click();
	await expect(page).toHaveURL(/definition=12/);
	await expect(detail.getByRole('heading', { name: 'SETTLE_MS', exact: true })).toBeVisible();
	await expect(detail.locator('.ownership')).toHaveText('Global - shared across Devices');

	await page.goto('/devices/WIFI_TEST?tab=criteria');
	const criterion = page.getByRole('row').filter({ hasText: 'POWER' }).last();
	await criterion.getByRole('cell').nth(2).click();
	await expect(page).toHaveURL(/definition=23/);
	await expect(detail.getByRole('heading', { name: 'POWER', exact: true })).toBeVisible();
	await expect(detail.locator('.ownership')).toHaveText('Criteria definition');
});

test('Device Examples Script and Meaning columns have equal widths', async ({ page }) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=examples');
	const columns = page.locator('table:visible').last().locator('thead th');
	await expect(columns).toHaveCount(4);

	for (const width of [1280, 390]) {
		await page.setViewportSize({ width, height: 844 });
		const columnWidths = await columns.evaluateAll((headers) =>
			headers.map((header) => header.getBoundingClientRect().width)
		);
		expect(Math.abs(columnWidths[1] - columnWidths[2])).toBeLessThanOrEqual(1);
	}
});

test('Device rename preserves counts and keeps conflict errors retryable', async ({
	page
}, testInfo) => {
	await mockDeviceHub(page);
	let devices = [
		{ device_name: 'WIFI_TEST', functions: 2, parameters: 3, criteria: 1, examples: 4 },
		{ device_name: 'GPIB', functions: 0, parameters: 0, criteria: 0, examples: 0 },
		{ device_name: '', functions: 0, parameters: 1, criteria: 0, examples: 0 }
	];
	const renames: string[] = [];
	await page.route('**/api/devices', (route) =>
		route.fulfill({ json: { items: devices, total: devices.length } })
	);
	await page.route('**/api/devices/WIFI_TEST', (route) => {
		expect(route.request().method()).toBe('PUT');
		const name = route.request().postDataJSON().device_name.trim().toUpperCase();
		renames.push(name);
		if (name === 'GPIB')
			return route.fulfill({
				status: 409,
				json: { error: 'A device with that name already exists.' }
			});
		devices = devices.map((row) =>
			row.device_name === 'WIFI_TEST' ? { ...row, device_name: name } : row
		);
		return route.fulfill({ json: { device_name: name } });
	});
	await page.goto('/devices');
	const tile = page.getByRole('article', { name: 'WIFI_TEST', exact: true });
	await tile.getByRole('button', { name: 'Edit WIFI_TEST', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Edit WIFI_TEST', exact: true });
	await expect(dialog.getByRole('button', { name: 'Save', exact: true })).toBeDisabled();
	await dialog.getByRole('textbox', { name: 'Device name' }).fill('CANCELLED');
	await dialog.getByRole('button', { name: 'Cancel' }).click();
	expect(renames).toHaveLength(0);
	await tile.getByRole('button', { name: 'Edit WIFI_TEST', exact: true }).click();
	await dialog.getByRole('textbox', { name: 'Device name' }).fill('GPIB');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog.getByRole('alert')).toContainText('already exists');
	await expect(tile).toBeVisible();
	await dialog.getByRole('textbox', { name: 'Device name' }).fill(' renamed_device ');
	await page.setViewportSize({ width: 390, height: 844 });
	const bounds = await dialog.boundingBox();
	expect(bounds!.x).toBeGreaterThanOrEqual(0);
	expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(390);
	await page.screenshot({ path: testInfo.outputPath('rename-device-mobile.png'), fullPage: true });
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	const renamed = page.getByRole('article', { name: 'RENAMED_DEVICE', exact: true });
	await expect(renamed).toBeVisible();
	await expect(renamed.getByRole('link', { name: '3 parameters' })).toBeVisible();
	await expect(renamed.getByRole('button', { name: 'Edit RENAMED_DEVICE' })).toBeVisible();
	await expect(renamed.getByRole('button', { name: 'Delete RENAMED_DEVICE' })).toHaveCount(0);
	const empty = page.getByRole('article', { name: 'GPIB', exact: true });
	const edit = empty.getByRole('button', { name: 'Edit GPIB' });
	const remove = empty.getByRole('button', { name: 'Delete GPIB' });
	const editBounds = await edit.boundingBox();
	const deleteBounds = await remove.boundingBox();
	const countBounds = await empty.getByRole('link', { name: '0 functions' }).boundingBox();
	expect(deleteBounds!.x).toBeGreaterThanOrEqual(editBounds!.x + editBounds!.width);
	expect(deleteBounds!.y + deleteBounds!.height).toBeLessThan(countBounds!.y);
	await expect(
		page.getByRole('region', { name: 'Global definitions' }).getByRole('button')
	).toHaveCount(0);
});

test('keyboard users can skip navigation to the page content', async ({ page }) => {
	await page.goto('/import');
	await expect(page.getByRole('heading', { name: 'Workbook import' })).toBeVisible();
	await page.keyboard.press('Tab');
	const skipLink = page.getByRole('link', { name: 'Skip to main content' });
	await expect(skipLink).toBeFocused();
	await page.keyboard.press('Enter');
	await expect(page.getByRole('main')).toBeFocused();
});

test('device hub browses functions beside scoped details and separate Global references', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices');
	await expect(page.getByRole('article', { name: 'WIFI_TEST', exact: true })).toBeVisible({
		timeout: 20_000
	});
	await page.screenshot({ path: test.info().outputPath('devices-desktop.png'), fullPage: true });
	await page.goto('/devices/WIFI_TEST');
	await expect(page.getByRole('heading', { name: 'WIFI_TEST', exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION', exact: true }).click();
	const details = page.getByRole('region', { name: 'Function details' });
	await expect(
		details.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await expect(details.getByRole('region', { name: 'WIFI_TEST parameters' })).toContainText('ANT');
	await expect(details.getByRole('region', { name: 'Global parameters' })).toContainText(
		'SETTLE_MS'
	);
	await expect(details.getByRole('region', { name: 'Global parameters' })).toContainText('TIMEOUT');
	await expect(
		details.getByRole('region', { name: 'Global parameters' }).getByRole('row', { name: /^ANT/ })
	).toHaveCount(0);
	await expect(
		details
			.getByRole('region', { name: 'WIFI_TEST parameters' })
			.getByRole('row', { name: /^UNRESOLVED_FIELD/ })
	).toContainText('Unresolved');
	await expect(details).not.toContainText('OTHER_DEVICE');
	await expect(details.getByText('Device transmit example', { exact: true })).toBeVisible();
	await expect(details.getByText('Shared transmit example', { exact: true })).toBeVisible();
	await page.screenshot({ path: test.info().outputPath('device-hub-desktop.png'), fullPage: true });
	await expect(page).toHaveURL(/function=1/);
	await page.getByRole('textbox', { name: 'Search device functions' }).fill('RX_PER');
	await expect(
		page.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION', exact: true })
	).toHaveCount(0);
	await expect(page.getByRole('button', { name: 'Select RX_PER', exact: true })).toBeVisible();
});

test('hybrid hub retains prototype typography and section navigation', async ({ page }) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices/WIFI_TEST?function=1');
	const identity = page.getByRole('heading', { name: 'WIFI_TEST', exact: true });
	await expect(identity).toHaveCSS('font-size', '28px');
	await expect(identity).toHaveCSS('font-family', /Cascadia Code/);
	const tab = page.getByRole('tab', { name: 'Functions', exact: true });
	await expect(tab).toHaveCSS('border-bottom-color', 'rgb(8, 127, 112)');
	await expect(tab).toHaveCSS('border-bottom-width', '2px');
	await expect(tab).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
	const details = page.getByRole('region', { name: 'Function details' });
	await expect(details.getByRole('heading', { name: 'TX_MULTI_VERIFICATION' })).toHaveCSS(
		'font-family',
		/Cascadia Code/
	);
	await expect(details.getByText('Device transmit example', { exact: true })).toBeVisible();
	const badge = details
		.getByRole('region', { name: 'WIFI_TEST parameters' })
		.getByRole('heading')
		.locator('span')
		.first();
	const contrast = await badge.evaluate((element) => {
		const style = getComputedStyle(element);
		const luminance = (color: string) => {
			const channels = color
				.match(/[\d.]+/g)!
				.slice(0, 3)
				.map(Number)
				.map((channel) => channel / 255)
				.map((channel) =>
					channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4
				);
			return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
		};
		const values = [luminance(style.color), luminance(style.backgroundColor)].sort(
			(first, second) => second - first
		);
		return (values[0] + 0.05) / (values[1] + 0.05);
	});
	expect(contrast).toBeGreaterThanOrEqual(4.5);
	const list = await page.getByRole('complementary', { name: 'Device functions' }).boundingBox();
	const detail = await details.boundingBox();
	expect(detail!.x).toBeGreaterThanOrEqual(list!.x + list!.width);
});

test('device directory and function hub fit a mobile viewport', async ({ page }) => {
	await mockDeviceHub(page);
	await page.setViewportSize({ width: 390, height: 844 });
	await page.goto('/devices');
	await expect(page.getByRole('article', { name: 'WIFI_TEST' })).toBeVisible({ timeout: 20_000 });
	expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
	await page.screenshot({ path: test.info().outputPath('devices-mobile.png'), fullPage: true });
	await page.getByRole('link', { name: 'WIFI_TEST', exact: true }).click();
	await expect(page.getByRole('region', { name: 'Function details' })).toBeVisible();
	await expect(page.getByText('Device transmit example', { exact: true })).toBeVisible();
	expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
	await page.screenshot({ path: test.info().outputPath('device-hub-mobile.png'), fullPage: true });
});

test('prototype styling handles large catalogs, long content and dark mode', async ({ page }) => {
	const functionName = 'FUNCTION_WITH_A_LONG_UNBROKEN_IDENTIFIER_FOR_TRANSMIT_VERIFICATION';
	await mockDeviceHub(page, functionName);
	const devices = Array.from({ length: 32 }, (_, index) => ({
		device_name:
			index === 0 ? 'WIFI_TEST' : `DEVICE_WITH_A_LONG_UNBROKEN_IDENTIFIER_${index}_REFERENCE`,
		functions: 101,
		parameters: 24,
		criteria: 12,
		examples: 48
	}));
	await page.route('**/api/devices', (route) =>
		route.fulfill({
			json: {
				items: [
					...devices,
					{ device_name: '', functions: 2, parameters: 3, criteria: 0, examples: 1 }
				],
				total: 33
			}
		})
	);
	await page.route('**/api/examples?*', (route) => {
		const device = new URL(route.request().url()).searchParams.get('device') ?? '';
		return route.fulfill({
			json: {
				items: [
					{
						id: device ? 31 : 32,
						function_source: 'DLL',
						function_name: functionName,
						device_name: device,
						script: `TX|VALUE=${'LONG_UNBROKEN_SCRIPT_VALUE_'.repeat(20)}`,
						meaning: 'Long script example',
						function_exists: true,
						created_at: '',
						updated_at: ''
					}
				],
				total: 1,
				page: 1,
				page_size: 20
			}
		});
	});
	for (const viewport of [
		{ width: 1440, height: 900 },
		{ width: 960, height: 900 },
		{ width: 390, height: 844 }
	]) {
		await page.setViewportSize(viewport);
		await page.goto('/devices');
		await expect(page.getByRole('article')).toHaveCount(32);
		await expect(page.getByRole('region', { name: 'Global definitions' })).toBeVisible();
		const tiles = page.getByRole('article');
		const firstTiles = await Promise.all(
			[0, 1, 2, 3].map((index) => tiles.nth(index).boundingBox())
		);
		expect(firstTiles.filter((bounds) => bounds!.y === firstTiles[0]!.y)).toHaveLength(
			viewport.width === 1440 ? 4 : viewport.width === 960 ? 2 : 1
		);
		await expect(tiles.first().getByRole('link', { name: /^\d+ / })).toHaveCount(4);
		for (const count of await tiles.first().getByRole('link', { name: /^\d+ / }).all()) {
			await expect(count).toBeVisible();
		}
		expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
			viewport.width
		);
		await page.screenshot({
			path: test.info().outputPath(`large-directory-${viewport.width}.png`),
			fullPage: false
		});
		await page.goto('/devices/WIFI_TEST?function=1');
		const details = page.getByRole('region', { name: 'Function details' });
		await expect(details.getByRole('heading', { name: functionName })).toBeVisible();
		await expect(details.getByText('Long script example', { exact: true }).first()).toBeVisible();
		expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
			viewport.width
		);
		await page.screenshot({
			path: test.info().outputPath(`long-content-hub-${viewport.width}.png`),
			fullPage: true
		});
		await page.getByRole('button', { name: 'Device: WIFI_TEST', exact: true }).click();
		const popup = page.getByRole('dialog', { name: 'Choose Device', exact: true });
		const search = popup.getByRole('combobox', { name: 'Search devices', exact: true });
		await expect(search).toBeFocused();
		await expect(popup.getByRole('option')).toHaveCount(33);
		const searchBounds = await search.boundingBox();
		const listBounds = await popup.getByRole('listbox').boundingBox();
		expect(searchBounds!.y + searchBounds!.height).toBeLessThanOrEqual(listBounds!.y);
		const scrollable = popup.getByRole('listbox').locator('> div');
		expect(
			await scrollable.evaluate((element) => element.scrollHeight > element.clientHeight)
		).toBe(true);
		const last = popup.getByRole('option', {
			name: 'DEVICE_WITH_A_LONG_UNBROKEN_IDENTIFIER_31_REFERENCE',
			exact: true
		});
		await last.scrollIntoViewIfNeeded();
		await expect(last).toBeVisible();
		const global = popup.getByRole('option', { name: 'Global definitions', exact: true });
		await global.scrollIntoViewIfNeeded();
		await expect(global).toBeVisible();
		await search.fill('identifier_31');
		await expect(popup.getByRole('option')).toHaveCount(1);
		expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
			viewport.width
		);
		await page.screenshot({ path: test.info().outputPath(`device-picker-${viewport.width}.png`) });
		await popup.getByRole('button', { name: 'Clear device search' }).click();
		await expect(popup.getByRole('option')).toHaveCount(33);
		await page.keyboard.press('Escape');
	}
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Toggle dark mode' }).click();
	const tile = page.getByRole('article', { name: 'WIFI_TEST', exact: true });
	await expect(tile.locator('svg').first()).toHaveCSS('color', 'rgb(119, 217, 188)');
	await page.screenshot({
		path: test.info().outputPath('device-directory-dark.png'),
		fullPage: true
	});
	await page.getByRole('link', { name: 'WIFI_TEST', exact: true }).click();
	await expect(page.getByRole('tab', { name: 'Functions', exact: true })).toHaveCSS(
		'border-bottom-color',
		'rgb(119, 217, 188)'
	);
	await expect(
		page
			.getByRole('region', { name: 'Function details' })
			.getByText('Long script example', { exact: true })
			.first()
	).toBeVisible();
	await page.screenshot({ path: test.info().outputPath('device-hub-dark.png'), fullPage: true });
	await page.getByRole('button', { name: 'Device: WIFI_TEST', exact: true }).click();
	const popup = page.getByRole('dialog', { name: 'Choose Device', exact: true });
	await expect(popup.getByRole('option', { name: 'WIFI_TEST', exact: true })).toHaveAttribute(
		'data-selected',
		''
	);
	await page.screenshot({ path: test.info().outputPath('device-picker-dark.png') });
});

test('Device selector searches inside its popup and restores keyboard focus', async ({ page }) => {
	const pageErrors: string[] = [];
	page.on('pageerror', (error) => pageErrors.push(error.message));
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=criteria');
	const trigger = page.getByRole('button', { name: 'Device: WIFI_TEST', exact: true });
	await expect(trigger.locator('..').locator(':scope > span')).toHaveCount(0);
	await trigger.focus();
	await page.keyboard.press('Enter');
	const popup = page.getByRole('dialog', { name: 'Choose Device', exact: true });
	const search = popup.getByRole('combobox', { name: 'Search devices', exact: true });
	await expect(search).toBeFocused();
	await expect(popup.getByRole('option')).toHaveCount(5);
	await expect(popup.getByRole('option', { name: 'WIFI_TEST', exact: true })).toHaveAttribute(
		'data-selected',
		''
	);
	await search.fill('no-such-device');
	await expect(popup.getByText('No devices match your search.')).toBeVisible();
	await popup.getByRole('button', { name: 'Clear device search' }).click();
	await expect(search).toBeFocused();
	await expect(popup.getByRole('option')).toHaveCount(5);
	await search.fill('g');
	await expect(popup.getByRole('option')).toHaveCount(3);
	await page.keyboard.press('ArrowDown');
	const firstMatch = await search.getAttribute('aria-activedescendant');
	expect(firstMatch).toBeTruthy();
	await page.keyboard.press('ArrowDown');
	await expect(search).not.toHaveAttribute('aria-activedescendant', firstMatch!);
	await page.keyboard.press('ArrowUp');
	await expect(search).toHaveAttribute('aria-activedescendant', firstMatch!);
	await search.fill('gOlD');
	await expect(popup.getByRole('option')).toHaveCount(1);
	await page.keyboard.press('ArrowDown');
	await page.keyboard.press('Enter');
	const selected = page.getByRole('button', { name: 'Device: GOLDENKV', exact: true });
	await expect(page).toHaveURL(/\/devices\/GOLDENKV$/);
	await expect(page.getByRole('tab', { name: 'Functions', exact: true })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await expect(popup).toBeHidden();
	await expect(selected).toBeFocused();
	await page.keyboard.press('Space');
	await expect(search).toHaveValue('');
	await expect(popup.getByRole('option')).toHaveCount(5);
	await page.keyboard.press('Escape');
	await expect(popup).toBeHidden();
	await expect(selected).toBeFocused();
	await page.keyboard.press('Enter');
	await search.fill('global');
	await expect(
		popup.getByRole('option', { name: 'Global definitions', exact: true })
	).toBeVisible();
	await page.keyboard.press('ArrowDown');
	await page.keyboard.press('Enter');
	await expect(page).toHaveURL(/\/devices\/_global$/);
	await expect(page.getByRole('button', { name: 'Device: Global definitions' })).toBeFocused();
	expect(pageErrors).toEqual([]);
});

test('Device heading opens the selector and function actions stay with their context', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST');
	const deviceHeading = page.locator('.device-heading h1');
	const deviceTrigger = deviceHeading.getByRole('button', {
		name: 'Device: WIFI_TEST',
		exact: true
	});
	await expect(page.getByRole('heading', { name: 'WIFI_TEST', exact: true })).toBeVisible();
	await expect(deviceTrigger).toBeVisible();
	await expect(page.locator('.device-heading [aria-label^="Device:"]')).toHaveCount(1);
	await expect(
		page.locator('.device-heading').getByRole('link', { name: /All-Device discovery/ })
	).toBeVisible();
	await expect(
		page.locator('.function-list .rail-heading').getByRole('button', { name: 'New function' })
	).toBeVisible();
	await expect(page.locator('.catalog-actions')).toHaveCount(0);

	await deviceTrigger.click();
	const popup = page.getByRole('dialog', { name: 'Choose Device', exact: true });
	await popup.getByRole('option', { name: 'GOLDENKV', exact: true }).click();
	await expect(page).toHaveURL(/\/devices\/GOLDENKV$/);
});

test('device switching keeps overrides, copy, history and empty states correct', async ({
	page,
	context
}) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	await switchDevice(page, 'GOLDENKV');
	const parameters = page.getByRole('region', { name: 'GOLDENKV parameters' });
	await expect(parameters.getByRole('row', { name: /^ANT/ }).getByRole('cell').nth(2)).toHaveText(
		'2'
	);
	await expect(
		page
			.getByRole('region', { name: 'GOLDENKV criteria' })
			.getByRole('row', { name: /^POWER/ })
			.getByRole('cell')
			.nth(2)
	).toHaveText('20');
	await page.getByRole('button', { name: 'Copy example 31', exact: true }).click();
	await expect(page.getByText('Script copied.', { exact: true })).toBeVisible();
	await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
	expect(await page.evaluate(() => navigator.clipboard.readText())).toBe('TX|ANT=2');
	await page.getByRole('button', { name: 'Select RX_PER', exact: true }).click();
	await expect(page).toHaveURL(/function=2/);
	await page.goBack();
	await expect(
		page.getByRole('heading', { name: 'TX_MULTI_VERIFICATION', exact: true })
	).toBeVisible();
	await page.goBack();
	await expect(page.getByRole('heading', { name: 'WIFI_TEST', exact: true })).toBeVisible();
	await page.goForward();
	await expect(page.getByRole('heading', { name: 'GOLDENKV', exact: true })).toBeVisible();
	await switchDevice(page, 'GPIB');
	await expect(page.getByText('No linked functions.', { exact: true })).toBeVisible();
	await expect(page.getByRole('region', { name: 'Function details' })).toHaveCount(0);
	await page.goto('/devices/WIFI_TEST?function=999');
	await expect(
		page.getByText('This function is not linked to this device. Select a function from the list.')
	).toBeVisible();
});

test('late definitions cannot replace the newly selected device', async ({ page }) => {
	await mockDeviceHub(page);
	let release: () => void = () => {};
	const gate = new Promise<void>((resolve) => {
		release = resolve;
	});
	await page.route('**/api/devices/WIFI_TEST/parameters?*', async (route) => {
		await gate;
		await route.fulfill({
			json: {
				items: [
					{
						id: 11,
						name: 'ANT',
						device_name: 'WIFI_TEST',
						type: 'integer',
						default_value: '1',
						allowed_values: '',
						cardinality: 'single',
						value_pattern: '',
						required: true
					}
				],
				total: 1,
				global_items: [],
				global_total: 0,
				page: 1,
				page_size: 20
			}
		});
	});
	const pending = page.waitForRequest('**/api/devices/WIFI_TEST/parameters?*');
	await page.goto('/devices/WIFI_TEST?tab=parameters');
	await pending;
	await switchDevice(page, 'GOLDENKV');
	await page.getByRole('tab', { name: 'Parameters', exact: true }).click();
	const defaultCell = page
		.getByRole('row')
		.filter({ has: page.getByRole('button', { name: 'Inspect ANT (GOLDENKV)', exact: true }) })
		.getByRole('cell')
		.nth(4);
	await expect(defaultCell).toHaveText('2');
	const response = page.waitForResponse('**/api/devices/WIFI_TEST/parameters?*');
	release();
	await response;
	await page.getByRole('tab', { name: 'Criteria', exact: true }).click();
	await page.getByRole('tab', { name: 'Parameters', exact: true }).click();
	await expect(defaultCell).toHaveText('2');
});

test('failed device and function detail requests can be retried', async ({ page }) => {
	await mockDeviceHub(page);
	let directoryAttempts = 0;
	await page.route('**/api/devices', async (route) => {
		if (directoryAttempts++ === 0)
			await route.fulfill({ status: 503, json: { error: 'Catalog unavailable' } });
		else await route.fallback();
	});
	await page.goto('/devices');
	await expect(page.getByRole('alert')).toBeVisible();
	await page.getByRole('button', { name: 'Retry', exact: true }).click();
	await expect(page.getByRole('article', { name: 'WIFI_TEST' })).toBeVisible();
	let detailAttempts = 0;
	await page.route('**/api/functions/1/parameter-links', async (route) => {
		if (detailAttempts++ === 0)
			await route.fulfill({ status: 503, json: { error: 'Function unavailable' } });
		else await route.fallback();
	});
	await page.getByRole('link', { name: 'WIFI_TEST', exact: true }).click();
	await page.getByRole('button', { name: 'Retry details', exact: true }).click();
	await expect(page.getByText('Device transmit example', { exact: true })).toBeVisible();
});

test('function search includes later API pages and opens the matching function', async ({
	page
}) => {
	await mockDeviceHub(page);
	const firstPage = Array.from({ length: 100 }, (_, index) => ({
		id: index + 1,
		name: `FUNCTION_${index + 1}`,
		source: 'DLL',
		interface: 'WIFI_TEST',
		category: '',
		description: '',
		updated_at: '2026-10-06'
	}));
	const last = { ...firstPage[0], id: 101, name: 'LAST_PAGE_FUNCTION' };
	await page.route('**/api/devices/WIFI_TEST/functions?*', (route) =>
		route.fulfill({
			json: {
				items: new URL(route.request().url()).searchParams.get('page') === '1' ? firstPage : [last],
				total: 101,
				page_size: 100
			}
		})
	);
	await page.goto('/devices/WIFI_TEST');
	await page.getByRole('textbox', { name: 'Search device functions' }).fill('LAST_PAGE_FUNCTION');
	await page.getByRole('button', { name: 'Select LAST_PAGE_FUNCTION', exact: true }).click();
	await expect(page).toHaveURL(/function=101/);
	await expect(
		page.getByRole('heading', { name: 'LAST_PAGE_FUNCTION', exact: true })
	).toBeVisible();
});

test('the device Examples tab exposes copyable device-scoped scripts', async ({
	page,
	context
}) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=examples');
	await page.getByRole('button', { name: 'Copy device example 31', exact: true }).click();
	await expect(page.getByText('Script copied.', { exact: true })).toBeVisible();
	await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
	expect(await page.evaluate(() => navigator.clipboard.readText())).toBe('TX|ANT=1');
	await expect(page.getByText('GLOBAL_REFERENCE', { exact: true })).toHaveCount(0);
});

test('Device Function detail keeps Device and Global examples in separate groups', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.route('**/api/examples?*', (route) => {
		const global = new URL(route.request().url()).searchParams.get('device') === '';
		const example = global
			? {
					id: 32,
					function_source: 'DLL',
					function_name: 'TX_MULTI_VERIFICATION',
					device_name: '',
					script: 'TX|GLOBAL_FIELD=1',
					meaning: 'Shared Global reference',
					function_exists: true,
					created_at: '',
					updated_at: ''
				}
			: {
					id: 31,
					function_source: 'DLL',
					function_name: 'TX_MULTI_VERIFICATION',
					device_name: 'WIFI_TEST',
					script: 'TX|ANT=1',
					meaning: 'Device-scoped reference',
					function_exists: true,
					created_at: '',
					updated_at: ''
				};
		return route.fulfill({ json: { items: [example], total: 1, page: 1, page_size: 20 } });
	});

	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	const deviceExamples = page.getByRole('region', { name: 'WIFI_TEST examples', exact: true });
	const globalExamples = page.getByRole('region', {
		name: 'Global references examples',
		exact: true
	});
	await expect(deviceExamples.getByText('Device-scoped reference', { exact: true })).toBeVisible();
	await expect(globalExamples.getByText('Shared Global reference', { exact: true })).toBeVisible();
});

test('Device Examples expose create, edit and delete actions in Device context', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=examples');

	await expect(page.getByRole('button', { name: 'New example', exact: true })).toBeVisible();
	const example = page.getByRole('row').filter({ hasText: 'TX_MULTI_VERIFICATION' });
	await expect(example.getByRole('button', { name: 'Edit example 31', exact: true })).toBeVisible();
	await expect(
		example.getByRole('button', { name: 'Delete example 31', exact: true })
	).toBeVisible();
});

test('Device Examples can retry a failed list request', async ({ page }) => {
	await mockDeviceHub(page);
	let requests = 0;
	await page.route(/\/api\/examples(?:\?.*)?$/, (route) => {
		requests++;
		return requests === 1
			? route.fulfill({ status: 503, json: { error: 'Examples unavailable' } })
			: route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 20 } });
	});

	await page.goto('/devices/WIFI_TEST?tab=examples');
	await expect(page.getByRole('alert')).toContainText('Examples unavailable');
	await page.getByRole('button', { name: 'Retry examples', exact: true }).click();
	await expect(page.getByText('No examples carry this device.', { exact: true })).toBeVisible();
	expect(requests).toBe(2);
});

test('Function Examples can retry a failed list request', async ({ page }) => {
	await mockFunctionCatalog(page);
	let requests = 0;
	await page.route(/\/api\/examples(?:\?.*)?$/, (route) => {
		requests++;
		return requests === 1
			? route.fulfill({ status: 503, json: { error: 'Examples unavailable' } })
			: route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 20 } });
	});

	await page.goto('/functions?function=1');
	await page.getByRole('tab', { name: 'Examples', exact: true }).click();
	await expect(page.getByRole('alert')).toContainText('Examples unavailable');
	await page.getByRole('button', { name: 'Retry examples', exact: true }).click();
	await expect(page.getByText('No examples yet.', { exact: true })).toBeVisible();
	expect(requests).toBe(2);
});

test('Device Example CRUD keeps the selected Device and list context', async ({ page }) => {
	await mockDeviceHub(page);
	let examples = [
		{
			id: 31,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: 'TX|ANT=1',
			meaning: 'Device example',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		}
	];
	const mutations: { method: string; body: Record<string, unknown> | null }[] = [];
	await page.route(/\/api\/examples(?:\/\d+)?(?:\?.*)?$/, async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		if (request.method() === 'POST') {
			const body = request.postDataJSON();
			mutations.push({ method: 'POST', body });
			const created = {
				...body,
				id: 32,
				function_exists: true,
				created_at: '2026-10-07',
				updated_at: '2026-10-07'
			};
			examples = [...examples, created];
			return route.fulfill({ status: 201, json: created });
		}
		if (request.method() === 'PUT') {
			const id = Number(url.pathname.split('/').pop());
			const body = request.postDataJSON();
			mutations.push({ method: 'PUT', body });
			examples = examples.map((row) => (row.id === id ? { ...row, ...body } : row));
			return route.fulfill({ json: examples.find((row) => row.id === id) });
		}
		if (request.method() === 'DELETE') {
			const id = Number(url.pathname.split('/').pop());
			mutations.push({ method: 'DELETE', body: null });
			examples = examples.filter((row) => row.id !== id);
			return route.fulfill({ status: 204 });
		}
		const device = url.searchParams.get('device');
		const items = examples.filter((row) => device === null || row.device_name === device);
		return route.fulfill({
			json: {
				items,
				total: items.length,
				page: Number(url.searchParams.get('page') ?? 1),
				page_size: Number(url.searchParams.get('page_size') ?? 20)
			}
		});
	});

	await page.goto('/devices/WIFI_TEST?tab=examples');
	await page.getByRole('button', { name: 'New example', exact: true }).click();
	let dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await expect(dialog).toContainText('Device: WIFI_TEST');
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=examples/);

	await page.getByRole('button', { name: 'New example', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await dialog.getByRole('textbox', { name: 'Function name', exact: true }).fill('CUSTOM_EXAMPLE');
	await dialog.getByRole('textbox', { name: 'Script', exact: true }).fill('CUSTOM|FIELD=1');
	await dialog
		.getByRole('textbox', { name: 'Meaning', exact: true })
		.fill('Created in Device context');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page.getByRole('row').filter({ hasText: 'CUSTOM_EXAMPLE' })).toBeVisible();
	expect(mutations[0]).toMatchObject({
		method: 'POST',
		body: {
			function_source: 'DLL',
			function_name: 'CUSTOM_EXAMPLE',
			device_name: 'WIFI_TEST',
			script: 'CUSTOM|FIELD=1',
			meaning: 'Created in Device context'
		}
	});

	await page.getByRole('button', { name: 'Edit example 32', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'Edit Example', exact: true });
	await expect(
		dialog.getByRole('button', { name: 'Device / Interface: WIFI_TEST', exact: true })
	).toBeEnabled();
	await dialog
		.getByRole('textbox', { name: 'Meaning', exact: true })
		.fill('Updated in Device context');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	expect(mutations[1]).toMatchObject({
		method: 'PUT',
		body: { device_name: 'WIFI_TEST', meaning: 'Updated in Device context' }
	});
	await expect(page.getByText('Updated in Device context', { exact: true })).toBeVisible();

	await page.getByRole('button', { name: 'Delete example 32', exact: true }).click();
	const confirmation = page.getByRole('alertdialog', { name: 'Delete Example', exact: true });
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page.getByRole('row').filter({ hasText: 'CUSTOM_EXAMPLE' })).toHaveCount(0);
	expect(mutations.map((mutation) => mutation.method)).toEqual(['POST', 'PUT', 'DELETE']);
});

test('Device Function Examples prefill Function and preserve Global ownership', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	let examples = [
		{
			id: 41,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: 'TX|ANT=1',
			meaning: 'Device transmit example',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		},
		{
			id: 42,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: '',
			script: 'TX|ANT=2',
			meaning: 'Global transmit example',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		}
	];
	const mutations: { method: string; body: Record<string, unknown> | null }[] = [];
	await page.route(/\/api\/examples(?:\/\d+)?(?:\?.*)?$/, async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		if (request.method() === 'POST') {
			const body = request.postDataJSON();
			mutations.push({ method: 'POST', body });
			const created = {
				...body,
				id: 43,
				function_exists: true,
				created_at: '2026-10-07',
				updated_at: '2026-10-07'
			};
			examples = [...examples, created];
			return route.fulfill({ status: 201, json: created });
		}
		if (request.method() === 'PUT') {
			const id = Number(url.pathname.split('/').pop());
			const body = request.postDataJSON();
			mutations.push({ method: 'PUT', body });
			examples = examples.map((row) => (row.id === id ? { ...row, ...body } : row));
			return route.fulfill({ json: examples.find((row) => row.id === id) });
		}
		if (request.method() === 'DELETE') {
			const id = Number(url.pathname.split('/').pop());
			mutations.push({ method: 'DELETE', body: null });
			examples = examples.filter((row) => row.id !== id);
			return route.fulfill({ status: 204 });
		}
		const device = url.searchParams.get('device');
		const items = examples.filter(
			(row) =>
				row.function_source === url.searchParams.get('function_source') &&
				row.function_name === url.searchParams.get('function_name') &&
				(device === null || row.device_name === device)
		);
		return route.fulfill({ json: { items, total: items.length, page: 1, page_size: 20 } });
	});

	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	const functionExamples = page.getByRole('region', { name: 'Function examples', exact: true });
	const deviceExamples = page.getByRole('region', { name: 'WIFI_TEST examples', exact: true });
	await functionExamples.getByRole('button', { name: 'New example', exact: true }).click();
	let dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await expect(dialog).toContainText('Function: DLL / TX_MULTI_VERIFICATION');
	await expect(dialog).toContainText('Device: WIFI_TEST');
	await dialog.getByRole('textbox', { name: 'Script', exact: true }).fill('TX|ANT=3');
	await dialog
		.getByRole('textbox', { name: 'Meaning', exact: true })
		.fill('Created from Function context');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(
		deviceExamples.getByText('Created from Function context', { exact: true })
	).toBeVisible();
	expect(mutations[0]).toMatchObject({
		method: 'POST',
		body: {
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST'
		}
	});

	const globalExamples = page.getByRole('region', {
		name: 'Global references examples',
		exact: true
	});
	await globalExamples.getByRole('button', { name: 'Edit Global example 42', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'Edit Example', exact: true });
	await expect(
		dialog.getByRole('button', {
			name: 'Device / Interface: Global (no Device / Interface)',
			exact: true
		})
	).toBeDisabled();
	await dialog.getByRole('textbox', { name: 'Meaning', exact: true }).fill('Edited as Global');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(globalExamples.getByText('Edited as Global', { exact: true })).toBeVisible();
	expect(mutations[1]).toMatchObject({ method: 'PUT', body: { device_name: '' } });

	await deviceExamples.getByRole('button', { name: 'Delete example 41', exact: true }).click();
	const confirmation = page.getByRole('alertdialog', { name: 'Delete Example', exact: true });
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(deviceExamples.getByText('Device transmit example', { exact: true })).toHaveCount(0);
	expect(mutations.map((mutation) => mutation.method)).toEqual(['POST', 'PUT', 'DELETE']);
});

test('Examples workspace detail preserves meaning, full script and Global ownership', async ({
	page,
	context
}) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	await mockDeviceHub(page);
	const script = 'BUILT-IN|RX_PER|WINDOW=10|PACKET_LOSS<1%|10000';
	await page.route('**/api/examples?*', (route) =>
		route.fulfill({
			json: {
				items: [
					{
						id: 52,
						function_source: 'BUILT-IN',
						function_name: 'RX_PER',
						device_name: '',
						script,
						meaning: 'Check packet loss remains under one percent.',
						function_exists: true,
						created_at: '2026-10-07',
						updated_at: '2026-10-07'
					}
				],
				total: 1,
				page: 1,
				page_size: 20
			}
		})
	);

	await page.goto('/examples?device=_global');
	const detail = page.getByRole('region', { name: 'Example details', exact: true });
	await expect(
		detail.getByRole('heading', { name: 'BUILT-IN / RX_PER', exact: true })
	).toBeVisible();
	await expect(
		detail.getByText('Check packet loss remains under one percent.', { exact: true })
	).toBeVisible();
	await expect(detail.locator('.script-scroll')).toHaveText(script);
	await expect(detail.locator('.ownership')).toHaveText('Global - shared across Devices');
	const bottomGaps = await page.locator('.examples-workspace').evaluate((workspace) => {
		const rail = workspace.querySelector('aside')!;
		const pager = rail.querySelector('[data-pagination-root]')!;
		return [
			workspace.getBoundingClientRect().bottom - rail.getBoundingClientRect().bottom,
			rail.getBoundingClientRect().bottom - pager.getBoundingClientRect().bottom
		];
	});
	for (const gap of bottomGaps) expect(Math.abs(gap)).toBeLessThan(2);
	await detail.getByRole('button', { name: 'Copy example 52', exact: true }).click();
	await expect(detail.getByRole('status')).toContainText('Script copied.');
	expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(script);
});

test('Examples workspace filters Global results and reaches later API pages', async ({ page }) => {
	await mockDeviceHub(page);
	const examples = Array.from({ length: 21 }, (_, index) => ({
		id: index + 1,
		function_source: 'DLL',
		function_name: 'TX_MATCH',
		device_name: '',
		script: `DLL|TX_MATCH|ANT=${index + 1}|POWER=10|10000`,
		meaning: `Meaning ${index + 1}`,
		function_exists: true,
		created_at: '2026-10-07',
		updated_at: '2026-10-07'
	}));
	await page.route('**/api/examples/function-names', (route) =>
		route.fulfill({ json: { items: ['TX_MATCH'] } })
	);
	const requests: URLSearchParams[] = [];
	await page.route(/\/api\/examples(?:\?.*)?$/, (route) => {
		const query = new URL(route.request().url()).searchParams;
		requests.push(query);
		const matches = examples.filter(
			(row) =>
				(query.get('device') === null || row.device_name === query.get('device')) &&
				(!query.get('function_source') || row.function_source === query.get('function_source')) &&
				(!query.get('function_name') || row.function_name === query.get('function_name')) &&
				(!query.get('search') || `${row.script} ${row.meaning}`.includes(query.get('search')!))
		);
		const pageNumber = Number(query.get('page') ?? 1);
		const pageSize = Number(query.get('page_size') ?? 20);
		return route.fulfill({
			json: {
				items: matches.slice((pageNumber - 1) * pageSize, pageNumber * pageSize),
				total: matches.length,
				page: pageNumber,
				page_size: pageSize
			}
		});
	});

	await page.goto('/examples');
	await expect(page.getByText('All-Device discovery', { exact: true })).toBeVisible();
	await page
		.getByRole('button', { name: 'Filter examples by Device: All Devices', exact: true })
		.click();
	const devicePicker = page.getByRole('dialog', {
		name: 'Choose Filter examples by Device',
		exact: true
	});
	await devicePicker.getByRole('option', { name: 'Global definitions', exact: true }).click();
	await expect(page).toHaveURL(/device=_global/);
	await expect(page.getByText('Global examples', { exact: true })).toBeVisible();

	await expect(
		page.getByRole('textbox', { name: 'Function source filter', exact: true })
	).toHaveCount(0);
	await page.getByRole('button', { name: 'Function name filter', exact: true }).click();
	const functionPicker = page.getByRole('dialog', {
		name: 'Choose Function name filter',
		exact: true
	});
	await functionPicker.getByRole('combobox', { name: 'Search options' }).fill('TX_MATCH');
	await functionPicker.getByRole('option', { name: 'TX_MATCH', exact: true }).click();
	await expect(page).toHaveURL(/function_name=TX_MATCH/);
	await expect(page.getByText('Page 1 of 2', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Next page', exact: true }).click();
	await expect(page).toHaveURL(/page=2/);
	const laterExample = page.getByRole('button', {
		name: 'Select example of TX_MATCH (Global)',
		exact: true
	});
	await expect(laterExample).toContainText('Meaning 21');
	await laterExample.click();
	await expect(page).toHaveURL(/example=21/);
	const detail = page.getByRole('region', { name: 'Example details', exact: true });
	await expect(detail.locator('.script-scroll')).toHaveText('DLL|TX_MATCH|ANT=21|POWER=10|10000');
	await expect(detail.getByText('Meaning 21', { exact: true })).toBeVisible();
	expect(Object.fromEntries(requests.at(-1) ?? [])).toMatchObject({
		device: '',
		function_name: 'TX_MATCH',
		page: '2',
		page_size: '20'
	});
	expect(requests.at(-1)?.has('function_source')).toBe(false);
});

test('Examples workspace ignores an out-of-range response after leaving the page', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	let releaseResponse: (() => void) | undefined;
	let signalRequest!: () => void;
	const requestStarted = new Promise<void>((resolve) => {
		signalRequest = resolve;
	});
	await page.route(/\/api\/examples(?:\?.*)?$/, async (route) => {
		const query = new URL(route.request().url()).searchParams;
		if (query.get('page') !== '2')
			return route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 20 } });
		await new Promise<void>((resolve) => {
			releaseResponse = () => {
				void route.fulfill({ json: { items: [], total: 0, page: 2, page_size: 20 } }).then(resolve);
			};
			signalRequest();
		});
	});

	await page.goto('/examples?page=2');
	await requestStarted;
	await page.getByRole('link', { name: 'Functions', exact: true }).click();
	await expect(page).toHaveURL(/\/functions$/);
	expect(releaseResponse).toBeDefined();
	releaseResponse?.();
	await expect(page).toHaveURL(/\/functions$/);
});

test('Examples workspace CRUD preserves filters and selects a valid record after delete', async ({
	page
}) => {
	await mockDeviceHub(page);
	let examples = [
		{
			id: 52,
			function_source: 'DLL',
			function_name: 'RX_PER',
			device_name: '',
			script: 'DLL|RX_PER|WINDOW=10',
			meaning: 'Original Global meaning',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		}
	];
	const mutations: { method: string; body: Record<string, unknown> | null }[] = [];
	await page.route(/\/api\/examples(?:\/\d+)?(?:\?.*)?$/, async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		if (request.method() === 'POST') {
			const body = request.postDataJSON();
			if (!body.function_name)
				return route.fulfill({
					status: 422,
					json: {
						error: 'Invalid example',
						fields: { function_name: 'Function name is required.' }
					}
				});
			mutations.push({ method: 'POST', body });
			const created = {
				...body,
				id: 53,
				function_exists: true,
				created_at: '2026-10-07',
				updated_at: '2026-10-07'
			};
			examples = [...examples, created];
			return route.fulfill({ status: 201, json: created });
		}
		if (request.method() === 'PUT') {
			const id = Number(url.pathname.split('/').pop());
			const body = request.postDataJSON();
			mutations.push({ method: 'PUT', body });
			examples = examples.map((row) => (row.id === id ? { ...row, ...body } : row));
			return route.fulfill({ json: examples.find((row) => row.id === id) });
		}
		if (request.method() === 'DELETE') {
			const id = Number(url.pathname.split('/').pop());
			mutations.push({ method: 'DELETE', body: null });
			examples = examples.filter((row) => row.id !== id);
			return route.fulfill({ status: 204 });
		}
		const device = url.searchParams.get('device');
		const items = examples.filter(
			(row) =>
				(device === null || row.device_name === device) &&
				(!url.searchParams.get('function_source') ||
					row.function_source === url.searchParams.get('function_source')) &&
				(!url.searchParams.get('function_name') ||
					row.function_name === url.searchParams.get('function_name'))
		);
		return route.fulfill({ json: { items, total: items.length, page: 1, page_size: 20 } });
	});

	await page.goto('/examples?device=_global&source=DLL&function_name=RX_PER');
	const detail = page.getByRole('region', { name: 'Example details', exact: true });
	await expect(detail.getByText('Original Global meaning', { exact: true })).toBeVisible();
	await detail.getByRole('button', { name: 'Edit Global example', exact: true }).click();
	let dialog = page.getByRole('dialog', { name: 'Edit Example', exact: true });
	await dialog.getByRole('textbox', { name: 'Meaning', exact: true }).fill('Edited Global meaning');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(detail.getByText('Edited Global meaning', { exact: true })).toBeVisible();
	await expect(page).toHaveURL(/device=_global.*function_name=RX_PER.*example=52/);
	await expect(page).not.toHaveURL(/source=/);
	expect(mutations[0]).toMatchObject({ method: 'PUT', body: { device_name: '' } });

	await page.getByRole('button', { name: 'New example', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await expect(dialog).toContainText('Global - shared across Devices');
	await dialog.getByRole('textbox', { name: 'Function name', exact: true }).fill('RX_PER');
	await dialog.getByRole('textbox', { name: 'Script', exact: true }).fill('DLL|RX_PER|WINDOW=20');
	await dialog.getByRole('textbox', { name: 'Meaning', exact: true }).fill('New Global meaning');
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(page).toHaveURL(/device=_global.*function_name=RX_PER.*example=52/);

	await page.getByRole('button', { name: 'New example', exact: true }).click();
	dialog = page.getByRole('dialog', { name: 'New Example', exact: true });
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(dialog.getByText('Function name is required.', { exact: true })).toBeVisible();
	await dialog.getByRole('textbox', { name: 'Function name', exact: true }).fill('RX_PER');
	await dialog.getByRole('textbox', { name: 'Script', exact: true }).fill('DLL|RX_PER|WINDOW=20');
	await dialog.getByRole('textbox', { name: 'Meaning', exact: true }).fill('New Global meaning');
	await dialog.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page).toHaveURL(/example=53/);
	expect(mutations[1]).toMatchObject({
		method: 'POST',
		body: { function_source: 'DLL', function_name: 'RX_PER', device_name: '' }
	});

	await page.getByRole('button', { name: 'Delete Global example', exact: true }).click();
	const confirmation = page.getByRole('alertdialog', { name: 'Delete Example', exact: true });
	await confirmation.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL(/device=_global.*function_name=RX_PER.*example=52/);
	await expect(detail.getByText('Edited Global meaning', { exact: true })).toBeVisible();
	expect(mutations.map((mutation) => mutation.method)).toEqual(['PUT', 'POST', 'DELETE']);
});

test('Examples workspace stays readable on desktop, mobile and dark theme', async ({
	page
}, testInfo) => {
	await mockDeviceHub(page);
	const script = `DLL|LONG_FUNCTION|FIELD=${'UNBROKEN_VALUE_'.repeat(60)}|NULL|10000`;
	await page.route(/\/api\/examples(?:\?.*)?$/, (route) =>
		route.fulfill({
			json: {
				items: [
					{
						id: 61,
						function_source: 'DLL',
						function_name: 'LONG_FUNCTION',
						device_name: '',
						script,
						meaning: 'A long example remains readable at narrow widths.',
						function_exists: true,
						created_at: '2026-10-07',
						updated_at: '2026-10-07'
					}
				],
				total: 1,
				page: 1,
				page_size: 20
			}
		})
	);

	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/examples?device=_global');
	await expect(page.getByText('1 example', { exact: true })).toBeVisible();
	const detail = page.getByRole('region', { name: 'Example details', exact: true });
	await expect(detail.locator('.script-scroll')).toBeVisible();
	await expect(page.locator('.examples-workspace')).toHaveCSS('border-radius', '6px');
	await expect(page.locator('.example-list .rail-heading h2')).toHaveText('Examples');
	const desktopGeometry = await page.evaluate(() => ({
		viewport: window.innerWidth,
		client: document.documentElement.clientWidth,
		scroll: document.documentElement.scrollWidth
	}));
	expect(desktopGeometry.scroll).toBeLessThanOrEqual(desktopGeometry.client);
	await page.screenshot({
		path: testInfo.outputPath('examples-desktop-light.png'),
		fullPage: true
	});

	await page.setViewportSize({ width: 390, height: 844 });
	const mobileGeometry = await page.evaluate(() => ({
		viewport: window.innerWidth,
		client: document.documentElement.clientWidth,
		scroll: document.documentElement.scrollWidth
	}));
	expect(mobileGeometry.scroll).toBeLessThanOrEqual(mobileGeometry.client);
	await expect(detail).toBeVisible();
	await page.screenshot({ path: testInfo.outputPath('examples-mobile-light.png'), fullPage: true });

	await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
	await expect(page.locator('html')).toHaveClass(/dark/);
	await page.screenshot({ path: testInfo.outputPath('examples-mobile-dark.png'), fullPage: true });
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.screenshot({ path: testInfo.outputPath('examples-desktop-dark.png'), fullPage: true });
});

test('Device and Function examples identify linked parameter and criteria fields', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	const examples = [
		{
			id: 41,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: 'DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=1&SETTLE_MS=200|POWER=10|10000|NEXT|FAIL',
			meaning: 'Transmit verification using device fields',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		},
		{
			id: 42,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: 'TX|ANT=2',
			meaning: 'Legacy parameter example',
			function_exists: true,
			created_at: '2026-10-07',
			updated_at: '2026-10-07'
		}
	];
	await page.route(/\/api\/examples(?:\?.*)?$/, (route) =>
		route.fulfill({ json: { items: examples, total: examples.length, page: 1, page_size: 20 } })
	);

	await page.goto('/devices/WIFI_TEST?tab=functions&function=1');
	const deviceExamples = page.getByRole('region', { name: 'WIFI_TEST examples' });
	const deviceUsage = deviceExamples.getByRole('group', { name: 'Fields used in script' });
	await expect(deviceUsage).toHaveCount(2);
	await expect(deviceUsage.nth(0)).toContainText('Parameters');
	await expect(deviceUsage.nth(0)).toContainText('ANT');
	await expect(deviceUsage.nth(0)).toContainText('SETTLE_MS');
	await expect(deviceUsage.nth(0)).toContainText('Criteria');
	await expect(deviceUsage.nth(0)).toContainText('POWER');
	await expect(deviceUsage.nth(1)).toContainText('Parameters');
	await expect(deviceUsage.nth(1)).toContainText('ANT');

	await page.goto('/functions?function=1');
	await page.getByRole('tab', { name: 'Examples', exact: true }).click();
	const functionUsage = page
		.getByRole('region', { name: 'Function details', exact: true })
		.getByRole('group', { name: 'Fields used in script' });
	await expect(functionUsage).toHaveCount(2);
	await expect(functionUsage.nth(0)).toContainText('Parameters');
	await expect(functionUsage.nth(0)).toContainText('ANT');
	await expect(functionUsage.nth(0)).toContainText('SETTLE_MS');
	await expect(functionUsage.nth(0)).toContainText('Criteria');
	await expect(functionUsage.nth(0)).toContainText('POWER');
	await expect(functionUsage.nth(1)).toContainText('ANT');
});

function sequenceStep(id: number, overrides: Partial<SequenceStep> = {}): SequenceStep {
	return {
		id,
		position: id,
		function_source: 'DLL',
		function_interface: 'WIFI_TEST',
		function_name: 'TX_MULTI_VERIFICATION',
		parameters: { ANT: '1' },
		criteria: { POWER: '10' },
		timeout_ms: '10000',
		flow: 'NEXT',
		on_fail: 'FAIL',
		device_name: 'WIFI_TEST',
		note: 'main',
		function_exists: true,
		...overrides
	};
}

function sequenceDetail(id: number, name: string, steps: SequenceStep[]): SequenceDetail {
	return { id, name, description: '', steps, created_at: '2026-10-07', updated_at: '2026-10-07' };
}

function retainedFields(step: SequenceStepInput) {
	return {
		function_source: step.function_source,
		function_interface: step.function_interface,
		function_name: step.function_name,
		parameters: step.parameters,
		criteria: step.criteria,
		timeout_ms: step.timeout_ms,
		flow: step.flow,
		on_fail: step.on_fail,
		device_name: step.device_name,
		note: step.note
	};
}

function emitSequenceText(sequence: SequenceDetail) {
	const steps = sequence.steps.map((step) => ({
		step_id: step.id,
		position: step.position,
		resolved: step.function_exists,
		script: `${step.function_source}${step.function_interface}|${step.function_name}|${step.timeout_ms}|${step.flow}|${step.on_fail}`,
		errors: {} as Record<string, string>
	}));
	return {
		text: steps
			.filter((step) => step.resolved)
			.map((step) => step.script)
			.join('\n'),
		steps
	};
}

// Stateful fake of the sequences API, including separate step and metadata writes.
async function mockSequenceStore(
	page: Page,
	initial: SequenceDetail[],
	options: {
		failListOnce?: boolean;
		failDetailOnce?: boolean;
		rejectSaveOnce?: { error: string; fields: Record<string, string> };
	} = {}
) {
	const sequences = [...initial];
	const requests: { method: string; path: string; search: string; body: unknown }[] = [];
	let failList = options.failListOnce ?? false;
	let failDetail = options.failDetailOnce ?? false;
	let rejectSave = options.rejectSaveOnce;
	let nextStepId = 1000;
	await page.route(/\/api\/sequences(?:\/\d+(?:\/(?:emit|steps))?)?(?:\?.*)?$/, async (route) => {
		const request = route.request();
		const url = new URL(request.url());
		const raw = request.postData();
		const body: unknown = raw ? JSON.parse(raw) : null;
		requests.push({ method: request.method(), path: url.pathname, search: url.search, body });
		const match = /^\/api\/sequences(?:\/(\d+)(?:\/(emit|steps))?)?$/.exec(url.pathname);
		if (!match) return route.fallback();
		if (!match[1]) {
			if (failList) {
				failList = false;
				return route.fulfill({ status: 500, json: { error: 'Sequence store is unavailable.' } });
			}
			const search = (url.searchParams.get('search') ?? '').toLowerCase();
			const matches = sequences.filter((sequence) =>
				`${sequence.name} ${sequence.description}`.toLowerCase().includes(search)
			);
			return route.fulfill({
				json: {
					items: matches.map((sequence) => ({
						id: sequence.id,
						name: sequence.name,
						description: sequence.description,
						step_count: sequence.steps.length,
						created_at: sequence.created_at,
						updated_at: sequence.updated_at
					})),
					total: matches.length,
					page: 1,
					page_size: Number(url.searchParams.get('page_size') ?? 20)
				}
			});
		}
		const sequence = sequences.find((row) => row.id === Number(match[1]));
		if (!sequence) return route.fulfill({ status: 404, json: { error: 'Sequence not found' } });
		if (match[2] === 'emit') return route.fulfill({ json: emitSequenceText(sequence) });
		if (request.method() === 'GET') {
			if (failDetail) {
				failDetail = false;
				return route.fulfill({ status: 500, json: { error: 'Sequence store is unavailable.' } });
			}
			return route.fulfill({ json: sequence });
		}
		if (request.method() === 'PUT' || request.method() === 'PATCH') {
			if (rejectSave) {
				const { error, fields } = rejectSave;
				rejectSave = undefined;
				return route.fulfill({ status: 422, json: { error, fields } });
			}
			const input = body as SequenceInput;
			if (match[2] !== 'steps') {
				sequence.name = input.name;
				sequence.description = input.description;
			}
			if (request.method() === 'PUT') {
				sequence.steps = input.steps.map((step, index) => {
					const stored = step as Partial<SequenceStep>;
					return {
						...step,
						id: nextStepId++,
						position: index + 1,
						function_exists: stored.function_exists ?? true
					};
				});
			}
			sequence.updated_at = '2026-10-08';
			return route.fulfill({ json: sequence });
		}
		if (request.method() === 'DELETE') {
			sequences.splice(sequences.indexOf(sequence), 1);
			return route.fulfill({ status: 204 });
		}
		return route.fulfill({ status: 405, json: { error: 'Unsupported' } });
	});
	return { requests, sequences };
}

test('Sequences list filters saved sequences and opens the selected detail', async ({ page }) => {
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [sequenceStep(1)]),
		sequenceDetail(2, 'RECEIVE_SWEEP', [sequenceStep(2), sequenceStep(3)])
	]);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/sequences');
	await expect(page.getByText('2 sequences', { exact: true })).toBeVisible({ timeout: 15_000 });
	await page.getByRole('searchbox', { name: 'Search sequences', exact: true }).fill('transmit');
	await expect(page.getByText('1 sequence', { exact: true })).toBeVisible();
	expect(store.requests.some((r) => r.search.includes('search=transmit'))).toBe(true);
	await page.getByRole('link', { name: 'TRANSMIT_CHECK', exact: true }).click();
	await expect(page).toHaveURL(/\/sequences\/1$/);
	await expect(page.getByRole('heading', { name: 'TRANSMIT_CHECK', exact: true })).toBeVisible();
});

test('Sequences list shows the empty state when no sequences exist', async ({ page }) => {
	await mockSequenceStore(page, []);
	await page.goto('/sequences');
	await expect(page.getByText('No sequences yet.', { exact: true })).toBeVisible({
		timeout: 15_000
	});
});

test('Sequences list retries a failed load without leaving the page', async ({ page }) => {
	await mockSequenceStore(page, [sequenceDetail(1, 'TRANSMIT_CHECK', [sequenceStep(1)])], {
		failListOnce: true
	});
	await page.goto('/sequences');
	await expect(page.locator('.banner-error')).toContainText('Sequence store is unavailable.', {
		timeout: 15_000
	});
	await expect(page.getByText('No sequences yet.', { exact: true })).toHaveCount(0);
	await page.getByRole('button', { name: 'Retry sequences', exact: true }).click();
	await expect(page.getByRole('link', { name: 'TRANSMIT_CHECK', exact: true })).toBeVisible();
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Sequence detail retries a failed load and then shows its steps', async ({ page }) => {
	await mockSequenceStore(page, [sequenceDetail(1, 'TRANSMIT_CHECK', [sequenceStep(1)])], {
		failDetailOnce: true
	});
	await page.goto('/sequences/1');
	await expect(page.locator('.banner-error')).toContainText('Sequence store is unavailable.', {
		timeout: 15_000
	});
	await page.getByRole('button', { name: 'Retry sequence', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'TRANSMIT_CHECK', exact: true })).toBeVisible();
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Sequence steps keep identity values and order through keyboard reorder and edit', async ({
	page
}) => {
	const txStep = sequenceStep(1, { parameters: { ANT: '1' }, criteria: { POWER: '10' } });
	const rxStep = sequenceStep(2, {
		function_interface: 'GPIB',
		function_name: 'RX_PER',
		parameters: {},
		criteria: {},
		timeout_ms: '2000',
		flow: 'QUIT',
		device_name: '',
		note: ''
	});
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [txStep, rxStep])
	]);
	await page.goto('/sequences/1');
	const steps = page.locator('.steps > li');
	await expect(steps).toHaveCount(2, { timeout: 15_000 });

	await steps.nth(0).getByRole('button', { name: 'Move step down', exact: true }).press('Enter');
	await expect(steps.nth(0)).toContainText('RX_PER');
	const reordered = store.requests.filter((r) => r.method === 'PUT').at(-1)!.body as SequenceInput;
	expect(reordered.steps.map((step) => step.function_name)).toEqual([
		'RX_PER',
		'TX_MULTI_VERIFICATION'
	]);
	expect(retainedFields(reordered.steps[0])).toEqual(retainedFields(rxStep));
	expect(retainedFields(reordered.steps[1])).toEqual(retainedFields(txStep));

	await steps.nth(1).getByRole('button', { name: 'Edit', exact: true }).press('Enter');
	await steps.nth(1).getByLabel('Parameters (one name=value per line)').fill('ANT=2\nBAND=5G');
	await steps.nth(1).getByLabel('Timeout-ms', { exact: true }).fill('2500');
	await steps.nth(1).getByLabel('Flow', { exact: true }).fill('QUIT');
	await steps.nth(1).getByLabel('Note', { exact: true }).fill('edited note');
	await steps.nth(1).getByRole('button', { name: 'Save step', exact: true }).press('Enter');
	await expect(page.getByText('Step saved', { exact: true })).toBeVisible();
	await expect(steps.nth(1)).toContainText('2500');
	await expect(page.locator('.emitted-preview')).toContainText('|2500|QUIT|FAIL');

	const edited = store.requests.filter((r) => r.method === 'PUT').at(-1)!.body as SequenceInput;
	expect(retainedFields(edited.steps[0])).toEqual(retainedFields(rxStep));
	expect(retainedFields(edited.steps[1])).toEqual({
		...retainedFields(txStep),
		parameters: { ANT: '2', BAND: '5G' },
		timeout_ms: '2500',
		flow: 'QUIT',
		note: 'edited note'
	});
});

test('Sequence removal and deletion keep their guards', async ({ page }) => {
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [
			sequenceStep(1),
			sequenceStep(2, { function_name: 'RX_PER' })
		])
	]);
	await page.goto('/sequences/1');
	const steps = page.locator('.steps > li');

	await steps.nth(1).getByRole('button', { name: 'Remove', exact: true }).click();
	await expect(steps).toHaveCount(1);
	const removal = store.requests.filter((r) => r.method === 'PUT').at(-1)!.body as SequenceInput;
	expect(removal.steps.map((step) => step.function_name)).toEqual(['TX_MULTI_VERIFICATION']);
	await expect(steps.nth(0).getByRole('button', { name: 'Remove', exact: true })).toBeDisabled();

	await page.getByRole('button', { name: 'Delete sequence', exact: true }).click();
	const confirm = page.getByRole('alertdialog');
	await confirm.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(confirm).not.toBeVisible();
	expect(store.requests.some((r) => r.method === 'DELETE')).toBe(false);

	await page.getByRole('button', { name: 'Delete sequence', exact: true }).click();
	await page.getByRole('alertdialog').getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL(/\/sequences$/);
	expect(store.requests.some((r) => r.method === 'DELETE' && r.path === '/api/sequences/1')).toBe(
		true
	);
});

test('Sequence reorder keeps an open step draft attached to its own step', async ({ page }) => {
	const txStep = sequenceStep(1);
	const rxStep = sequenceStep(2, {
		function_name: 'RX_PER',
		parameters: {},
		criteria: {},
		device_name: ''
	});
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [txStep, rxStep])
	]);
	await page.goto('/sequences/1');
	const steps = page.locator('.steps > li');
	await steps.nth(0).getByRole('button', { name: 'Edit', exact: true }).click();
	await steps.nth(0).getByLabel('Timeout-ms', { exact: true }).fill('2500');
	await steps.nth(0).getByRole('button', { name: 'Move step down', exact: true }).click();
	await expect(steps.nth(0)).toContainText('RX_PER', { timeout: 15_000 });
	await expect(steps.nth(1).getByLabel('Timeout-ms', { exact: true })).toHaveValue('2500');
	await steps.nth(1).getByRole('button', { name: 'Save step', exact: true }).click();
	await expect(steps.nth(1)).toContainText('2500');
	const saved = store.requests.filter((r) => r.method === 'PUT').at(-1)!.body as SequenceInput;
	expect(saved.steps.map((step) => step.function_name)).toEqual([
		'RX_PER',
		'TX_MULTI_VERIFICATION'
	]);
	expect(retainedFields(saved.steps[0])).toEqual(retainedFields(rxStep));
	expect(retainedFields(saved.steps[1])).toEqual({ ...retainedFields(txStep), timeout_ms: '2500' });
});

test('Sequence step editor rejects malformed parameter lines before saving', async ({ page }) => {
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [
			sequenceStep(1),
			sequenceStep(2, { function_name: 'RX_PER' })
		])
	]);
	await page.goto('/sequences/1');
	const steps = page.locator('.steps > li');
	await steps.nth(0).getByRole('button', { name: 'Edit', exact: true }).click();
	await steps.nth(0).getByLabel('Parameters (one name=value per line)').fill('ANT 1');
	await steps.nth(0).getByRole('button', { name: 'Save step', exact: true }).click();
	await expect(steps.nth(0).getByRole('alert')).toContainText('ANT 1');
	expect(store.requests.some((r) => r.method === 'PUT')).toBe(false);
});

test('Sequence rename keeps the ordered steps unchanged', async ({ page }) => {
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [
			sequenceStep(1),
			sequenceStep(2, { function_name: 'RX_PER' })
		])
	]);
	await page.goto('/sequences/1');
	const firstStep = page.locator('.steps > li').first();
	await firstStep.getByRole('button', { name: 'Edit', exact: true }).click();
	await firstStep.getByLabel('Note', { exact: true }).fill('Unsaved draft');
	await page.getByRole('button', { name: 'Rename', exact: true }).click();
	await page.getByLabel('Sequence name', { exact: true }).fill('TRANSMIT_CHECK_V2');
	await page.getByLabel('Description', { exact: true }).fill('bench sequence');
	await page.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page.getByText('Sequence updated', { exact: true })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'TRANSMIT_CHECK_V2', exact: true })).toBeVisible();
	await expect(firstStep.getByLabel('Note', { exact: true })).toHaveValue('Unsaved draft');
	const renames = store.requests.filter((r) => r.method === 'PATCH' || r.method === 'PUT');
	expect(renames).toMatchObject([
		{ method: 'PATCH', body: { name: 'TRANSMIT_CHECK_V2', description: 'bench sequence' } }
	]);
	expect(renames[0].body).not.toHaveProperty('steps');
	expect(store.sequences[0].steps.map((step) => step.function_name)).toEqual([
		'TX_MULTI_VERIFICATION',
		'RX_PER'
	]);
});

test('Sequence step validation errors stay beside the open step editor', async ({ page }) => {
	await mockSequenceStore(
		page,
		[
			sequenceDetail(1, 'TRANSMIT_CHECK', [
				sequenceStep(1),
				sequenceStep(2, { function_name: 'RX_PER' })
			])
		],
		{
			rejectSaveOnce: {
				error: 'Step rejected',
				fields: { 'steps[1].timeout_ms': 'Timeout must be a number.' }
			}
		}
	);
	await page.goto('/sequences/1');
	const steps = page.locator('.steps > li');
	await steps.nth(1).getByRole('button', { name: 'Edit', exact: true }).click();
	await steps.nth(1).getByLabel('Timeout-ms', { exact: true }).fill('soon');
	await steps.nth(1).getByRole('button', { name: 'Save step', exact: true }).click();
	await expect(steps.nth(1).getByRole('alert')).toContainText(
		'Step rejected — Timeout must be a number.'
	);
	await expect(steps.nth(1).getByLabel('Timeout-ms', { exact: true })).toHaveValue('soon');
	await expect(page.getByRole('alert')).toHaveCount(1);
});

test('Composer adds a step without restoring old Sequence metadata', async ({
	page
}) => {
	await mockDeviceHub(page);
	const store = await mockSequenceStore(page, [
		sequenceDetail(1, 'TRANSMIT_CHECK', [sequenceStep(1)])
	]);
	await page.route('**/api/sequences/1', (route) => {
		if (route.request().method() !== 'GET') return route.fallback();
		const sequence = store.sequences[0];
		const snapshot = { ...sequence };
		// Another tab saves metadata between the composer's read and write.
		sequence.name = 'RENAMED_CHECK';
		sequence.description = 'Saved in another tab';
		return route.fulfill({ json: snapshot });
	});
	await page.route('**/api/functions/1/compose', (route) =>
		route.fulfill({ json: { script: 'TX|ANT=1' } })
	);
	await page.goto('/devices/WIFI_TEST?tab=functions');
	await page.getByRole('button', { name: 'Compose script', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Script Composer' });
	await expect(dialog).toBeVisible();
	await dialog.getByText('Add to sequence').click();
	await expect(dialog.getByRole('link', { name: 'sequences', exact: true })).toHaveAttribute(
		'href',
		'/sequences'
	);
	await dialog.getByRole('button', { name: 'Target sequence' }).click();
	await page.getByRole('option', { name: 'TRANSMIT_CHECK (1 steps)' }).click();
	await dialog.getByRole('button', { name: 'Add step' }).click();
	await expect(dialog.getByText('Added as a new step.')).toBeVisible();
	expect(store.sequences[0]).toMatchObject({
		name: 'RENAMED_CHECK',
		description: 'Saved in another tab'
	});
	const appended = store.requests.filter((r) => r.method === 'PUT').at(-1)!.body as SequenceInput;
	expect(appended.steps.map((step) => step.function_name)).toEqual([
		'TX_MULTI_VERIFICATION',
		'TX_MULTI_VERIFICATION'
	]);
	expect(retainedFields(appended.steps[0])).toEqual(retainedFields(sequenceStep(1)));
});

test('Sequence identifiers share the catalog identifier typography', async ({ page }) => {
	await mockSequenceStore(page, [sequenceDetail(1, 'TRANSMIT_CHECK', [sequenceStep(1)])]);
	await page.goto('/sequences/1');
	const heading = page.getByRole('heading', { name: 'TRANSMIT_CHECK', exact: true });
	await expect(heading).toBeVisible({ timeout: 15_000 });
	expect(await heading.evaluate((el) => getComputedStyle(el).fontFamily)).toContain(
		'Cascadia Code'
	);
});

test('Sequence pages stay readable with long steps on desktop, mobile and dark theme', async ({
	page
}, testInfo) => {
	await mockDeviceHub(page);
	const name = `SEQUENCE_${'LONG_NAME_'.repeat(8)}`;
	const value = 'UNBROKEN_VALUE_'.repeat(20);
	await mockSequenceStore(page, [
		sequenceDetail(1, name, [
			sequenceStep(1, {
				function_name: `TX_${'LONG_IDENTIFIER_'.repeat(10)}`,
				parameters: { ANT: value },
				note: value
			}),
			sequenceStep(2, { function_name: 'RX_PER', device_name: '' })
		])
	]);
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	const fits = () =>
		page.evaluate(
			() => document.documentElement.scrollWidth <= document.documentElement.clientWidth
		);
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: 900 });
		await page.goto('/sequences/1');
		await expect(page.getByRole('heading', { name, exact: true })).toBeVisible({
			timeout: 15_000
		});
		await expect(page.locator('.emitted-preview')).toBeVisible({ timeout: 15_000 });
		expect(await fits(), `${width}px detail overflow`).toBe(true);
		await page.screenshot({
			path: testInfo.outputPath(`sequence-detail-${width}.png`),
			fullPage: true
		});

		const firstStep = page.locator('.steps > li').first();
		await firstStep.getByRole('button', { name: 'Edit', exact: true }).click();
		await expect(page.getByRole('button', { name: 'Save step', exact: true })).toBeVisible();
		expect(await fits(), `${width}px step editor overflow`).toBe(true);
		await page.screenshot({
			path: testInfo.outputPath(`sequence-editor-${width}.png`),
			fullPage: true
		});
		await firstStep.getByRole('button', { name: 'Cancel', exact: true }).click();

		await page.goto('/sequences');
		await expect(page.getByRole('link', { name, exact: true })).toBeVisible({
			timeout: 15_000
		});
		expect(await fits(), `${width}px list overflow`).toBe(true);
		await page.screenshot({
			path: testInfo.outputPath(`sequences-list-${width}.png`),
			fullPage: true
		});
	}
	await page.goto('/sequences/1');
	await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
	await expect(page.locator('html')).toHaveClass(/dark/);
	await page.screenshot({
		path: testInfo.outputPath('sequence-detail-dark-390.png'),
		fullPage: true
	});
	await page.goto('/sequences');
	await expect(page.getByRole('link', { name, exact: true })).toBeVisible({ timeout: 15_000 });
	await page.screenshot({
		path: testInfo.outputPath('sequences-list-dark-390.png'),
		fullPage: true
	});
	expect(errors).toEqual([]);
});

type OrphanStore = {
	parameterLinks: OrphanLinkRow[];
	criteriaLinks: OrphanLinkRow[];
	examples: OrphanExampleRow[];
};

// Representative missing-target groups: a Function and a Parameter definition
// missing together, a Global link with a missing definition, a Criteria
// definition missing on a Device, and an Example whose Function is missing.
function orphanStoreFixture(): OrphanStore {
	return {
		parameterLinks: [
			{
				id: 101,
				function_source: 'DLL',
				function_name: 'TX_MULTI_VERIFICATION',
				function_exists: false,
				function_id: null,
				name: 'ANT',
				definition_exists: false,
				device_name: 'WIFI_TEST',
				position: 1,
				required: true
			},
			{
				id: 102,
				function_source: 'DLL',
				function_name: 'RX_PER',
				function_exists: true,
				function_id: 2,
				name: 'SETTLE_MS',
				definition_exists: false,
				device_name: '',
				position: 2,
				required: false
			}
		],
		criteriaLinks: [
			{
				id: 201,
				function_source: 'DLL',
				function_name: 'TX_MULTI_VERIFICATION',
				function_exists: true,
				function_id: 1,
				name: 'POWER',
				definition_exists: false,
				device_name: 'GOLDENKV',
				position: 1,
				required: true
			}
		],
		examples: [
			{
				id: 301,
				function_source: 'DLL',
				function_name: 'LEGACY_TX',
				device_name: 'WIFI_TEST',
				script: 'TX|LEGACY=1',
				meaning: 'Legacy transmit usage'
			}
		]
	};
}

function orphanReport(store: OrphanStore): OrphanReport {
	const groups: OrphanReport['groups'] = [
		{ kind: 'parameter-links', count: store.parameterLinks.length, items: store.parameterLinks },
		{ kind: 'criteria-links', count: store.criteriaLinks.length, items: store.criteriaLinks },
		{ kind: 'examples', count: store.examples.length, items: store.examples }
	];
	return { total: groups.reduce((sum, group) => sum + group.count, 0), groups };
}

// A row leaves the report only when its Function and definition both exist.
function settleOrphanStore(store: OrphanStore) {
	const resolved = (link: OrphanLinkRow) => link.function_exists && link.definition_exists;
	store.parameterLinks = store.parameterLinks.filter((link) => !resolved(link));
	store.criteriaLinks = store.criteriaLinks.filter((link) => !resolved(link));
}

// Stateful fake of the orphan report and its three resolution endpoints. Rules
// mirror the server: a Function resolves by (source, name); a definition
// resolves a link by name when its scope is the link's Device or Global.
async function mockOrphanStore(
	page: Page,
	store: OrphanStore,
	options: {
		failLoadOnce?: boolean;
		failCreateOnce?: { status: number; body: { error: string; fields?: Record<string, string> } };
	} = {}
) {
	const loads = { count: 0 };
	const mutations: { method: string; path: string; body: Record<string, unknown> }[] = [];
	let failLoad = options.failLoadOnce ?? false;
	let failCreate = options.failCreateOnce;
	await page.route(/\/api\/orphans(?:\?.*)?$/, (route) => {
		if (route.request().method() !== 'GET') return route.fallback();
		loads.count++;
		if (failLoad) {
			failLoad = false;
			return route.fulfill({ status: 503, json: { error: 'Orphan report is unavailable.' } });
		}
		return route.fulfill({ json: orphanReport(store) });
	});
	await page.route(/\/api\/functions(?:\?.*)?$/, (route) => {
		if (route.request().method() !== 'POST') return route.fallback();
		const body = route.request().postDataJSON() as { source: string; name: string };
		mutations.push({ method: 'POST', path: '/api/functions', body });
		if (failCreate) {
			const failure = failCreate;
			failCreate = undefined;
			return route.fulfill({ status: failure.status, json: failure.body });
		}
		for (const link of [...store.parameterLinks, ...store.criteriaLinks]) {
			if (link.function_source === body.source && link.function_name === body.name) {
				link.function_exists = true;
				link.function_id = 9;
			}
		}
		store.examples = store.examples.filter(
			(row) => !(row.function_source === body.source && row.function_name === body.name)
		);
		settleOrphanStore(store);
		return route.fulfill({ status: 201, json: { id: 9, ...body } });
	});
	await page.route(/\/api\/(?:parameter|criteria)-definitions(?:\?.*)?$/, (route) => {
		if (route.request().method() !== 'POST') return route.fallback();
		const url = new URL(route.request().url());
		const body = route.request().postDataJSON() as { name: string; device_name?: string };
		mutations.push({ method: 'POST', path: url.pathname, body });
		if (failCreate) {
			const failure = failCreate;
			failCreate = undefined;
			return route.fulfill({ status: failure.status, json: failure.body });
		}
		const links = url.pathname.startsWith('/api/criteria')
			? store.criteriaLinks
			: store.parameterLinks;
		const scope = body.device_name ?? '';
		for (const link of links) {
			if (link.name === body.name && (scope === '' || scope === link.device_name))
				link.definition_exists = true;
		}
		settleOrphanStore(store);
		return route.fulfill({ status: 201, json: { id: 900, ...body } });
	});
	return { loads, mutations };
}

function orphanRow(page: Page, cell: string) {
	return page.getByRole('row').filter({ has: page.getByRole('cell', { name: cell, exact: true }) });
}

test('Orphan report identifies missing targets, usage records, and Device or Global context', async ({
	page
}) => {
	await mockDeviceHub(page);
	await mockOrphanStore(page, orphanStoreFixture());
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.goto('/orphans');
	await expect(page.getByRole('heading', { name: 'Orphan report', exact: true })).toBeVisible({
		timeout: 20_000
	});
	await expect(page.getByText('4 unresolved references', { exact: true })).toBeVisible();
	await expect(page.getByText('Missing Functions: 2', { exact: true })).toBeVisible();
	await expect(page.getByText('Missing Parameter definitions: 2', { exact: true })).toBeVisible();
	await expect(page.getByText('Missing Criteria definitions: 1', { exact: true })).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'Parameter links (2)', exact: true })
	).toBeVisible();
	await expect(
		page.getByRole('heading', { name: 'Criteria links (1)', exact: true })
	).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Examples (1)', exact: true })).toBeVisible();

	const ant = orphanRow(page, 'ANT');
	await expect(ant.getByText('DLL/TX_MULTI_VERIFICATION', { exact: true })).toBeVisible();
	await expect(ant.getByRole('link')).toHaveCount(0);
	await expect(ant.getByText('Function missing', { exact: true })).toBeVisible();
	await expect(ant.getByText('Parameter definition missing', { exact: true })).toBeVisible();
	await expect(ant.getByText('WIFI_TEST', { exact: true })).toBeVisible();
	await expect(
		ant.getByRole('button', { name: 'Create missing Function', exact: true })
	).toBeVisible();

	const settle = orphanRow(page, 'SETTLE_MS');
	await expect(settle.getByRole('link', { name: 'DLL/RX_PER', exact: true })).toHaveAttribute(
		'href',
		'/functions/2'
	);
	await expect(settle.getByText('Parameter definition missing', { exact: true })).toBeVisible();
	await expect(settle.getByText('Function missing', { exact: true })).toHaveCount(0);
	await expect(settle.getByText('Global', { exact: true })).toBeVisible();
	await expect(
		settle.getByRole('button', { name: 'Create missing Parameter definition', exact: true })
	).toBeVisible();

	const power = orphanRow(page, 'POWER');
	await expect(
		power.getByRole('link', { name: 'DLL/TX_MULTI_VERIFICATION', exact: true })
	).toHaveAttribute('href', '/functions/1');
	await expect(power.getByText('GOLDENKV', { exact: true })).toBeVisible();
	await expect(power.getByText('Criteria definition missing', { exact: true })).toBeVisible();

	const example = orphanRow(page, 'TX|LEGACY=1');
	await expect(example.getByText('DLL/LEGACY_TX', { exact: true })).toBeVisible();
	await expect(example.getByText('WIFI_TEST', { exact: true })).toBeVisible();
	await expect(example.getByText('Function missing', { exact: true })).toBeVisible();

	await page.getByRole('button', { name: 'Criteria links (1)', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'Parameter links (2)', exact: true })).toHaveCount(
		0
	);
	await expect(
		page.getByRole('heading', { name: 'Criteria links (1)', exact: true })
	).toBeVisible();
	await page.getByRole('button', { name: 'All (4)', exact: true }).click();
	await expect(
		page.getByRole('heading', { name: 'Parameter links (2)', exact: true })
	).toBeVisible();
});

test('Orphan repair prefills affected context, opens from the keyboard, and cancelling keeps every row', async ({
	page
}) => {
	await mockDeviceHub(page);
	const { mutations } = await mockOrphanStore(page, orphanStoreFixture());
	await page.goto('/orphans');

	const repair = orphanRow(page, 'SETTLE_MS').getByRole('button', {
		name: 'Create missing Parameter definition',
		exact: true
	});
	await repair.focus();
	await page.keyboard.press('Enter');
	const parameter = page.getByRole('dialog', {
		name: 'Create parameter definition "SETTLE_MS"',
		exact: true
	});
	await expect(parameter).toBeVisible();
	await expect(parameter.getByRole('textbox', { name: 'Name', exact: true })).toHaveValue(
		'SETTLE_MS'
	);
	await expect(parameter.getByText('Parameter link SETTLE_MS', { exact: true })).toBeVisible();
	await expect(parameter.getByText('DLL/RX_PER', { exact: true })).toBeVisible();
	await expect(parameter.getByText('Global', { exact: true })).toBeVisible();
	await expect(
		parameter.getByRole('button', { name: 'Device / Interface: Global definitions', exact: true })
	).toBeDisabled();
	await page.keyboard.press('Escape');
	await expect(parameter).toHaveCount(0);
	await expect(repair).toBeFocused();

	await orphanRow(page, 'POWER')
		.getByRole('button', { name: 'Create missing Criteria definition', exact: true })
		.press('Enter');
	const criteria = page.getByRole('dialog', {
		name: 'Create criteria definition "POWER"',
		exact: true
	});
	await expect(
		criteria.getByRole('button', { name: 'Device / Interface: GOLDENKV', exact: true })
	).toBeDisabled();
	await criteria.getByRole('button', { name: 'Cancel', exact: true }).press('Enter');
	await expect(criteria).toHaveCount(0);
	await expect(
		orphanRow(page, 'POWER').getByRole('button', {
			name: 'Create missing Criteria definition',
			exact: true
		})
	).toBeFocused();

	expect(mutations).toHaveLength(0);
	await expect(page.getByText('4 unresolved references', { exact: true })).toBeVisible();
	await expect(orphanRow(page, 'SETTLE_MS')).toHaveCount(1);
	await expect(orphanRow(page, 'POWER')).toHaveCount(1);
});

test('Successful Function repair refreshes the report instead of assuming related orphans resolved', async ({
	page
}) => {
	await mockDeviceHub(page);
	const { loads, mutations } = await mockOrphanStore(page, orphanStoreFixture());
	await page.goto('/orphans');
	await expect.poll(() => loads.count).toBe(1);

	const ant = orphanRow(page, 'ANT');
	await ant.getByRole('button', { name: 'Create missing Function', exact: true }).click();
	const dialog = page.getByRole('dialog', {
		name: 'Create function DLL/TX_MULTI_VERIFICATION',
		exact: true
	});
	await expect(dialog.getByRole('textbox', { name: 'Name', exact: true })).toHaveValue(
		'TX_MULTI_VERIFICATION'
	);
	await expect(dialog.getByText('Parameter link ANT', { exact: true })).toBeVisible();
	await expect(dialog.getByText('WIFI_TEST', { exact: true })).toBeVisible();
	await dialog.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(dialog).toHaveCount(0);

	expect(mutations).toHaveLength(1);
	expect(mutations[0]).toMatchObject({
		path: '/api/functions',
		body: { source: 'DLL', name: 'TX_MULTI_VERIFICATION', interface: '' }
	});
	await expect.poll(() => loads.count).toBe(2);
	// ANT still waits on its Parameter definition, so its row stays and changes action.
	await expect(
		ant.getByRole('button', { name: 'Create missing Parameter definition', exact: true })
	).toBeVisible();
	await expect(ant.getByText('Function missing', { exact: true })).toHaveCount(0);
	await expect(page.getByText('Missing Functions: 1', { exact: true })).toBeVisible();
	await expect(page.getByText('4 unresolved references', { exact: true })).toBeVisible();
});

test('Creating a missing Function for an Example refreshes the report and removes only that usage', async ({
	page
}) => {
	await mockDeviceHub(page);
	await mockOrphanStore(page, orphanStoreFixture());
	await page.goto('/orphans');
	await orphanRow(page, 'TX|LEGACY=1')
		.getByRole('button', { name: 'Create missing Function', exact: true })
		.click();
	const dialog = page.getByRole('dialog', { name: 'Create Function DLL/LEGACY_TX', exact: true });
	await expect(dialog.getByRole('textbox', { name: 'Name', exact: true })).toHaveValue('LEGACY_TX');
	await expect(dialog.getByText('Example', { exact: true })).toBeVisible();
	await expect(dialog.getByText('WIFI_TEST', { exact: true })).toBeVisible();
	await dialog.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	await expect(page.getByRole('heading', { name: 'Examples (1)', exact: true })).toHaveCount(0);
	await expect(page.getByText('3 unresolved references', { exact: true })).toBeVisible();
	await expect(page.getByRole('button', { name: 'All (3)', exact: true })).toBeVisible();
});

test('Definition repair targets the affected Device or Global scope and resolves only matching usage', async ({
	page
}) => {
	await mockDeviceHub(page);
	const { mutations } = await mockOrphanStore(page, orphanStoreFixture());
	await page.goto('/orphans');

	await orphanRow(page, 'POWER')
		.getByRole('button', { name: 'Create missing Criteria definition', exact: true })
		.click();
	const criteria = page.getByRole('dialog', {
		name: 'Create criteria definition "POWER"',
		exact: true
	});
	await criteria.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(criteria).toHaveCount(0);
	expect(mutations.at(-1)).toMatchObject({
		path: '/api/criteria-definitions',
		body: { name: 'POWER', device_name: 'GOLDENKV' }
	});
	await expect(page.getByRole('heading', { name: 'Criteria links (1)', exact: true })).toHaveCount(
		0
	);
	await expect(page.getByText('3 unresolved references', { exact: true })).toBeVisible();

	await orphanRow(page, 'SETTLE_MS')
		.getByRole('button', { name: 'Create missing Parameter definition', exact: true })
		.click();
	const parameter = page.getByRole('dialog', {
		name: 'Create parameter definition "SETTLE_MS"',
		exact: true
	});
	await parameter.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(parameter).toHaveCount(0);
	expect(mutations.at(-1)).toMatchObject({
		path: '/api/parameter-definitions',
		body: { name: 'SETTLE_MS', device_name: '' }
	});
	await expect(orphanRow(page, 'SETTLE_MS')).toHaveCount(0);
	await expect(
		page.getByRole('heading', { name: 'Parameter links (1)', exact: true })
	).toBeVisible();
	await expect(orphanRow(page, 'ANT')).toHaveCount(1);
});

test('Rejected repair keeps unresolved rows, shows the field error, and succeeds on a corrected retry', async ({
	page
}) => {
	await mockDeviceHub(page);
	const { loads, mutations } = await mockOrphanStore(page, orphanStoreFixture(), {
		failCreateOnce: {
			status: 422,
			body: { error: 'Invalid function', fields: { name: 'Name is required.' } }
		}
	});
	await page.goto('/orphans');
	await orphanRow(page, 'ANT')
		.getByRole('button', { name: 'Create missing Function', exact: true })
		.click();
	const dialog = page.getByRole('dialog', {
		name: 'Create function DLL/TX_MULTI_VERIFICATION',
		exact: true
	});
	await dialog.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(dialog.getByRole('alert')).toContainText('Invalid function');
	await expect(dialog.getByRole('alert')).toContainText('stay unresolved');
	await expect(dialog.getByText('Name is required.', { exact: true })).toBeVisible();
	expect(loads.count).toBe(1);
	await expect(orphanRow(page, 'ANT')).toHaveCount(1);

	await dialog.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(dialog).toHaveCount(0);
	expect(mutations).toHaveLength(2);
	await expect.poll(() => loads.count).toBe(2);
	await expect(
		orphanRow(page, 'ANT').getByRole('button', {
			name: 'Create missing Parameter definition',
			exact: true
		})
	).toBeVisible();
});

test('Server failure keeps the unresolved definition row and lets the user cancel without refreshing', async ({
	page
}) => {
	await mockDeviceHub(page);
	const { loads, mutations } = await mockOrphanStore(page, orphanStoreFixture(), {
		failCreateOnce: { status: 500, body: { error: 'Repair could not be saved.' } }
	});
	await page.goto('/orphans');
	await orphanRow(page, 'SETTLE_MS')
		.getByRole('button', { name: 'Create missing Parameter definition', exact: true })
		.click();
	const dialog = page.getByRole('dialog', {
		name: 'Create parameter definition "SETTLE_MS"',
		exact: true
	});
	await dialog.getByRole('button', { name: 'Create and resolve', exact: true }).click();
	await expect(dialog.getByRole('alert')).toContainText('Repair could not be saved.');
	await expect(
		dialog.getByRole('button', { name: 'Create and resolve', exact: true })
	).toBeEnabled();
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(dialog).toHaveCount(0);

	expect(mutations).toHaveLength(1);
	expect(loads.count).toBe(1);
	await expect(orphanRow(page, 'SETTLE_MS')).toHaveCount(1);
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Orphan report shows the fully linked state when no unresolved references remain', async ({
	page
}) => {
	await mockDeviceHub(page);
	await mockOrphanStore(page, { parameterLinks: [], criteriaLinks: [], examples: [] });
	await page.goto('/orphans');
	await expect(
		page.getByText('No unresolved references. The knowledge base is fully linked.', {
			exact: true
		})
	).toBeVisible({ timeout: 15_000 });
	await expect(page.getByRole('button', { name: /^Create missing/ })).toHaveCount(0);
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Orphan report retries a failed load before presenting any rows', async ({ page }) => {
	await mockDeviceHub(page);
	await mockOrphanStore(page, orphanStoreFixture(), { failLoadOnce: true });
	await page.goto('/orphans');
	await expect(page.getByRole('alert')).toContainText('Orphan report is unavailable.', {
		timeout: 15_000
	});
	await expect(orphanRow(page, 'ANT')).toHaveCount(0);
	await page.getByRole('button', { name: 'Retry orphan report', exact: true }).click();
	await expect(orphanRow(page, 'ANT')).toBeVisible();
	await expect(page.locator('.banner-error')).toHaveCount(0);
});

test('Orphan report keeps long identities in local scrolling on desktop, mobile and dark theme', async ({
	page
}, testInfo) => {
	await mockDeviceHub(page);
	const store = orphanStoreFixture();
	store.parameterLinks[0].function_name = `TX_${'LONG_IDENTIFIER_'.repeat(10)}`;
	store.parameterLinks[0].name = `ANT_${'LONG_DEFINITION_NAME_'.repeat(4)}`;
	store.examples[0].script = `TX|LEGACY=${'UNBROKEN_SCRIPT_VALUE_'.repeat(20)}`;
	await mockOrphanStore(page, store);
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	const fits = () =>
		page.evaluate(
			() => document.documentElement.scrollWidth <= document.documentElement.clientWidth
		);
	for (const width of [1440, 390]) {
		await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
		await page.goto('/orphans');
		await expect(page.getByRole('heading', { name: 'Orphan report', exact: true })).toBeVisible();
		await expect(orphanRow(page, store.examples[0].script)).toBeVisible();
		expect(await fits(), `${width}px orphan page overflow`).toBe(true);
		const table = await page.locator('.orphan-table').first().boundingBox();
		expect(table!.x + table!.width).toBeLessThanOrEqual(width);
		await expect(page.locator('.banner-error')).toHaveCount(0);
		await page.screenshot({ path: testInfo.outputPath(`orphans-${width}.png`), fullPage: true });
	}
	await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
	await expect(page.locator('html')).toHaveClass(/dark/);
	await page.screenshot({ path: testInfo.outputPath('orphans-390-dark.png'), fullPage: true });
	await orphanRow(page, store.examples[0].script)
		.getByRole('button', { name: 'Create missing Function', exact: true })
		.click();
	await expect(page.getByRole('dialog', { name: /^Create Function/ })).toBeVisible();
	await page.screenshot({ path: testInfo.outputPath('orphan-repair-390-dark.png') });
	await page.keyboard.press('Escape');
	await expect(page.getByRole('dialog')).toHaveCount(0);
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.screenshot({ path: testInfo.outputPath('orphans-1440-dark.png'), fullPage: true });
	expect(errors).toEqual([]);
});

async function simulateHttpClipboard(page: Page, denyModern = false) {
	await page.addInitScript((denyModern) => {
		const clipboard = navigator.clipboard;
		Object.assign(window, { readCopiedText: () => clipboard.readText() });
		Object.defineProperty(navigator, 'clipboard', {
			configurable: true,
			value: denyModern
				? { writeText: () => Promise.reject(new Error('Clipboard permission denied')) }
				: undefined
		});
	}, denyModern);
}

for (const surface of ['Function examples', 'Device examples', 'Composer', 'Sequence'] as const) {
	test(`UI controls: HTTP clipboard fallback copies ${surface} and displays a toast`, async ({
		page,
		context
	}) => {
		await context.grantPermissions(['clipboard-read', 'clipboard-write']);
		await simulateHttpClipboard(page);
		await mockDeviceHub(page);
		const script =
			surface === 'Composer' || surface === 'Sequence' ? 'FIRST|ANT=1\nSECOND|ANT=2' : 'TX|ANT=1';
		let button;
		if (surface === 'Sequence') {
			await page.route('**/api/sequences/1', (route) =>
				route.fulfill({
					json: {
						id: 1,
						name: 'COPY_SEQUENCE',
						description: '',
						created_at: '',
						updated_at: '',
						steps: ['FIRST', 'SECOND'].map((function_name, index) => ({
							id: index + 1,
							position: index + 1,
							function_source: 'DLL',
							function_interface: 'WIFI_TEST',
							function_name,
							parameters: { ANT: String(index + 1) },
							criteria: {},
							timeout_ms: '',
							flow: '',
							on_fail: '',
							device_name: 'WIFI_TEST',
							note: '',
							function_exists: true
						}))
					}
				})
			);
			await page.route('**/api/sequences/1/emit', (route) =>
				route.fulfill({
					json: {
						text: script,
						steps: script.split('\n').map((script, index) => ({
							step_id: index + 1,
							position: index + 1,
							resolved: true,
							script,
							errors: {}
						}))
					}
				})
			);
			await page.goto('/sequences/1');
			button = page.getByRole('button', { name: 'Copy all', exact: true });
		} else {
			await page.route('**/api/functions/1/compose', (route) =>
				route.fulfill({ json: { script } })
			);
			await page.goto(
				`/devices/WIFI_TEST?tab=${surface === 'Device examples' ? 'examples' : 'functions'}`
			);
			if (surface === 'Composer') {
				await page.getByRole('button', { name: 'Compose script', exact: true }).click();
				await expect(page.getByRole('dialog').locator('code')).toHaveText(script);
				button = page.getByRole('dialog').getByRole('button', { name: 'Copy', exact: true });
			} else {
				button = page.getByRole('button', {
					name: surface === 'Device examples' ? 'Copy device example 31' : 'Copy example 31',
					exact: true
				});
			}
		}
		await button.click();
		await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
		const copied = await page.evaluate(() =>
			(window as unknown as { readCopiedText: () => Promise<string> }).readCopiedText()
		);
		expect(copied.replace(/\r\n/g, '\n')).toBe(script);
		await expect(button).toBeFocused();
		if (surface === 'Composer') await expect(page.getByRole('dialog')).toBeVisible();
	});
}

test('UI controls: denied modern clipboard permissions use the fallback', async ({
	page,
	context
}) => {
	await context.grantPermissions(['clipboard-read', 'clipboard-write']);
	await simulateHttpClipboard(page, true);
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST');
	await page.getByRole('button', { name: 'Copy example 31', exact: true }).click();
	await expect(page.getByText('Copied to clipboard', { exact: true })).toBeVisible();
	expect(
		await page.evaluate(() =>
			(window as unknown as { readCopiedText: () => Promise<string> }).readCopiedText()
		)
	).toBe('TX|ANT=1');
});

test('UI controls: failed copies show an error instead of a success toast', async ({ page }) => {
	await page.addInitScript(() => {
		Object.defineProperty(navigator, 'clipboard', { configurable: true, value: undefined });
		document.execCommand = () => false;
	});
	await mockDeviceHub(page);
	await page.goto('/devices/WIFI_TEST?tab=examples');
	await page.getByRole('button', { name: 'Copy device example 31', exact: true }).click();
	await expect(
		page.getByText('Copy failed. Select the text and copy it manually.', { exact: true })
	).toBeVisible();
	await expect(page.getByText('Copied to clipboard', { exact: true })).toHaveCount(0);
	await expect(page.getByText('Script copied.', { exact: true })).toHaveCount(0);
});

test('UI controls: dropdown search supports empty results, keyboard selection and blank choices', async ({
	page
}) => {
	await mockFunctionCatalog(page);
	await page.goto('/functions');
	const trigger = page.getByRole('button', { name: 'Filter by source', exact: true });
	await trigger.click();
	const popup = page.getByRole('dialog', { name: 'Choose Filter by source', exact: true });
	const search = popup.getByRole('combobox', { name: 'Search options', exact: true });
	await expect(search).toBeFocused();
	await search.fill('no matching source');
	await expect(popup.getByRole('option')).toHaveCount(0);
	await expect(popup.getByRole('status')).toHaveText('No options match your search.');
	await search.fill('  buILT  ');
	await expect(popup.getByRole('option')).toHaveCount(1);
	await page.keyboard.press('ArrowDown');
	await page.keyboard.press('Enter');
	await expect(trigger).toHaveText('BUILT-IN');
	await expect(page).toHaveURL(/source=BUILT-IN/);
	await expect(trigger).toBeFocused();
	await trigger.click();
	await expect(search).toHaveValue('');
	await search.fill('all sources');
	await popup.getByRole('option', { name: 'All sources', exact: true }).click();
	await expect(trigger).toHaveText('All sources');
	await expect(page).not.toHaveURL(/source=/);
});

test('UI controls: searchable dropdown works inside a modal without submitting its form', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Create device', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Create Device', exact: true });
	const trigger = dialog.getByRole('button', { name: 'Type', exact: true });
	await trigger.click();
	const popup = page.getByRole('dialog', { name: 'Choose Type', exact: true });
	const search = popup.getByRole('combobox', { name: 'Search options', exact: true });
	await search.fill('dLl');
	await expect(popup.getByRole('option')).toHaveCount(1);
	await page.keyboard.press('ArrowDown');
	await page.keyboard.press('Enter');
	await expect(trigger).toHaveText('DLL');
	await expect(trigger).toBeFocused();
	await expect(dialog).toBeVisible();
	await expect(dialog.getByRole('textbox', { name: 'Device name', exact: true })).toHaveValue('');
});

for (const viewport of [
	{ width: 1440, height: 900 },
	{ width: 390, height: 844 }
]) {
	test(`UI controls: Device sorting dropdown search fits ${viewport.width}px screens`, async ({
		page
	}) => {
		await page.setViewportSize(viewport);
		await mockDeviceHub(page);
		await page.goto('/devices');
		await page.getByRole('button', { name: 'Sort devices', exact: true }).click();
		const popup = page.getByRole('dialog', { name: 'Choose Sort devices', exact: true });
		await popup.getByRole('combobox', { name: 'Search options', exact: true }).fill('EXAMPLES');
		await expect(popup.getByRole('option')).toHaveCount(1);
		await expect(popup.getByRole('option', { name: 'Examples', exact: true })).toBeVisible();
		const bounds = await popup.boundingBox();
		expect(bounds!.x).toBeGreaterThanOrEqual(0);
		expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(viewport.width);
		await page.screenshot({ path: test.info().outputPath(`dropdown-${viewport.width}.png`) });
		await popup.getByRole('option', { name: 'Examples', exact: true }).click();
		await expect(page.getByRole('button', { name: 'Sort devices', exact: true })).toHaveText(
			'Examples'
		);
		expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
			viewport.width
		);
	});
}

async function assertIndependentSidebar(page: Page, railSelector: string, paneSelector: string) {
	const rail = page.locator(railSelector);
	await expect
		.poll(() => rail.evaluate((element) => element.scrollHeight > element.clientHeight))
		.toBe(true);
	await rail.hover();
	const before = await page.locator(paneSelector).evaluate((element) => ({
		top: element.getBoundingClientRect().top,
		pageScroll: window.scrollY
	}));
	await page.mouse.wheel(0, 1200);
	await expect.poll(() => rail.evaluate((element) => element.scrollTop)).toBeGreaterThan(0);
	const after = await page.locator(paneSelector).evaluate((element) => ({
		top: element.getBoundingClientRect().top,
		pageScroll: window.scrollY
	}));
	expect(after).toEqual(before);
	await page.screenshot({ path: test.info().outputPath('sidebar-scrolled.png') });
}

for (const viewport of [
	{ width: 1440, height: 900 },
	{ width: 390, height: 844 }
]) {
	for (const catalog of ['Functions', 'Device Functions'] as const) {
		test(`UI controls: ${catalog} sidebar scrolls independently at ${viewport.width}px`, async ({
			page
		}) => {
			await page.setViewportSize(viewport);
			const { functions } = await mockFunctionCatalog(page);
			functions.push(
				...Array.from({ length: 28 }, (_value, index) => ({
					...functions[0],
					id: index + 10,
					name: `FUNCTION_${String(index).padStart(2, '0')}`
				}))
			);
			await page.goto(catalog === 'Functions' ? '/functions' : '/devices/WIFI_TEST');
			await assertIndependentSidebar(page, '.function-list', '.detail-pane');
			expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(
				viewport.width
			);
		});
	}
}

for (const kind of ['parameter-definitions', 'criteria-definitions'] as const) {
	test(`catalog workspace: ${kind} has flush pagination and consistent controls`, async ({
		page
	}) => {
		const { rows } = await mockDictionary(page, kind);
		rows.push(
			...Array.from({ length: 330 }, (_, index) => ({
				...rows[0],
				id: index + 10,
				name: `PAGED_${index}`
			}))
		);
		for (const width of [1440, 390]) {
			await page.setViewportSize({ width, height: 900 });
			await page.goto(`/${kind}`);
			const rail = page.locator('.definition-list');
			await expect(page.locator('.dictionary-workspace')).toHaveCSS('border-radius', '6px');
			await expect(rail.locator('.rail-heading h2')).toHaveText(
				kind === 'parameter-definitions' ? 'Parameters' : 'Criteria'
			);
			await expect(rail.locator('[data-pagination-root]')).toBeVisible();
			expect(
				await rail
					.getByRole('button', { name: 'Page size', exact: true })
					.locator('span')
					.evaluate((element) => element.scrollWidth <= element.clientWidth)
			).toBe(true);
			await expect(page.locator('.banner-error')).toHaveCount(0);
			const geometry = await rail.evaluate((element) => {
				const box = element.getBoundingClientRect();
				const pager = element.querySelector('[data-pagination-root]')!;
				const counter = pager.querySelector('[aria-hidden="true"].tabular-nums')!;
				return {
					bottomGap: element.parentElement!.getBoundingClientRect().bottom - box.bottom,
					pagerGap: box.bottom - pager.getBoundingClientRect().bottom,
					counterHeight: counter.getBoundingClientRect().height,
					navWidth: pager.querySelector('[data-pagination-next]')!.getBoundingClientRect().width,
					overflow: document.documentElement.scrollWidth > document.documentElement.clientWidth
				};
			});
			if (width > 850) expect(Math.abs(geometry.bottomGap)).toBeLessThan(2);
			expect(Math.abs(geometry.pagerGap)).toBeLessThan(2);
			expect(geometry.counterHeight).toBeLessThan(22);
			expect(geometry.navWidth).toBe(32);
			expect(geometry.overflow).toBe(false);
			await page.screenshot({
				path: test.info().outputPath(`${kind}-${width}.png`),
				fullPage: true
			});
		}
	});

	test(`UI controls: ${kind} dictionary sidebar scrolls independently`, async ({ page }) => {
		await page.setViewportSize({ width: 1440, height: 900 });
		const { rows } = await mockDictionary(page, kind);
		rows.push(
			...Array.from({ length: 28 }, (_value, index) => ({
				...rows[0],
				id: index + 10,
				name: `DEFINITION_${String(index).padStart(2, '0')}`
			}))
		);
		await page.goto(`/${kind}`);
		await assertIndependentSidebar(page, '.definition-list', '.detail-pane');
	});
}

for (const kind of ['parameters', 'criteria'] as const) {
	test(`UI controls: Device ${kind} table scrolls independently from its detail pane`, async ({
		page
	}) => {
		await page.setViewportSize({ width: 1440, height: 900 });
		await mockDeviceHub(page);
		await page.route(`**/api/devices/WIFI_TEST/${kind}?*`, (route) =>
			route.fulfill({
				json: {
					items: Array.from({ length: 30 }, (_value, index) => ({
						id: index + 50,
						name: `DEFINITION_${index}`,
						device_name: 'WIFI_TEST',
						description: '',
						type: 'int',
						unit: '',
						default_value: '1',
						allowed_values: '1,2',
						cardinality: 'SINGLE',
						value_pattern: '',
						required: true
					})),
					total: 30,
					global_items: [],
					global_total: 0,
					page: 1,
					page_size: 100
				}
			})
		);
		await page.goto(`/devices/WIFI_TEST?tab=${kind}`);
		await assertIndependentSidebar(page, '.definition-tables:visible', '.definition-pane:visible');
	});
}

for (const kind of ['parameter-definitions', 'criteria-definitions'] as const) {
	for (const theme of ['light', 'dark'] as const) {
		test(`Appearance: ${kind} checkbox is visible and operable in ${theme} mode`, async ({
			page
		}) => {
			const { mutations } = await mockDictionary(page, kind);
			await page.emulateMedia({ colorScheme: theme });
			await page.goto(`/${kind}`);
			await page.getByRole('button', { name: 'New', exact: true }).click();
			const dialog = page.getByRole('dialog');
			const checkbox = dialog.getByRole('checkbox', { name: 'Required', exact: true });
			const label = dialog.locator('label[for="d-required"]');
			await expect(checkbox).not.toBeChecked();
			const contrast = async () =>
				checkbox.evaluate((element) => {
					const luminance = (color: string) => {
						const channels = color
							.match(/[\d.]+/g)!
							.slice(0, 3)
							.map(Number)
							.map((channel) => {
								const value = channel / 255;
								return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
							});
						return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
					};
					const ratio = (first: string, second: string) => {
						const values = [luminance(first), luminance(second)].sort((a, b) => b - a);
						return (values[0] + 0.05) / (values[1] + 0.05);
					};
					const style = getComputedStyle(element);
					const surface = getComputedStyle(element.closest('[role="dialog"]')!).backgroundColor;
					return Math.min(
						ratio(style.borderColor, surface),
						element.getAttribute('aria-checked') === 'true'
							? ratio(style.color, style.backgroundColor)
							: ratio(style.borderColor, style.backgroundColor)
					);
				});
			await expect.poll(contrast).toBeGreaterThanOrEqual(3);
			await label.click();
			await expect(checkbox).toBeChecked();
			await expect(checkbox.locator('svg')).toHaveCount(1);
			await expect.poll(contrast).toBeGreaterThanOrEqual(3);
			const boxBounds = await checkbox.boundingBox();
			const labelBounds = await label.boundingBox();
			expect(labelBounds!.x).toBeGreaterThan(boxBounds!.x + boxBounds!.width);
			expect(
				Math.abs(boxBounds!.y + boxBounds!.height / 2 - labelBounds!.y - labelBounds!.height / 2)
			).toBeLessThan(2);
			await checkbox.focus();
			await page.keyboard.press('Space');
			await expect(checkbox).not.toBeChecked();
			await expect(checkbox).toBeFocused();
			expect(await checkbox.evaluate((element) => getComputedStyle(element).outlineWidth)).toBe(
				'2px'
			);
			await page.keyboard.press('Space');
			await expect(checkbox).toBeChecked();
			await page.screenshot({ path: test.info().outputPath(`checkbox-${kind}-${theme}.png`) });
			await dialog.getByRole('textbox', { name: 'Name', exact: true }).fill('VISIBLE_CHECKBOX');
			await dialog.getByRole('button', { name: 'Save', exact: true }).click();
			await expect(dialog).not.toBeVisible();
			expect(mutations.at(-1)?.body?.required).toBe(true);
		});
	}
}

for (const noun of ['parameter', 'criteria'] as const) {
	test(`Appearance: ${noun} link checkboxes save changes and use clickable form labels`, async ({
		page
	}) => {
		await mockFunctionCatalog(page);
		await page.emulateMedia({ colorScheme: noun === 'criteria' ? 'dark' : 'light' });
		const kind = `${noun}-links`;
		const row = {
			id: 101,
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			name: 'ANT',
			position: 1,
			required: true,
			source_file: '',
			resolved: true,
			definition_id: 11,
			created_at: '',
			updated_at: ''
		};
		const updates: Record<string, unknown>[] = [];
		await page.route(`**/api/functions/1/${kind}`, (route) =>
			route.fulfill({ json: { items: [row] } })
		);
		await page.route(`**/api/functions/1/${kind}/101`, (route) => {
			const body = route.request().postDataJSON();
			updates.push(body);
			Object.assign(row, body);
			return route.fulfill({ json: row });
		});
		await page.route(`**/api/${noun}-definitions?*`, (route) =>
			route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 100 } })
		);
		await page.goto('/functions/1');
		if (noun === 'criteria') await page.getByRole('tab', { name: 'Criteria', exact: true }).click();
		const checkbox = page
			.getByRole('checkbox', { name: 'Toggle required for ANT', exact: true })
			.filter({ visible: true });
		await expect(checkbox).toBeChecked();
		await checkbox.focus();
		await page.keyboard.press('Space');
		await expect.poll(() => updates.at(-1)?.required).toBe(false);
		await expect(checkbox).not.toBeChecked();
		await checkbox.click();
		await expect.poll(() => updates.at(-1)?.required).toBe(true);
		await expect(checkbox).toBeChecked();
		await page.getByRole('button', { name: `Attach ${noun}`, exact: true }).click();
		const dialog = page.getByRole('dialog', {
			name: `Attach ${noun[0].toUpperCase()}${noun.slice(1)}`,
			exact: true
		});
		const field = dialog.getByRole('checkbox', { name: 'Required', exact: true });
		await expect(field).not.toBeChecked();
		await dialog.locator('label[for="add-required"]').click();
		await expect(field).toBeChecked();
		await field.focus();
		await page.keyboard.press('Space');
		await expect(field).not.toBeChecked();
		await page.screenshot({ path: test.info().outputPath(`attach-${noun}.png`) });
	});
}

for (const theme of ['light', 'dark'] as const) {
	test(`Final audit: Function browse filter edit link and Composer support the keyboard in ${theme} mode`, async ({
		page
	}) => {
		const { mutations } = await mockFunctionCatalog(page);
		await page.route('**/api/parameter-definitions?*', (route) =>
			route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 100 } })
		);
		await page.route('**/api/functions/1/compose', (route) =>
			route.fulfill({ json: { script: 'SERVER_EMITTED_SCRIPT' } })
		);
		await page.emulateMedia({ colorScheme: theme });
		await page.goto('/functions?function=1');
		await page.getByRole('button', { name: 'Select RX_PER', exact: true }).press('Enter');
		await expect(page.getByRole('heading', { name: 'RX_PER', exact: true })).toBeVisible();
		await page.getByRole('button', { name: 'Filter by source', exact: true }).press('Enter');
		const search = page.getByRole('combobox', { name: 'Search options', exact: true });
		await expect(search).toBeFocused();
		await search.fill('DLL');
		await page.keyboard.press('ArrowDown');
		await page.keyboard.press('Enter');
		await expect(page).toHaveURL(/source=DLL/);
		await page
			.getByRole('button', { name: 'Select TX_MULTI_VERIFICATION', exact: true })
			.press('Enter');
		const edit = page.getByRole('button', { name: 'Edit function', exact: true });
		await edit.press('Enter');
		const dialog = page.getByRole('dialog', { name: 'Edit Function #1', exact: true });
		await expect(dialog.locator(':focus')).toHaveCount(1);
		await page.keyboard.press('Escape');
		await expect(dialog).toHaveCount(0);
		await expect(edit).toBeFocused();
		await page.getByRole('link', { name: 'Manage function', exact: true }).press('Enter');
		await expect(page).toHaveURL(/\/functions\/1$/);
		const actions = page.getByRole('button', { name: 'Actions for SETTLE_MS', exact: true });
		await actions.press('Enter');
		const menuEdit = page.getByRole('menuitem', { name: 'Edit', exact: true });
		await expect(menuEdit).toBeFocused();
		await page.keyboard.press('ArrowDown');
		await expect(
			page.getByRole('menuitem', { name: 'Delete definition', exact: true })
		).toBeFocused();
		await page.keyboard.press('Escape');
		await expect(page.getByRole('menu')).toHaveCount(0);
		await expect(actions).toBeFocused();
		await page.keyboard.press('Enter');
		await menuEdit.press('Enter');
		const linkRow = page.getByRole('row').filter({
			has: page.getByRole('spinbutton', { name: 'Position', exact: true })
		});
		await expect(linkRow).toContainText('SETTLE_MS');
		await linkRow.getByRole('button', { name: 'Cancel', exact: true }).press('Enter');
		await expect(actions).toBeVisible();
		const attach = page.getByRole('button', { name: 'Attach parameter', exact: true });
		await attach.press('Enter');
		await expect(page.getByRole('dialog', { name: 'Attach Parameter', exact: true })).toBeVisible();
		await page.keyboard.press('Escape');
		await expect(attach).toBeFocused();
		const compose = page.getByRole('button', { name: 'Compose script', exact: true });
		await compose.press('Enter');
		const composer = page.getByRole('dialog', { name: 'Script Composer', exact: true });
		await expect(composer.locator('code')).toHaveText('SERVER_EMITTED_SCRIPT');
		await expect(composer.getByRole('alert')).toHaveCount(0);
		await page.screenshot({ path: test.info().outputPath(`composer-keyboard-${theme}.png`) });
		await page.keyboard.press('Escape');
		await expect(composer).toHaveCount(0);
		await expect(compose).toBeFocused();
		expect(mutations).toEqual([]);
	});
}

test('Appearance: theme follows the system, persists manual changes and ignores old density', async ({
	page
}) => {
	await mockDeviceHub(page);
	await page.addInitScript(() => localStorage.setItem('density', 'comfortable'));
	await page.emulateMedia({ colorScheme: 'light' });
	await page.goto('/devices');
	const toggle = page.getByRole('button', { name: 'Toggle dark mode', exact: true });
	await expect(toggle).toHaveAttribute('aria-pressed', 'false');
	await expect(toggle).toHaveAttribute('title', 'Switch to dark mode');
	await expect(toggle.locator('svg')).toHaveCount(1);
	await expect(
		page.getByRole('button', { name: /Toggle table density|Compact|Comfortable/ })
	).toHaveCount(0);
	await expect(page.locator('html')).not.toHaveClass(/density-comfortable/);
	await page.emulateMedia({ colorScheme: 'dark' });
	await expect(toggle).toHaveAttribute('aria-pressed', 'true');
	await toggle.click();
	await expect(toggle).toHaveAttribute('aria-pressed', 'false');
	await page.emulateMedia({ colorScheme: 'light' });
	await page.emulateMedia({ colorScheme: 'dark' });
	await expect(page.locator('html')).not.toHaveClass(/dark/);
	await page.reload();
	await expect(toggle).toHaveAttribute('aria-pressed', 'false');
	await toggle.focus();
	await page.keyboard.press('Enter');
	await expect(toggle).toHaveAttribute('aria-pressed', 'true');
	await expect(toggle).toHaveAttribute('title', 'Switch to light mode');
	await page.reload();
	await expect(page.locator('html')).toHaveClass(/dark/);
	await expect(page.locator('html')).not.toHaveClass(/density-comfortable/);
});

test('Appearance: theme remains usable when saving preferences is blocked', async ({ page }) => {
	await mockDeviceHub(page);
	await page.addInitScript(() => {
		Object.defineProperty(Storage.prototype, 'setItem', {
			value: () => {
				throw new Error('Storage blocked');
			}
		});
	});
	await page.emulateMedia({ colorScheme: 'light' });
	await page.goto('/devices');
	const toggle = page.getByRole('button', { name: 'Toggle dark mode', exact: true });
	await toggle.click();
	await expect(toggle).toHaveAttribute('aria-pressed', 'true');
	await expect(page.locator('html')).toHaveClass(/dark/);
	await toggle.click();
	await expect(page.locator('html')).not.toHaveClass(/dark/);
});

for (const viewport of [
	{ width: 1920, height: 1080 },
	{ width: 390, height: 844 }
]) {
	for (const route of [
		'/devices',
		'/devices/WIFI_TEST',
		'/functions',
		'/functions/1',
		'/parameter-definitions',
		'/criteria-definitions',
		'/examples',
		'/sequences',
		'/sequences/1',
		'/orphans',
		'/import'
	]) {
		test(`Appearance audit: ${route} uses a centered ${viewport.width}px workspace without overflow`, async ({
			page
		}) => {
			await page.setViewportSize(viewport);
			await page.emulateMedia({ colorScheme: 'light' });
			const browserErrors: string[] = [];
			page.on('pageerror', (error) => browserErrors.push(error.message));
			for (const theme of ['light', 'dark'] as const) {
				if (route.startsWith('/functions')) await mockFunctionCatalog(page);
				else if (route === '/parameter-definitions' || route === '/criteria-definitions')
					await mockDictionary(page, route.slice(1) as DefinitionKind);
				else await mockDeviceHub(page);
				await page.route('**/api/sequences?*', (request) =>
					request.fulfill({
						json: {
							items: [
								{ id: 1, name: 'TRANSMIT_CHECK', description: '', created_at: '', updated_at: '' }
							],
							total: 1,
							page: 1,
							page_size: 20
						}
					})
				);
				await page.route('**/api/sequences/1', (request) =>
					request.fulfill({
						json: {
							id: 1,
							name: 'TRANSMIT_CHECK',
							description: '',
							created_at: '',
							updated_at: '',
							steps: []
						}
					})
				);
				await page.route('**/api/sequences/1/emit', (request) =>
					request.fulfill({ json: { text: '', steps: [] } })
				);
				await page.route('**/api/orphans', (request) =>
					request.fulfill({ json: { total: 0, groups: [] } })
				);
				await page.route('**/api/import/runs', (request) =>
					request.fulfill({ json: { items: [] } })
				);
				await page.goto(route);
				const toggle = page.getByRole('button', { name: 'Toggle dark mode', exact: true });
				await expect(toggle).toHaveAttribute('aria-pressed', theme === 'dark' ? 'true' : 'false');
				await expect(page.getByRole('heading').first()).toBeVisible();
				await expect(page.locator('.banner-error')).toHaveCount(0);
				const activeRoute = page.locator(
					'header nav[aria-label="Main navigation"] a[aria-current="page"]'
				);
				await expect(activeRoute).toHaveCount(1);
				await expect(activeRoute).toHaveAttribute('href', `/${route.split('/')[1]}`);
				await page.keyboard.press('Tab');
				const skipLink = page.getByRole('link', { name: 'Skip to main content', exact: true });
				await expect(skipLink).toBeFocused();
				await expect(skipLink).toBeInViewport();
				await page.keyboard.press('Enter');
				await expect(page.getByRole('main')).toBeFocused();
				const geometry = await page.evaluate(() => {
					const root = document.documentElement;
					const main = document.querySelector('main')!.getBoundingClientRect();
					const header = document.querySelector('header')!.getBoundingClientRect();
					const control = document.querySelector('.theme-toggle')!.getBoundingClientRect();
					return {
						width: root.clientWidth,
						scrollWidth: root.scrollWidth,
						mainWidth: main.width,
						mainLeft: main.x,
						headerRight: header.right,
						controlRight: control.right,
						controlWidth: control.width,
						controlHeight: control.height
					};
				});
				expect(geometry.mainWidth).toBe(Math.min(geometry.width, 1504));
				expect(geometry.mainLeft).toBe((geometry.width - geometry.mainWidth) / 2);
				expect(geometry.scrollWidth, `${route} ${theme} horizontal overflow`).toBeLessThanOrEqual(
					geometry.width
				);
				expect(geometry.headerRight).toBeLessThanOrEqual(geometry.width);
				expect(geometry.controlRight).toBeLessThanOrEqual(geometry.width);
				expect(geometry.controlWidth).toBeGreaterThanOrEqual(44);
				expect(geometry.controlHeight).toBeGreaterThanOrEqual(44);
				await page.screenshot({
					path: test
						.info()
						.outputPath(`${route.replaceAll('/', '_')}-${theme}-${viewport.width}.png`),
					fullPage: true
				});
				if (theme === 'light') {
					await toggle.focus();
					await page.keyboard.press('Enter');
					await expect(toggle).toHaveAttribute('aria-pressed', 'true');
				}
			}
			expect(browserErrors).toEqual([]);
		});
	}
}

for (const route of ['functions', 'parameter-definitions', 'criteria-definitions'] as const) {
	for (const width of [1440, 390]) {
		test(`Filter sizing: ${route} controls align and remain searchable at ${width}px`, async ({
			page
		}) => {
			await page.setViewportSize({ width, height: 900 });
			await page.emulateMedia({ colorScheme: 'light' });
			if (route === 'functions') await mockFunctionCatalog(page);
			else await mockDictionary(page, route);
			await page.goto(`/${route}`);
			await expect(page.locator('.banner-error')).toHaveCount(0);
			const filter = page.locator('.filters').getByRole('button', {
				name: /^Filter (?:definitions )?by Device:/
			});
			const sort = page.getByRole('button', {
				name: route === 'functions' ? 'Sort functions' : 'Sort definitions by',
				exact: true
			});
			for (const theme of ['light', 'dark']) {
				await expect(filter).toBeVisible();
				const fieldBox = (await filter.boundingBox())!;
				const sortBox = (await sort.boundingBox())!;
				expect(fieldBox.height).toBe(route === 'functions' ? 32 : 40);
				expect(fieldBox.height).toBe(sortBox.height);
				expect(fieldBox.width).toBe(sortBox.width);
				expect(fieldBox.x).toBeGreaterThanOrEqual(16);
				expect(fieldBox.x + fieldBox.width).toBeLessThanOrEqual(width - 16);
				if (width === 1440) expect(fieldBox.y).toBe(sortBox.y);
				await page.screenshot({
					path: test.info().outputPath(`filters-${route}-${width}-${theme}.png`),
					fullPage: true
				});
				if (theme === 'light') {
					await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
					await expect(page.locator('html')).toHaveClass(/dark/);
				}
			}
			await filter.click();
			const popup = page.getByRole('dialog', {
				name: /^Choose Filter (?:definitions )?by Device$/
			});
			await popup.getByRole('combobox', { name: 'Search devices' }).fill('wifi');
			await expect(popup.getByRole('option')).toHaveCount(1);
			await popup.getByRole('option', { name: 'WIFI_TEST', exact: true }).click();
			await expect(page).toHaveURL(/device=WIFI_TEST/);
			await expect(filter).toContainText('WIFI_TEST');
			await expect(filter).toBeFocused();
			const longDevice = 'DEVICE_INTERFACE_WITH_A_VERY_LONG_CATALOG_NAME';
			await page.goto(`/${route}?device=${longDevice}`);
			await expect(filter).toContainText(longDevice);
			await expect(filter).toHaveAttribute('title', longDevice);
			expect((await filter.boundingBox())!.height).toBe(route === 'functions' ? 32 : 40);
			expect(
				await filter.locator('span').evaluate((span) => getComputedStyle(span).textOverflow)
			).toBe('ellipsis');
			await page.screenshot({
				path: test.info().outputPath(`filters-${route}-${width}-long.png`),
				fullPage: true
			});
		});
	}
}

for (const width of [1440, 390]) {
	test(`Scrollbars: themed native areas reveal only while scrolling at ${width}px`, async ({
		page
	}) => {
		await page.setViewportSize({ width, height: width === 390 ? 844 : 900 });
		await page.emulateMedia({ colorScheme: 'light' });
		const { functions } = await mockFunctionCatalog(page);
		functions[0].description = 'A detailed function description for this catalog. '.repeat(140);
		functions.push(
			...Array.from({ length: 28 }, (_value, index) => ({
				...functions[0],
				id: index + 10,
				name: `FUNCTION_${String(index).padStart(2, '0')}`
			}))
		);
		await page.route(/\/api\/devices(?:\?.*)?$/, (request) =>
			request.fulfill({
				json: {
					items: Array.from({ length: 30 }, (_value, index) => ({
						device_name: `DEVICE_${String(index).padStart(2, '0')}`
					})),
					total: 30,
					page: 1,
					page_size: 100
				}
			})
		);
		await page.goto('/functions');
		await expect(page.locator('.banner-error')).toHaveCount(0);
		const rail = page.locator('.function-list');
		const root = page.locator('html');
		const idleColor = 'rgba(0, 0, 0, 0) rgba(0, 0, 0, 0)';
		const scrollWithActivity = async (area: Locator, scroll: () => Promise<void>) => {
			await area.evaluate((element) => {
				element.removeAttribute('data-audit-scroll-active');
				const target = element === document.documentElement ? document : element;
				target.addEventListener(
					'scroll',
					() => {
						const style = getComputedStyle(element);
						const color = document.createElement('span').style;
						color.color = style.getPropertyValue('--muted-foreground');
						element.setAttribute(
							'data-audit-scroll-active',
							String(element.getAttribute('data-scroll-active') === 'true')
						);
						element.setAttribute('data-audit-scroll-color', style.scrollbarColor);
						element.setAttribute('data-audit-scroll-expected', `${color.color} rgba(0, 0, 0, 0)`);
						element.setAttribute(
							'data-audit-scroll-root-active',
							String(document.documentElement.getAttribute('data-scroll-active') === 'true')
						);
					},
					{ once: true }
				);
			});
			await scroll();
			await expect(area).toHaveAttribute('data-audit-scroll-active', 'true');
			return area.evaluate((element) => ({
				actual: element.getAttribute('data-audit-scroll-color'),
				expected: element.getAttribute('data-audit-scroll-expected'),
				rootActive: element.getAttribute('data-audit-scroll-root-active') === 'true'
			}));
		};
		const activeColors: string[] = [];
		for (const theme of ['light', 'dark']) {
			if (theme === 'dark') {
				await page.getByRole('button', { name: 'Toggle dark mode', exact: true }).click();
				await expect(root).toHaveClass(/dark/);
			}
			await expect
				.poll(() => rail.evaluate((area) => area.scrollHeight > area.clientHeight))
				.toBe(true);
			await rail.evaluate((area) => (area.scrollTop = 0));
			await rail.hover();
			await expect(rail).not.toHaveAttribute('data-scroll-active', 'true');
			await expect(root).not.toHaveAttribute('data-scroll-active', 'true');
			expect(await rail.evaluate((area) => getComputedStyle(area).scrollbarWidth)).toBe('thin');
			expect(await rail.evaluate((area) => getComputedStyle(area).scrollbarColor)).toBe(idleColor);
			const before = await rail.evaluate((area) => ({
				width: area.clientWidth,
				height: area.clientHeight
			}));
			const colors = await scrollWithActivity(rail, () => page.mouse.wheel(0, 160));
			await expect.poll(() => rail.evaluate((area) => area.scrollTop)).toBeGreaterThan(0);
			expect(colors.rootActive).toBe(false);
			expect(colors.actual).toBe(colors.expected);
			activeColors.push(colors.actual!);
			await page.screenshot({
				path: test.info().outputPath(`scrollbar-${width}-${theme}-active.png`)
			});
			await scrollWithActivity(rail, () => page.mouse.wheel(0, 120));
			await expect(rail).not.toHaveAttribute('data-scroll-active', 'true');
			expect(await rail.evaluate((area) => getComputedStyle(area).scrollbarColor)).toBe(idleColor);
			expect(
				await rail.evaluate((area) => ({ width: area.clientWidth, height: area.clientHeight }))
			).toEqual(before);
			await page.screenshot({
				path: test.info().outputPath(`scrollbar-${width}-${theme}-idle.png`)
			});
		}
		expect(activeColors[0]).not.toBe(activeColors[1]);
		await rail.getByRole('button').first().focus();
		await scrollWithActivity(rail, () => page.keyboard.press('PageDown'));
		await expect.poll(() => rail.evaluate((area) => area.scrollTop)).toBeGreaterThan(0);
		await expect(rail).not.toHaveAttribute('data-scroll-active', 'true');
		await page.locator('main').focus();
		await scrollWithActivity(root, () => page.keyboard.press('PageDown'));
		await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(0);
		await expect(rail).not.toHaveAttribute('data-scroll-active', 'true');
		await expect(root).not.toHaveAttribute('data-scroll-active', 'true');
		await page.getByRole('button', { name: /^Filter by Device:/ }).click();
		const popup = page.getByRole('dialog', { name: 'Choose Filter by Device', exact: true });
		const viewport = popup.locator('[data-combobox-viewport]');
		await expect
			.poll(() => viewport.evaluate((area) => area.scrollHeight > area.clientHeight))
			.toBe(true);
		await scrollWithActivity(viewport, async () => {
			await viewport.evaluate((area) => (area.scrollTop = 120));
		});
		expect(await viewport.evaluate((area) => getComputedStyle(area).scrollbarWidth)).toBe('thin');
		await expect(viewport).not.toHaveAttribute('data-scroll-active', 'true');
		expect(await viewport.evaluate((area) => getComputedStyle(area).scrollbarColor)).toBe(
			idleColor
		);
		await page.keyboard.press('Escape');
		await page.emulateMedia({ forcedColors: 'active' });
		expect(await rail.evaluate((area) => getComputedStyle(area).scrollbarColor)).toBe('auto');
	});
}
