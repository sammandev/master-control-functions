import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// ARIA state assertions (ticket 08, spec stories 30/31): sort-state
// exposure on sortable headers and live-region announcements are part of
// the polish contract, so they are asserted through the browser e2e seam
// like every other external behavior — role- and attribute-based, never
// class-coupled.

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

test('sortable headers expose their sort state through aria-sort', async ({ page }) => {
	await page.goto('/functions');
	await expect(page.getByRole('heading', { name: 'Functions' })).toBeVisible();

	const nameHead = page.getByRole('columnheader', { name: /Name/ });
	const interfaceHead = page.getByRole('columnheader', { name: /Interface/ });

	// The list loads sorted by name ascending: the default sort state is
	// already exposed before any interaction.
	await expect(nameHead).toHaveAttribute('aria-sort', 'ascending');
	await expect(interfaceHead).not.toHaveAttribute('aria-sort');

	// Clicking the sorted header flips the direction.
	await nameHead.getByRole('button').click();
	await expect(nameHead).toHaveAttribute('aria-sort', 'descending');

	// Sorting another column moves the state with it.
	await interfaceHead.getByRole('button').click();
	await expect(interfaceHead).toHaveAttribute('aria-sort', 'ascending');
	await expect(nameHead).not.toHaveAttribute('aria-sort');
});

test('the newly sortable list tables expose aria-sort too', async ({ page }) => {
	// Ticket 09 extended sorting to every list table; the state exposure
	// travels with it through the shared header mechanism.
	await page.goto('/devices');
	const deviceSort = page.getByRole('combobox', { name: 'Sort devices' });
	await deviceSort.selectOption('device');
	await expect(deviceSort).toHaveValue('device');
	await expect(page.getByRole('combobox', { name: 'Sort order', exact: true })).toHaveValue('asc');

	await page.goto('/examples');
	const scriptHead = page.getByRole('columnheader', { name: 'Script' });
	await scriptHead.getByRole('button').click();
	await expect(scriptHead).toHaveAttribute('aria-sort', 'ascending');

	await page.goto('/sequences');
	const nameHead = page.getByRole('columnheader', { name: 'Name' });
	await nameHead.getByRole('button').click();
	await expect(nameHead).toHaveAttribute('aria-sort', 'ascending');
});

test('error banners are announced through role="alert"', async ({ page }) => {
	await page.goto('/import');
	await expect(page.getByRole('heading', { name: 'Workbook import' })).toBeVisible();

	// Import without choosing a file fails inline; the banner is a live
	// region so screen readers announce it the moment it appears.
	const importButton = page.getByRole('button', { name: 'Import', exact: true });
	await importButton.click();
	const alert = page.getByRole('alert');
	await expect(alert).toBeVisible();
	await expect(alert).toHaveText('Choose a workbook file first.');
});

test('import completion is announced politely', async ({ page }) => {
	await page.goto('/import');
	await expect(page.getByRole('heading', { name: 'Workbook import' })).toBeVisible();

	// A UI-driven import finishes with a polite toast; blocking errors stay
	// in the persistent alert banner. The fixture is large enough that the
	// import itself takes a few seconds — wait generously for completion.
	await page.locator('input[type=file]').setInputFiles(FIXTURE);
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	await expect(page.getByRole('heading', { name: /Last import/ })).toBeVisible({ timeout: 15000 });
	const toast = page
		.locator('section[aria-live=polite] [data-sonner-toast]')
		.filter({ hasText: 'Import complete' });
	await expect(toast).toBeVisible({ timeout: 5000 });
});

test('empty states are polite live regions', async ({ page }) => {
	await page.goto('/examples');
	await expect(page.getByRole('heading', { name: 'Examples' })).toBeVisible();

	// A no-match search empties the table; the empty-state cell announces
	// itself as a polite live region.
	const search = page.getByLabel('Search examples');
	await search.fill('zzz-no-match');
	const liveCell = page.locator('td[aria-live="polite"]');
	await expect(liveCell).toBeVisible();
	await expect(liveCell).toHaveText('No examples yet.');
});
