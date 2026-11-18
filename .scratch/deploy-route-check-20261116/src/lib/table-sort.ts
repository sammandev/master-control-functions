// Shared list-table sorting (ticket 09): one mechanism for every table
// instead of per-page reimplementations. Functions and the dictionaries
// sort server-side through their list APIs; every other list table sorts
// the loaded rows client-side through sortRows(). The header presentation
// lives in SortHead.svelte; this module holds the state transitions and
// the aria-sort/arrow derivations the existing functions/dictionary
// headers also use.

export type SortDir = 'asc' | 'desc';

// Flip semantics shared by every sortable header: clicking the active
// column flips direction; clicking another column sorts it ascending.
export function nextSort(
	currentKey: string,
	currentDir: SortDir,
	col: string
): { key: string; dir: SortDir } {
	if (currentKey === col) {
		return { key: col, dir: currentDir === 'asc' ? 'desc' : 'asc' };
	}
	return { key: col, dir: 'asc' };
}

// aria-sort on the active column header only; undefined drops the
// attribute for the rest (Svelte omits undefined attributes).
export function ariaSortValue(
	col: string,
	activeKey: string,
	dir: SortDir
): 'ascending' | 'descending' | undefined {
	if (activeKey !== col) return undefined;
	return dir === 'asc' ? 'ascending' : 'descending';
}

// Direction arrow shown in the active header only (both themes: the
// arrow inherits the button's currentColor, no color coupling).
export function sortArrow(col: string, activeKey: string, dir: SortDir): string {
	if (activeKey !== col) return '';
	return dir === 'asc' ? '▲' : '▼';
}

// Sortable header button styles: presentation shared with SortHead and
// with the server-sorted pages so the affordance is identical everywhere.
export const SORT_BUTTON_CLASS =
	'inline-flex cursor-pointer select-none items-center gap-0.5 border-none bg-transparent p-0 text-xs font-semibold uppercase tracking-[0.03em] text-muted-foreground outline-none transition-colors hover:text-foreground focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary';

export type SortValue = string | number | null | undefined;

function compare(a: SortValue, b: SortValue): number {
	const av = a ?? '';
	const bv = b ?? '';
	if (typeof av === 'number' && typeof bv === 'number') return av - bv;
	return String(av).localeCompare(String(bv), undefined, { numeric: true, sensitivity: 'base' });
}

// Stable client-side sort of the loaded rows. Missing values are treated
// as empty strings, so a blank device (Global) sorts first ascending —
// the same position it holds in the API's default order.
export function sortRows<T>(rows: T[], keyFn: (row: T) => SortValue, dir: SortDir): T[] {
	const factor = dir === 'asc' ? 1 : -1;
	return [...rows].sort((a, b) => factor * compare(keyFn(a), keyFn(b)));
}
