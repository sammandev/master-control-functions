<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLThAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Table header cell: the uppercase muted header language. Extra
	// attributes (e.g. aria-sort on sortable columns) pass through.
	let { class: className, children, ...rest }: HTMLThAttributes & { children?: Snippet } = $props();
</script>

<th
	class={cn(
		'bg-popover text-muted-foreground h-9 px-3 text-left align-middle text-xs font-semibold tracking-[0.03em] uppercase',
		className
	)}
	{...rest}
>
	{@render children?.()}
</th>
