<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import {
		ChevronRight,
		Download,
		Layers,
		Plus,
		RefreshCw,
		Search,
		Upload,
		X
	} from '@lucide/svelte';
	import { listDevices, type DeviceSummary } from '#lib/devices.ts';
	import { createLatestRequestGuard } from '#lib/request-guard.ts';
	import {
		categoryLabel,
		deleteTroubleshootingGuide,
		getTroubleshootingGuide,
		listRelatedTroubleshootingGuides,
		listTroubleshootingGuides,
		type TroubleshootingGuideDetail,
		type TroubleshootingGuideListItem
	} from '#lib/troubleshooting.ts';
	import {
		listTroubleshootingCategories,
		listTroubleshootingSeverities,
		type TroubleshootingCategory,
		type TroubleshootingSeverity as RegistrySeverity
	} from '#lib/troubleshooting-registry.ts';
	import DevicePicker from '#lib/components/DevicePicker.svelte';
	import TroubleshootingDetail from '#lib/components/TroubleshootingDetail.svelte';
	import TroubleshootingDialog, {
		type TroubleshootingDialogMode
	} from '#lib/components/TroubleshootingDialog.svelte';
	import TroubleshootingRegistryDialog from '#lib/components/TroubleshootingRegistryDialog.svelte';
	import TroubleshootingBulkImportDialog from '#lib/components/TroubleshootingBulkImportDialog.svelte';
	import TroubleshootingExportDialog from '#lib/components/TroubleshootingExportDialog.svelte';
	import AlertDialog from '#lib/components/ui/alert-dialog.svelte';
	import Badge from '#lib/components/ui/badge.svelte';
	import Button from '#lib/components/ui/button.svelte';
	import Input from '#lib/components/ui/input.svelte';
	import Pagination from '#lib/components/ui/pagination.svelte';
	import Select from '#lib/components/ui/select.svelte';

	let search = $state('');
	let debouncedSearch = $state('');
	let category = $state('_all');
	let device = $state('_all');
	let errorCodeFilter = $state('');
	let page = $state(1);
	let pageSize = $state(20);

	let items = $state<TroubleshootingGuideListItem[]>([]);
	let total = $state(0);
	let loadingList = $state(true);
	let listError = $state('');
	let bulkImportOpen = $state(false);
	let exportSelection = $state<{ guideId?: number }>();

	let selectedId = $state<number | null>(null);
	let selectedDetail = $state<TroubleshootingGuideDetail | null>(null);
	let loadingDetail = $state(false);
	let detailError = $state('');

	let deviceSummaries = $state<DeviceSummary[]>([]);
	let categories = $state<TroubleshootingCategory[]>([]);
	let severities = $state<RegistrySeverity[]>([]);
	let categoryError = $state('');
	let severityError = $state('');
	let relatedGuides = $state<TroubleshootingGuideListItem[]>([]);
	let relatedError = $state('');
	const listRequestGuard = createLatestRequestGuard();
	const detailRequestGuard = createLatestRequestGuard();

	let manageOpen = $state(false);
	let dialogOpen = $state(false);
	let dialogMode = $state<TroubleshootingDialogMode>('create');
	let dialogTarget = $state<TroubleshootingGuideDetail | null>(null);

	let deleteTarget = $state<TroubleshootingGuideDetail | null>(null);
	let deleteBusy = $state(false);

	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	function handleSearchInput(value: string) {
		search = value;
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			debouncedSearch = value.trim();
			page = 1;
			void refreshList();
		}, 220);
	}

	onDestroy(() => clearTimeout(searchTimer));

	async function loadDeviceOptions() {
		try {
			const res = await listDevices();
			deviceSummaries = res.items;
		} catch {
			// DevicePicker also self-loads if summaries are empty
		}
	}

	async function loadCategories(): Promise<boolean> {
		categoryError = '';
		try {
			categories = (await listTroubleshootingCategories()).items;
			return true;
		} catch (err) {
			categoryError = err instanceof Error ? err.message : String(err);
			return false;
		}
	}

	async function loadSeverities(): Promise<boolean> {
		severityError = '';
		try {
			severities = (await listTroubleshootingSeverities()).items;
			return true;
		} catch (err) {
			severityError = err instanceof Error ? err.message : String(err);
			return false;
		}
	}

	async function loadMetadataResults() {
		return Promise.all([loadCategories(), loadSeverities()]);
	}

	async function loadMetadata(): Promise<void> {
		await loadMetadataResults();
	}

	async function retryMetadata() {
		const stableSelectedId = selectedId;
		const metadataLoaded = await loadMetadataResults();
		if (!metadataLoaded.every(Boolean)) return;
		if (category !== '_all' && !categories.some((item) => item.key === category)) {
			category = '_all';
			page = 1;
		}
		await refreshList(stableSelectedId ?? undefined, true);
	}

	async function handleRegistryChanged() {
		await retryMetadata();
	}

	async function refreshList(preferId?: number, preservePreferred = false) {
		const request = listRequestGuard.begin();
		detailRequestGuard.begin();
		loadingList = true;
		loadingDetail = false;
		listError = '';
		try {
			const res = await listTroubleshootingGuides({
				search: debouncedSearch || undefined,
				category: category !== '_all' ? category : undefined,
				device: device !== '_all' ? device : undefined,
				error_code: errorCodeFilter || undefined,
				page,
				page_size: pageSize
			});
			if (!listRequestGuard.isCurrent(request)) return;
			items = res.items;
			total = res.total;

			const requestedId = preferId ?? selectedId;
			const targetId =
				(requestedId !== null &&
				(preservePreferred || res.items.some((item) => item.id === requestedId))
					? requestedId
					: res.items[0]?.id) ?? null;
			if (targetId !== null) {
				await selectGuide(targetId);
			} else {
				selectedId = null;
				selectedDetail = null;
				relatedGuides = [];
			}
		} catch (err) {
			if (listRequestGuard.isCurrent(request)) {
				listError = err instanceof Error ? err.message : String(err);
			}
		} finally {
			if (listRequestGuard.isCurrent(request)) loadingList = false;
		}
	}

	async function selectGuide(id: number) {
		const request = detailRequestGuard.begin();
		selectedId = id;
		loadingDetail = true;
		detailError = '';
		try {
			const detail = await getTroubleshootingGuide(id);
			if (!detailRequestGuard.isCurrent(request)) return;
			selectedDetail = detail;
			relatedError = '';
			try {
				const related = await listRelatedTroubleshootingGuides(id);
				if (!detailRequestGuard.isCurrent(request)) return;
				relatedGuides = related.items;
			} catch (err) {
				if (detailRequestGuard.isCurrent(request)) {
					relatedGuides = [];
					relatedError = err instanceof Error ? err.message : String(err);
				}
			}
		} catch (err) {
			if (detailRequestGuard.isCurrent(request)) {
				detailError = err instanceof Error ? err.message : String(err);
			}
		} finally {
			if (detailRequestGuard.isCurrent(request)) loadingDetail = false;
		}
	}

	function clearFilters() {
		clearTimeout(searchTimer);
		search = '';
		debouncedSearch = '';
		category = '_all';
		device = '_all';
		errorCodeFilter = '';
		page = 1;
		void refreshList();
	}

	function openCreate() {
		dialogMode = 'create';
		dialogTarget = null;
		dialogOpen = true;
	}

	function openEdit(guide: TroubleshootingGuideDetail) {
		dialogMode = 'edit';
		dialogTarget = guide;
		dialogOpen = true;
	}

	function openAddAlternate(guide: TroubleshootingGuideDetail) {
		dialogMode = 'alternate';
		dialogTarget = guide;
		dialogOpen = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		deleteBusy = true;
		try {
			await deleteTroubleshootingGuide(deleteTarget.id);
			toast.success(`Deleted guide "${deleteTarget.title}"`);
			deleteTarget = null;
			selectedId = null;
			await refreshList();
		} catch (err) {
			toast.error(err instanceof Error ? err.message : String(err));
		} finally {
			deleteBusy = false;
		}
	}

	const hasActiveFilters = $derived(
		Boolean(debouncedSearch || category !== '_all' || device !== '_all' || errorCodeFilter)
	);

	onMount(() => {
		void loadDeviceOptions();
		void loadMetadata();
		void refreshList();
	});
</script>

<div class="directory-heading">
	<div>
		<div class="directory-title">
			<h1>Troubleshooting Guides</h1>
			<span>{total} {total === 1 ? 'guide' : 'guides'}</span>
		</div>
		<p class="text-muted-foreground mt-1 text-sm">
			Step-by-step resolution guides for troubleshooting. Multiple solutions can share the same
			issue, error code, or function to cover distinct failure causes.
		</p>
	</div>
	<div class="flex flex-wrap items-center gap-2">
		<Button variant="secondary" onclick={() => (manageOpen = true)}>Manage Labels</Button>
		<Button variant="outline" onclick={() => (bulkImportOpen = true)}>
			<Upload size={16} /> Import
		</Button>
		<Button variant="outline" onclick={() => (exportSelection = {})}>
			<Download size={16} /> Export Guides
		</Button>
		<Button onclick={openCreate}>
			<Plus size={16} />
			<span>New Guide</span>
		</Button>
	</div>
</div>

<div class="filters">
	<div class="filter-field">
		<span>Testing Stage / Category</span>
		<Select
			ariaLabel={`Filter by category: ${categories.find((item) => item.key === category)?.label ?? 'All Stages'}`}
			class="registry-filter"
			value={category}
			options={[
				{ value: '_all', label: 'All Stages' },
				...categories.map((item) => ({ value: item.key, label: item.label }))
			]}
			onchange={(value) => {
				category = value;
				page = 1;
				void refreshList();
			}}
		/>
	</div>

	<div class="filter-field">
		<span>Device / Interface</span>
		<DevicePicker
			id="troubleshooting-device-filter"
			label="Device / Interface"
			compact
			comfortable
			blankLabel="Global (All Devices)"
			allowAll
			{device}
			summaries={deviceSummaries}
			fullWidth
			onValueChange={(value) => {
				device = value;
				page = 1;
				void refreshList();
			}}
		/>
	</div>

	{#if errorCodeFilter}
		<div class="filter-field">
			<span>Error Code Filter</span>
			<div class="flex h-8 items-center">
				<Badge variant="default" class="gap-1.5 px-2.5 py-1 font-mono text-xs">
					<span>{errorCodeFilter}</span>
					<button
						type="button"
						aria-label="Clear error code filter"
						class="hover:text-foreground focus-visible:ring-ring rounded-full focus-visible:ring-2 focus-visible:outline-none"
						onclick={() => {
							errorCodeFilter = '';
							page = 1;
							void refreshList();
						}}
					>
						<X size={12} />
					</button>
				</Badge>
			</div>
		</div>
	{/if}

	{#if hasActiveFilters}
		<Button variant="secondary" class="clear-filters" onclick={clearFilters}>Clear filters</Button>
	{/if}
</div>

{#if categoryError || severityError}
	<div class="metadata-error banner-error" role="alert">
		<span>
			{[
				categoryError && `Categories: ${categoryError}`,
				severityError && `Severities: ${severityError}`
			]
				.filter(Boolean)
				.join(' · ')}
		</span>
		<Button variant="outline" size="sm" onclick={retryMetadata}>
			<RefreshCw size={14} />Retry metadata
		</Button>
	</div>
{/if}

<div class="troubleshooting-workspace">
	<aside aria-label="Troubleshooting guides" class="guide-list">
		<div class="rail-controls">
			<div class="rail-heading">
				<h2>Troubleshooting Catalog</h2>
				<span>{total} total</span>
			</div>
			<div class="catalog-controls">
				<div class="guide-search">
					<Search size={16} />
					<Input
						type="search"
						placeholder="Search error code, symptom, function…"
						aria-label="Search troubleshooting guides"
						value={search}
						oninput={(e) => handleSearchInput(e.currentTarget.value)}
					/>
				</div>
			</div>
		</div>

		{#if loadingList}
			<p class="banner-empty" role="status">Loading troubleshooting guides…</p>
		{:else if listError}
			<div class="banner-error" role="alert">
				<span>{listError}</span>
				<Button variant="outline" size="sm" onclick={() => refreshList()}>
					<RefreshCw size={14} />
					<span>Retry</span>
				</Button>
			</div>
		{:else if items.length === 0}
			<div class="banner-empty" role="status">
				<p>No troubleshooting guides match the current filters.</p>
				<div class="mt-2 flex justify-center gap-2">
					{#if hasActiveFilters}
						<Button variant="secondary" size="sm" onclick={clearFilters}>Reset filters</Button>
					{/if}
					<Button size="sm" onclick={openCreate}>
						<Plus size={14} />
						<span>Create Guide</span>
					</Button>
				</div>
			</div>
		{:else}
			{#each items as row (row.id)}
				<div class="catalog-row">
					<button
						type="button"
						class={[
							'focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-none',
							{ chosen: selectedId === row.id }
						]}
						aria-pressed={selectedId === row.id}
						onclick={() => selectGuide(row.id)}
					>
						<div class="flex flex-wrap items-center gap-1.5">
							<div class="guide-labels flex flex-wrap items-center gap-1.5">
								{#if row.error_code}
									<Badge variant="default"><span class="font-mono">{row.error_code}</span></Badge>
								{/if}
								<Badge variant="outline">{row.device_name || 'Global'}</Badge>
								{#if row.function_name}
									<Badge variant="outline">
										<code class="font-mono">{row.function_name}</code>
									</Badge>
								{/if}
							</div>
							{#if row.related_count > 0}
								<span
									class="text-primary inline-flex items-center gap-1 text-[11px] font-semibold"
									title="{row.related_count + 1} solutions available for this problem / error code"
								>
									<Layers size={11} />
									+{row.related_count} alt
								</span>
							{/if}
						</div>
						<span class="sr-only">Select guide </span>
						<strong>{row.title}</strong>
						<span class="meta-line">
							{categoryLabel(row.category, categories)}
							· {row.step_count}
							{row.step_count === 1 ? 'step' : 'steps'}
						</span>
						<ChevronRight size={16} class="chevron-icon" />
					</button>
				</div>
			{/each}

			<Pagination
				variant="rail"
				pageSizeLabel={`${pageSize} / page, Page size`}
				{page}
				{total}
				{pageSize}
				onPageChange={(nextPage) => {
					page = nextPage;
					void refreshList();
				}}
				onPageSizeChange={(nextSize) => {
					pageSize = nextSize;
					page = 1;
					void refreshList();
				}}
			/>
		{/if}
	</aside>

	<div class="detail-pane">
		{#if loadingDetail}
			<p class="banner-empty" role="status">Loading step-by-step guide…</p>
		{:else if detailError}
			<div class="banner-error" role="alert">
				<span>{detailError}</span>
				{#if selectedId !== null}
					<Button variant="outline" size="sm" onclick={() => selectGuide(selectedId!)}>
						<RefreshCw size={14} />
						<span>Retry</span>
					</Button>
				{/if}
			</div>
		{:else if selectedDetail}
			{#if relatedError}
				<div class="banner-error" role="alert">
					<span>{relatedError}</span>
					<Button variant="outline" size="sm" onclick={() => selectGuide(selectedDetail!.id)}>
						Retry related guides
					</Button>
				</div>
			{/if}
			<TroubleshootingDetail
				guide={selectedDetail}
				{relatedGuides}
				{categories}
				{severities}
				onSelectGuide={(id) => selectGuide(id)}
				onFilterErrorCode={(code) => {
					errorCodeFilter = code;
					page = 1;
					void refreshList(selectedDetail?.id);
				}}
				onAddAlternate={openAddAlternate}
				onEdit={openEdit}
				onExport={(guide) => (exportSelection = { guideId: guide.id })}
				onDelete={(guide) => {
					deleteTarget = guide;
				}}
			/>
		{:else}
			<p class="banner-empty" role="status">
				Select a troubleshooting guide from the catalog or create a new one.
			</p>
		{/if}
	</div>
</div>

{#if bulkImportOpen}
	<TroubleshootingBulkImportDialog
		onClose={() => (bulkImportOpen = false)}
		onImported={({ created, updated }) => {
			toast.success(`Import complete: ${created} created, ${updated} updated`);
			void refreshList();
		}}
	/>
{/if}

{#if exportSelection}
	<TroubleshootingExportDialog
		initialGuideId={exportSelection.guideId}
		{severities}
		filters={{ search: search.trim(), category, device, error_code: errorCodeFilter }}
		onClose={() => (exportSelection = undefined)}
	/>
{/if}

{#if dialogOpen}
	<TroubleshootingDialog
		open={dialogOpen}
		mode={dialogMode}
		initialGuide={dialogTarget}
		{deviceSummaries}
		{categories}
		{severities}
		onRegistryChanged={loadMetadata}
		onOpenChange={(o) => {
			dialogOpen = o;
		}}
		onSaved={async (saved) => {
			toast.success(
				dialogMode === 'edit'
					? `Updated guide "${saved.title}"`
					: dialogMode === 'alternate'
						? `Added alternate solution "${saved.title}"`
						: `Created guide "${saved.title}"`
			);
			await refreshList(saved.id);
		}}
	/>
{/if}

<TroubleshootingRegistryDialog
	open={manageOpen}
	{categories}
	{severities}
	onOpenChange={(value) => {
		manageOpen = value;
	}}
	onChanged={handleRegistryChanged}
/>

<AlertDialog
	open={deleteTarget !== null}
	onOpenChange={(o) => {
		if (!o && !deleteBusy) deleteTarget = null;
	}}
	title="Delete Troubleshooting Guide?"
	description={deleteTarget
		? `Delete "${deleteTarget.title}" and all ${deleteTarget.steps.length} resolution steps? Sibling solutions for ${deleteTarget.error_code || 'this problem'} will remain intact.`
		: ''}
	confirmLabel={deleteBusy ? 'Deleting…' : 'Delete Guide'}
	onConfirm={confirmDelete}
/>

<style>
	.directory-heading,
	.directory-title {
		display: flex;
		align-items: center;
		gap: 16px;
	}
	.directory-heading {
		justify-content: space-between;
		flex-wrap: wrap;
		margin: 8px 0 22px;
	}
	.directory-title {
		gap: 8px;
		align-items: baseline;
	}
	h1 {
		font-size: 28px;
		font-weight: 650;
		line-height: 1.3;
	}
	.directory-title span {
		color: var(--muted-foreground);
		font-size: 12px;
		background: var(--muted);
		border-radius: 4px;
		padding: 4px 8px;
	}
	.filters {
		display: flex;
		flex-wrap: wrap;
		align-items: end;
		gap: 10px;
		margin-bottom: 20px;
	}
	.filter-field {
		display: grid;
		gap: 5px;
		flex: 1 1 180px;
		min-width: 0;
		color: var(--muted-foreground);
		font-size: 11px;
		font-weight: 600;
	}
	.filters :global(.clear-filters) {
		flex: 0 0 auto;
	}
	.filters :global(.registry-filter) {
		height: 40px;
	}
	.metadata-error {
		padding: 10px 16px;
		margin: -8px 0 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-wrap: wrap;
	}
	.guide-search {
		position: relative;
		min-width: 0;
	}
	.guide-search :global(svg) {
		position: absolute;
		left: 12px;
		top: 9px;
		color: var(--muted-foreground);
	}
	.guide-search :global(input) {
		padding-left: 36px;
	}
	.catalog-controls {
		display: grid;
		gap: 10px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--border);
	}
	.troubleshooting-workspace {
		display: grid;
		grid-template-columns: minmax(270px, 350px) minmax(0, 1fr);
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--card);
		min-height: 620px;
		height: max(620px, calc(100dvh - 240px));
		grid-template-rows: minmax(0, 1fr);
	}
	.guide-list {
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
		overflow-y: auto;
		overscroll-behavior-y: contain;
		scrollbar-gutter: stable;
		border-right: 1px solid var(--border);
	}
	.rail-controls {
		flex-shrink: 0;
		position: sticky;
		top: 0;
		z-index: 2;
		background: var(--card);
	}
	.rail-heading {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 18px 16px 6px;
	}
	.rail-heading h2 {
		font-size: 14px;
		font-weight: 650;
	}
	.rail-heading span {
		font-size: 12px;
		color: var(--muted-foreground);
	}
	.catalog-row {
		position: relative;
		flex-shrink: 0;
	}
	.catalog-row > button {
		flex-shrink: 0;
		display: grid;
		position: relative;
		gap: 6px;
		width: 100%;
		min-width: 0;
		text-align: left;
		padding: 14px 28px 14px 14px;
		border-left: 3px solid transparent;
		border-top: 1px solid var(--border);
		cursor: pointer;
	}
	.guide-list strong {
		font-size: 13px;
		font-weight: 600;
		line-height: 1.4;
		color: var(--foreground);
		overflow-wrap: anywhere;
	}
	.guide-list .meta-line {
		font-size: 11px;
		color: var(--muted-foreground);
	}
	.catalog-row > button :global(.chevron-icon) {
		position: absolute;
		top: 50%;
		right: 6px;
		transform: translateY(-50%);
		color: var(--muted-foreground);
	}
	.catalog-row button:hover,
	.catalog-row button.chosen {
		background: var(--device-soft);
	}
	.catalog-row button.chosen {
		border-left-color: var(--device-accent);
	}
	.detail-pane {
		min-width: 0;
		min-height: 0;
		overflow-y: auto;
		padding: 18px;
		background: var(--background);
	}
	.banner-empty,
	.banner-error {
		padding: 28px 16px;
		text-align: center;
		font-size: 13px;
		color: var(--muted-foreground);
	}
	.banner-error {
		display: grid;
		justify-items: center;
		gap: 10px;
		color: var(--destructive);
	}
	@media (max-width: 850px) {
		.troubleshooting-workspace {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto auto;
			height: auto;
		}
		.guide-list {
			position: static;
			border-right: 0;
			border-bottom: 1px solid var(--border);
			max-height: 360px;
			overflow-y: auto;
		}
		.detail-pane {
			overflow-y: visible;
		}
	}
</style>
