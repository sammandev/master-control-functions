<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Table row: bottom border + hover tint, matching the legacy global
	// tr:hover look without scoped styles.
	let {
		class: className,
		children,
		...rest
	}: HTMLAttributes<HTMLTableRowElement> & { children?: Snippet } = $props();
</script>

<tr class={cn('border-border hover:bg-accent border-b transition-colors', className)} {...rest}>
	{@render children?.()}
</tr>
