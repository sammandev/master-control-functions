<script lang="ts">
	import { Pagination } from 'bits-ui';
	import Select from './select.svelte';
	import { cn } from '#lib/utils.ts';

	// Pagination wrapper (shadcn-svelte pattern, compact density): one
	// pager for every list surface (ticket 06) — page indicator, previous/
	// next buttons over bits-ui's keyboard-operable primitive, and the
	// unified page-size select. Pages keep owning their list state; the
	// wrapper only presents it. A page-size change resets to page 1, the
	// shared behavior the list pages previously each hand-rolled.
	let {
		page = $bindable(1),
		total,
		pageSize = $bindable(20),
		pageSizeOptions = [10, 20, 50],
		onPageChange,
		onPageSizeChange,
		class: className
	}: {
		page?: number;
		/** Total item count across all pages. */
		total: number;
		pageSize?: number;
		pageSizeOptions?: number[];
		onPageChange?: (page: number) => void;
		onPageSizeChange?: (size: number) => void;
		class?: string;
	} = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / Math.max(1, pageSize))));

	const rootBase = 'mt-3 flex flex-wrap items-center justify-end gap-2';
	const navBtnBase =
		'inline-flex h-8 cursor-pointer select-none items-center justify-center gap-1 whitespace-nowrap rounded-[4px] border border-border bg-card px-3 text-sm font-medium text-foreground transition-colors hover:bg-accent focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring disabled:cursor-default disabled:opacity-50';
</script>

<Pagination.Root
	bind:page
	{onPageChange}
	count={total}
	perPage={Math.max(1, pageSize)}
	class={cn(rootBase, className)}
>
	{#snippet children({ currentPage })}
		<span class="text-muted-foreground text-xs">Page {currentPage} of {totalPages}</span>
		<Pagination.PrevButton class={navBtnBase} disabled={page <= 1} aria-label="Previous page">
			◀ Prev
		</Pagination.PrevButton>
		<Pagination.NextButton class={navBtnBase} disabled={page >= totalPages} aria-label="Next page">
			Next ▶
		</Pagination.NextButton>
		<Select
			class="w-28"
			ariaLabel="Page size"
			value={String(pageSize)}
			options={pageSizeOptions.map((n) => ({ value: String(n), label: `${n} / page` }))}
			onchange={(v) => {
				pageSize = Number(v);
				page = 1;
				onPageSizeChange?.(pageSize);
			}}
		/>
	{/snippet}
</Pagination.Root>
