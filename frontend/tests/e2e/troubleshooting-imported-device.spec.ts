import { expect, test } from '@playwright/test';

const device = 'IMPORTED BENCH';
const guides = '/api/troubleshooting';

test.afterEach(async ({ request }) => {
	const listed = await request.get(guides, { params: { device } });
	for (const guide of (await listed.json()).items) {
		await request.delete(`${guides}/${guide.id}`);
	}
});

test('a Device only named by an imported guide can be picked in the Troubleshooting filter', async ({
	page,
	request
}) => {
	const title = 'IMPORTED BENCH loses its link';
	const catalog = await request.get('/api/devices');
	expect(JSON.stringify(await catalog.json())).not.toContain(`"device_name":"${device}"`);
	const file = {
		guides: [
			{
				title,
				category: 'connection',
				severity: 'warning',
				device_name: device,
				symptom: 'IMPORTED BENCH drops its link after startup.',
				steps: [{ title: 'Reseat the cable', instruction: 'Reseat the IMPORTED BENCH cable.' }]
			}
		]
	};

	await page.goto('/troubleshooting');
	await page.getByRole('button', { name: 'Import', exact: true }).click();
	const bulk = page.getByRole('dialog', { name: 'Import Guides' });
	await bulk.getByLabel('Guide file', { exact: true }).setInputFiles({
		name: 'guides.json',
		mimeType: 'application/json',
		buffer: Buffer.from(JSON.stringify(file))
	});
	await bulk.getByRole('button', { name: 'Preview import' }).click();
	await expect(bulk.getByText('Create: 1')).toBeVisible();
	await bulk.getByRole('button', { name: 'Confirm import' }).click();
	await expect(bulk).toBeHidden();

	await page.getByRole('button', { name: 'Device / Interface: All Devices' }).click();
	await page.getByRole('combobox', { name: 'Search devices' }).fill(device);
	const screenshot = process.env.TROUBLESHOOTING_IMPORTED_DEVICE_SCREENSHOT;
	if (screenshot) {
		await page.waitForTimeout(300);
		await page.screenshot({ path: screenshot });
	}
	const filtered = page.waitForResponse((response) => {
		const url = new URL(response.url());
		return url.pathname === guides && url.searchParams.get('device') === device;
	});
	await page.getByRole('option', { name: device, exact: true }).click();
	expect((await filtered).status()).toBe(200);
	await expect(page.getByRole('button', { name: `Device / Interface: ${device}` })).toBeVisible();
	await expect(page.getByRole('heading', { name: title, exact: true })).toBeVisible();
});
