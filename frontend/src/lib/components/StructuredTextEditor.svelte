<script lang="ts">
	import type { HTMLTextareaAttributes } from 'svelte/elements';
	import { getListEdit } from '#lib/structured-text.ts';
	import Textarea from '#lib/components/ui/textarea.svelte';

	let {
		value = $bindable(''),
		onkeydown,
		...rest
	}: Omit<HTMLTextareaAttributes, 'value'> & { value?: string } = $props();

	const handleKeydown: NonNullable<HTMLTextareaAttributes['onkeydown']> = (event) => {
		onkeydown?.(event);
		const textarea = event.currentTarget;
		if (
			event.defaultPrevented ||
			event.isComposing ||
			event.altKey ||
			event.ctrlKey ||
			event.metaKey ||
			event.shiftKey ||
			(event.key !== 'Enter' && event.key !== 'Backspace') ||
			textarea.readOnly ||
			textarea.selectionStart !== textarea.selectionEnd
		)
			return;

		const edit = getListEdit(textarea.value, textarea.selectionStart, event.key);
		if (!edit) return;
		event.preventDefault();
		textarea.setRangeText(edit.text, edit.start, edit.end, 'end');
		value = textarea.value;
	};
</script>

<Textarea {...rest} bind:value onkeydown={handleKeydown} />
