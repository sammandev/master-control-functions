<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		getFunction,
		listFunctionOptions,
		ApiError,
		type FunctionOption,
		type FunctionRow
	} from '#lib/api.ts';
	import Select from '#lib/components/ui/select.svelte';
	import FunctionDetail from '#lib/components/FunctionDetail.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Breadcrumb from '#lib/components/ui/breadcrumb.svelte';
	import { RefreshCw } from '@lucide/svelte';
	const fnId = $derived(Number(page.params.id));
	let functions = $state.raw<FunctionOption[] | null>(null);
	let switchError = $state('');
	let optionsRequest = $state.raw<AbortController | null>(null);
	async function loadOptions() {
		if (functions || optionsRequest) return;
		const { signal } = (optionsRequest = new AbortController());
		switchError = '';
		try {
			const { items } = await listFunctionOptions(signal);
			if (!signal.aborted) functions = items;
		} catch (error) {
			if (!signal.aborted) switchError = error instanceof Error ? error.message : String(error);
		} finally {
			if (!signal.aborted) optionsRequest = null;
		}
	}

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

<svelte:head>
	{#await content}
		<title>Functions | MasterControl Functions Database</title>
	{:then fn}
		<title>{fn.name} | Functions | MasterControl Functions Database</title>
	{:catch}
		<title>Functions | MasterControl Functions Database</title>
	{/await}
</svelte:head>

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

				if (row) {
					if (row.name !== fn.name || row.source !== fn.source || row.interface !== fn.interface) {
						optionsRequest?.abort();
						optionsRequest = null;
						functions = null;
						switchError = '';
					}
					retry++;
				} else await goto(resolve('functions'));
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
	{@const rows = functions ?? []}
	<h1 class="function-heading">
		<Select
			heading
			ariaLabel="Switch function"
			value={String(fn.id)}
			displayValue={fn.name}
			loadingLabel={optionsRequest ? 'Loading functions…' : undefined}
			onopen={loadOptions}
			options={(rows.some((row) => row.id === fn.id) ? rows : [fn, ...rows]).map(
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
