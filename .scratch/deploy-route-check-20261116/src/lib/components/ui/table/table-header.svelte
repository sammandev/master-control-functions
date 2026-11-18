<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Table header row group; row borders come from TableRow.
	let {
		class: className,
		children,
		...rest
	}: HTMLAttributes<HTMLTableSectionElement> & { children?: Snippet } = $props();
</script>

<thead class={cn(className)} {...rest}>
	{@render children?.()}
</thead>
