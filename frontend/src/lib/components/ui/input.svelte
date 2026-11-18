<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Input wrapper (shadcn-svelte pattern, compact density): replaces the
	// bare global input element styling for migrated pages (ticket 01).
	let {
		value = $bindable(''),
		type = 'text',
		class: className,
		...rest
	}: HTMLInputAttributes = $props();

	const base =
		'h-8 w-full rounded-[4px] border border-border bg-card px-2.5 text-sm text-foreground transition-colors placeholder:text-muted-foreground focus-visible:border-primary focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring/40 disabled:cursor-default disabled:opacity-50';
</script>

<input {type} bind:value class={cn(base, className)} {...rest} />
