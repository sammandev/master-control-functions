<script lang="ts">
	import { Pagination } from 'bits-ui';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
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
		pageSizeLabel = 'Page size',
		variant = 'default',
		class: className
	}: {
		page?: number;
		/** Total item count across all pages. */
		total: number;
		pageSize?: number;
		pageSizeOptions?: number[];
		onPageChange?: (page: number) => void;
		onPageSizeChange?: (size: number) => void;
		pageSizeLabel?: string;
		/** `rail` is the compact sticky footer used by list/detail sidebars. */
		variant?: 'default' | 'rail';
		class?: string;
	} = $props();

	const totalPages = $derived(Math.max(1, Math.ceil(total / Math.max(1, pageSize))));
	const rail = $derived(variant === 'rail');

	const rootBase = 'mt-3 flex flex-wrap items-center justify-end gap-2';
	const railBase =
		'bg-card border-border text-muted-foreground sticky bottom-0 z-[1] mt-auto flex shrink-0 items-center gap-1 border-t px-2 py-2.5 text-xs';
	const navBtnBase =
		'inline-flex h-8 cursor-pointer select-none items-center justify-center gap-1 whitespace-nowrap rounded-[4px] border border-border bg-card px-3 text-sm font-medium text-foreground transition-colors hover:bg-accent focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring disabled:cursor-default disabled:opacity-50';
	const railBtnBase =
		'inline-grid size-8 shrink-0 cursor-pointer place-items-center rounded-[4px] text-foreground transition-colors hover:bg-accent focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring disabled:cursor-default disabled:opacity-40';
	const sizeOptions = $derived(
		pageSizeOptions.map((n) => ({ value: String(n), label: `${n} / page` }))
	);

	function changeSize(value: string) {
		pageSize = Number(value);
		page = 1;
		onPageSizeChange?.(pageSize);
	}
</script>

<Pagination.Root
	bind:page
	{onPageChange}
	count={total}
	perPage={Math.max(1, pageSize)}
	class={cn(rail ? railBase : rootBase, className)}
>
	{#snippet children({ currentPage })}
		{#if rail}
			<span class="mr-auto flex shrink-0 items-center gap-1.5">
				<span aria-hidden="true">Rows</span>
				<Select
					class="w-28"
					ariaLabel={pageSizeLabel}
					value={String(pageSize)}
					options={sizeOptions}
					onchange={changeSize}
				/>
			</span>
			<span class="sr-only">Page {currentPage} of {totalPages}</span>
			<span class="shrink-0 whitespace-nowrap tabular-nums" aria-hidden="true"
				>{currentPage} / {totalPages}</span
			>
			<Pagination.PrevButton class={railBtnBase} disabled={page <= 1} aria-label="Previous page">
				<ChevronLeft size={16} aria-hidden="true" />
			</Pagination.PrevButton>
			<Pagination.NextButton
				class={railBtnBase}
				disabled={page >= totalPages}
				aria-label="Next page"
			>
				<ChevronRight size={16} aria-hidden="true" />
			</Pagination.NextButton>
		{:else}
			<span class="text-muted-foreground text-xs">Page {currentPage} of {totalPages}</span>
			<Pagination.PrevButton class={navBtnBase} disabled={page <= 1} aria-label="Previous page">
				<ChevronLeft size={14} aria-hidden="true" /> Prev
			</Pagination.PrevButton>
			<Pagination.NextButton
				class={navBtnBase}
				disabled={page >= totalPages}
				aria-label="Next page"
			>
				Next <ChevronRight size={14} aria-hidden="true" />
			</Pagination.NextButton>
			<Select
				class="w-28"
				ariaLabel={pageSizeLabel}
				value={String(pageSize)}
				options={sizeOptions}
				onchange={changeSize}
			/>
		{/if}
	{/snippet}
</Pagination.Root>
