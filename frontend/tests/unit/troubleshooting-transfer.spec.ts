import { afterEach, expect, test, vi } from 'vitest';
import {
	getTroubleshootingImportDraft,
	getTroubleshootingPreview,
	requestTroubleshootingExport
} from '../../src/lib/troubleshooting-transfer';

afterEach(() => vi.unstubAllGlobals());

test('uploads the original file for preview and keeps errored batches invalid', async () => {
	const result = { valid: false, rows: [{ index: 1, action: 'error', reasons: ['Unknown id'] }] };
	const fetchMock = vi.fn(async (_url: string, init: RequestInit) => {
		expect((init.body as FormData).get('file')).toBe(file);
		return Response.json(result);
	});
	vi.stubGlobal('fetch', fetchMock);
	const file = new File(['{}'], 'guides.json');
	await expect(getTroubleshootingPreview(file)).resolves.toEqual(result);
	expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/import/preview', {
		method: 'POST',
		body: expect.any(FormData)
	});
});

test('returns a single guide draft without importing its identity', async () => {
	const steps = [{ title: 'One', instruction: '- First\n- Second' }];
	vi.stubGlobal(
		'fetch',
		vi.fn(async () => Response.json({ id: 100, title: 'Draft', steps }))
	);
	const draft = await getTroubleshootingImportDraft(new File(['{}'], 'guide.xlsx'));
	expect(draft.title).toBe('Draft');
	expect(draft.steps).toEqual(steps);
	expect(draft).not.toHaveProperty('id');
});

test('matching export sends catalog filters but no pagination', async () => {
	const fetchMock = vi.fn(async () => new Response('export'));
	vi.stubGlobal('fetch', fetchMock);
	await requestTroubleshootingExport(
		'both',
		{ scope: 'matching' },
		{ search: 'prompt', device: '' }
	);
	expect(fetchMock).toHaveBeenCalledWith('/api/troubleshooting/export?search=prompt&device=', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: '{"format":"both","scope":"matching"}'
	});
});
