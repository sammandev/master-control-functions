import { describe, expect, test } from 'vitest';
import { getListEdit, parseStructuredText } from '../../src/lib/structured-text';

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

describe('getListEdit', () => {
	test('increments a numbered marker on the current line', () => {
		expect(getListEdit('Before\n  9. Inspect DUT\nAfter', 23, 'Enter')).toEqual({
			start: 23,
			end: 23,
			text: '\n  10. '
		});
	});

	test.each(['Enter', 'Backspace'] as const)('%s removes only an empty marker', (key) => {
		expect(getListEdit('Keep\n- \nDone', 7, key)).toEqual({
			start: 5,
			end: 7,
			text: ''
		});
	});

	test('leaves plain text and unspaced markers to native editing', () => {
		expect(getListEdit('Check 1. Retry', 14, 'Enter')).toBeUndefined();
		expect(getListEdit('-', 1, 'Enter')).toBeUndefined();
		expect(getListEdit('1. ', 0, 'Backspace')).toBeUndefined();
	});
});
