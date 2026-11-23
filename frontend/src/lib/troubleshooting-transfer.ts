import { ApiError, request } from './client';
import {
	troubleshootingFilterParams,
	type TroubleshootingFilters,
	type TroubleshootingGuideInput
} from './troubleshooting';

export const troubleshootingTransferFileTypes = '.json,.xlsx';

export interface TroubleshootingPreview {
	valid: boolean;
	errors?: string[];
	rows: {
		index: number;
		id?: number;
		title: string;
		action: 'create' | 'update' | 'error';
		reasons?: string[];
	}[];
}

export interface TroubleshootingImportResult {
	created: number;
	updated: number;
}

export type TroubleshootingExportFormat = 'json' | 'xlsx' | 'both';
export type TroubleshootingExportSelection =
	{ scope: 'selected'; ids: number[] } | { scope: 'matching' | 'all' };

const transferBase = '/api/troubleshooting';

export async function downloadTroubleshootingTemplate(format: 'json' | 'xlsx'): Promise<void> {
	const response = await fetch(`${transferBase}/templates/${format}`);
	if (!response.ok) {
		const body: { error?: string } = await response.json().catch(() => ({}));
		throw new ApiError(
			response.status,
			body.error ?? `Template download failed (${response.status})`
		);
	}
	if (!response.headers.get('Content-Disposition')?.startsWith('attachment')) {
		throw new ApiError(response.status, 'The server did not return a troubleshooting template.');
	}
	saveDownload(await response.blob(), `troubleshooting-guide-template.${format}`);
}

function upload<T>(file: File, operation: 'single' | 'preview' | 'commit'): Promise<T> {
	const body = new FormData();
	body.set('file', file);
	return request<T>(`${transferBase}/import/${operation}`, { method: 'POST', body });
}

export async function getTroubleshootingImportDraft(
	file: File
): Promise<TroubleshootingGuideInput> {
	const {
		title,
		category,
		severity,
		device_name,
		function_name,
		error_code,
		symptom,
		root_cause,
		steps
	} = await upload<TroubleshootingGuideInput>(file, 'single');
	return {
		title,
		category,
		severity,
		device_name,
		function_name,
		error_code,
		symptom,
		root_cause,
		steps
	};
}

export function getTroubleshootingPreview(file: File): Promise<TroubleshootingPreview> {
	return upload(file, 'preview');
}

export function commitTroubleshootingImport(file: File): Promise<TroubleshootingImportResult> {
	return upload(file, 'commit');
}

export async function requestTroubleshootingExport(
	format: TroubleshootingExportFormat,
	selection: TroubleshootingExportSelection,
	filters: TroubleshootingFilters
): Promise<Blob> {
	const query =
		selection.scope === 'matching' ? troubleshootingFilterParams(filters).toString() : '';
	const response = await fetch(`${transferBase}/export${query ? `?${query}` : ''}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ format, ...selection })
	});
	if (!response.ok) {
		const body: { error?: string } = await response.json().catch(() => ({}));
		throw new ApiError(response.status, body.error ?? `Export failed (${response.status})`);
	}
	return response.blob();
}

export function downloadTroubleshootingExport(blob: Blob, format: TroubleshootingExportFormat) {
	saveDownload(blob, `troubleshooting-guides.${format === 'both' ? 'zip' : format}`);
}

function saveDownload(blob: Blob, filename: string) {
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.href = url;
	link.download = filename;
	link.click();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}
