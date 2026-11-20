<script lang="ts">
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Dialog from '#lib/components/ui/dialog.svelte';

	let { kind }: { kind: 'dialog' | 'alert' } = $props();
</script>

{#if kind === 'dialog'}
	<Dialog open title="Example of test" description="Long form content">
		<div style="height: 4000px">Long dialog content</div>
	</Dialog>
{:else}
	<AlertDialog open title="Delete test value" description="Choose a replacement before deletion">
		<div style="height: 4000px">Long alert content</div>
	</AlertDialog>
{/if}
