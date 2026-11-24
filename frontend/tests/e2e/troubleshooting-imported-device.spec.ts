import { expect, test } from '@playwright/test';

test('a Device only named by an imported guide can be picked in the Troubleshooting filter', async ({
	page,
	request
}) => {
	const device = 'IMPORTED BENCH';
	const title = 'IMPORTED BENCH loses its link';
	const guides = '/api/troubleshooting';
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
	const imported = await request.post(`${guides}/import/commit`, {
		multipart: {
			file: {
				name: 'guides.json',
				mimeType: 'application/json',
				buffer: Buffer.from(JSON.stringify(file))
			}
		}
	});
	expect(imported.status()).toBe(200);
	const listed = await request.get(guides, { params: { device } });
	const guideId: number = (await listed.json()).items[0].id;

	try {
		await page.goto('/troubleshooting');
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
	} finally {
		await request.delete(`${guides}/${guideId}`);
	}
});
