<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Badge wrapper (shadcn-svelte pattern): the "warn" variant replaces
	// the old global .badge.warn orphan marker (ticket 01).
	type Variant = 'default' | 'warn' | 'destructive' | 'outline';

	let {
		variant = 'default',
		class: className,
		children,
		...rest
	}: HTMLAttributes<HTMLSpanElement> & { variant?: Variant; children?: Snippet } = $props();

	const variants: Record<Variant, string> = {
		default: 'border border-primary/20 bg-primary/10 text-primary',
		warn: 'border border-warn/40 bg-warn/10 text-warn',
		destructive: 'border border-destructive bg-destructive text-destructive-foreground',
		outline: 'border border-border text-foreground'
	};

	const base =
		'inline-flex items-center whitespace-nowrap rounded-full border px-2 py-0.5 text-[11px] font-semibold';
</script>

<span class={cn(base, variants[variant], className)} {...rest}>
	{@render children?.()}
</span>
