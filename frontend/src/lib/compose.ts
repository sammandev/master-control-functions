// Typed client for the script composer (ticket 07).
import { ApiError, request } from './client';

export { ApiError };

// One composer form field: a link enriched with its definition's
// constraints. resolved=false means the linked definition is missing.
export interface ComposeField {
	name: string;
	position: number | null;
	required: boolean;
	device_name: string;
	resolved: boolean;
	type: string;
	unit: string;
	default_value: string;
	allowed_values: string;
	cardinality: string;
}

export interface ComposeFields {
	function: { source: string; interface: string; name: string };
	parameters: ComposeField[];
	criteria: ComposeField[];
}

export interface ComposeRequest {
	parameters: Record<string, string>;
	criteria: Record<string, string>;
	timeout_ms: string;
	flow: string;
	on_fail: string;
}

export async function getComposeFields(functionId: number): Promise<ComposeFields> {
	return request<ComposeFields>(`/api/functions/${functionId}/compose/fields`);
}

// 200 → {script}; 422 → ApiError with field-level messages.
export async function composeScript(
	functionId: number,
	req: ComposeRequest
): Promise<{ script: string }> {
	return request<{ script: string }>(`/api/functions/${functionId}/compose`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(req)
	});
}
