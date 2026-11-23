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

	test('keeps existing guides with repeated numbered markers as one list', () => {
		expect(parseStructuredText('1. Inspect\n1. Retry')).toEqual([
			{ kind: 'ordered-list', items: ['Inspect', 'Retry'] }
		]);
	});
});

describe('getListEdit', () => {
	test('keeps a blank boundary so a numbered list started after exiting renders separately', () => {
		const value = '1. Inspect\n2. Retry\n3. ';
		const edit = getListEdit(value, value.length, 'Enter')!;
		const next = value.slice(0, edit.start) + edit.text + value.slice(edit.end);
		expect(parseStructuredText(next + '1. Reconnect')).toEqual([
			{ kind: 'ordered-list', items: ['Inspect', 'Retry'] },
			{ kind: 'ordered-list', items: ['Reconnect'] }
		]);
	});

	test('increments a numbered marker on the current line', () => {
		expect(getListEdit('Before\n  9. Inspect DUT\nAfter', 23, 'Enter')).toEqual({
			start: 23,
			end: 23,
			text: '\n  10. '
		});
	});

	test('Backspace removes only the empty marker', () => {
		expect(getListEdit('Keep\n- \nDone', 7, 'Backspace')).toEqual({
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
