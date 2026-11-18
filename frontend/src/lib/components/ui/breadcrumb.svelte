<script lang="ts">
	import { resolve } from '$app/paths';
	import { cn } from '#lib/utils.ts';
	type ResolvedHref = ReturnType<typeof resolve>;

	let {
		items,
		class: className
	}: { items: { label: string; href?: ResolvedHref }[]; class?: string } = $props();
</script>

<nav
	aria-label="Breadcrumb"
	class={cn('text-muted-foreground flex items-center gap-1.5 text-[13px]', className)}
>
	{#each items as item, i (i)}
		{#if i > 0}<span aria-hidden="true">/</span>{/if}
		{#if i < items.length - 1 && item.href}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={item.href} class="hover:text-foreground">{item.label}</a>
		{:else if i === items.length - 1}
			<span aria-current="page" class="text-foreground">{item.label}</span>
		{:else}
			<span class="text-foreground">{item.label}</span>
		{/if}
	{/each}
</nav>
