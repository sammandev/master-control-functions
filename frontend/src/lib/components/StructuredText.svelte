<script lang="ts">
	import { cn } from '#lib/utils.ts';
	import { parseStructuredText } from '#lib/structured-text.ts';

	let { text, class: className }: { text: string; class?: string } = $props();
	const blocks = $derived(parseStructuredText(text));
</script>

<div class={cn('structured-text text-foreground grid gap-2 text-sm leading-relaxed', className)}>
	{#each blocks as block, blockIndex (blockIndex)}
		{#if block.kind === 'paragraph'}
			<p class="m-0 whitespace-pre-line">{block.text}</p>
		{:else if block.kind === 'unordered-list'}
			<ul class="m-0 grid list-disc gap-1 pl-5">
				{#each block.items as item, itemIndex (itemIndex)}<li>{item}</li>{/each}
			</ul>
		{:else}
			<ol class="m-0 grid list-decimal gap-1 pl-5">
				{#each block.items as item, itemIndex (itemIndex)}<li>{item}</li>{/each}
			</ol>
		{/if}
	{/each}
</div>
