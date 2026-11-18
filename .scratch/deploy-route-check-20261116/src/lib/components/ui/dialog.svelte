<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { cn } from '#lib/utils.ts';
	import type { Snippet } from 'svelte';

	// Shared modal (ticket 02): portal overlay over bits-ui Dialog. Escape
	// closes, focus is trapped while open and restored to the trigger on
	// close, backdrop/outside clicks close. Open state is controlled: the
	// parent passes `open` and receives every close/open request through
	// `onOpenChange`, so derived state (e.g. `resolveKind !== null`) works.
	let {
		open,
		onOpenChange = () => {},
		title,
		description = '',
		class: className = '',
		children
	}: {
		open: boolean;
		onOpenChange?: (open: boolean) => void;
		title: string;
		description?: string;
		class?: string;
		children?: Snippet;
	} = $props();
</script>

<Dialog.Root {open} onOpenChange={(o) => onOpenChange(o)}>
	<Dialog.Portal>
		<Dialog.Overlay class="bg-overlay/45 fixed inset-0 z-50" />
		<Dialog.Content
			class={cn(
				'border-border bg-card fixed top-12 left-1/2 z-50 flex max-h-[calc(100vh-6rem)] w-[calc(100%-2rem)] max-w-[520px] -translate-x-1/2 flex-col overflow-auto rounded-lg border p-5 shadow-lg',
				className
			)}
		>
			<Dialog.Title class="mb-3 text-base font-semibold">{title}</Dialog.Title>
			{#if description}
				<Dialog.Description class="text-muted-foreground mb-3 text-[13px]"
					>{description}</Dialog.Description
				>
			{/if}
			{@render children?.()}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
