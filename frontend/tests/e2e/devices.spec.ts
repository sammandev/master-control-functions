import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect } from '@playwright/test';

// Device browser (ADR-0009/ADR-0010): the index lists every distinct
// device_name verbatim with per-device counts, "(Global)" first; a device's
// detail page shows its linked functions and its definitions split from
// the global ones its links resolve to. The fixture's devices include both
// Device names from both scoped rows and Function interfaces; these
// values share one catalog.
//
// Assertions are role- and label-based; counts are derived from the API at
// runtime (the derive-from-API pattern), never pinned to literals.

const E2E_DIR = dirname(fileURLToPath(import.meta.url));
const FIXTURE = join(
	E2E_DIR,
	'..',
	'..',
	'..',
	'backend',
	'tests',
	'testdata',
	'Master_Control_V2_Script_Definition_Knowledge.xlsx'
);

test.beforeAll(async ({ request }) => {
	const workbook = readFileSync(FIXTURE);
	const res = await request.post('/api/import/reset', {
		multipart: {
			file: {
				name: 'Master_Control_V2_Script_Definition_Knowledge.xlsx',
				mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
				buffer: workbook
			}
		}
	});
	expect(res.status(), 'fixture import').toBe(200);
});

interface DeviceSummary {
	device_name: string;
	functions: number;
	parameters: number;
	criteria: number;
	examples: number;
}

test('devices index: tiles match the API, Global is separate', async ({ page, request }) => {
	const api = await (await request.get('/api/devices')).json();
	const wifi = api.items.find((s: DeviceSummary) => s.device_name === 'WIFI_TEST');
	expect(wifi, 'WIFI_TEST present in API response').toBeTruthy();

	await page.goto('/devices');
	await expect(page.getByRole('heading', { name: 'Devices' })).toBeVisible();

	const global = page.getByRole('region', { name: 'Global definitions' });
	await expect(global).toBeVisible();
	await expect(global).toContainText(
		String(api.items.find((item: DeviceSummary) => item.device_name === '').parameters)
	);

	const row = page.getByRole('article', { name: /WIFI_TEST/ });
	await expect(row).toBeVisible();
	await expect(row).toContainText(String(wifi.functions));
	await expect(row).toContainText(String(wifi.examples));
});

test('devices detail: click-through shows functions and definition groups', async ({
	page,
	request
}) => {
	const api = await (await request.get('/api/devices')).json();
	const wifi = api.items.find((s: DeviceSummary) => s.device_name === 'WIFI_TEST');
	expect(wifi, 'WIFI_TEST present in API response').toBeTruthy();

	await page.goto('/devices');
	await page.getByRole('link', { name: 'WIFI_TEST' }).click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST$/);
	await expect(page.getByRole('heading', { name: 'WIFI_TEST' })).toBeVisible();

	// Functions tab is the default: the API total renders.
	await expect(
		page.getByText(`${wifi.functions} functions with links in this device.`)
	).toBeVisible();

	// Parameters tab: scoped rows plus the resolved-from-Global group.
	await page.getByRole('tab', { name: 'Parameters' }).click();
	await expect(
		page.getByText(new RegExp(`${wifi.parameters} parameter definitions? scoped to WIFI_TEST`))
	).toBeVisible();
	if (wifi.parameters > 0) {
		await expect(page.getByRole('heading', { name: 'Scoped' })).toBeVisible();
	}
});

test('devices detail: ?tab= deep-link lands on the requested tab', async ({ page, request }) => {
	const api = await (await request.get('/api/devices')).json();
	const wifi = api.items.find((s: DeviceSummary) => s.device_name === 'WIFI_TEST');
	expect(wifi, 'WIFI_TEST present in API response').toBeTruthy();

	// A shared link with ?tab= opens directly on that tab, data loaded.
	await page.goto('/devices/WIFI_TEST?tab=parameters');
	await expect(page.getByRole('tab', { name: 'Parameters' })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await expect(
		page.getByText(new RegExp(`${wifi.parameters} parameter definitions? scoped to WIFI_TEST`))
	).toBeVisible();

	// An unknown tab value falls back to the default (Functions).
	await page.goto('/devices/WIFI_TEST?tab=bogus');
	await expect(page.getByRole('tab', { name: 'Functions' })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await expect(
		page.getByText(`${wifi.functions} functions with links in this device.`)
	).toBeVisible();
});

test('devices: function search previews a match without replacing the bookmark', async ({ page }) => {
	const evidence = process.env.DEVICE_SEARCH_SCREENSHOT_DIR;
	if (evidence) await page.setViewportSize({ width: 1280, height: 900 });
	await page.goto('/devices/WIFI_TEST');
	const rail = page.getByRole('complementary', { name: 'Device functions' });
	const search = rail.getByRole('textbox', { name: 'Search device functions' });
	const details = page.getByRole('region', { name: 'Function details' });
	const remove = rail.getByRole('button', { name: 'Select DUT_REMOVE', exact: true });
	const initialize = rail.getByRole('button', { name: 'Select DUT_INITIALIZE', exact: true });
	const removeDetail = details.getByRole('heading', { name: 'DUT_REMOVE', exact: true });
	const initializeDetail = details.getByRole('heading', { name: 'DUT_INITIALIZE', exact: true });

	await remove.click();
	await expect(page).toHaveURL(/[?&]function=\d+$/);
	await expect(removeDetail).toBeVisible();
	const bookmark = page.url();

	await search.fill('DUT_REMOVE');
	await expect(remove).toHaveAttribute('aria-pressed', 'true');
	await expect(removeDetail).toBeVisible();

	await search.fill('DUT_INITIALIZE');
	await expect(rail.getByRole('button', { name: /^Select / })).toHaveCount(1);
	await expect(initialize).toBeVisible();
	await expect.soft(initializeDetail).toBeVisible();
	if (evidence) {
		await expect(page.getByRole('img', { name: 'DLL type' })).toBeVisible();
		await expect(details.getByText('Loading function details...')).toBeHidden();
		await page.evaluate(() => window.scrollTo(0, 0));
		await page.screenshot({ path: join(evidence, 'functions.png') });
	}
	await expect.soft(initialize).toHaveAttribute('aria-pressed', 'true');

	await search.fill('no-such-device-function');
	await expect(rail.getByText('No matching functions.')).toBeVisible();
	await expect.soft(details).toBeHidden();
	await expect.soft(
		page.getByRole('button', { name: /^(Edit function|Delete function|Compose script)$/ })
	).toHaveCount(0);
	await expect.soft(page.getByText('This function is not linked to this device.')).toBeHidden();

	await search.clear();
	await expect(page).toHaveURL(bookmark);
	await expect(removeDetail).toBeVisible();
	await page.reload();
	await expect(removeDetail).toBeVisible();
});

test('devices detail: tab switches land in history', async ({ page }) => {
	await page.goto('/devices/WIFI_TEST?tab=parameters');

	await page.getByRole('tab', { name: 'Criteria' }).click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=criteria$/);

	// Back returns to the previously active tab, forward re-advances it.
	await page.goBack();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=parameters$/);
	await expect(page.getByRole('tab', { name: 'Parameters' })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await page.goForward();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=criteria$/);
	await expect(page.getByRole('tab', { name: 'Criteria' })).toHaveAttribute(
		'aria-selected',
		'true'
	);
});

test('devices index: count cells link to the matching tab', async ({ page, request }) => {
	const api = await (await request.get('/api/devices')).json();
	const wifi = api.items.find((s: DeviceSummary) => s.device_name === 'WIFI_TEST');
	expect(wifi, 'WIFI_TEST present in API response').toBeTruthy();

	await page.goto('/devices');
	const row = page.getByRole('article', { name: /WIFI_TEST/ });
	await expect(row).toBeVisible();

	// The parameters count link (aria-labelled "N parameters") opens the
	// device page with the parameters tab preselected.
	await row.getByRole('link', { name: / parameters$/ }).click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST\?tab=parameters$/);
	await expect(page.getByRole('tab', { name: 'Parameters' })).toHaveAttribute(
		'aria-selected',
		'true'
	);
	await expect(
		page.getByText(new RegExp(`${wifi.parameters} parameter definitions? scoped to WIFI_TEST`))
	).toBeVisible();
});

test('devices: create a device through the UI', async ({ page, request }) => {
	const name = `E2E_UI_${Date.now()}`;
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Create device' }).click();

	const dialog = page.getByRole('dialog');
	await dialog.getByLabel('Device name').fill(` ${name.toLowerCase()} `);
	// The dialog previews the uppercase-normalized name (ADR-0002 rule).
	await expect(dialog.getByText(`Will be created as ${name}`)).toBeVisible();

	await dialog.getByRole('button', { name: 'Create', exact: true }).click();
	await expect(dialog).toBeHidden();
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Device created' })
	).toBeVisible();

	// The new device appears in the list with zero counts…
	const row = page.getByRole('article', { name: new RegExp(name) });
	await expect(row).toBeVisible();
	await expect(row).toContainText('0');

	// …and is registered server-side.
	const api = await (await request.get('/api/devices')).json();
	expect(api.items.map((s: DeviceSummary) => s.device_name)).toContain(name);
});

test('devices: duplicate create shows the API error inline', async ({ page }) => {
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Create device' }).click();

	const dialog = page.getByRole('dialog');
	await dialog.getByLabel('Device name').fill(' wifi_test ');
	await dialog.getByRole('button', { name: 'Create', exact: true }).click();

	// The API's 409 message surfaces inline; the dialog stays open and the
	// input stays editable so the user can correct the name.
	await expect(dialog.getByText(/already exists/i)).toBeVisible();
	await expect(dialog.getByLabel('Device name')).toBeEditable();
});

test('devices: delete an empty registered device', async ({ page, request }) => {
	const name = `E2E_DEL_${Date.now()}`;
	const res = await request.post('/api/devices', { data: { device_name: name } });
	expect(res.status(), 'seed a registered empty device').toBe(201);

	await page.goto('/devices');
	const row = page.getByRole('article', { name: new RegExp(name) });
	await expect(row).toBeVisible();

	await row.getByRole('button', { name: 'Delete' }).click();
	const confirm = page.getByRole('dialog', { name: `Delete ${name}?` });
	await expect(confirm.getByRole('button', { name: 'Delete', exact: true })).toBeDisabled();
	await confirm.getByLabel(`Type DELETE ${name} to confirm`).fill(`DELETE ${name}`);
	await confirm.getByRole('button', { name: 'Delete' }).click();

	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Device deleted' })
	).toBeVisible();
	await expect(row).toHaveCount(0);
	const api = await (await request.get('/api/devices')).json();
	expect(api.items.map((s: DeviceSummary) => s.device_name)).not.toContain(name);
});

test('devices: populated devices offer deletion but Cancel preserves content', async ({
	page,
	request
}) => {
	await page.goto('/devices');
	const row = page.getByRole('article', { name: /WIFI_TEST/ });
	await expect(row).toBeVisible();
	await row.getByRole('button', { name: 'Delete' }).click();
	const confirm = page.getByRole('dialog', { name: 'Delete WIFI_TEST?' });
	await expect(confirm.getByRole('button', { name: 'Delete', exact: true })).toBeDisabled();
	await confirm.getByRole('button', { name: 'Cancel' }).click();
	await expect(confirm).toBeHidden();
	await expect(row).toBeVisible();
	const functions = await (await request.get('/api/devices/WIFI_TEST/functions')).json();
	expect(functions.total).toBeGreaterThan(0);
});

test('devices: deleting a device used by a Sequence is refused and keeps it', async ({
	page,
	request
}, testInfo) => {
	const sequenceName = `E2E uses WIFI_TEST ${Date.now()}`;
	const created = await request.post('/api/sequences', {
		data: {
			name: sequenceName,
			steps: [
				{
					function_source: 'DLL',
					function_interface: 'WIFI_TEST',
					function_name: 'TX_MULTI_VERIFICATION'
				}
			]
		}
	});
	expect(created.status(), 'seed a Sequence using WIFI_TEST').toBe(201);
	const { id } = (await created.json()) as { id: number };
	try {
		const functionsBefore = await (await request.get('/api/devices/WIFI_TEST/functions')).json();

		await page.goto('/devices');
		const row = page.getByRole('article', { name: /WIFI_TEST/ });
		await row.getByRole('button', { name: 'Delete' }).click();
		const confirm = page.getByRole('dialog', { name: 'Delete WIFI_TEST?' });
		await confirm.getByLabel('Type DELETE WIFI_TEST to confirm').fill('DELETE WIFI_TEST');
		await confirm.getByRole('button', { name: 'Delete', exact: true }).click();

		await expect(confirm.getByRole('alert')).toHaveText(
			`WIFI_TEST is used by Sequences: ${sequenceName}. Remove those steps first.`
		);
		await page.screenshot({ path: testInfo.outputPath('device-delete-refused.png') });
		await confirm.getByRole('button', { name: 'Cancel' }).click();
		await expect(row).toBeVisible();

		const functionsAfter = await (await request.get('/api/devices/WIFI_TEST/functions')).json();
		expect(functionsAfter.total).toBe(functionsBefore.total);
		const sequence = await (await request.get(`/api/sequences/${id}`)).json();
		expect(sequence.steps).toHaveLength(1);
	} finally {
		await request.delete(`/api/sequences/${id}`);
	}
});

test('devices: a dictionary device badge deep-links to the device page', async ({ page }) => {
	await page.goto('/parameter-definitions');
	await page.getByPlaceholder('Search name or description…').fill('ANT');
	const row = page.getByRole('row', { name: /ANT/ }).first();
	await expect(row).toBeVisible();

	await row.getByRole('link', { name: 'WIFI_TEST' }).click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST$/);
	await expect(page.getByRole('heading', { name: 'WIFI_TEST' })).toBeVisible();
});

test('devices: add a parameter definition from the device page', async ({ page, request }) => {
	const name = `E2E_DEF_${Date.now()}`;
	const param = `E2E_DEF_P_${Date.now()}`;

	// Create the device through the UI, then populate it from its own page.
	await page.goto('/devices');
	await page.getByRole('button', { name: 'Create device' }).click();
	const createDialog = page.getByRole('dialog');
	await createDialog.getByLabel('Device name').fill(name);
	await createDialog.getByRole('button', { name: 'Create' }).click();
	await expect(createDialog).toBeHidden();

	await page.goto(`/devices/${name}?tab=parameters`);
	await page.getByRole('button', { name: 'Add parameter definition' }).click();

	const dialog = page.getByRole('dialog');
	// The Device field is prefilled and visibly locked to the device in
	// context; the definition cannot silently land on another device.
	await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toBeDisabled();
	await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(name);

	await dialog.getByLabel('Name').fill(param);
	await dialog.getByRole('button', { name: 'Save' }).click();
	await expect(dialog).toBeHidden();

	// The new definition appears in the device's scoped list…
	const scopedRow = page.getByRole('row', { name: param });
	await expect(scopedRow).toBeVisible();

	// …and is registered server-side, scoped to the device.
	const defs = await (await request.get(`/api/devices/${name}/parameters?page_size=100`)).json();
	expect(defs.items.map((d: { name: string }) => d.name)).toContain(param);
});

test('functions: a link device value navigates to the device browser', async ({
	page,
	request
}) => {
	// Derive from the API a function that carries links in WIFI_TEST.
	const fns = await (await request.get('/api/devices/WIFI_TEST/functions')).json();
	expect(fns.items.length, 'WIFI_TEST has linked functions in the fixture').toBeGreaterThan(0);
	const fnId = fns.items[0].id;

	await page.goto(`/functions/${fnId}`);

	// The links panel renders the device value as a link into the device
	// browser (the function→device back-path); scoped by title so only the
	// panel's device links match (a function may carry several).
	const link = page.getByTitle('Browse device WIFI_TEST').first();
	await expect(link).toBeVisible();
	await link.click();
	await expect(page).toHaveURL(/\/devices\/WIFI_TEST$/);
	await expect(page.getByRole('heading', { name: 'WIFI_TEST' })).toBeVisible();
});

test('devices: the parameters tab pager reloads on page change', async ({ page, request }) => {
	// Regression: the tab's lazy loader used to gate on its loaded flag only,
	// so the pager rendered but Next/Prev/page-size did nothing.
	const name = `E2E_PAG_${Date.now()}`;
	const res = await request.post('/api/devices', { data: { device_name: name } });
	expect(res.status(), 'seed a registered empty device').toBe(201);

	// 25 scoped parameter definitions → 2 pages at the default page size 20.
	for (let i = 1; i <= 25; i++) {
		const defRes = await request.post('/api/parameter-definitions', {
			data: {
				name: `${name}_P${String(i).padStart(2, '0')}`,
				type: 'string',
				cardinality: 'SINGLE',
				device_name: name
			}
		});
		expect(defRes.status(), `seed definition ${i}`).toBe(201);
	}

	await page.goto(`/devices/${name}?tab=parameters`);
	const scopedRows = page.getByRole('row', { name: new RegExp(`${name}_P`) });
	await expect(scopedRows).toHaveCount(20);

	// Page 2 loads the remaining definitions.
	await page.getByRole('button', { name: 'Next page' }).click();
	await expect(page.getByText('Page 2 of 2')).toBeVisible();
	await expect(scopedRows).toHaveCount(5);
});

test('devices: renaming a resolved Global definition keeps scoped details usable', async ({
	page,
	request
}) => {
	const suffix = Date.now();
	const device = `E2E_GLOBAL_RENAME_${suffix}`;
	const global = `Global parameter ${suffix}`;
	const scoped = 'Scoped parameter';
	const fn = await request.post('/api/functions', {
		data: { source: 'DLL', interface: device, name: 'Global rename test' }
	});
	expect(fn.status()).toBe(201);
	for (const [name, deviceName] of [[global, ''], [scoped, device]]) {
		const response = await request.post('/api/parameter-definitions', {
			data: { name, device_name: deviceName, type: 'string', cardinality: 'SINGLE' }
		});
		expect(response.status()).toBe(201);
	}
	const { id } = await fn.json();
	const link = await request.post(`/api/functions/${id}/parameter-links`, {
		data: { name: global, device_name: device, position: 1 }
	});
	expect(link.status()).toBe(201);

	await page.goto(`/devices/${device}?tab=parameters`);
	const globalRow = page.getByRole('button', { name: `Inspect ${global} (Global)` });
	await globalRow.click();
	await page.getByRole('button', { name: 'Edit Global definition' }).click();
	const dialog = page.getByRole('dialog');
	await dialog.getByLabel('Name', { exact: true }).fill(`${global} renamed`);
	await dialog.getByRole('button', { name: 'Save' }).click();
	await expect(dialog).toBeHidden();

	await expect(globalRow).toBeHidden();
	await expect(page.getByRole('heading', { name: scoped, exact: true })).toBeVisible();
});

test('devices: examples search finds a match on page five', async ({ page, request }) => {
	const device = 'E2E_EXAMPLES_SEARCH';
	const functionName = 'READ_SENSOR';
	const matchMeaning = 'Read the backup sensor';
	const evidence = process.env.DEVICE_SEARCH_SCREENSHOT_DIR;
	const fn = await request.post('/api/functions', {
		data: { source: 'DLL', interface: device, name: functionName }
	});
	expect(fn.status()).toBe(201);
	for (let i = 1; i <= 81; i++) {
		const example = await request.post('/api/examples', {
			data: {
				function_source: 'DLL',
				function_name: functionName,
				device_name: device,
				script: `DLL${device}|${functionName}||${5000 + i}|NEXT|FAIL`,
				meaning: i === 81 ? matchMeaning : `Read sensor sample ${i}`
			}
		});
		expect(example.status(), `seed example ${i}`).toBe(201);
	}

	if (evidence) await page.setViewportSize({ width: 1280, height: 850 });
	await page.goto(`/devices/${device}?tab=examples`);
	const panel = page.getByRole('tabpanel', { name: /^examples/i });
	const search = panel.getByRole('textbox', { name: 'Search device examples' });
	const match = panel.getByRole('cell', { name: matchMeaning, exact: true });
	const next = panel.getByRole('button', { name: 'Next page' });
	const previous = panel.getByRole('button', { name: 'Previous page' });

	for (let i = 2; i <= 5; i++) {
		await next.click();
		await expect(panel.getByText(`Page ${i} of 5`)).toBeVisible();
	}
	await expect(match).toBeVisible();
	for (let i = 4; i >= 2; i--) {
		await previous.click();
		await expect(panel.getByText(`Page ${i} of 5`)).toBeVisible();
	}
	await expect(match).toBeHidden();
	await search.fill('backup sensor');

	await expect.soft(match).toBeVisible();
	await expect.soft(panel.getByText('Page 1 of 1')).toBeVisible();
	if (evidence) {
		await page.evaluate(() => window.scrollTo(0, 0));
		await page.screenshot({ path: join(evidence, 'examples.png') });
	}

	await search.clear();
	await expect(panel.getByText('Page 1 of 5')).toBeVisible();
	await expect(panel.getByRole('cell', { name: 'Read sensor sample 1', exact: true })).toBeVisible();
	await expect(match).toBeHidden();
});

for (const [tab, kind, noun] of [
	['parameters', 'parameter-definitions', 'Parameter'],
	['criteria', 'criteria-definitions', 'Criteria']
] as const) {
	test(`devices: ${tab} search finds off-page definitions without stale detail`, async ({
		page,
		request
	}) => {
		const device = `E2E_${noun.toUpperCase()}_SEARCH_${Date.now()}`;
		const firstName = `${noun} 01`;
		const matchName = `${noun} 21`;
		const evidence = process.env.DEVICE_SEARCH_SCREENSHOT_DIR;
		const created = await request.post('/api/devices', { data: { device_name: device } });
		expect(created.status()).toBe(201);
		for (let i = 1; i <= 21; i++) {
			const response = await request.post(`/api/${kind}`, {
				data: {
					name: `${noun} ${String(i).padStart(2, '0')}`,
					type: 'string',
					cardinality: 'SINGLE',
					default_value: i === 21 ? 'Off-page value' : '',
					device_name: device
				}
			});
			expect(response.status(), `seed definition ${i}`).toBe(201);
		}

		if (evidence) await page.setViewportSize({ width: 1280, height: 900 });
		await page.goto(`/devices/${device}?tab=${tab}`);
		const first = page.getByRole('button', { name: `Inspect ${firstName} (${device})` });
		const match = page.getByRole('button', { name: `Inspect ${matchName} (${device})` });
		const firstDetail = page.getByRole('heading', { name: firstName, exact: true });
		const search = page.getByRole('textbox', { name: `Search ${noun.toLowerCase()} definitions` });

		await page.getByRole('button', { name: 'Next page' }).click();
		await expect(match).toBeVisible();
		await page.getByRole('button', { name: 'Previous page' }).click();
		await first.click();
		await expect(firstDetail).toBeVisible();
		await expect(match).toBeHidden();
		await search.fill('off-page value');

		await expect.soft(match).toBeVisible();
		await expect.soft(firstDetail).toBeHidden();
		if (evidence) {
			if (await match.isVisible()) {
				await expect(page.getByRole('heading', { name: matchName, exact: true })).toBeVisible();
			}
			await page.evaluate(() => window.scrollTo(0, 0));
			await page.screenshot({ path: join(evidence, `${tab}.png`) });
		}

		const previousSelection = page.url();
		await match.click();
		await expect(page).not.toHaveURL(previousSelection);
		await search.clear();
		await expect(page.getByText('Page 2 of 2')).toBeVisible();
		await expect(match).toBeVisible();
		await expect(page.getByRole('heading', { name: matchName, exact: true })).toBeVisible();

		const panel = page.getByRole('tabpanel', { name: new RegExp(`^${tab}`, 'i') });
		await search.fill('off-page value');
		await expect(panel.getByText('Page 1 of 1')).toBeVisible();
		await page.getByRole('button', { name: 'Edit definition' }).click();
		const dialog = page.getByRole('dialog');
		await dialog.getByLabel('Default value').fill('Updated value');
		await dialog.getByRole('button', { name: 'Save' }).click();
		await expect(dialog).toBeHidden();
		await expect(search).toHaveValue('');
		await expect(match).toBeVisible();
		const details = page.getByRole('region', { name: 'Definition details' });
		await expect(details.getByText('Updated value')).toBeVisible();

		await search.fill('updated value');
		await expect(panel.getByText('Page 1 of 1')).toBeVisible();
		await page.getByRole('button', { name: `Add ${noun.toLowerCase()} definition` }).click();
		const addedName = `${noun} 22`;
		await dialog.getByLabel('Name', { exact: true }).fill(addedName);
		await dialog.getByRole('button', { name: 'Save' }).click();
		await expect(dialog).toBeHidden();
		await expect(search).toHaveValue('');
		await expect(page.getByRole('button', { name: `Inspect ${addedName} (${device})` })).toBeVisible();
		await expect(details.getByRole('heading', { name: addedName, exact: true })).toBeVisible();
	});
}
