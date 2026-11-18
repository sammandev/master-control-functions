// Typed client for the two definition dictionaries. Shape is identical
// for parameters and criteria (CONTEXT.md); only the endpoint path differs.
// HTTP plumbing comes from the shared client.
import { ApiError, request } from './client';

export { ApiError };

export interface DefinitionRow {
	id: number;
	name: string;
	// Device scope (ADR-0009): blank = global/shared definition;
	// an explicit scope overrides the global one for matching
	// device-scoped links.
	device_name: string;
	description: string;
	type: string;
	unit: string;
	default_value: string | null;
	allowed_values: string;
	cardinality: string;
	value_pattern: string;
	required: boolean;
	usage_count: number;
	// Links referencing this definition whose Function does not exist
	// (ADR-0003 orphan usage).
	orphan_usage_count: number;
	created_at: string;
	updated_at: string;
}

export interface DefinitionInput {
	name: string;
	device_name?: string;
	description?: string;
	type: string;
	unit?: string;
	default_value?: string | null;
	allowed_values?: string;
	cardinality: string;
	value_pattern?: string;
	required?: boolean;
}

export interface DefinitionListResult {
	items: DefinitionRow[];
	total: number;
	page: number;
	page_size: number;
}

export type DefinitionKind = 'parameter-definitions' | 'criteria-definitions';

export async function listDefinitions(
	kind: DefinitionKind,
	params: {
		search?: string;
		sort?: string;
		order?: 'asc' | 'desc';
		page?: number;
		page_size?: number;
	}
): Promise<DefinitionListResult> {
	const qs = new URLSearchParams();
	if (params.search) qs.set('search', params.search);
	if (params.sort) qs.set('sort', params.sort);
	if (params.order) qs.set('order', params.order);
	if (params.page) qs.set('page', String(params.page));
	if (params.page_size) qs.set('page_size', String(params.page_size));
	const s = qs.toString();
	return request<DefinitionListResult>(`/api/${kind}${s ? `?${s}` : ''}`);
}

export async function getDefinition(kind: DefinitionKind, id: number): Promise<DefinitionRow> {
	return request<DefinitionRow>(`/api/${kind}/${id}`);
}

export async function createDefinition(
	kind: DefinitionKind,
	input: DefinitionInput
): Promise<DefinitionRow> {
	return request<DefinitionRow>(`/api/${kind}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function updateDefinition(
	kind: DefinitionKind,
	id: number,
	input: DefinitionInput
): Promise<DefinitionRow> {
	return request<DefinitionRow>(`/api/${kind}/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function deleteDefinition(
	kind: DefinitionKind,
	id: number,
	force = false
): Promise<void> {
	await request<void>(`/api/${kind}/${id}${force ? '?force=true' : ''}`, {
		method: 'DELETE'
	});
}
