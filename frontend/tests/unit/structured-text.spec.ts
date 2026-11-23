import { describe, expect, test } from 'vitest';
import { applyListMarkup, parseStructuredText } from '../../src/lib/structured-text';

describe('parseStructuredText', () => {
	test('parses mixed paragraphs, bullets, and ordered lists', () => {
		expect(
			parseStructuredText(
				'Failure occurs during login.\nThe prompt never arrives.\n\n- Check the DUT connection\n- Check the shell prompt\n\n1. Reconnect the cable\n2. Retry the flow'
			)
		).toEqual([
			{
				kind: 'paragraph',
				text: 'Failure occurs during login.\nThe prompt never arrives.'
			},
			{
				kind: 'unordered-list',
				items: ['Check the DUT connection', 'Check the shell prompt']
			},
			{
				kind: 'ordered-list',
				items: ['Reconnect the cable', 'Retry the flow']
			}
		]);
	});

	test('keeps non-list text as an escaped-safe plain-text paragraph', () => {
		expect(parseStructuredText('<img src=x onerror=alert(1)>')).toEqual([
			{ kind: 'paragraph', text: '<img src=x onerror=alert(1)>' }
		]);
	});
});

describe('applyListMarkup', () => {
	test('numbers only the selected lines and restores the selection', () => {
		expect(applyListMarkup('Before\nInspect DUT\nRetry flow\nAfter', 7, 29, 'ordered')).toEqual({
			value: 'Before\n1. Inspect DUT\n2. Retry flow\nAfter',
			selectionStart: 7,
			selectionEnd: 35
		});
	});

	test('toggles bullet markers without touching neighboring lines', () => {
		expect(applyListMarkup('Keep\n- Inspect DUT\n- Retry flow\nDone', 5, 31, 'bullet')).toEqual({
			value: 'Keep\nInspect DUT\nRetry flow\nDone',
			selectionStart: 5,
			selectionEnd: 27
		});
	});
});
