<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tabs as TabsPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	let {
		class: className,
		children,
		...rest
	}: { class?: string; children?: Snippet } & TabsPrimitive.TriggerProps = $props();
</script>

<!-- Trigger renders a <button role="tab">: bits-ui supplies aria-selected,
		roving focus and ArrowLeft/Right/Up/Down activation. -->
<TabsPrimitive.Trigger
	{...rest}
	class={cn(
		'text-muted-foreground hover:text-foreground border-b-2 border-transparent px-3.5 py-2 text-sm',
		'data-[state=active]:border-primary data-[state=active]:text-primary data-[state=active]:font-semibold',
		className
	)}
>
	{@render children?.()}
</TabsPrimitive.Trigger>
