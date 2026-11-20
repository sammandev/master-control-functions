<script lang="ts">
	import { AlertDialog } from 'bits-ui';
	import Button from './button.svelte';
	import { dialogTitleCase } from '#lib/utils.ts';
	import type { Snippet } from 'svelte';

	// Accessible destructive confirmation (ticket 02), replacing every
	// window.confirm. Controlled open like dialog.svelte. bits-ui's
	// AlertDialog.Action carries no close behavior of its own (v2.19.4),
	// so the confirm button explicitly closes the dialog and fires
	// `onConfirm` so the parent runs the destructive action only on
	// explicit confirmation.
	let {
		open,
		onOpenChange = () => {},
		title,
		description = '',
		confirmLabel = 'Confirm',
		cancelLabel = 'Cancel',
		destructive = true,
		onConfirm = () => {},
		children
	}: {
		open: boolean;
		onOpenChange?: (open: boolean) => void;
		title: string;
		description?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		destructive?: boolean;
		onConfirm?: () => void;
		children?: Snippet;
	} = $props();
</script>

<AlertDialog.Root {open} onOpenChange={(o) => onOpenChange(o)}>
	<AlertDialog.Portal>
		<AlertDialog.Overlay class="bg-overlay/45 fixed inset-0 z-50" />
		<AlertDialog.Content
			class="border-border bg-card fixed top-12 left-1/2 z-50 flex max-h-[calc(100dvh-6rem)] w-[calc(100%-2rem)] max-w-[440px] -translate-x-1/2 flex-col overflow-hidden rounded-lg border p-5 shadow-lg"
		>
			<AlertDialog.Title class="mb-2 text-base font-semibold"
				>{dialogTitleCase(title)}</AlertDialog.Title
			>
			{#if description}
				<AlertDialog.Description class="text-muted-foreground text-[13px]"
					>{description}</AlertDialog.Description
				>
			{/if}
			<div class="alert-dialog-body min-h-0 flex-1 overflow-y-auto overscroll-contain">
				{@render children?.()}
			</div>
			<div class="mt-4 flex shrink-0 justify-end gap-2">
				<AlertDialog.Cancel>
					{#snippet child({ props })}
						<Button {...props} variant="outline">{cancelLabel}</Button>
					{/snippet}
				</AlertDialog.Cancel>
				<AlertDialog.Action>
					{#snippet child({ props })}
						<Button
							{...props}
							variant={destructive ? 'destructive' : 'default'}
							onclick={() => {
								onConfirm();
								onOpenChange(false);
							}}
						>
							{confirmLabel}
						</Button>
					{/snippet}
				</AlertDialog.Action>
			</div>
		</AlertDialog.Content>
	</AlertDialog.Portal>
</AlertDialog.Root>
