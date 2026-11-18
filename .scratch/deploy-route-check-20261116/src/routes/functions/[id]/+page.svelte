<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { getFunction, ApiError } from '#lib/api.ts';
	import FunctionDetail from '#lib/components/FunctionDetail.svelte';
	import FunctionDialog from '#lib/components/FunctionDialog.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Breadcrumb from '#lib/components/ui/breadcrumb.svelte';
	import { RefreshCw } from '@lucide/svelte';
	const fnId = $derived(Number(page.params.id));

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
		/>
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
