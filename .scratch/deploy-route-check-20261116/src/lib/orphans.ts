// Typed client for the orphan report endpoint (ticket 05). The report is
// read-only: resolution happens through the regular create endpoints, so
// this module only lists and groups. HTTP plumbing from the shared client.
import { ApiError, request } from './client';

export { ApiError };

export type OrphanKind = 'parameter-links' | 'criteria-links' | 'examples';

export interface OrphanLinkRow {
	id: number;
	function_source: string;
	function_name: string;
	function_exists: boolean;
	// The Function's row id for jump-to-row; null when the function itself
	// is the missing target (multi-interface functions are safe: the API
	// resolves one representative id via a scalar subquery).
	function_id: number | null;
	name: string;
	definition_exists: boolean;
	device_name: string;
	position: number | null;
	required: boolean;
}

export interface OrphanExampleRow {
	id: number;
	function_source: string;
	function_name: string;
	device_name: string;
	script: string;
	meaning: string;
}

export interface OrphanGroup {
	kind: OrphanKind;
	count: number;
	items: (OrphanLinkRow | OrphanExampleRow)[];
}

export interface OrphanReport {
	total: number;
	groups: OrphanGroup[];
}

export async function listOrphans(): Promise<OrphanReport> {
	return request<OrphanReport>('/api/orphans');
}
