import { expect, test } from '@playwright/test';

test('keeps the latest sequence search results when an older response arrives', async ({
	page,
	request
}, testInfo) => {
	const oldName = 'A Scout 001';
	const latestName = 'ZZ Scout final target';
	const sequenceIds: number[] = [];
	const held = Promise.withResolvers<boolean>();
	const release = Promise.withResolvers<void>();
	const oldSearch = (url: URL) =>
		url.pathname === '/api/sequences' && url.searchParams.get('search') === oldName;

	try {
		for (const name of [oldName, latestName]) {
			const response = await request.post('/api/sequences', {
				data: {
					name,
					steps: [
						{
							function_source: 'DLL',
							function_interface: 'GPIB',
							function_name: 'E2E_SEQUENCE_SEARCH'
						}
					]
				}
			});
			expect(response.status()).toBe(201);
			const { id } = (await response.json()) as { id: number };
			sequenceIds.push(id);
		}

		await page.route(oldSearch, async (route) => {
			const response = await route.fetch();
			held.resolve(response.ok());
			await release.promise;
			await route.fulfill({ response });
		});

		await page.goto('/sequences');
		await expect(page.getByText(/^\d+ sequences?$/)).toBeVisible();
		const search = page.getByRole('searchbox', { name: 'Search sequences' });
		await search.fill(oldName);
		expect(await held.promise).toBe(true);

		await search.fill(latestName);
		const latestResult = page.getByRole('link', { name: latestName, exact: true });
		const oldResult = page.getByRole('link', { name: oldName, exact: true });
		await expect(latestResult).toBeVisible();
		await expect(oldResult).toBeHidden();

		const delayedResponse = page.waitForResponse((response) => oldSearch(new URL(response.url())));
		release.resolve();
		await (await delayedResponse).finished();
		await page.screenshot({ path: testInfo.outputPath('sequence-search.png') });

		await expect(search).toHaveValue(latestName);
		await expect(latestResult).toBeVisible();
		await expect(oldResult).toBeHidden();
	} finally {
		release.resolve();
		for (const id of sequenceIds) await request.delete(`/api/sequences/${id}`);
	}
});
