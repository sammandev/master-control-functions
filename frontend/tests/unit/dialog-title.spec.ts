import { describe, expect, it } from 'vitest';
import { dialogTitleCase } from '#lib/utils.ts';

describe('dialogTitleCase', () => {
	it.each([
		['Manage troubleshooting lists', 'Manage Troubleshooting Lists'],
		['Example of test', 'Example of Test'],
		['Script composer', 'Script Composer'],
		['Delete sequence', 'Delete Sequence'],
		[
			'Add Alternate Solution for Existing Problem / Error Code',
			'Add Alternate Solution for Existing Problem / Error Code'
		],
		['Delete WIFI_TEST?', 'Delete WIFI_TEST?'],
		['Force delete device_name?', 'Force Delete device_name?'],
		['Open iPhone settings', 'Open iPhone Settings']
	])('formats %s', (input, expected) => {
		expect(dialogTitleCase(input)).toBe(expected);
	});
});
