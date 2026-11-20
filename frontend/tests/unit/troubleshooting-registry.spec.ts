import { afterEach, describe, expect, test, vi } from 'vitest';
import {
	createTroubleshootingCategory,
	createTroubleshootingSeverity,
	deleteTroubleshootingCategory,
	deleteTroubleshootingSeverity,
	listTroubleshootingCategories,
	listTroubleshootingSeverities,
	updateTroubleshootingCategory,
	updateTroubleshootingSeverity
} from '../../src/lib/troubleshooting-registry';

function mockJsonResponse(body: unknown, status = 200) {
	const fetchMock = vi.fn(
		async () =>
			new Response(JSON.stringify(body), {
				status,
				headers: { 'Content-Type': 'application/json' }
			})
	);
	vi.stubGlobal('fetch', fetchMock);
	return fetchMock;
}

afterEach(() => vi.unstubAllGlobals());

describe('troubleshooting category registry client', () => {
	test('lists categories from the exact collection route', async () => {
		const fetchMock = mockJsonResponse({ items: [], total: 0 });

		await expect(listTroubleshootingCategories()).resolves.toEqual({ items: [], total: 0 });
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/categories', undefined);
	});

	test('creates a category with the JSON body unchanged', async () => {
		const input = { label: 'Line Setup', description: 'Station wiring' };
		const fetchMock = mockJsonResponse({ key: 'line-setup', ...input });

		await createTroubleshootingCategory(input);
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/categories', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	});

	test('updates a category using its case-preserved encoded key', async () => {
		const input = { label: 'Line Setup', description: 'Updated', sort_order: 4 };
		const fetchMock = mockJsonResponse({ key: 'Line Setup/Primary', ...input });

		await updateTroubleshootingCategory('Line Setup/Primary', input);
		expect(fetchMock).toHaveBeenCalledWith(
			'/api/troubleshooting/categories/Line%20Setup%2FPrimary',
			{
				method: 'PUT',
				body: JSON.stringify(input)
			}
		);
	});

	test('deletes a category with encoded key and replacement query', async () => {
		const fetchMock = vi.fn(async () => new Response(null, { status: 204 }));
		vi.stubGlobal('fetch', fetchMock);

		await deleteTroubleshootingCategory('foo-bar', 'connection');
		expect(fetchMock).toHaveBeenCalledWith(
			'/api/troubleshooting/categories/foo-bar?reassign_to=connection',
			{ method: 'DELETE' }
		);
	});
});

describe('troubleshooting severity registry client', () => {
	test('lists severities from the exact collection route', async () => {
		const fetchMock = mockJsonResponse({ items: [], total: 0 });

		await expect(listTroubleshootingSeverities()).resolves.toEqual({ items: [], total: 0 });
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/severities', undefined);
	});

	test('creates a severity with the JSON body unchanged', async () => {
		const input = {
			label: 'Needs Review',
			description: 'Operator review required',
			tone: 'warn'
		};
		const fetchMock = mockJsonResponse({ key: 'needs-review', ...input });

		await createTroubleshootingSeverity(input);
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/severities', {
			method: 'POST',
			body: JSON.stringify(input)
		});
	});

	test('updates a severity using its case-preserved encoded key', async () => {
		const input = {
			label: 'Needs Review',
			description: 'Updated',
			sort_order: 5,
			tone: 'outline'
		};
		const fetchMock = mockJsonResponse({ key: 'Review/Level', ...input });

		await updateTroubleshootingSeverity('Review/Level', input);
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/severities/Review%2FLevel', {
			method: 'PUT',
			body: JSON.stringify(input)
		});
	});

	test('deletes a severity with encoded key and replacement query', async () => {
		const fetchMock = vi.fn(async () => new Response(null, { status: 204 }));
		vi.stubGlobal('fetch', fetchMock);

		await deleteTroubleshootingSeverity('Urgent/Now', 'review level');
		expect(fetchMock).toHaveBeenCalledWith(
			'/api/troubleshooting/severities/Urgent%2FNow?reassign_to=review%20level',
			{ method: 'DELETE' }
		);
	});
});
