import { page, userEvent } from 'vitest/browser';
import { describe, expect, test } from 'vitest';
import { render } from 'vitest-browser-svelte';
import StructuredText from '../../src/lib/components/StructuredText.svelte';
import StructuredTextEditor from '../../src/lib/components/StructuredTextEditor.svelte';

describe('StructuredTextEditor', () => {
	test('continues typed list markers and exits when the next item is empty', async () => {
		await render(StructuredTextEditor, {
			id: 'structured-notes',
			'aria-label': 'Troubleshooting notes'
		});
		const textarea = page.getByRole('textbox', { name: 'Troubleshooting notes' });
		await textarea.click();
		await userEvent.keyboard('- Inspect DUT{Enter}Retry flow{Enter}{Enter}');
		await expect.element(textarea).toHaveValue('- Inspect DUT\n- Retry flow\n');
		await userEvent.keyboard('1. Reconnect{Enter}Retest{Enter}{Backspace}');
		await expect
			.element(textarea)
			.toHaveValue('- Inspect DUT\n- Retry flow\n1. Reconnect\n2. Retest\n');
		await expect
			.element(page.getByRole('button', { name: /bulleted list|numbered list/i }))
			.not.toBeInTheDocument();
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
