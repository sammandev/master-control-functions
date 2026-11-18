// Client for the MCP catalog contract (ADR-0012): audited script grammar,
// script-key audits and readiness for script-audit-mcp.
import { request } from './client';

export type ReadinessStatus = 'ready' | 'bundle_ready' | 'not_ready';
export type GrammarScope = 'line' | 'bundle';

export interface ScriptIssue {
	field: string;
	code: string;
	message: string;
}

export interface Grammar {
	encoding: 'raw' | 'named';
	scope: GrammarScope;
	audited_by: string;
	audited_at: string;
	note: string;
}

export interface ScriptKey {
	kind: 'parameter' | 'criterion';
	name: string;
	scope: GrammarScope;
	audited_by: string;
	audited_at: string;
	note: string;
}

export interface ReadinessField {
	name: string;
	device_name: string;
	required: boolean;
	has_definition: boolean;
	script_key_scope: '' | GrammarScope;
}

export interface Readiness {
	function_id: number;
	device: string;
	status: ReadinessStatus;
	grammar: Grammar | null;
	line: { ready: boolean; issues: ScriptIssue[] };
	bundle: { ready: boolean; issues: ScriptIssue[] };
	parameters: ReadinessField[];
	criteria: ReadinessField[];
	warnings: ScriptIssue[];
}

const json = { 'Content-Type': 'application/json' };

export function getReadiness(fnId: number, device = ''): Promise<Readiness> {
	const qs = device ? `?device=${encodeURIComponent(device)}` : '';
	return request<Readiness>(`/api/functions/${fnId}/mcp-readiness${qs}`);
}

export function getGrammar(
	fnId: number
): Promise<{ grammar: Grammar | null; script_keys: ScriptKey[] }> {
	return request(`/api/functions/${fnId}/grammar`);
}

export function saveGrammar(
	fnId: number,
	input: Pick<Grammar, 'encoding' | 'scope' | 'audited_by' | 'note'>
): Promise<Grammar> {
	return request(`/api/functions/${fnId}/grammar`, {
		method: 'PUT',
		headers: json,
		body: JSON.stringify(input)
	});
}

export async function deleteGrammar(fnId: number): Promise<void> {
	await request<void>(`/api/functions/${fnId}/grammar`, { method: 'DELETE' });
}

export function saveScriptKey(
	fnId: number,
	input: Pick<ScriptKey, 'kind' | 'name' | 'scope' | 'audited_by' | 'note'>
): Promise<ScriptKey> {
	return request(`/api/functions/${fnId}/script-keys`, {
		method: 'PUT',
		headers: json,
		body: JSON.stringify(input)
	});
}

export async function deleteScriptKey(
	fnId: number,
	kind: ScriptKey['kind'],
	name: string
): Promise<void> {
	await request<void>(`/api/functions/${fnId}/script-keys/${kind}/${encodeURIComponent(name)}`, {
		method: 'DELETE'
	});
}

export const statusLabel: Record<ReadinessStatus, string> = {
	ready: 'Ready for MCP',
	bundle_ready: 'Bundle only',
	not_ready: 'Not ready'
};
