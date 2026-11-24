import { randomUUID } from 'node:crypto';
import { test, expect } from '@playwright/test';

for (const { kind, label, device } of [
	{ kind: 'parameter', label: 'Parameter', device: 'WIFI_TEST' },
	{ kind: 'criteria', label: 'Criteria', device: '' }
]) {
	test(`Orphan ${kind} repair creates the exact missing definition after a typing attempt`, async ({
		page,
		request
	}) => {
		const suffix = randomUUID().slice(0, 8).toUpperCase();
		const name = `REPAIR_TARGET_${suffix}`;
		const definitionPath = `/api/${kind}-definitions`;
		const data = { name, device_name: device, type: 'string' };
		const createdFunction = await request.post('/api/functions', {
			data: { source: 'DLL', name: `REPAIR_FIXTURE_${suffix}`, interface: device }
		});
		expect(createdFunction.status()).toBe(201);
		const { id: functionId } = await createdFunction.json();
		const linksPath = `/api/functions/${functionId}/${kind}-links`;
		let linkId: number | undefined;
		let definitionId: number | undefined;
		try {
			const seededDefinition = await request.post(definitionPath, { data });
			expect(seededDefinition.status()).toBe(201);
			const seeded = await seededDefinition.json();
			definitionId = seeded.id;
			const link = await request.post(linksPath, { data: { name, device_name: device } });
			expect(link.status()).toBe(201);
			linkId = (await link.json()).id;
			expect((await request.delete(`${definitionPath}/${definitionId}?force=true`)).status()).toBe(
				204
			);
			definitionId = undefined;

			await page.goto('/orphans');
			const row = page
				.getByRole('row')
				.filter({ has: page.getByRole('cell', { name, exact: true }) });
			await row.getByRole('button', { name: `Create missing ${label} definition` }).click();
			const dialog = page.getByRole('dialog');
			const nameField = dialog.getByRole('textbox', { name: 'Name', exact: true });
			await expect(nameField).toHaveValue(name);
			await nameField.click({ force: true });
			await page.keyboard.press('End');
			await page.keyboard.press('X');
			expect.soft(await nameField.isDisabled(), 'definition name is locked').toBe(true);

			const [response] = await Promise.all([
				page.waitForResponse(
					(response) =>
						new URL(response.url()).pathname === definitionPath &&
						response.request().method() === 'POST'
				),
				dialog.getByRole('button', { name: 'Create and resolve' }).click()
			]);
			expect(response.status()).toBe(201);
			const created = await response.json();
			definitionId = created.id;
			expect.soft(created).toMatchObject(data);
			await expect(row).toHaveCount(0);
		} finally {
			if (linkId !== undefined) await request.delete(`${linksPath}/${linkId}`);
			if (definitionId !== undefined) await request.delete(`${definitionPath}/${definitionId}`);
			await request.delete(`/api/functions/${functionId}`);
		}
	});
}
