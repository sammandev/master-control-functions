<script lang="ts">
	import { resolve } from '$app/paths';
	import ExamplesWorkspace from '#lib/components/ExamplesWorkspace.svelte';

	// Global Examples catalog (catalog-workspace ticket 04): the same
	// filterable list/detail workspace the Functions and dictionary pages
	// use, with search and Function/Device filters beside complete script
	// details and explicit CRUD.
</script>

<svelte:head><title>Examples | MasterControl Functions Database</title></svelte:head>

<h1 class="mt-2 mb-4 text-xl font-semibold">Examples</h1>

<p class="text-muted-foreground mb-4 text-xs">
	Script call examples per function. Examples whose function is missing are flagged as orphans and
	can also be resolved from the

	<a class="text-primary hover:underline" href={resolve('orphans')}>orphan report</a>

	.
</p>

<ExamplesWorkspace />
