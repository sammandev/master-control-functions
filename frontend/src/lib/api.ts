// Typed client for the app's REST API. Errors are normalized into
// ApiError with field-level messages from the server's validation shape.

export interface FunctionRow {
	id: number;
	source: string;
	interface: string;
	name: string;
	description: string;
	convention: number | null;
	category: string;
	execution_role: string;
	execution_phase: string;
	created_at: string;
	updated_at: string;
}

export interface FunctionInput {
	source: string;
	interface: string;
	name: string;
	description?: string;
	convention?: number | null;
	category?: string;
	execution_role?: string;
	execution_phase?: string;
}

export interface ListResult {
	items: FunctionRow[];
	total: number;
	page: number;
	page_size: number;
}

export interface ListParams {
	search?: string;
	source?: string;
	interface?: string;
	category?: string;
	execution_role?: string;
	execution_phase?: string;
	sort?: string;
	order?: 'asc' | 'desc';
	page?: number;
	page_size?: number;
	function_id?: number;
}

// Shared client plumbing lives in client.ts; re-exported here so existing
// import sites keep working.
import { ApiError, request } from './client';

export { ApiError };

function withQuery(params: ListParams): string {
	const qs = new URLSearchParams();
	if (params.search) qs.set('search', params.search);
	if (params.source) qs.set('source', params.source);
	if (params.interface) qs.set('interface', params.interface);
	if (params.category) qs.set('category', params.category);
	if (params.execution_role) qs.set('execution_role', params.execution_role);
	if (params.execution_phase) qs.set('execution_phase', params.execution_phase);
	if (params.sort) qs.set('sort', params.sort);
	if (params.order) qs.set('order', params.order);
	if (params.page) qs.set('page', String(params.page));
	if (params.page_size) qs.set('page_size', String(params.page_size));
	if (params.function_id) qs.set('function_id', String(params.function_id));
	const s = qs.toString();
	return s ? `?${s}` : '';
}

export async function listFunctions(params: ListParams = {}): Promise<ListResult> {
	return request<ListResult>(`/api/functions${withQuery(params)}`);
}

// Distinct enum values for the functions list filter bar (ticket 09).
export interface FunctionFacets {
	sources: string[];
	interfaces: string[];
	categories: string[];
	execution_roles: string[];
	execution_phases: string[];
}

export async function getFunctionFacets(): Promise<FunctionFacets> {
	return request<FunctionFacets>('/api/functions/facets');
}

export async function getFunction(id: number): Promise<FunctionRow> {
	return request<FunctionRow>(`/api/functions/${id}`);
}

export async function createFunction(input: FunctionInput): Promise<FunctionRow> {
	return request<FunctionRow>('/api/functions', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function updateFunction(id: number, input: FunctionInput): Promise<FunctionRow> {
	return request<FunctionRow>(`/api/functions/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function deleteFunction(id: number): Promise<void> {
	await request<void>(`/api/functions/${id}`, { method: 'DELETE' });
}
