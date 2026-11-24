<script lang="ts">
	import { Combobox, Popover } from 'bits-ui';
	import { Check, ChevronDown, Globe, Search, X } from '@lucide/svelte';
	import { devicePath, listDevices, type DeviceSummary } from '#lib/devices.ts';
	import { cn } from '#lib/utils.ts';

	let {
		device,
		summaries,
		id,
		label = 'Device',
		blankLabel = 'Global definitions',
		allowAll = false,
		disabled = false,
		invalid = false,
		describedBy,
		fullWidth = false,
		compact = false,
		comfortable = false,
		heading = false,
		onValueChange
	}: {
		device: string;
		summaries?: DeviceSummary[];
		id?: string;
		label?: string;
		blankLabel?: string;
		allowAll?: boolean;
		disabled?: boolean;
		invalid?: boolean;
		describedBy?: string;
		fullWidth?: boolean;
		compact?: boolean;
		comfortable?: boolean;
		heading?: boolean;
		onValueChange: (device: string) => void;
	} = $props();

	let open = $state(false);
	let term = $state('');
	let input = $state<HTMLInputElement | null>(null);
	let loaded = $state<DeviceSummary[]>([]);
	let loading = $state(false);
	let error = $state('');
	const currentLabel = $derived(
		allowAll && device === '_all' ? 'All Devices' : device || blankLabel
	);
	const names = $derived([
		'',
		...new Set(
			(summaries?.length ? summaries : loaded).map((item) => item.device_name).filter(Boolean)
		)
	]);
	const options = $derived(
		[
			...(allowAll ? [{ value: '_all', label: 'All Devices', device: '_all' }] : []),
			...names.map((name) => ({
				value: devicePath(name),
				label: name || blankLabel,
				device: name
			}))
		].filter((item) => item.label.toLowerCase().includes(term.trim().toLowerCase()))
	);
	async function loadOptions() {
		if (summaries?.length || loading) return;
		loading = true;
		error = '';
		try {
			loaded = (await listDevices()).items;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : String(cause);
		} finally {
			loading = false;
		}
	}
</script>

<svelte:element
	this={heading ? 'h1' : 'div'}
	aria-label={heading ? currentLabel : undefined}
	class={heading
		? 'device-picker-heading'
		: fullWidth
			? 'grid w-full min-w-0 gap-1.5'
			: compact
				? comfortable
					? 'grid w-44 max-w-full min-w-0'
					: 'grid w-36 max-w-full min-w-0'
				: 'grid w-[280px] max-w-full gap-1.5'}
>
	<Popover.Root
		bind:open
		onOpenChange={(next) => {
			term = '';
			if (next) void loadOptions();
		}}
	>
		<Popover.Trigger
			{id}
			{disabled}
			aria-haspopup="dialog"
			aria-expanded={open}
			aria-label={`${label}: ${currentLabel}`}
			aria-invalid={invalid || undefined}
			aria-describedby={describedBy}
			title={currentLabel}
			class={cn(
				'border-border bg-card text-foreground flex w-full min-w-0 items-center justify-between rounded-[4px] border text-sm',
				compact
					? 'focus-visible:border-primary focus-visible:ring-ring/40 h-8 gap-1.5 px-2.5 transition-colors outline-none focus-visible:ring-1 disabled:cursor-default disabled:opacity-50'
					: 'focus-visible:outline-ring min-h-10 gap-2 px-3 py-2 focus-visible:outline-2 focus-visible:outline-offset-2',
				heading && 'device-picker-heading-trigger',
				compact && comfortable && 'h-10'
			)}
		>
			<span class={compact ? 'min-w-0 truncate text-left' : 'text-left font-mono break-all'}
				>{currentLabel}</span
			>
			<ChevronDown size={compact ? 14 : 16} class={cn('shrink-0', compact && 'opacity-50')} />
		</Popover.Trigger>
		<Popover.Portal>
			<Popover.Content
				role="dialog"
				aria-label={`Choose ${label}`}
				sideOffset={6}
				align="end"
				class="border-border bg-card text-foreground z-50 w-[360px] max-w-[calc(100vw-32px)] rounded-[4px] border p-2 shadow-md"
				onOpenAutoFocus={(event) => {
					event.preventDefault();
					input?.focus({ preventScroll: true });
				}}
			>
				<Combobox.Root
					type="single"
					value={devicePath(device)}
					inputValue={term}
					open={true}
					allowDeselect={false}
					onOpenChange={(next) => {
						if (!next) open = false;
					}}
					onValueChange={(value) => {
						const selected = options.find((item) => item.value === value);
						if (selected) onValueChange(selected.device);
						open = false;
					}}
				>
					<div class="border-border relative mb-2 border-b pb-2">
						<Search size={16} class="text-muted-foreground absolute top-2.5 left-2.5" />
						<Combobox.Input
							bind:ref={input}
							aria-label="Search devices"
							placeholder="Search devices"
							autocomplete="off"
							class="border-border bg-background placeholder:text-muted-foreground focus-visible:outline-ring h-9 w-full rounded-[4px] border pr-9 pl-9 text-sm focus-visible:outline-2"
							oninput={(event) => (term = event.currentTarget.value)}
						/>
						{#if term}
							<button
								type="button"
								aria-label="Clear device search"
								title="Clear device search"
								class="text-muted-foreground focus-visible:outline-ring absolute top-0.5 right-0.5 grid size-8 place-items-center rounded-sm focus-visible:outline-2"
								onclick={() => {
									term = '';
									input?.focus();
								}}
							>
								<X size={16} />
							</button>
						{/if}
					</div>
					{#if loading}<p role="status" class="text-muted-foreground px-2 py-3 text-sm">
							Loading devices...
						</p>{/if}
					{#if error}
						<div role="alert" class="text-destructive px-2 py-3 text-sm">
							{error}
							<button type="button" class="ml-2 underline" onclick={loadOptions}>Retry</button>
						</div>
					{/if}
					<Combobox.ContentStatic>
						<Combobox.Viewport class="max-h-64 overflow-y-auto overscroll-contain">
							{#each options as option (option.value)}
								<Combobox.Item
									value={option.value}
									label={option.label}
									class="data-[highlighted]:bg-accent relative flex min-h-9 cursor-pointer items-center justify-between gap-2 rounded-sm px-2 py-2 text-sm outline-none data-[selected]:font-semibold"
								>
									{#snippet children({ selected })}
										<span class="flex min-w-0 items-center gap-2">
											{#if option.device === ''}<Globe size={16} class="shrink-0" />{/if}
											<span class="break-all">{option.label}</span>
										</span>
										{#if selected}<Check size={16} class="shrink-0" />{/if}
									{/snippet}
								</Combobox.Item>
							{:else}
								<p role="status" class="text-muted-foreground px-2 py-3 text-sm">
									No devices match your search.
								</p>
							{/each}
						</Combobox.Viewport>
					</Combobox.ContentStatic>
				</Combobox.Root>
			</Popover.Content>
		</Popover.Portal>
	</Popover.Root>
</svelte:element>

<style>
	.device-picker-heading {
		width: max-content;
		max-width: 100%;
		margin: 0;
		font-family: var(--mono);
		font-size: 28px;
		font-weight: 650;
		line-height: 1.3;
		overflow-wrap: anywhere;
	}
	:global(.device-picker-heading-trigger) {
		min-height: 44px;
		padding: 0;
		border: 0;
		border-radius: 0;
		background: transparent;
		color: inherit;
		font: inherit;
		text-align: left;
	}
	:global(.device-picker-heading-trigger:hover) {
		background: transparent;
		text-decoration: underline;
		text-decoration-color: var(--border);
		text-underline-offset: 5px;
	}
	:global(.device-picker-heading-trigger:focus-visible) {
		outline: 2px solid var(--ring);
		outline-offset: 4px;
	}
	@media (max-width: 600px) {
		.device-picker-heading {
			font-size: 23px;
		}
	}
</style>
