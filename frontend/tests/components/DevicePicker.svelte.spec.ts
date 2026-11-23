import { page } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DevicePicker from '../../src/lib/components/DevicePicker.svelte';

afterEach(() => vi.restoreAllMocks());

describe('DevicePicker', () => {
	it('loads devices when an initial summary list is empty', async () => {
		const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
			Response.json({
				items: [{ device_name: 'WIFI_TEST', functions: 0, parameters: 0, criteria: 0, examples: 0 }]
			})
		);

		render(DevicePicker, {
			device: '',
			summaries: [],
			onValueChange: vi.fn()
		});

		await page.getByRole('button', { name: 'Device: Global definitions' }).click();
		await expect.element(page.getByRole('option', { name: 'WIFI_TEST' })).toBeVisible();
		expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/devices');
	});

	it('keeps the trigger as a button while showing an invalid state', async () => {
		render(DevicePicker, {
			device: '',
			summaries: [{ device_name: '', functions: 0, parameters: 0, criteria: 0, examples: 0 }],
			invalid: true,
			describedBy: 'device-error',
			onValueChange: vi.fn()
		});

		const trigger = page.getByRole('button', { name: 'Device: Global definitions' });
		await expect.element(trigger).toHaveAttribute('aria-invalid', 'true');
		await expect.element(trigger).toHaveAttribute('aria-describedby', 'device-error');
	});
});
