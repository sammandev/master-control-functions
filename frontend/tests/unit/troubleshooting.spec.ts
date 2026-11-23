import { expect, test } from 'vitest';
import { createLatestRequestGuard } from '../../src/lib/request-guard';
import {
	categoryLabel,
	findRelatedGuides,
	formatTroubleshootingAiContext,
	moveDraftStep,
	severityLabel,
	severityTone,
	validateTroubleshootingDraftSteps,
	type TroubleshootingGuideDetail,
	type TroubleshootingGuideListItem
} from '../../src/lib/troubleshooting';
import type {
	TroubleshootingCategory,
	TroubleshootingSeverity
} from '../../src/lib/troubleshooting-registry';

const baseGuide: TroubleshootingGuideDetail = {
	id: 4,
	title: 'TX_MULTI_VERIFICATION Power Out of Spec Due to Tolerance vs Absolute Bounds',
	category: 'measurement',
	severity: 'critical',
	device_name: 'WIFI_TEST',
	function_name: 'TX_MULTI_VERIFICATION',
	error_code: 'ERRCODE:104',
	symptom: 'TX power fails criteria check immediately even though measured dBm is near target.',
	root_cause: 'POWER criterion was entered as absolute bounds instead of relative tolerance.',
	steps: [
		{
			id: 10,
			position: 1,
			title: 'Check TX_TARGET_POWER and POWER criterion',
			instruction: 'Verify POWER in TEST_SPEC.ini uses relative dB delta (e.g. -1.5~1.5).',
			script_snippet:
				'DLL:WIFI_TEST,TX_MULTI_VERIFICATION,ANT=1 TX_TARGET_POWER=17,POWER=-1.5~1.5,10000,NEXT,FAIL',
			verification: 'Pass window evaluates to 15.5 dBm .. 18.5 dBm.'
		}
	],
	created_at: '2026-11-19T00:00:00Z',
	updated_at: '2026-11-19T00:00:00Z'
};

const catalogItems: TroubleshootingGuideListItem[] = [
	{
		...baseGuide,
		step_count: 1,
		related_count: 2
	},
	{
		id: 5,
		title: 'TX_MULTI_VERIFICATION Power Low Across All Channels Due to RF Path Loss Table',
		category: 'measurement',
		severity: 'warning',
		device_name: 'WIFI_TEST',
		function_name: 'TX_MULTI_VERIFICATION',
		error_code: 'ERRCODE:104',
		symptom: 'Measured TX power is 4 dB lower across all frequencies on a station.',
		root_cause: 'Station RF cable loss compensation table is stale.',
		step_count: 2,
		related_count: 2,
		created_at: '2026-11-19T00:00:00Z',
		updated_at: '2026-11-19T00:00:00Z'
	},
	{
		id: 8,
		title: 'TX_MULTI_VERIFICATION EVM Degrades at High Target Power',
		category: 'measurement',
		severity: 'info',
		device_name: 'WIFI_TEST',
		function_name: 'TX_MULTI_VERIFICATION',
		error_code: 'ERRCODE:109',
		symptom: 'EVM fails on MCS11 while power passes.',
		root_cause: 'PA compression at high target power.',
		step_count: 1,
		related_count: 2,
		created_at: '2026-11-19T00:00:00Z',
		updated_at: '2026-11-19T00:00:00Z'
	},
	{
		id: 1,
		title: 'SSH Prompt Timeout Due to RD_SETTING Prompt Mismatch',
		category: 'connection',
		severity: 'critical',
		device_name: 'SSH',
		function_name: 'SENDSTR_AND_WAITSTR',
		error_code: 'ERRCODE:201',
		symptom: 'Timeout waiting for shell prompt.',
		root_cause: 'Prompt mismatch.',
		step_count: 2,
		related_count: 0,
		created_at: '2026-11-19T00:00:00Z',
		updated_at: '2026-11-19T00:00:00Z'
	}
];

test('findRelatedGuides groups sibling solutions sharing error_code or function_name (same error_code first)', () => {
	const related = findRelatedGuides(baseGuide, catalogItems);
	expect(related.map((item) => item.id)).toEqual([5, 8]);
});

test('formatTroubleshootingAiContext includes problem tags, steps, script snippets, and sibling solutions', () => {
	const related = findRelatedGuides(baseGuide, catalogItems);
	const context = formatTroubleshootingAiContext(baseGuide, related);
	expect(context).toContain(
		'# MCV2 Troubleshooting Guide: TX_MULTI_VERIFICATION Power Out of Spec'
	);
	expect(context).toContain('ERRCODE:104');
	expect(context).toContain('WIFI_TEST');
	expect(context).toContain('Step 1: Check TX_TARGET_POWER and POWER criterion');
	expect(context).toContain('DLL:WIFI_TEST,TX_MULTI_VERIFICATION');
	expect(context).toContain('Alternative Solutions (2)');
	expect(context).toContain(
		'TX_MULTI_VERIFICATION Power Low Across All Channels Due to RF Path Loss Table'
	);
});

test('categoryLabel uses registry metadata and preserves legacy one-argument calls', () => {
	const categories: TroubleshootingCategory[] = [
		{
			key: 'measurement-plus',
			label: 'Advanced Measurement',
			description: '',
			sort_order: 1,
			built_in: false,
			guide_count: 0
		}
	];

	expect(categoryLabel('measurement-plus', categories)).toBe('Advanced Measurement');
	expect(categoryLabel('measurement')).toBe('Measurement & Criteria');
	expect(categoryLabel('signal_quality')).toBe('Signal Quality');
	expect(categoryLabel('')).toBe('');
});

test('latest request guard rejects older asynchronous responses', () => {
	const guard = createLatestRequestGuard();
	const first = guard.begin();
	const second = guard.begin();

	expect(guard.isCurrent(first)).toBe(false);
	expect(guard.isCurrent(second)).toBe(true);
});

test('validateTroubleshootingDraftSteps identifies only the invalid step', () => {
	const errors = validateTroubleshootingDraftSteps([
		{
			client_id: 'draft-a',
			title: 'Valid step',
			instruction: 'Valid instruction',
			script_snippet: '',
			verification: ''
		},
		{
			client_id: 'draft-b',
			title: '',
			instruction: ' ',
			script_snippet: '',
			verification: ''
		}
	]);

	expect(errors).toEqual({
		'steps.1.title': 'Step 2: step title is required',
		'steps.1.instruction': 'Step 2: instruction is required'
	});
});

test('moveDraftStep moves a complete stable record without rekeying fields', () => {
	const first = {
		client_id: 'draft-a',
		title: 'First',
		instruction: 'Inspect first',
		script_snippet: 'FIRST=1',
		verification: 'first verified'
	};
	const second = {
		client_id: 'draft-b',
		title: 'Second',
		instruction: 'Inspect second',
		script_snippet: 'SECOND=2',
		verification: 'second verified'
	};
	const steps = [first, second];

	const moved = moveDraftStep(steps, 1, -1);

	expect(moved).toEqual([second, first]);
	expect(moved[0]).toBe(second);
	expect(steps).toEqual([first, second]);
});

test('severity helpers use registry metadata and neutral fallbacks', () => {
	const severities: TroubleshootingSeverity[] = [
		{
			key: 'review',
			label: 'Needs Review',
			description: '',
			sort_order: 1,
			built_in: false,
			guide_count: 0,
			tone: 'warn'
		}
	];

	expect(severityLabel('review', severities)).toBe('Needs Review');
	expect(severityTone('review', severities)).toBe('warn');
	expect(severityLabel('critical')).toBe('Critical');
	expect(severityLabel('operator_attention')).toBe('Operator Attention');
	expect(severityTone('unknown')).toBe('outline');
	expect(severityLabel('')).toBe('');
	expect(severityTone('')).toBe('outline');
});
