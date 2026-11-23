import { afterEach, describe, expect, test, vi } from 'vitest';
import {
	createTroubleshootingGuide,
	deleteTroubleshootingGuide,
	getTroubleshootingGuide,
	listTroubleshootingGuides,
	updateTroubleshootingGuide,
	type TroubleshootingGuideDetail,
	type TroubleshootingGuideInput
} from '../../src/lib/troubleshooting';

const guide: TroubleshootingGuideDetail = {
	id: 1,
	title: 'SSH prompt mismatch',
	category: 'connection',
	severity: 'critical',
	device_name: 'SSH',
	function_name: 'SENDSTR_AND_WAITSTR',
	error_code: 'ERRCODE:201',
	symptom: 'Prompt timeout',
	root_cause: 'Wrong prompt',
	steps: [
		{
			id: 1,
			position: 1,
			title: 'Inspect prompt',
			instruction: 'Open a terminal',
			script_snippet: '[SSH]\nPROMPT=root@OpenWrt:~#',
			verification: 'Prompt matches'
		}
	],
	created_at: '2026-11-21T00:00:00Z',
	updated_at: '2026-11-21T00:00:00Z'
};

const input: TroubleshootingGuideInput = {
	title: guide.title,
	category: guide.category,
	severity: guide.severity,
	device_name: guide.device_name,
	function_name: guide.function_name,
	error_code: guide.error_code,
	symptom: guide.symptom,
	root_cause: guide.root_cause,
	steps: guide.steps.map(({ title, instruction, script_snippet, verification }) => ({
		title,
		instruction,
		script_snippet,
		verification
	}))
};

function mockResponse(body: unknown, status = 200) {
	const fetchMock = vi.fn(async () => {
		if (status === 204) return new Response(null, { status });
		return Response.json(body, { status });
	});
	vi.stubGlobal('fetch', fetchMock);
	return fetchMock;
}

afterEach(() => vi.unstubAllGlobals());

describe('troubleshooting guide CRUD client', () => {
	test('lists and reads guides', async () => {
		const listMock = mockResponse({ items: [guide], total: 1, page: 1, page_size: 20 });
		await expect(listTroubleshootingGuides({ search: 'prompt' })).resolves.toMatchObject({
			total: 1
		});
		expect(listMock).toHaveBeenCalledWith('/api/troubleshooting?search=prompt', undefined);

		vi.unstubAllGlobals();
		const getMock = mockResponse(guide);
		await expect(getTroubleshootingGuide(1)).resolves.toMatchObject({ id: 1 });
		expect(getMock).toHaveBeenCalledWith('/api/troubleshooting/1', undefined);
	});

	test('creates and updates guides with JSON content type', async () => {
		const fetchMock = mockResponse(guide);

		await createTroubleshootingGuide(input);
		expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/troubleshooting', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		});

		await updateTroubleshootingGuide(1, input);
		expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/troubleshooting/1', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		});
	});

	test('deletes a guide from the exact item route', async () => {
		const fetchMock = mockResponse(undefined, 204);
		await deleteTroubleshootingGuide(1);
		expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/1', { method: 'DELETE' });
	});
});
