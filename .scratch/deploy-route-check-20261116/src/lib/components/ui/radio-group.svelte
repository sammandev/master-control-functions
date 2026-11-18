<script lang="ts">
	import { RadioGroup as RadioGroupPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	// RadioGroup root wrapper over the bits-ui RadioGroup primitive
	// (roving-focus arrow-key navigation for free). Item is in
	// radio-group-item.svelte (ticket 01).
	let {
		ref = $bindable(null),
		value = $bindable(''),
		class: className,
		...rest
	}: RadioGroupPrimitive.RootProps = $props();
</script>

<RadioGroupPrimitive.Root bind:ref bind:value class={cn('grid gap-2', className)} {...rest} />
