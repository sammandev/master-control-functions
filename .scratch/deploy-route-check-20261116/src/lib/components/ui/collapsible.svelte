<script lang="ts">
	import { Collapsible } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { cn } from '#lib/utils.ts';

	// Collapsible wrapper (shadcn-svelte pattern, compact density): one
	// progressive-disclosure control for the composer's detail groups and
	// the import history rows (ticket 07). Trigger and content are owned
	// here so open/close styling, the rotating chevron and keyboard
	// operation behave identically at every call site; `open` stays
	// bindable so pages can drive or observe the state.
	let {
		open = $bindable(false),
		onOpenChange,
		title,
		children,
		class: className
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		title: Snippet;
		children?: Snippet;
		class?: string;
	} = $props();

	const rootBase =
		'rounded-[6px] border border-border bg-card transition-colors data-[state=open]:bg-popover';
	const triggerBase =
		'group flex w-full cursor-pointer select-none items-center justify-between gap-2 rounded-[6px] px-3 py-2 text-left text-sm font-medium text-foreground outline-none focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring';
	const contentBase = 'px-3 pb-3';
</script>

<Collapsible.Root bind:open {onOpenChange} class={cn(rootBase, className)}>
	<Collapsible.Trigger class={triggerBase}>
		{@render title()}
		<svg
			aria-hidden="true"
			xmlns="http://www.w3.org/2000/svg"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			class="text-muted-foreground size-4 shrink-0 transition-transform group-data-[state=open]:rotate-180"
		>
			<path d="m6 9 6 6 6-6" />
		</svg>
	</Collapsible.Trigger>
	<Collapsible.Content class={contentBase}>
		{@render children?.()}
	</Collapsible.Content>
</Collapsible.Root>
