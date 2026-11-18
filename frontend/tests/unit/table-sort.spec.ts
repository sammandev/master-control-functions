import { describe, expect, it } from 'vitest';
import { ariaSortValue, nextSort, sortArrow, sortRows } from '../../src/lib/table-sort';

describe('table sorting', () => {
	it('cycles ascending, descending, then the original row order', () => {
		const rows = [{ name: 'Bravo' }, { name: 'Alpha' }, { name: 'Charlie' }];
		let sorting = nextSort('', 'asc', 'name');
		expect(sorting).toEqual({ key: 'name', dir: 'asc' });
		expect(sortRows(rows, (row) => row.name, sorting.dir).map((row) => row.name)).toEqual([
			'Alpha',
			'Bravo',
			'Charlie'
		]);
		sorting = nextSort(sorting.key, sorting.dir, 'name');
		expect(sorting).toEqual({ key: 'name', dir: 'desc' });
		sorting = nextSort(sorting.key, sorting.dir, 'name');
		expect(sorting).toEqual({ key: '', dir: 'asc' });
		expect(ariaSortValue('name', sorting.key, sorting.dir)).toBeUndefined();
		expect(sortArrow('name', sorting.key, sorting.dir)).toBe('');
		expect(rows.map((row) => row.name)).toEqual(['Bravo', 'Alpha', 'Charlie']);
	});

	it('starts a different column in ascending order', () => {
		expect(nextSort('name', 'desc', 'type')).toEqual({ key: 'type', dir: 'asc' });
	});
});
