<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { cn } from '#lib/utils.ts';

	// Button wrapper (shadcn-svelte pattern, compact density): variants
	// driven by the canonical app theme tokens (--primary/--destructive/...).
	type Variant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost';
	type Size = 'default' | 'sm' | 'icon';

	let {
		variant = 'default',
		size = 'default',
		type = 'button',
		class: className,
		children,
		...rest
	}: HTMLButtonAttributes & { variant?: Variant; size?: Size; children?: Snippet } = $props();

	const variants: Record<Variant, string> = {
		default: 'border border-primary bg-primary text-primary-foreground hover:bg-primary/90',
		destructive:
			'border border-destructive bg-destructive text-destructive-foreground hover:bg-destructive/90',
		outline: 'border border-border bg-card text-foreground hover:bg-accent',
		secondary: 'border border-border bg-popover text-foreground hover:bg-accent',
		ghost: 'text-foreground hover:bg-accent'
	};

	const sizes: Record<Size, string> = {
		default: 'h-8 px-3 text-sm',
		sm: 'h-7 px-2.5 text-xs',
		icon: 'size-8 p-0'
	};

	const base =
		'inline-flex cursor-pointer items-center justify-center gap-1.5 whitespace-nowrap rounded-[4px] font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring disabled:cursor-default disabled:opacity-50';
</script>

<button {type} class={cn(base, variants[variant], sizes[size], className)} {...rest}>
	{@render children?.()}
</button>
