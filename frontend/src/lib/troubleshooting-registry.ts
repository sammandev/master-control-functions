import { request } from './client';

export interface TroubleshootingCategory {
	key: string;
	label: string;
	description: string;
	sort_order: number;
	built_in: boolean;
	guide_count: number;
}

export interface TroubleshootingSeverity {
	key: string;
	label: string;
	description: string;
	sort_order: number;
	built_in: boolean;
	guide_count: number;
	tone: string;
}

export interface TroubleshootingCategoryInput {
	label: string;
	description?: string;
	sort_order?: number;
}

export interface TroubleshootingSeverityInput extends TroubleshootingCategoryInput {
	tone?: string;
}

export interface TroubleshootingRegistryListResult<T> {
	items: T[];
	total: number;
}

const categoriesUrl = '/api/troubleshooting/categories';
const severitiesUrl = '/api/troubleshooting/severities';
const jsonHeaders = { 'Content-Type': 'application/json' };

export function listTroubleshootingCategories(): Promise<
	TroubleshootingRegistryListResult<TroubleshootingCategory>
> {
	return request<TroubleshootingRegistryListResult<TroubleshootingCategory>>(categoriesUrl);
}

export function createTroubleshootingCategory(
	body: TroubleshootingCategoryInput
): Promise<TroubleshootingCategory> {
	return request<TroubleshootingCategory>(categoriesUrl, {
		method: 'POST',
		headers: jsonHeaders,
		body: JSON.stringify(body)
	});
}

export function updateTroubleshootingCategory(
	key: string,
	body: TroubleshootingCategoryInput
): Promise<TroubleshootingCategory> {
	return request<TroubleshootingCategory>(`${categoriesUrl}/${encodeURIComponent(key)}`, {
		method: 'PUT',
		headers: jsonHeaders,
		body: JSON.stringify(body)
	});
}

export function reorderTroubleshootingCategory(
	key: string,
	withKey: string
): Promise<{ items: TroubleshootingCategory[] }> {
	return request<{ items: TroubleshootingCategory[] }>(`${categoriesUrl}/reorder`, {
		method: 'POST',
		headers: jsonHeaders,
		body: JSON.stringify({ key, with_key: withKey })
	});
}

export async function deleteTroubleshootingCategory(
	key: string,
	reassignTo: string
): Promise<void> {
	await request<null>(
		`${categoriesUrl}/${encodeURIComponent(key)}?reassign_to=${encodeURIComponent(reassignTo)}`,
		{ method: 'DELETE' }
	);
}

export function listTroubleshootingSeverities(): Promise<
	TroubleshootingRegistryListResult<TroubleshootingSeverity>
> {
	return request<TroubleshootingRegistryListResult<TroubleshootingSeverity>>(severitiesUrl);
}

export function createTroubleshootingSeverity(
	body: TroubleshootingSeverityInput
): Promise<TroubleshootingSeverity> {
	return request<TroubleshootingSeverity>(severitiesUrl, {
		method: 'POST',
		headers: jsonHeaders,
		body: JSON.stringify(body)
	});
}

export function updateTroubleshootingSeverity(
	key: string,
	body: TroubleshootingSeverityInput
): Promise<TroubleshootingSeverity> {
	return request<TroubleshootingSeverity>(`${severitiesUrl}/${encodeURIComponent(key)}`, {
		method: 'PUT',
		headers: jsonHeaders,
		body: JSON.stringify(body)
	});
}

export function reorderTroubleshootingSeverity(
	key: string,
	withKey: string
): Promise<{ items: TroubleshootingSeverity[] }> {
	return request<{ items: TroubleshootingSeverity[] }>(`${severitiesUrl}/reorder`, {
		method: 'POST',
		headers: jsonHeaders,
		body: JSON.stringify({ key, with_key: withKey })
	});
}

export async function deleteTroubleshootingSeverity(
	key: string,
	reassignTo: string
): Promise<void> {
	await request<null>(
		`${severitiesUrl}/${encodeURIComponent(key)}?reassign_to=${encodeURIComponent(reassignTo)}`,
		{ method: 'DELETE' }
	);
}
