import { page } from 'vitest/browser';
import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import TroubleshootingRegistryPicker from '../../src/lib/components/TroubleshootingRegistryPicker.svelte';

const items = [
	{
		key: 'connection',
		label: 'DUT Connection',
		description: 'Connection and setup failures',
		sort_order: 10,
		built_in: true,
		guide_count: 4
	},
	{
		key: 'execution',
		label: 'Flow Execution',
		description: 'Script and flow failures',
		sort_order: 20,
		built_in: false,
		guide_count: 2
	}
];

function renderPicker(
	onCreate = vi.fn(async (label: string) => ({
		key: label.toLowerCase().replace(/\s+/g, '-'),
		label,
		description: '',
		sort_order: 30,
		built_in: false,
		guide_count: 0
	}))
) {
	render(TroubleshootingRegistryPicker, {
		items,
		value: 'connection',
		kind: 'category',
		onCreate
	});
	return onCreate;
}

describe('TroubleshootingRegistryPicker', () => {
	it('selects a matching item and closes the picker', async () => {
		renderPicker();

		await page.getByRole('button', { name: 'Category: DUT Connection' }).click();
		const search = page.getByRole('combobox', { name: 'Search categories' });
		await search.fill('Flow');
		await page.getByRole('option', { name: /Flow Execution/ }).click();

		await expect.element(search).not.toBeInTheDocument();
		await expect
			.element(page.getByRole('button', { name: 'Category: Flow Execution' }))
			.toBeVisible();
	});

	it('offers creation only for an unmatched search term', async () => {
		const onCreate = renderPicker();

		await page.getByRole('button', { name: 'Category: DUT Connection' }).click();
		await page.getByRole('combobox', { name: 'Search categories' }).fill('Production');

		await expect.element(page.getByRole('button', { name: 'Create “Production”' })).toBeVisible();
		expect(onCreate).not.toHaveBeenCalled();
	});

	it('selects a successfully created item', async () => {
		const onCreate = renderPicker(
			vi.fn(async () => ({
				key: 'production',
				label: 'Production',
				description: '',
				sort_order: 30,
				built_in: false,
				guide_count: 0
			}))
		);

		await page.getByRole('button', { name: 'Category: DUT Connection' }).click();
		await page.getByRole('combobox', { name: 'Search categories' }).fill('Production');
		await page.getByRole('button', { name: 'Create “Production”' }).click();

		await expect.element(page.getByRole('button', { name: 'Category: Production' })).toBeVisible();
		expect(onCreate).toHaveBeenCalledWith('Production');
	});

	it('keeps the typed term visible when creation fails', async () => {
		const onCreate = vi.fn(async () => {
			throw new Error('Category creation failed');
		});
		renderPicker(onCreate);

		await page.getByRole('button', { name: 'Category: DUT Connection' }).click();
		const search = page.getByRole('combobox', { name: 'Search categories' });
		await search.fill('Production');
		await page.getByRole('button', { name: 'Create “Production”' }).click();

		await expect.element(page.getByRole('alert')).toHaveTextContent('Category creation failed');
		await expect.element(search).toHaveValue('Production');
		await expect.element(page.getByRole('button', { name: 'Create “Production”' })).toBeEnabled();
	});

	it('does not offer creation for an exact label or key match', async () => {
		renderPicker();
		await page.getByRole('button', { name: 'Category: DUT Connection' }).click();
		const search = page.getByRole('combobox', { name: 'Search categories' });

		for (const term of ['DUT Connection', 'execution']) {
			await search.fill(term);
			await expect.element(page.getByRole('button', { name: /^Create / })).not.toBeInTheDocument();
		}
	});
});
