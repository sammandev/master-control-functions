import { expect, test } from '@playwright/test';

test('an empty Sequence catalog shows its catalog empty state', async ({ page }) => {
	await page.route(
		(url) => url.pathname === '/api/sequences',
		(route) => route.fulfill({ json: { items: [], total: 0, page: 1, page_size: 20 } })
	);

	await page.goto('/sequences');
	await expect(page.getByText('0 sequences', { exact: true })).toBeVisible();
	const emptyCatalog = page.getByText('No sequences yet.', { exact: true });
	const noMatches = page.getByText('No sequences match the search.', { exact: true });
	const search = page.getByRole('searchbox', { name: 'Search sequences' });
	await expect(emptyCatalog).toBeVisible();
	await expect(noMatches).toBeHidden();

	await search.fill('does-not-match-any-sequence');
	await expect(noMatches).toBeVisible();

	await search.fill('%');
	await expect(emptyCatalog).toBeVisible();
	await expect(noMatches).toBeHidden();
});

test('a settled search with no Sequence matches does not look like an empty catalog', async ({
	page,
	request
}) => {
	const name = 'E2E Sequence empty search';
	const searchTerm = 'does-not-match-any-sequence';
	const created = await request.post('/api/sequences', {
		data: {
			name,
			steps: [
				{
					function_source: 'DLL',
					function_interface: 'WIFI_TEST',
					function_name: 'DUT_REMOVE',
					timeout_ms: '5000',
					flow: 'NEXT',
					on_fail: 'FAIL'
				}
			]
		}
	});
	expect(created.status()).toBe(201);
	const { id } = (await created.json()) as { id: number };

	try {
		await page.goto('/sequences');
		const sequence = page.getByRole('link', { name, exact: true });
		const search = page.getByRole('searchbox', { name: 'Search sequences' });
		const noMatches = page.getByText('No sequences match the search.', { exact: true });
		const emptyCatalog = page.getByText('No sequences yet.', { exact: true });
		await expect(sequence).toBeVisible();

		await search.fill(searchTerm);
		await expect(page.getByText('0 sequences', { exact: true })).toBeVisible();
		await expect(noMatches).toBeVisible();
		await expect(emptyCatalog).toBeHidden();
		await expect(sequence).toBeHidden();

		await search.clear();
		await expect(sequence).toBeVisible();
		await expect(noMatches).toBeHidden();
		await expect(emptyCatalog).toBeHidden();

		await page.route(
			(url) => url.pathname === '/api/sequences' && url.searchParams.get('search') === searchTerm,
			(route) => route.fulfill({ status: 503, json: { error: 'Sequence list unavailable' } })
		);
		await search.fill(searchTerm);
		await expect(page.getByText('Count unavailable', { exact: true })).toBeVisible();
		await expect(page.getByRole('alert')).toContainText('Sequence list unavailable');
		await expect(page.getByText('Sequences could not be loaded.', { exact: true })).toBeVisible();
		await expect(noMatches).toBeHidden();
		await expect(emptyCatalog).toBeHidden();
	} finally {
		await request.delete(`/api/sequences/${id}`);
	}
});
