<script lang="ts">
	import type { HTMLTextareaAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Textarea wrapper (shadcn-svelte pattern, compact density): replaces
	// the bare global textarea styling for migrated pages (ticket 01).
	let { value = $bindable(''), class: className, ...rest }: HTMLTextareaAttributes = $props();

	const base =
		'min-h-16 w-full rounded-[4px] border border-border bg-card px-2.5 py-1.5 text-sm text-foreground transition-colors placeholder:text-muted-foreground focus-visible:border-primary focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring/40 disabled:cursor-default disabled:opacity-50';
</script>

<textarea bind:value class={cn(base, className)} {...rest}></textarea>
