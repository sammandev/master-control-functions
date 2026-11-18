<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLTdAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Table data cell: compact padding. Caller classes (text-right,
	// text-muted-foreground, banner-empty + text-center for empty states) merge over
	// the defaults.
	let { class: className, children, ...rest }: HTMLTdAttributes & { children?: Snippet } = $props();
</script>

<td class={cn('px-3 py-2 text-left align-middle', className)} {...rest}>
	{@render children?.()}
</td>
