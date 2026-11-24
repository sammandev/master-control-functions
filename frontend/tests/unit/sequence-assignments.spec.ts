import { describe, expect, test } from 'vitest';
import { parseAssignments } from '../../src/lib/sequence-assignments';

describe('parseAssignments', () => {
	test.each(['=42', ' \t=42', 'timeout=10\n=42'])('rejects an empty name in %j', (input) => {
		expect(parseAssignments(input)).toEqual({ error: 'Use name=value on each line: "=42"' });
	});

	test('rejects a line without an equals sign', () => {
		expect(parseAssignments('timeout=10\nmissing')).toEqual({
			error: 'Use name=value on each line: "missing"'
		});
	});

	test('ignores blank lines and keeps empty values and equals signs in values', () => {
		expect(parseAssignments('\n timeout = 10 \n \t\n empty=\n expression = a=b \n')).toEqual({
			values: { timeout: '10', empty: '', expression: 'a=b' }
		});
	});
});
