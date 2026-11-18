<script lang="ts">
	import type { SortDir } from '#lib/table-sort.ts';
	import { SORT_BUTTON_CLASS, ariaSortValue, sortArrow } from '#lib/table-sort.ts';
	import TableHead from '#lib/components/ui/table/table-head.svelte';

	// Sortable table header (ticket 09): one shared presentation for every
	// list table — a header button with a direction arrow and aria-sort on
	// the th. The sort state itself lives in the page (server-side sort
	// params or client-side sortRows); onclick flips it there.
	let {
		label,
		col,
		active,
		dir,
		onclick
	}: {
		label: string;
		col: string;
		active: string;
		dir: SortDir;
		onclick: () => void;
	} = $props();
</script>

<TableHead aria-sort={ariaSortValue(col, active, dir)}>
	<button type="button" class={SORT_BUTTON_CLASS} {onclick}>
		{label}
		{sortArrow(col, active, dir)}
	</button>
</TableHead>
