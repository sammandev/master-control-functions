// Shared HTTP client for the app's REST API: one error type, one request
// wrapper. Both the functions client (api.ts) and the dictionary client
// (definitions.ts) build on this so error handling stays identical.

export class ApiError extends Error {
	status: number;
	fields: Record<string, string>;

	constructor(status: number, message: string, fields: Record<string, string> = {}) {
		super(message);
		this.status = status;
		this.fields = fields;
	}
}

export async function request<T>(url: string, init?: RequestInit): Promise<T> {
	const res = await fetch(url, init);
	if (res.ok) {
		if (res.status === 204) return undefined as T;
		return (await res.json()) as T;
	}
	let message = `request failed (${res.status})`;
	let fields: Record<string, string> = {};
	try {
		const body = (await res.json()) as { error?: string; fields?: Record<string, string> };
		if (body.error) message = body.error;
		if (body.fields) fields = body.fields;
	} catch {
		// non-JSON error body: keep the generic message
	}
	throw new ApiError(res.status, message, fields);
}

// One display form for API errors: the message, plus the first field
// error after an em dash when the backend sent field details. Dialogs
// and pages render this directly so error copy stays identical
// everywhere.
export function formatApiError(e: unknown): string {
	if (e instanceof ApiError) {
		const field = Object.values(e.fields)[0];
		return field ? `${e.message}: ${field}` : e.message;
	}
	return e instanceof Error ? e.message : String(e);
}
