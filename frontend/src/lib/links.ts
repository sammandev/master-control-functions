// Typed client for the function editor's parameter/criteria link
// endpoints. The two link families share one shape (CONTEXT.md); only the
// endpoint segment differs. HTTP plumbing comes from the shared client.
import { ApiError, request } from './client';
import type { DefinitionInput } from './definitions';

export { ApiError };

export type LinkKind = 'parameter-links' | 'criteria-links';

export interface LinkRow {
	id: number;
	function_source: string;
	function_name: string;
	device_name: string;
	name: string;
	position: number | null;
	required: boolean;
	source_file: string;
	resolved: boolean;
	definition_id: number | null;
	created_at: string;
	updated_at: string;
}

// Attach payload: either name (pick an existing definition) or definition
// (create it inline and link it in one flow).
export interface LinkAttachInput {
	name?: string;
	definition?: DefinitionInput;
	device_name?: string;
	position?: number | null;
	required?: boolean;
}

// Update payload: only the editable link attributes (ADR-0002).
export interface LinkUpdateInput {
	device_name?: string;
	position?: number | null;
	required?: boolean;
}

export async function listLinks(fnId: number, kind: LinkKind): Promise<LinkRow[]> {
	const res = await request<{ items: LinkRow[] }>(`/api/functions/${fnId}/${kind}`);
	return res.items;
}

export async function attachLink(
	fnId: number,
	kind: LinkKind,
	input: LinkAttachInput
): Promise<LinkRow> {
	return request<LinkRow>(`/api/functions/${fnId}/${kind}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function updateLink(
	fnId: number,
	kind: LinkKind,
	linkId: number,
	input: LinkUpdateInput
): Promise<LinkRow> {
	return request<LinkRow>(`/api/functions/${fnId}/${kind}/${linkId}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(input)
	});
}

export async function detachLink(fnId: number, kind: LinkKind, linkId: number): Promise<void> {
	await request<void>(`/api/functions/${fnId}/${kind}/${linkId}`, { method: 'DELETE' });
}
