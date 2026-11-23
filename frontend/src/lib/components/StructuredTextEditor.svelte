<script lang="ts">
	import { tick } from 'svelte';
	import type { HTMLTextareaAttributes } from 'svelte/elements';
	import { List, ListOrdered } from '@lucide/svelte';
	import { applyListMarkup } from '#lib/structured-text.ts';
	import Button from '#lib/components/ui/button.svelte';
	import Textarea from '#lib/components/ui/textarea.svelte';
	import Tooltip from '#lib/components/ui/tooltip.svelte';

	let {
		value = $bindable(''),
		id,
		class: className,
		...rest
	}: Omit<HTMLTextareaAttributes, 'value'> & { value?: string } = $props();

	async function applyMarkup(kind: 'bullet' | 'ordered') {
		const textarea = document.getElementById(id ?? '') as HTMLTextAreaElement | null;
		if (!textarea) return;
		const result = applyListMarkup(value, textarea.selectionStart, textarea.selectionEnd, kind);
		value = result.value;
		await tick();
		textarea.focus();
		textarea.setSelectionRange(result.selectionStart, result.selectionEnd);
	}
</script>

<div class="structured-text-editor grid gap-1.5">
	<div
		class="border-border bg-muted/35 flex items-center justify-end gap-1 rounded-[4px] border px-1 py-1"
	>
		<Tooltip content="Bulleted list">
			{#snippet children(props)}
				<Button
					{...props}
					type="button"
					variant="ghost"
					size="icon"
					class="size-7"
					aria-label="Bulleted list"
					onclick={() => applyMarkup('bullet')}
				>
					<List data-icon="inline-start" />
				</Button>
			{/snippet}
		</Tooltip>
		<Tooltip content="Numbered list">
			{#snippet children(props)}
				<Button
					{...props}
					type="button"
					variant="ghost"
					size="icon"
					class="size-7"
					aria-label="Numbered list"
					onclick={() => applyMarkup('ordered')}
				>
					<ListOrdered data-icon="inline-start" />
				</Button>
			{/snippet}
		</Tooltip>
	</div>
	<Textarea {id} bind:value class={className} {...rest} />
</div>
