<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';
	import { getFunction, listFunctions, ApiError, type FunctionRow } from '#lib/api.ts';
	import Select from '#lib/components/ui/select.svelte';
	import FunctionDetail from '#lib/components/FunctionDetail.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Breadcrumb from '#lib/components/ui/breadcrumb.svelte';
	import { RefreshCw } from '@lucide/svelte';
	const fnId = $derived(Number(page.params.id));
	let functions = $state<FunctionRow[]>([]);
	let switchError = $state('');
	let switchingOptions = $state(false);
	async function loadOptions() {
		if (switchingOptions) return;
		switchingOptions = true;
		switchError = '';
		try {
			const result = await listFunctions({ page: 1, page_size: 100, sort: 'name', order: 'asc' });
			const rows = [...result.items];
			for (let next = 2; rows.length < result.total; next++) {
				const more = await listFunctions({
					page: next,
					page_size: 100,
					sort: 'name',
					order: 'asc'
				});
				if (!more.items.length) throw new Error('The function list changed. Please retry.');
				rows.push(...more.items);
			}
			functions = rows;
		} catch (error) {
			switchError = error instanceof Error ? error.message : String(error);
		} finally {
			switchingOptions = false;
		}
	}
	onMount(() => {
		void loadOptions();
	});

	let retry = $state(0);
	const content = $derived.by(() => {
		void retry;
		return getFunction(fnId);
	});
	let dialogMode: 'edit' | 'delete' | null = $derived.by(() => {
		void fnId;
		return null;
	});
</script>

{#await content}
	<p class="muted" role="status">Loading function...</p>
{:then fn}
	<Breadcrumb items={[{ label: 'Functions', href: resolve('functions') }, { label: fn.name }]} />

	{#key fn}
		<FunctionDetail
			{fn}
			standalone={true}
			onEdit={() => (dialogMode = 'edit')}
			onDelete={() => (dialogMode = 'delete')}
		>
			{#snippet heading()}{@render functionHeading(fn)}{/snippet}
		</FunctionDetail>
	{/key}

	{#if dialogMode}
		<FunctionDialog
			mode={dialogMode}
			{fn}
			onClose={() => (dialogMode = null)}
			onSaved={async (row) => {
				dialogMode = null;

				if (row) retry++;
				else await goto(resolve('functions'));
			}}
		/>{/if}
{:catch error}
	<div class="banner-error" role="alert">
		{error instanceof ApiError && error.status === 404
			? 'This function does not exist.'
			: error instanceof Error
				? error.message
				: String(error)}<Button variant="outline" onclick={() => retry++}
			><RefreshCw size={16} /> Retry function</Button
		>
	</div>
	<a href={resolve('functions')}>Back to functions</a>
{/await}

{#snippet functionHeading(fn: FunctionRow)}
	<h1 class="function-heading">
		<Select
			heading
			ariaLabel="Switch function"
			value={String(fn.id)}
			displayValue={fn.name}
			options={(functions.some((row) => row.id === fn.id) ? functions : [fn, ...functions]).map(
				(row) => ({
					value: String(row.id),
					label: `${row.name} (${row.source} / ${row.interface || 'Global'})`
				})
			)}
			onchange={(id) => {
				void goto(resolve('/functions/[id]', { id }));
			}}
		/>
	</h1>
	{#if switchError}<div role="alert" class="banner-error">
			{switchError}<Button variant="outline" onclick={loadOptions}
				><RefreshCw data-icon="inline-start" /> Retry functions</Button
			>
		</div>{/if}
{/snippet}
