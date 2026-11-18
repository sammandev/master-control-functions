<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tooltip as TooltipPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	// Composite tooltip: the caller's snippet IS the trigger — it receives the
	// trigger props (aria-describedby, data-state, pointer/focus handlers) and
	// spreads them onto its own element (a Button, a span badge, …) so we never
	// nest a second button. Opens on hover AND focus (keyboard reachable).
	let {
		content,
		class: className,
		children,
		delayDuration = 300
	}: {
		content: string;
		class?: string;
		children?: Snippet<[Record<string, unknown>]>;
		delayDuration?: number;
	} = $props();
</script>

<!-- Provider scopes delayDuration; without it Tooltip.Root throws
		"Context Tooltip.Provider not found" at render time. -->
<TooltipPrimitive.Provider {delayDuration}>
	<TooltipPrimitive.Root>
		<TooltipPrimitive.Trigger>
			{#snippet child(tp)}
				{@render children?.(tp.props)}
			{/snippet}
		</TooltipPrimitive.Trigger>
		<TooltipPrimitive.Portal>
			<TooltipPrimitive.Content
				sideOffset={4}
				role="tooltip"
				class={cn(
					'border-border bg-card text-foreground z-50 overflow-hidden rounded-md border px-3 py-1.5 text-xs shadow-md',
					className
				)}
			>
				{content}
			</TooltipPrimitive.Content>
		</TooltipPrimitive.Portal>
	</TooltipPrimitive.Root>
</TooltipPrimitive.Provider>
