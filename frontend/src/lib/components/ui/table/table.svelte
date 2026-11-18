<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLTableAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Table wrapper (shadcn-svelte pattern, compact density): shared table
	// chrome for every list surface (ticket 06). The bordered, data-dense
	// look of the legacy global table styles, expressed as utilities so
	// migrated pages drop their scoped table styles.
	let {
		class: className,
		children,
		...rest
	}: HTMLTableAttributes & { children?: Snippet } = $props();
</script>

<table
	class={cn('border-border bg-card w-full border-collapse border text-sm', className)}
	{...rest}
>
	{@render children?.()}
</table>
