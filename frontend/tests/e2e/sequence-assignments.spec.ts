import { expect, test } from '@playwright/test';

const duplicate = 'SCOUT=1\nSCOUT=2';

for (const [label, field] of [
	['Parameters', 'parameters'],
	['Criteria', 'criteria']
] as const) {
	test(`rejects duplicate Sequence step ${label.toLowerCase()} without saving`, async ({
		page,
		request
	}) => {
		const created = await request.post('/api/sequences', {
			data: {
				name: `Duplicate ${field} test ${Date.now()}`,
				steps: [
					{ function_source: 'DLL', function_name: 'E2E_ASSIGNMENTS', [field]: { SCOUT: '0' } }
				]
			}
		});
		expect(created.status()).toBe(201);
		const { id } = await created.json();
		const path = `/sequences/${id}`;

		try {
			await page.goto(path);
			await page.getByRole('button', { name: 'Edit', exact: true }).click();
			const assignments = page.getByLabel(`${label} (one name=value per line)`);
			await assignments.fill(duplicate);
			await page.getByRole('button', { name: 'Save step', exact: true }).click();

			await expect(page.getByRole('alert')).toHaveText('Duplicate assignment name: "SCOUT"');
			await expect(assignments).toHaveValue(duplicate);
			await expect(page.getByText('Step saved', { exact: true })).toHaveCount(0);

			const saved = await request.get(`/api${path}`);
			expect(saved.ok()).toBe(true);
			expect(await saved.json()).toMatchObject({ steps: [{ [field]: { SCOUT: '0' } }] });
		} finally {
			await request.delete(`/api${path}`);
		}
	});
}
