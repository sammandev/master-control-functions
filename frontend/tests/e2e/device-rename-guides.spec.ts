import { expect, test } from '@playwright/test';

test('renaming a Device keeps its Troubleshooting guide under the new filter', async ({
	page,
	request
}) => {
	const original = 'RADIO BENCH';
	const renamed = 'RADIO STATION';
	const guide = {
		title: 'RADIO BENCH loses connection',
		category: 'connection',
		severity: 'warning',
		device_name: original,
		symptom: 'RADIO BENCH disconnects during startup.',
		root_cause: 'Loose Ethernet cable at RADIO BENCH.',
		steps: [
			{
				title: 'Check RADIO BENCH cabling',
				instruction: 'Reseat the Ethernet cable at RADIO BENCH.',
				script_snippet: 'ping RADIO BENCH',
				verification: 'RADIO BENCH responds.'
			}
		]
	};
	const devices = '/api/devices';
	const guides = '/api/troubleshooting';
	const registered = await request.post(devices, { data: { device_name: original } });
	expect(registered.status()).toBe(201);
	let guideId: number | undefined;
	let device = original;

	async function filterByDevice(name: string) {
		await page.goto('/troubleshooting');
		await page.getByRole('button', { name: 'Device / Interface: All Devices' }).click();
		await page.getByRole('combobox', { name: 'Search devices' }).fill(name);
		const filtered = page.waitForResponse((response) => {
			const url = new URL(response.url());
			return url.pathname === guides && url.searchParams.get('device') === name;
		});
		await page.getByRole('option', { name, exact: true }).click();
		expect((await filtered).status()).toBe(200);
		await expect(page.getByRole('button', { name: `Device / Interface: ${name}` })).toBeVisible();
		await expect(page.getByText('Loading troubleshooting guides…')).toBeHidden();
	}

	try {
		const created = await request.post(guides, { data: guide });
		expect(created.status()).toBe(201);
		const saved = await created.json();
		guideId = saved.id;

		const guideHeading = page.getByRole('heading', { name: guide.title, exact: true });
		await filterByDevice(original);
		await expect(guideHeading).toBeVisible();

		await page.goto('/devices');
		await page.getByRole('button', { name: `Edit ${original}`, exact: true }).click();
		const dialog = page.getByRole('dialog', { name: `Edit ${original}` });
		await dialog.getByLabel('Device name').fill(renamed);
		await dialog.getByRole('button', { name: 'Save', exact: true }).click();
		await expect(page.getByRole('link', { name: renamed, exact: true })).toBeVisible();
		device = renamed;

		await filterByDevice(renamed);
		const screenshot = process.env.DEVICE_RENAME_GUIDES_SCREENSHOT;
		if (screenshot) await page.screenshot({ path: screenshot });
		await expect(guideHeading).toBeVisible();

		const detail = await request.get(`${guides}/${guideId}`);
		expect(detail.status()).toBe(200);
		expect(await detail.json()).toMatchObject({
			...guide,
			device_name: renamed,
			steps: saved.steps
		});
		const oldScope = await request.get(guides, { params: { device: original } });
		expect(oldScope.status()).toBe(200);
		expect(await oldScope.json()).toMatchObject({ items: [], total: 0 });
	} finally {
		if (guideId !== undefined) await request.delete(`${guides}/${guideId}`);
		await request.delete(`${devices}/${encodeURIComponent(device)}`, {
			data: { confirmation: `DELETE ${device}` }
		});
	}
});
