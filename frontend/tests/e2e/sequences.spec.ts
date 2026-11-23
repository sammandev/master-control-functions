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

test('sorting by Steps orders every sequence before paging', async ({
	page,
	request
}, testInfo) => {
	const step = {
		function_source: 'DLL',
		function_interface: 'GPIB',
		function_name: 'E2E_SEQUENCE_SORT'
	};
	const longest = 'ZZ Sort longest';
	const sequences = [
		...Array.from({ length: 20 }, (_, i) => ({
			name: `Sort ${String(i + 1).padStart(2, '0')}`,
			steps: [step]
		})),
		{ name: longest, steps: [step, step, step] }
	];
	const sequenceIds: number[] = [];

	try {
		for (const sequence of sequences) {
			const response = await request.post('/api/sequences', { data: sequence });
			expect(response.status()).toBe(201);
			sequenceIds.push(((await response.json()) as { id: number }).id);
		}

		await page.goto('/sequences');
		await expect(page.getByText('21 sequences')).toBeVisible();
		// The longest sequence's name sorts last, so it starts on page 2.
		await expect(page.getByRole('link', { name: longest })).toBeHidden();

		const stepsHead = page.getByRole('columnheader', { name: 'Steps' });
		await stepsHead.getByRole('button').click();
		await stepsHead.getByRole('button').click();
		await expect(stepsHead).toHaveAttribute('aria-sort', 'descending');

		const firstRow = page.getByRole('row').nth(1);
		await expect(firstRow.getByRole('link')).toHaveText(longest);
		await expect(firstRow.getByRole('cell').nth(2)).toHaveText('3');
		await page.screenshot({ path: testInfo.outputPath('sequence-steps-sort.png') });
	} finally {
		for (const id of sequenceIds) await request.delete(`/api/sequences/${id}`);
	}
});
