// Client for workbook import, reset, and the change-log views.
import { request, ApiError } from './client';

export { ApiError };

export interface ImportResult {
	run_id: number;
	file_name: string;
	reset: boolean;
	counts: Record<string, number>;
	skipped: { table: string; row_key: string; reason: string }[];
	normalizations: {
		table: string;
		row_key: string;
		column: string;
		source: string;
		normalized: string;
		rule: string;
	}[];
}

export interface ImportRun {
	id: number;
	file_name: string;
	reset: boolean;
	started_at: string;
	finished_at: string;
	counts: string;
	skipped: string;
}

export interface ImportNormalization {
	id: number;
	run_id: number;
	table_name: string;
	row_key: string;
	column_name: string;
	source_value: string;
	normalized_value: string;
	rule: string;
}

async function upload(path: string, file: File): Promise<ImportResult> {
	const fd = new FormData();
	fd.append('file', file);
	const raw = await request<ImportResult>(path, { method: 'POST', body: fd });
	// The Go server marshals nil slices as null — a clean import produces
	// no normalizations or skips and arrives as null. Normalize here so
	// callers can rely on the declared array types.
	return { ...raw, skipped: raw.skipped ?? [], normalizations: raw.normalizations ?? [] };
}

export function importWorkbook(file: File): Promise<ImportResult> {
	return upload('/api/import', file);
}

export function importReset(file: File): Promise<ImportResult> {
	return upload('/api/import/reset', file);
}

export async function downloadWorkbook(): Promise<void> {
	const response = await fetch('/api/export');
	if (!response.ok) {
		const body: { error?: string } = await response.json().catch(() => ({}));
		throw new ApiError(response.status, body.error ?? `Export failed (${response.status})`);
	}
	const url = URL.createObjectURL(await response.blob());
	const link = document.createElement('a');
	link.href = url;
	link.download = 'knowledge_workbook.xlsx';
	link.click();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export async function listImportRuns(): Promise<{ items: ImportRun[] }> {
	return request<{ items: ImportRun[] }>('/api/import/runs');
}

export async function getImportRun(
	id: number
): Promise<{ run: ImportRun; normalizations: ImportNormalization[] }> {
	return request<{ run: ImportRun; normalizations: ImportNormalization[] }>(
		`/api/import/runs/${id}`
	);
}
