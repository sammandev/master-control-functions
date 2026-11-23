// Typed client for the sequence builder (ticket 10). HTTP plumbing from
// the shared client, mirroring examples.ts/compose.ts.
import { ApiError, request } from './client';

export { ApiError };

// One ordered step of a sequence: a Function reference by natural key,
// name-keyed raw values, per-step fixed segments, device scope and note.
// function_exists=false means the referenced Function is missing (orphan,
// ADR-0003).
export interface SequenceStep {
	id: number;
	position: number;
	function_source: string;
	function_interface: string;
	function_name: string;
	parameters: Record<string, string>;
	criteria: Record<string, string>;
	timeout_ms: string;
	flow: string;
	on_fail: string;
	device_name: string;
	note: string;
	function_exists: boolean;
}

export interface SequenceDetail {
	id: number;
	name: string;
	description: string;
	steps: SequenceStep[];
	created_at: string;
	updated_at: string;
}

export interface SequenceListItem {
	id: number;
	name: string;
	description: string;
	step_count: number;
	created_at: string;
	updated_at: string;
}

export interface SequenceListResult {
	items: SequenceListItem[];
	total: number;
	page: number;
	page_size: number;
}

// The write payload: PUT is a full replace of meta + ordered steps.
export interface SequenceStepInput {
	function_source: string;
	function_interface: string;
	function_name: string;
	parameters: Record<string, string>;
	criteria: Record<string, string>;
	timeout_ms: string;
	flow: string;
	on_fail: string;
	device_name: string;
	note: string;
}

export interface SequenceInput {
	name: string;
	description: string;
	steps: SequenceStepInput[];
}

export async function listSequences(
	params: { search?: string; page?: number; page_size?: number } = {}
): Promise<SequenceListResult> {
	const qs = new URLSearchParams();
	if (params.search) qs.set('search', params.search);
	if (params.page) qs.set('page', String(params.page));
	if (params.page_size) qs.set('page_size', String(params.page_size));
	const suffix = qs.size > 0 ? `?${qs.toString()}` : '';
	return request<SequenceListResult>(`/api/sequences${suffix}`);
}

export async function getSequence(id: number): Promise<SequenceDetail> {
	return request<SequenceDetail>(`/api/sequences/${id}`);
}

export async function createSequence(body: SequenceInput): Promise<SequenceDetail> {
	return request<SequenceDetail>('/api/sequences', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// Full replace: name/description plus the ordered step array in one write.
export async function updateSequence(id: number, body: SequenceInput): Promise<SequenceDetail> {
	return request<SequenceDetail>(`/api/sequences/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

export async function updateSequenceMeta(
	id: number,
	body: Pick<SequenceInput, 'name' | 'description'>
): Promise<SequenceDetail> {
	return request<SequenceDetail>(`/api/sequences/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

export async function deleteSequence(id: number): Promise<void> {
	await request<void>(`/api/sequences/${id}`, { method: 'DELETE' });
}

// One step's emission result: resolved=false when the Function is gone;
// non-empty errors when the stored values fail live re-validation.
export interface StepEmit {
	step_id: number;
	position: number;
	resolved: boolean;
	script: string;
	errors: Record<string, string>;
}

export interface EmitResult {
	// The newline-joined emission of the clean steps — "emitted text" per
	// CONTEXT.md ("script" is reserved for a single Function's string).
	text: string;
	steps: StepEmit[];
}

export async function emitSequence(id: number): Promise<EmitResult> {
	return request<EmitResult>(`/api/sequences/${id}/emit`, { method: 'POST' });
}
