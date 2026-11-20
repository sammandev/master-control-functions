import { page } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import TroubleshootingDialog from '../../src/lib/components/TroubleshootingDialog.svelte';

const categories = [
	{
		key: 'connection',
		label: 'Connection',
		description: '',
		sort_order: 1,
		built_in: true,
		guide_count: 1
	}
];
const severities = [
	{
		key: 'critical',
		label: 'Critical',
		description: '',
		sort_order: 1,
		built_in: true,
		guide_count: 1,
		tone: 'destructive'
	}
];

async function renderDialog() {
	await render(TroubleshootingDialog, {
		open: true,
		mode: 'create',
		initialGuide: null,
		categories,
		severities,
		deviceSummaries: [
			{ device_name: 'WIFI_TEST', functions: 0, parameters: 0, criteria: 0, examples: 0 }
		],
		onOpenChange: vi.fn(),
		onSaved: vi.fn(),
		onRegistryChanged: vi.fn()
	});
}

afterEach(() => vi.unstubAllGlobals());

describe('TroubleshootingDialog accessibility', () => {
	it('orders problem context fields for rapid triage', async () => {
		await renderDialog();
		const section = document.querySelector('.form-section');
		const labels = Array.from(section?.querySelectorAll('label') ?? []).map((label) =>
			label.textContent?.replace(/\s+/g, ' ').trim()
		);
		expect(labels).toEqual([
			'Solution / Manifestation Title *',
			'Error Code / Problem Tag',
			'Device / Interface Scope',
			'Related Function Name',
			'Testing Stage / Category',
			'Severity'
		]);
	});

	it('renders structured field errors and wires every invalid control to an exact error id', async () => {
		const fetchMock = vi.fn(
			async () =>
				new Response(
					JSON.stringify({
						error: 'validation failed',
						fields: {
							title: 'Title server error',
							category: 'Category server error',
							severity: 'Severity server error',
							device_name: 'Device server error',
							symptom: 'Symptom server error',
							'steps.title': 'Step title server error',
							'steps.instruction': 'Instruction server error'
						}
					}),
					{ status: 422, headers: { 'content-type': 'application/json' } }
				)
		);
		vi.stubGlobal('fetch', fetchMock);
		await renderDialog();

		await page
			.getByRole('textbox', { name: /Solution \/ Manifestation Title/ })
			.fill('Valid title');
		await page.getByRole('textbox', { name: /Observed Symptom/ }).fill('Valid symptom');
		await page.getByRole('textbox', { name: /Action Headline/ }).fill('Valid step');
		await page.getByRole('textbox', { name: /Detailed Instruction/ }).fill('Valid instruction');
		const invalidControls = Array.from(
			document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement>(':invalid')
		).map((control) => control.id || control.name || control.outerHTML.slice(0, 80));
		expect(invalidControls).toEqual([]);
		const submitButton = Array.from(document.querySelectorAll('button')).find(
			(button) => button.textContent?.trim() === 'Create Guide'
		);
		expect(submitButton?.type).toBe('submit');
		submitButton?.form?.requestSubmit(submitButton);
		await expect.poll(() => fetchMock.mock.calls.length).toBe(1);

		const expectations = [
			[page.getByRole('textbox', { name: /Solution \/ Manifestation Title/ }), 'guide-title-error'],
			[page.getByRole('combobox', { name: /^Category:/ }), 'guide-category-error'],
			[page.getByRole('combobox', { name: /^Severity:/ }), 'guide-severity-error'],
			[page.getByRole('combobox', { name: /^Device \/ Interface Scope:/ }), 'guide-device-error'],
			[page.getByRole('textbox', { name: /Observed Symptom/ }), 'guide-symptom-error'],
			[page.getByRole('textbox', { name: /Action Headline/ }), 'step-title-error-0'],
			[page.getByRole('textbox', { name: /Detailed Instruction/ }), 'step-instruction-error-0']
		] as const;

		for (const [control, errorId] of expectations) {
			await expect.element(control).toHaveAttribute('aria-invalid', 'true');
			await expect.element(control).toHaveAttribute('aria-describedby', errorId);
			expect(document.getElementById(errorId)).not.toBeNull();
		}
	});
});
