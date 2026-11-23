import { page } from 'vitest/browser';
import { describe, expect, test } from 'vitest';
import { render } from 'vitest-browser-svelte';
import StructuredText from '../../src/lib/components/StructuredText.svelte';
import StructuredTextEditor from '../../src/lib/components/StructuredTextEditor.svelte';

describe('StructuredTextEditor', () => {
	test('converts the selected textarea lines to bullets and numbered lists', async () => {
		await render(StructuredTextEditor, {
			id: 'structured-notes',
			'aria-label': 'Troubleshooting notes',
			value: 'Inspect DUT\nRetry flow'
		});
		const textarea = page.getByRole('textbox', { name: 'Troubleshooting notes' });
		const selectAll = () => {
			const element = textarea.element() as HTMLTextAreaElement;
			element.setSelectionRange(0, element.value.length);
		};
		selectAll();

		await page.getByRole('button', { name: 'Bulleted list' }).click();
		await expect.element(textarea).toHaveValue('- Inspect DUT\n- Retry flow');

		selectAll();
		await page.getByRole('button', { name: 'Numbered list' }).click();
		await expect.element(textarea).toHaveValue('1. Inspect DUT\n2. Retry flow');
	});
});

describe('StructuredText', () => {
	test('renders semantic lists and escapes HTML-like input', async () => {
		const view = await render(StructuredText, {
			text: '- Check prompt\n- Check connection\n\n1. Reconnect\n2. Retry'
		});

		expect(document.querySelectorAll('ul, ol')).toHaveLength(2);
		expect(document.querySelectorAll('li')).toHaveLength(4);
		await expect.element(page.getByText('Check prompt')).toBeVisible();

		await view.rerender({ text: '<img src=x onerror=alert(1)>' });
		await expect.element(page.getByText('<img src=x onerror=alert(1)>')).toBeVisible();
		expect(document.querySelector('img')).toBeNull();
	});
});
