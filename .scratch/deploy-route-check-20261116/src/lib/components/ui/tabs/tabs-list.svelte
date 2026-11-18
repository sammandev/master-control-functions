<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tabs as TabsPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	let {
		class: className,
		children,
		...rest
	}: { class?: string; children?: Snippet } & TabsPrimitive.ListProps = $props();
</script>

<TabsPrimitive.List {...rest} class={cn('border-border flex gap-1 border-b', className)}>
	{@render children?.()}
</TabsPrimitive.List>
