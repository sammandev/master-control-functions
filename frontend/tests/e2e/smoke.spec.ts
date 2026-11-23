import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { test, expect, type APIRequestContext } from '@playwright/test';

// E2E smoke suite for the four happy paths (ticket 09), driven against
// the seeded database: the real knowledge workbook is imported through
// /api/import/reset before any test runs, so the suite exercises the
// same data shape as production. Expected counts are derived from the
// API at runtime — a workbook batch requires zero test edits.

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

let fnPath = '';

test.beforeAll(async ({ request }) => {
	await seed(request);
	// Resolve the id of the fixture function the suite drives — never
	// assume id 1 (import order decides).
	const res = await request.get('/api/functions?search=TX_MULTI_VERIFICATION&page_size=1');
	expect(res.status()).toBe(200);
	const body = (await res.json()) as { items: Array<{ id: number; name: string }> };
	expect(body.items.length, 'TX_MULTI_VERIFICATION present').toBe(1);
	fnPath = `/functions/${body.items[0].id}`;
});

test('search a function and open its detail page', async ({ page, request }) => {
	// Headline total derived from the API at runtime — the UI must agree
	// with the API whatever the current workbook batch contains.
	const api = (await (await request.get('/api/functions?page_size=1')).json()) as {
		total: number;
	};
	await page.goto('/functions');
	await expect(page.getByRole('heading', { name: 'Functions' })).toBeVisible();

	// The full dataset is loaded: the UI total equals the API total.
	await expect(page.getByText(`${api.total} functions`)).toBeVisible();

	// Search narrows to the one function, then the link opens the detail
	// page with its parameter links tab.
	const search = page.getByPlaceholder('Search name or description…');
	await search.fill('TX_MULTI_VERIFICATION');
	await expect(page.getByRole('link', { name: 'TX_MULTI_VERIFICATION' })).toBeVisible();
	await page.getByRole('link', { name: 'TX_MULTI_VERIFICATION' }).click();
	await expect(page).toHaveURL(/\/functions\/\d+$/);
	await expect(page.getByText('TX_MULTI_VERIFICATION').first()).toBeVisible();
});

test('edit a parameter link on the function detail page', async ({ page, request }) => {
	const registered = await request.post('/api/devices', { data: { device_name: 'WIFI_TEST_E2E' } });
	expect([201, 409]).toContain(registered.status());
	await page.goto(fnPath);
	await page.getByRole('tab', { name: 'Parameters' }).click();

	// Edit the ANT link inline. The sentinel device scope is a value the
	// fixture row provably does NOT have (it carries WIFI_TEST), so the
	// post-save assertion can only pass if the edit actually landed.
	// Row actions live behind the row's dropdown menu (ticket 04).
	const row = page.getByRole('row', { name: /ANT/ }).first();
	await row.getByRole('button', { name: /Actions for/ }).click();
	await page.getByRole('menuitem', { name: 'Edit' }).click();
	// bits-ui Tabs keeps inactive panels mounted-but-hidden, so scope to the
	// visible Parameters tabpanel (ExamplesPanel carries a "Device
	// filter" input that would otherwise collide).
	const paramsPanel = page.getByRole('tabpanel', { name: 'Parameters' });
	await paramsPanel.getByLabel('Position').fill('1');
	await paramsPanel.getByRole('button', { name: /^Device:/ }).click();
	const picker = page.getByRole('dialog', { name: 'Choose Device', exact: true });
	await picker.getByRole('combobox', { name: 'Search devices' }).fill('WIFI_TEST_E2E');
	await picker.getByRole('option', { name: 'WIFI_TEST_E2E', exact: true }).click();
	await row.getByRole('checkbox').check();
	await row.getByRole('button', { name: 'Save' }).click();

	// Back to display mode with the edited values, no error banner.
	await expect(page.locator('.banner-error')).toHaveCount(0);
	await expect(page.getByRole('row', { name: /WIFI_TEST_E2E/ })).toBeVisible();
});

test('compose a script and save it as an example', async ({ page }) => {
	await page.goto(fnPath);
	await page.getByRole('button', { name: 'Compose script' }).click();

	const dialog = page.getByRole('dialog', { name: 'Script Composer' });
	await expect(dialog).toBeVisible();

	// The form is generated from the function's parameter links; fill
	// every required field the server enforces. Anchored regexes: labels
	// read "NAME · type · cardinality · required" and names share
	// prefixes (POWER vs TX_TARGET_POWER).
	await dialog.getByLabel(/^ANT ·/).fill('15');
	await dialog.getByLabel(/^STANDARD ·/).fill('11AX');
	await dialog.getByLabel(/^FREQ ·/).fill('6165');
	await dialog.getByLabel(/^BW ·/).fill('20');
	await dialog.getByLabel(/^DATA_RATE ·/).fill('7');
	await dialog.getByLabel(/^TX_TARGET_POWER ·/).fill('15');
	await dialog.getByLabel(/^POWER ·/).fill('5dB');

	// Live preview emits the workbook grammar (server-owned): prefix +
	// function + assignments + criteria + timeout + trailing segments.
	// Parameter order follows the links' position order, so anchor on the
	// stable head and tail.
	await expect(
		dialog.getByText(/^DLLWIFI_TEST\|TX_MULTI_VERIFICATION\|.+\|POWER=5dB\|10000\|NEXT\|FAIL$/)
	).toBeVisible({ timeout: 10_000 });

	// Save as example (device scope prefilled from the links, script from
	// the preview — both editable).
	await dialog.getByText('Save as example').click();
	await dialog.getByLabel('Meaning').fill('composed in the smoke test');
	await dialog.getByRole('button', { name: 'Save example' }).click();
	await expect(dialog.getByText('Saved as an example')).toBeVisible();
	// The compose-and-save flow also announces through the global toast.
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Saved as an example' })
	).toBeVisible();
});

test('view the orphan report', async ({ page }) => {
	await page.goto('/orphans');

	// The current fixture workbook has zero orphans of every kind: the
	// report renders the fully-linked empty state (no kind toolbar,
	// sections only appear for groups with orphans).
	await expect(page.getByRole('heading', { name: 'Orphan report' })).toBeVisible();
	await expect(
		page.getByText('No unresolved references. The knowledge base is fully linked.')
	).toBeVisible();
	await expect(page.getByRole('button', { name: /All \(\d+\)/ })).toHaveCount(0);
});

test('enum filter bar drives the functions list', async ({ page, request }) => {
	// Expected outcome read from the API: pick the first source facet and
	// its server-side filtered total, then assert the UI select reproduces
	// it — this is the UI→API→render wiring the HTTP-seam tests cannot see.
	const facets = (await (await request.get('/api/functions/facets')).json()) as {
		sources: string[];
	};
	const source = facets.sources[0];
	expect(source, 'at least one source facet exists').toBeTruthy();
	const list = (await (
		await request.get(`/api/functions?source=${encodeURIComponent(source)}`)
	).json()) as { total: number };

	await page.goto('/functions');
	// bits-ui Select (ticket 03): trigger is a button with a portaled
	// listbox of role=option items, driven by pointer clicks.
	await page.getByRole('button', { name: 'Filter by source' }).click();
	await page.getByRole('option', { name: source, exact: true }).click();
	await expect(page.getByText(`${list.total} functions`)).toBeVisible();
});

test('sequence: emit, reorder and add a composer step', async ({ page, request }) => {
	// Ticket 10 happy path, seeded through the API for determinism: a
	// two-step sequence (the fixture function + a synthetic one), then the
	// UI must emit, reorder, and accept a composer-added step. Unique names
	// per run: reset does NOT clear sequences (ADR-0005).
	const stamp = Date.now();
	const helper = `E2E_SEQ_HELPER_${stamp}`;
	const seqName = `E2E bringup ${stamp}`;

	const res = await request.post('/api/functions', {
		data: { source: 'DLL', interface: 'GPIB', name: helper }
	});
	expect(res.status()).toBe(201);
	const tx = (
		await (await request.get('/api/functions?search=TX_MULTI_VERIFICATION&page_size=1')).json()
	).items[0];
	const created = await request.post('/api/sequences', {
		data: {
			name: seqName,
			description: 'smoke sequence',
			steps: [
				{
					function_source: tx.source,
					function_interface: tx.interface,
					function_name: tx.name,
					parameters: { ANT: '15' },
					criteria: {},
					timeout_ms: '10000',
					flow: 'NEXT',
					on_fail: 'FAIL',
					device_name: '',
					note: 'main'
				},
				{
					function_source: 'DLL',
					function_interface: 'GPIB',
					function_name: helper,
					parameters: {},
					criteria: {},
					timeout_ms: '2000',
					flow: 'QUIT',
					on_fail: 'FAIL',
					device_name: '',
					note: ''
				}
			]
		}
	});
	expect(created.status(), 'sequence created').toBe(201);
	const { id } = (await created.json()) as { id: number };

	await page.goto(`/sequences/${id}`);
	await expect(page.getByText('Emitted text', { exact: true })).toBeVisible();
	// Step 2's function has no links, so it always emits exactly; step 1's
	// validity depends on the fixture's criteria requirements, so anchor on
	// the helper line only.
	const helperLine = `DLLGPIB|${helper}||NULL|2000|QUIT|FAIL`;
	await expect(page.locator('.emitted-preview')).toContainText(helperLine);
	await expect(page.getByRole('button', { name: 'Copy all' })).toBeVisible();

	// Reorder: move step 2 (the helper) up, the helper line must lead.
	await page.getByRole('button', { name: 'Move step up' }).nth(1).click();
	await expect(page.locator('.steps li').first()).toContainText(helper);
	const emitted = (await page.locator('.emitted-preview').textContent()) ?? '';
	expect(emitted.startsWith('DLLGPIB')).toBe(true);

	// Composer wiring: the dialog's "Add to sequence" appends a step to
	// this sequence.
	await page.goto(fnPath);
	await page.getByRole('button', { name: 'Compose script' }).click();
	const dialog = page.getByRole('dialog', { name: 'Script Composer' });
	await expect(dialog).toBeVisible();
	await dialog.getByText('Add to sequence').click();
	await dialog.getByRole('button', { name: 'Target sequence' }).click();
	// Select listboxes portal to the document body, outside the dialog.
	await page.getByRole('option', { name: `${seqName} (2 steps)` }).click();
	await dialog.getByRole('button', { name: 'Add step' }).click();
	await expect(dialog.getByText('Added as a new step.')).toBeVisible();
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Added as a new step' })
	).toBeVisible();

	const detail = (await (await request.get(`/api/sequences/${id}`)).json()) as { steps: unknown[] };
	expect(detail.steps.length, 'composer step appended').toBe(3);

	// Durable saves ride the polite toast channel; sequence delete confirms
	// through the alert dialog first (ui-refresh 07).
	await page.goto(`/sequences/${id}`);
	await page.getByRole('button', { name: 'Rename' }).click();
	await page.getByLabel('Sequence name').fill(`${seqName} renamed`);
	await page.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Sequence updated' })
	).toBeVisible();

	await page.getByRole('button', { name: 'Delete sequence' }).click();
	const confirm = page.getByRole('alertdialog');
	await expect(confirm).toBeVisible();
	await confirm.getByRole('button', { name: 'Delete' }).click();
	// The confirm itself must close on confirm — a stuck alertdialog would
	// invite a double-delete.
	await expect(confirm).not.toBeVisible();
	await expect(
		page
			.locator('section[aria-live=polite] [data-sonner-toast]')
			.filter({ hasText: 'Sequence deleted' })
	).toBeVisible();
	await expect(page).toHaveURL(/\/sequences$/);
});

test('Composer can find and append to a sequence beyond the first 100', async ({
	page,
	request
}) => {
	const prefix = `Composer sequence ${Date.now()}`;
	const targetName = `${prefix} 103`;
	const sequenceIds: number[] = [];
	try {
		for (let index = 1; index <= 103; index++) {
			const name = `${prefix} ${String(index).padStart(3, '0')}`;
			const created = await request.post('/api/sequences', {
				data: {
					name,
					description: '',
					steps: [
						{
							function_source: 'DLL',
							function_interface: 'WIFI_TEST',
							function_name: 'TX_MULTI_VERIFICATION'
						}
					]
				}
			});
			expect(created.status(), name).toBe(201);
			const { id } = (await created.json()) as { id: number };
			sequenceIds.push(id);
		}

		await page.goto(fnPath);
		await page.getByRole('button', { name: 'Compose script' }).click();
		const dialog = page.getByRole('dialog', { name: 'Script Composer' });
		await dialog.getByText('Add to sequence').click();
		await dialog.getByRole('button', { name: 'Target sequence' }).click();
		const picker = page.getByRole('dialog', { name: 'Choose Target sequence' });
		await picker.getByRole('combobox', { name: 'Search options' }).fill(targetName);
		const target = picker.getByRole('option', { name: `${targetName} (1 steps)` });
		await expect(target).toBeVisible();
		await target.click();
		await dialog.getByRole('button', { name: 'Add step' }).click();
		await expect(dialog.getByText('Added as a new step.')).toBeVisible();

		const saved = await request.get(`/api/sequences/${sequenceIds.at(-1)}`);
		expect(saved.status()).toBe(200);
		const detail = (await saved.json()) as { steps: { function_name: string }[] };
		expect(detail.steps).toHaveLength(2);
		expect(detail.steps[1].function_name).toBe('TX_MULTI_VERIFICATION');
	} finally {
		for (const id of sequenceIds) {
			await request.delete(`/api/sequences/${id}`);
		}
	}
});
