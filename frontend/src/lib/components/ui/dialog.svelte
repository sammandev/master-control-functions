<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { cn, dialogTitleCase } from '#lib/utils.ts';
	import type { Snippet } from 'svelte';
	import { X } from '@lucide/svelte';

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
				'border-border bg-card dialog-shell fixed top-4 left-1/2 z-50 flex max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-[640px] -translate-x-1/2 flex-col overflow-hidden rounded-lg border p-5 shadow-lg',
				className
			)}
		>
			<div class="dialog-heading">
				<Dialog.Title class="pr-10 text-base font-semibold">{dialogTitleCase(title)}</Dialog.Title>
				{#if description}<Dialog.Description class="text-muted-foreground mt-1 text-[13px]"
						>{description}</Dialog.Description
					>{/if}
				<Dialog.Close
					aria-label="Close dialog"
					title="Close dialog"
					class="hover:bg-accent focus-visible:outline-ring absolute top-0 right-0 inline-grid size-8 place-items-center rounded-[4px] focus-visible:outline-2"
					><X size={16} /></Dialog.Close
				>
			</div>
			<div class="dialog-body min-h-0 flex-1 overflow-y-auto overscroll-contain">
				{@render children?.()}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	.dialog-heading {
		position: relative;
		flex-shrink: 0;
		padding-bottom: 16px;
		background: var(--card);
	}
	:global(.dialog-shell .crud-form) {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 16px;
	}
	:global(.dialog-shell .crud-form .field) {
		display: grid;
		align-content: start;
		gap: 6px;
		min-width: 0;
		margin: 0;
	}
	:global(.dialog-shell .crud-form .field-wide),
	:global(.dialog-shell .dialog-actions) {
		grid-column: 1 / -1;
	}
	:global(.dialog-shell .dialog-actions) {
		position: sticky;
		bottom: 0;
		z-index: 1;
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		margin-top: 4px;
		border-top: 1px solid var(--border);
		padding: 16px 0 4px;
		background: var(--card);
	}
	@media (max-width: 600px) {
		:global(.dialog-shell .crud-form) {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
