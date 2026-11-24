// Typed client for the Examples CRUD API (ticket 06). HTTP plumbing from
// the shared client, mirroring links.ts/definitions.ts.
import { ApiError, request } from './client';
import type { SortDir } from './table-sort';

export { ApiError };

export interface ExampleRow {
	id: number;
	function_source: string;
	function_name: string;
	device_name: string;
	script: string;
	meaning: string;
	// False when the referenced Function is missing (orphan, ADR-0003).
	function_exists: boolean;
	created_at: string;
	updated_at: string;
}

export interface ExampleListParams {
	function_source?: string;
	function_name?: string;
	device?: string;
	search?: string;
	sort?: string;
	order?: SortDir;
	page?: number;
	page_size?: number;
	example_id?: number;
}

export interface ExampleListResult {
	items: ExampleRow[];
	total: number;
	page: number;
	page_size: number;
}

// Create payload: the function reference is required and validated.
export interface ExampleInput {
	function_source: string;
	function_name: string;
	device_name: string;
	script: string;
	meaning: string;
}

// Update payload: the function reference is immutable on edit.
export interface ExampleUpdateInput {
	device_name: string;
	script: string;
	meaning: string;
}

export async function listExamples(params: ExampleListParams = {}): Promise<ExampleListResult> {
	const qs = new URLSearchParams();
	if (params.function_source) qs.set('function_source', params.function_source);
	if (params.function_name) qs.set('function_name', params.function_name);
	// device is sent whenever the key is present, so an empty string
	// filters blank-scope (global) examples instead of meaning "no filter".
	if (params.device !== undefined) qs.set('device', params.device);
	if (params.search) qs.set('search', params.search);
	if (params.sort) qs.set('sort', params.sort);
	if (params.order) qs.set('order', params.order);
	if (params.page) qs.set('page', String(params.page));
	if (params.page_size) qs.set('page_size', String(params.page_size));
	if (params.example_id) qs.set('example_id', String(params.example_id));
	const suffix = qs.size > 0 ? `?${qs.toString()}` : '';
	return request<ExampleListResult>(`/api/examples${suffix}`);
}

export async function listExampleFunctionNames(): Promise<{ items: string[] }> {
	return request<{ items: string[] }>('/api/examples/function-names');
}

export async function createExample(body: ExampleInput): Promise<ExampleRow> {
	return request<ExampleRow>('/api/examples', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

export async function updateExample(id: number, body: ExampleUpdateInput): Promise<ExampleRow> {
	return request<ExampleRow>(`/api/examples/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

export async function deleteExample(id: number): Promise<void> {
	await request<void>(`/api/examples/${id}`, { method: 'DELETE' });
}
