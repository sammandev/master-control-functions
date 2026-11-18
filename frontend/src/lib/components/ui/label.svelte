<script lang="ts">
	import { Label as LabelPrimitive } from 'bits-ui';
	import { cn } from '#lib/utils.ts';

	// Label wrapper over the bits-ui Label primitive (accessible name
	// association via the `for` attribute passthrough).
	let { ref = $bindable(null), class: className, ...rest }: LabelPrimitive.RootProps = $props();
</script>

<LabelPrimitive.Root
	bind:ref
	class={cn('text-muted-foreground text-xs leading-none font-semibold select-none', className)}
	{...rest}
/>
