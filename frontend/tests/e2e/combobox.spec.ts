import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// Combobox link-picker tests (ticket 04): the attach dialog's searchable
// definition picker must be fully keyboard operable — type to search the
// server, ArrowDown to highlight, Enter to commit the link, Escape to
// dismiss the listbox without committing (and without closing the dialog),
// and an inline-creation escape hatch from empty results. The listbox
// portals next to the dialog, so options are looked up at page level;
// assertions stay role- and label-based.

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

test.beforeAll(async ({ request }) => {
	await seed(request);
});

/** Seed a link-less function and a parameter definition, return their ids. */
async function seedFunctionAndParam(request: APIRequestContext, stamp: number) {
	const fnRes = await request.post('/api/functions', {
		data: { source: 'DLL', interface: 'GPIB', name: `E2E_CB_FN_${stamp}` }
	});
	expect(fnRes.status(), 'function seeded').toBe(201);
	const fn = (await fnRes.json()) as { id: number };

	const defRes = await request.post('/api/parameter-definitions', {
		data: { name: `E2E_CB_PARAM_${stamp}`, type: 'int', cardinality: 'SINGLE' }
	});
	expect(defRes.status(), 'parameter definition seeded').toBe(201);

	return { fnId: fn.id };
}

test('combobox: search, arrow down and Enter commit the link', async ({ page, request }) => {
	const stamp = Date.now();
	const { fnId } = await seedFunctionAndParam(request, stamp);

	await page.goto(`/functions/${fnId}`);
	await page.getByRole('tab', { name: 'Parameters' }).click();
	await page.getByRole('button', { name: 'Attach parameter' }).click();

	const dialog = page.getByRole('dialog', { name: 'Attach Parameter' });
	await expect(dialog).toBeVisible();

	// Click into the picker: focus opens the listbox with the first page
	// of definitions, then typing narrows it server-side.
	const input = dialog.getByRole('combobox', { name: 'Definition' });
	await input.click();
	await expect(page.getByRole('listbox')).toBeVisible();
	// Expanded/active-option ARIA comes free from bits-ui; assert it.
	await expect(input).toHaveAttribute('aria-expanded', 'true');
	await input.fill(`E2E_CB_PARAM_${stamp}`);

	// Debounced server search returns the seeded definition; select it
	// with the keyboard. Selection fills the name and closes the listbox
	// (the device-scope fields stay editable) — committing the link is
	// the form's Attach action.
	await page.getByRole('option', { name: /E2E_CB_PARAM_/ }).waitFor();
	await page.keyboard.press('ArrowDown');
	await expect(input).toHaveAttribute('aria-activedescendant', /.+/);
	await page.keyboard.press('Enter');
	await expect(page.getByRole('listbox')).not.toBeVisible();
	await expect(dialog).toBeVisible();
	await expect(input).toHaveValue(`E2E_CB_PARAM_${stamp}`);

	// The picker input sits inside the form: a second Enter submits it.
	await page.keyboard.press('Enter');
	await expect(dialog).not.toBeVisible();

	// The link row appears, fully resolved.
	await expect(page.getByRole('row', { name: /E2E_CB_PARAM_/ })).toBeVisible();
	await expect(page.getByText('1 linked definitions')).toBeVisible();
});

test('combobox: Escape dismisses the listbox but keeps the dialog', async ({ page, request }) => {
	const stamp = Date.now();
	const { fnId } = await seedFunctionAndParam(request, stamp);

	await page.goto(`/functions/${fnId}`);
	await page.getByRole('tab', { name: 'Parameters' }).click();
	await page.getByRole('button', { name: 'Attach parameter' }).click();

	const dialog = page.getByRole('dialog', { name: 'Attach Parameter' });
	await expect(dialog).toBeVisible();

	const input = dialog.getByRole('combobox', { name: 'Definition' });
	await input.click();
	await input.fill('TX');
	await expect(page.getByRole('option').first()).toBeVisible();

	// First Escape: the layer stack hands it to the open listbox, which
	// closes without committing anything and without closing the dialog.
	await page.keyboard.press('Escape');
	await expect(page.getByRole('listbox')).not.toBeVisible();
	await expect(dialog).toBeVisible();

	// Second Escape: the dialog itself closes.
	await page.keyboard.press('Escape');
	await expect(dialog).not.toBeVisible();
});

test('combobox: empty results reach inline creation without leaving the dialog', async ({
	page,
	request
}) => {
	const stamp = Date.now();
	const { fnId } = await seedFunctionAndParam(request, stamp);
	const missing = `E2E_CB_MISSING_${stamp}`;

	await page.goto(`/functions/${fnId}`);
	await page.getByRole('tab', { name: 'Parameters' }).click();
	await page.getByRole('button', { name: 'Attach parameter' }).click();

	const dialog = page.getByRole('dialog', { name: 'Attach Parameter' });
	await expect(dialog).toBeVisible();

	const input = dialog.getByRole('combobox', { name: 'Definition' });
	await input.click();
	await input.fill(missing);
	await expect(page.getByText('No matching definitions.')).toBeVisible();

	// The empty state offers the inline flow in place — no dialog exit.
	await page.getByRole('button', { name: 'Create new definition' }).click();
	await expect(dialog.getByLabel('Name')).toBeVisible();

	await dialog.getByLabel('Name').fill(missing);
	await dialog.getByRole('button', { name: 'Attach' }).click();

	// Inline attach creates the definition and links it atomically.
	await expect(dialog).not.toBeVisible();
	await expect(page.getByRole('row', { name: new RegExp(missing) })).toBeVisible();
	// The attach success announces through the global toast.
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'parameter link attached' })
	).toBeVisible();
});
