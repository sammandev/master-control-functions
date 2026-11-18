import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// Dialog + AlertDialog interaction tests (ticket 02). The examples page is
// the migrated tracer surface: Escape close, focus restore, focus trapping
// and backdrop close are asserted through real keyboard/pointer events, and
// the destructive confirms run against rows seeded through the API.
// Assertions are role- and label-based only — portal-rendered dialogs are
// found by role, never by DOM location or class.

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

async function seed(request: APIRequestContext) {
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
}

async function seedExample(request: APIRequestContext, meaning: string): Promise<void> {
	const created = await request.post('/api/examples', {
		data: {
			function_source: 'DLL',
			function_name: 'TX_MULTI_VERIFICATION',
			device_name: 'WIFI_TEST',
			script: `DLLWIFI_TEST|TX_MULTI_VERIFICATION|ANT=15|10000|NEXT|FAIL`,
			meaning
		}
	});
	expect(created.status(), 'example seeded').toBe(201);
}

test.beforeAll(async ({ request }) => {
	await seed(request);
});

function focusIsInside(page: import('@playwright/test').Page): Promise<boolean> {
	return page.evaluate(() => {
		const dialog = document.querySelector('[role="dialog"], [role="alertdialog"]');
		return dialog !== null && dialog.contains(document.activeElement);
	});
}

test('examples dialog: Escape closes and focus returns to the trigger', async ({ page }) => {
	await page.goto('/examples');
	const trigger = page.getByRole('button', { name: 'New example' });
	await trigger.click();
	const dialog = page.getByRole('dialog', { name: 'New example' });
	await expect(dialog).toBeVisible();

	// Focus moves into the dialog when it opens (bits-ui autofocus).
	expect(await focusIsInside(page)).toBe(true);

	await page.keyboard.press('Escape');
	await expect(dialog).not.toBeVisible();

	// Focus is restored to the trigger so keyboard flow is never broken.
	await expect(trigger).toBeFocused();
});

test('examples dialog: focus stays trapped while open', async ({ page }) => {
	await page.goto('/examples');
	await page.getByRole('button', { name: 'New example' }).click();
	const dialog = page.getByRole('dialog', { name: 'New example' });
	await expect(dialog).toBeVisible();

	// Tab past the number of focusable elements in the dialog: a trapped
	// focus keeps cycling inside; an untrapped one escapes to the page.
	for (let i = 0; i < 12; i++) {
		await page.keyboard.press('Tab');
	}
	expect(await focusIsInside(page)).toBe(true);
});

test('examples dialog: backdrop click closes', async ({ page }) => {
	await page.goto('/examples');
	await page.getByRole('button', { name: 'New example' }).click();
	const dialog = page.getByRole('dialog', { name: 'New example' });
	await expect(dialog).toBeVisible();

	// The overlay covers the viewport; the far-left edge is backdrop.
	await page.mouse.click(4, 300);
	await expect(dialog).not.toBeVisible();
});

test('examples delete: confirm dialog deletes the seeded row', async ({ page, request }) => {
	const meaning = `e2e delete me ${Date.now()}`;
	await seedExample(request, meaning);

	await page.goto('/examples');
	const search = page.getByLabel('Search examples');
	await search.fill(meaning);
	await expect(page.getByRole('row', { name: /TX_MULTI_VERIFICATION/ })).toBeVisible();

	await page.getByRole('button', { name: 'Delete' }).click();
	const confirm = page.getByRole('alertdialog');
	await expect(confirm).toBeVisible();
	await expect(confirm.getByText(/This cannot be undone/)).toBeVisible();

	await confirm.getByRole('button', { name: 'Delete' }).click();
	// The confirm itself must close on confirm — a stuck alertdialog would
	// block the page and invite a double-delete.
	await expect(confirm).not.toBeVisible();
	await expect(page.getByText('No examples yet.')).toBeVisible();
});

test('examples delete: cancel keeps the row', async ({ page, request }) => {
	const meaning = `e2e keep me ${Date.now()}`;
	await seedExample(request, meaning);

	await page.goto('/examples');
	const search = page.getByLabel('Search examples');
	await search.fill(meaning);
	const row = page.getByRole('row', { name: /TX_MULTI_VERIFICATION/ });
	await expect(row).toBeVisible();

	await page.getByRole('button', { name: 'Delete' }).click();
	const confirm = page.getByRole('alertdialog');
	await expect(confirm).toBeVisible();
	await confirm.getByRole('button', { name: 'Cancel' }).click();
	await expect(confirm).not.toBeVisible();
	await expect(row).toBeVisible();
});

test('functions delete: confirm dialog deletes the function', async ({ page, request }) => {
	const name = `E2E_DELETE_FN_${Date.now()}`;
	const created = await request.post('/api/functions', {
		data: { source: 'DLL', interface: 'GPIB', name }
	});
	expect(created.status(), 'function seeded').toBe(201);

	await page.goto('/functions');
	await page.getByPlaceholder('Search name or description…').fill(name);
	await expect(page.getByRole('link', { name })).toBeVisible();

	// Row actions live behind the per-row dropdown menu (ticket 03).
	await page
		.getByRole('row', { name })
		.getByRole('button', { name: `Actions for ${name}` })
		.click();
	await page.getByRole('menuitem', { name: 'Delete' }).click();
	const confirm = page.getByRole('alertdialog');
	await expect(confirm).toBeVisible();

	await confirm.getByRole('button', { name: 'Delete' }).click();
	// The confirm itself must close on confirm — a stuck alertdialog would
	// block the page and invite a double-delete.
	await expect(confirm).not.toBeVisible();
	// Successful delete rides the polite toast channel.
	await expect(page.locator('[data-sonner-toast]')).toContainText('Function deleted');
	await expect(
		page.getByText('No functions match. Adjust the search or create one.')
	).toBeVisible();
});

test('functions row actions: dropdown menu is keyboard operable', async ({ page, request }) => {
	const name = `E2E_KBD_FN_${Date.now()}`;
	const created = await request.post('/api/functions', {
		data: { source: 'DLL', interface: 'GPIB', name }
	});
	expect(created.status(), 'function seeded').toBe(201);
	const { id } = (await created.json()) as { id: number };

	await page.goto('/functions');
	await page.getByPlaceholder('Search name or description…').fill(name);

	// Keyboard-only path: focus the row trigger, open with Enter, walk the
	// items with the arrow keys, activate with Enter.
	const trigger = page.getByRole('button', { name: `Actions for ${name}` });
	await trigger.focus();
	await page.keyboard.press('Enter');
	await expect(page.getByRole('menu')).toBeVisible();
	await expect(page.getByRole('menuitem', { name: 'Edit' })).toBeFocused();
	await page.keyboard.press('ArrowDown');
	await expect(page.getByRole('menuitem', { name: 'Delete' })).toBeFocused();
	await page.keyboard.press('ArrowUp');
	await expect(page.getByRole('menuitem', { name: 'Edit' })).toBeFocused();
	await page.keyboard.press('Enter');

	const dialog = page.getByRole('dialog', { name: `Edit function #${id}` });
	await expect(dialog).toBeVisible();
});

test('functions save: success toast announces politely', async ({ page }) => {
	await page.goto('/functions');
	await page.getByRole('button', { name: 'New function' }).click();
	const dialog = page.getByRole('dialog', { name: 'New function' });
	await dialog.getByRole('button', { name: /^Device \/ Interface:/ }).click();
	const picker = page.getByRole('dialog', { name: 'Choose Device / Interface', exact: true });
	await picker.getByRole('combobox', { name: 'Search devices' }).fill('GPIB');
	await picker.getByRole('option', { name: 'GPIB', exact: true }).click();
	await dialog.getByLabel('Name').fill(`E2E_TOAST_FN_${Date.now()}`);
	await dialog.getByRole('button', { name: 'Save' }).click();
	await expect(dialog).not.toBeVisible();

	// Success rides the polite toast channel; validation errors keep the
	// persistent role="alert" banners.
	const politeToast = page.locator('section[aria-live="polite"] [data-sonner-toast]');
	await expect(politeToast).toHaveCount(1);
	await expect(politeToast).toContainText('Function created');
});

test('parameter definitions: device is editable in the modal and survives save', async ({
	page,
	request
}) => {
	// Seed a device-scoped definition through the API (fixture data stays
	// untouched); the modal edit below must round-trip its device. Before
	// the Device field existed, saving an edit through the modal silently
	// downgraded a scoped definition to Global (full-replace PUT with no
	// device_name in the body).
	const name = `E2E_DEVICE_${Date.now()}`;
	const created = await request.post('/api/parameter-definitions', {
		data: { name, type: 'string', cardinality: 'SINGLE', device_name: 'WIFI_TEST' }
	});
	expect(created.status(), 'device-scoped definition seeded').toBe(201);

	await page.goto('/parameter-definitions');
	await page.getByRole('textbox', { name: 'Search parameter definitions' }).fill(name);
	await page.getByRole('button', { name: `Select ${name} (WIFI_TEST)`, exact: true }).click();
	const details = page.getByRole('region', { name: 'Definition details', exact: true });
	await details.getByRole('button', { name: 'Edit definition', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: `Edit ${name}` });
	await expect(dialog).toBeVisible();
	// The existing device arrives in the modal instead of a blank field.
	await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText(
		'WIFI_TEST'
	);

	// Move the device — the change must stick, not be dropped.
	await dialog.getByRole('button', { name: /^Device \/ Interface:/ }).click();
	const picker = page.getByRole('dialog', { name: 'Choose Device / Interface', exact: true });
	await picker.getByRole('combobox', { name: 'Search devices' }).fill('OSP');
	await picker.getByRole('option', { name: 'OSP', exact: true }).click();
	await dialog.getByRole('button', { name: 'Save' }).click();
	await expect(dialog).not.toBeVisible();
	// Success rides the polite toast channel.
	await expect(page.locator('[data-sonner-toast]')).toContainText(`${name} updated`);
	await expect(details.getByText('OSP', { exact: true })).toBeVisible();

	// Re-open: the moved device persists.
	await details.getByRole('button', { name: 'Edit definition', exact: true }).click();
	await expect(dialog).toBeVisible();
	await expect(dialog.getByRole('button', { name: /^Device \/ Interface:/ })).toContainText('OSP');
	await page.keyboard.press('Escape');
	await expect(dialog).not.toBeVisible();
});

test('dictionary row actions: dropdown menu is keyboard operable', async ({ page, request }) => {
	const name = `E2E_KBD_DEF_${Date.now()}`;
	const created = await request.post('/api/parameter-definitions', {
		data: { name, type: 'string', cardinality: 'SINGLE' }
	});
	expect(created.status(), 'definition seeded').toBe(201);

	await page.goto('/parameter-definitions');
	await page.getByPlaceholder('Search name or description…').fill(name);

	// Keyboard-only path: focus the row trigger, open with Enter, walk the
	// items with the arrow keys, activate with Enter.
	const trigger = page.getByRole('button', { name: `Actions for ${name}` });
	await trigger.focus();
	await page.keyboard.press('Enter');
	await expect(page.getByRole('menu')).toBeVisible();
	await expect(page.getByRole('menuitem', { name: 'Edit' })).toBeFocused();
	await page.keyboard.press('ArrowDown');
	await expect(page.getByRole('menuitem', { name: 'Delete' })).toBeFocused();
	await page.keyboard.press('ArrowUp');
	await expect(page.getByRole('menuitem', { name: 'Edit' })).toBeFocused();
	await page.keyboard.press('Enter');

	const dialog = page.getByRole('dialog', { name: `Edit ${name}` });
	await expect(dialog).toBeVisible();
});

test('dictionary delete: confirm dialog deletes the definition', async ({ page, request }) => {
	const name = `E2E_DELETE_DEF_${Date.now()}`;
	const created = await request.post('/api/parameter-definitions', {
		data: { name, type: 'string', cardinality: 'SINGLE' }
	});
	expect(created.status(), 'definition seeded').toBe(201);

	await page.goto('/parameter-definitions');
	await page.getByPlaceholder('Search name or description…').fill(name);
	const row = page.getByRole('row', { name: new RegExp(name) });
	await expect(row).toBeVisible();

	// Row actions live behind the per-row dropdown menu (ticket 05).
	await row.getByRole('button', { name: `Actions for ${name}` }).click();
	await page.getByRole('menuitem', { name: 'Delete' }).click();
	const confirm = page.getByRole('alertdialog');
	await expect(confirm).toBeVisible();

	await confirm.getByRole('button', { name: 'Delete' }).click();
	// The confirm itself must close on confirm — a stuck alertdialog would
	// block the page and invite a double-delete.
	await expect(confirm).not.toBeVisible();
	// Successful delete rides the polite toast channel.
	await expect(page.locator('[data-sonner-toast]')).toContainText('Definition deleted');
	await expect(page.getByText('No entries match. Adjust the search or create one.')).toBeVisible();
});
