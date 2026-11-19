// Typed client and AI-context formatter for the Troubleshooting Playbook & AI Knowledge Base.
// Multiple guides can share the same error_code, function_name, or device_name because one
// error code or symptom can manifest in multiple ways with distinct step-by-step solutions.
import { ApiError, request } from './client';

export { ApiError };

export type TroubleshootingCategory = 'connection' | 'execution' | 'measurement' | 'production';
export type TroubleshootingSeverity = 'critical' | 'warning' | 'info';

export interface TroubleshootingStep {
	id: number;
	position: number;
	title: string;
	instruction: string;
	script_snippet: string;
	verification: string;
}

export interface TroubleshootingGuideDetail {
	id: number;
	title: string;
	category: TroubleshootingCategory;
	severity: TroubleshootingSeverity;
	device_name: string;
	function_name: string;
	error_code: string;
	symptom: string;
	root_cause: string;
	steps: TroubleshootingStep[];
	created_at: string;
	updated_at: string;
}

export interface TroubleshootingGuideListItem {
	id: number;
	title: string;
	category: TroubleshootingCategory;
	severity: TroubleshootingSeverity;
	device_name: string;
	function_name: string;
	error_code: string;
	symptom: string;
	root_cause: string;
	step_count: number;
	related_count: number;
	created_at: string;
	updated_at: string;
}

export interface TroubleshootingListResult {
	items: TroubleshootingGuideListItem[];
	total: number;
	page: number;
	page_size: number;
}

export interface TroubleshootingStepInput {
	title: string;
	instruction: string;
	script_snippet: string;
	verification: string;
}

export interface TroubleshootingGuideInput {
	title: string;
	category: TroubleshootingCategory;
	severity: TroubleshootingSeverity;
	device_name: string;
	function_name: string;
	error_code: string;
	symptom: string;
	root_cause: string;
	steps: TroubleshootingStepInput[];
}

export const TROUBLESHOOTING_CATEGORIES: {
	value: TroubleshootingCategory;
	label: string;
	description: string;
}[] = [
	{
		value: 'connection',
		label: 'DUT Connection & Setup',
		description: 'SSH, Telnet, UART, ADB, socket, or instrument initialization problems'
	},
	{
		value: 'execution',
		label: 'Script & Flow Execution',
		description: '*V:Section-Key# resolution, 7/8-field grammar, macros, loops, and jump targets'
	},
	{
		value: 'measurement',
		label: 'Measurement & Criteria',
		description: 'RF TX/RX verification, ANT= bitmasks, tolerance windows, and path loss'
	},
	{
		value: 'production',
		label: 'Production Run & Handoff',
		description: 'FINAL FLOW cleanup, shopfloor handoff, fixture locks, and station repeatability'
	}
];

export const TROUBLESHOOTING_SEVERITIES: {
	value: TroubleshootingSeverity;
	label: string;
}[] = [
	{ value: 'critical', label: 'Critical (Blocks Flow)' },
	{ value: 'warning', label: 'Warning (Intermittent / Margin)' },
	{ value: 'info', label: 'Info (Best Practice)' }
];

export function categoryLabel(category: string): string {
	return TROUBLESHOOTING_CATEGORIES.find((c) => c.value === category)?.label ?? category;
}

export function severityLabel(severity: string): string {
	switch (severity) {
		case 'critical':
			return 'Critical';
		case 'warning':
			return 'Warning';
		case 'info':
			return 'Info';
		default:
			return severity;
	}
}

export interface ListTroubleshootingQuery {
	search?: string;
	category?: string;
	device?: string;
	error_code?: string;
	page?: number;
	page_size?: number;
}

export async function listTroubleshootingGuides(
	q: ListTroubleshootingQuery = {}
): Promise<TroubleshootingListResult> {
	const params = new URLSearchParams();
	if (q.search) params.set('search', q.search);
	if (q.category && q.category !== '_all') params.set('category', q.category);
	if (q.device !== undefined && q.device !== '_all') params.set('device', q.device);
	if (q.error_code) params.set('error_code', q.error_code);
	if (q.page) params.set('page', String(q.page));
	if (q.page_size) params.set('page_size', String(q.page_size));
	const qs = params.toString();
	return request<TroubleshootingListResult>(`/api/troubleshooting${qs ? `?${qs}` : ''}`);
}

export async function getTroubleshootingGuide(id: number): Promise<TroubleshootingGuideDetail> {
	return request<TroubleshootingGuideDetail>(`/api/troubleshooting/${id}`);
}

export async function createTroubleshootingGuide(
	body: TroubleshootingGuideInput
): Promise<TroubleshootingGuideDetail> {
	return request<TroubleshootingGuideDetail>('/api/troubleshooting', {
		method: 'POST',
		body: JSON.stringify(body)
	});
}

export async function updateTroubleshootingGuide(
	id: number,
	body: TroubleshootingGuideInput
): Promise<TroubleshootingGuideDetail> {
	return request<TroubleshootingGuideDetail>(`/api/troubleshooting/${id}`, {
		method: 'PUT',
		body: JSON.stringify(body)
	});
}

export async function deleteTroubleshootingGuide(id: number): Promise<void> {
	await request<null>(`/api/troubleshooting/${id}`, { method: 'DELETE' });
}

/**
 * Finds sibling/alternative solutions for the same problem: guides that share the same
 * non-empty `error_code` or the same non-empty `function_name` (excluding `current.id`).
 * Guides matching the exact `error_code` are ranked first.
 */
export function findRelatedGuides(
	current: Pick<TroubleshootingGuideDetail, 'id' | 'error_code' | 'function_name'>,
	allItems: TroubleshootingGuideListItem[]
): TroubleshootingGuideListItem[] {
	const errorCode = current.error_code.trim().toUpperCase();
	const fnName = current.function_name.trim().toUpperCase();
	if (!errorCode && !fnName) return [];

	return allItems
		.filter((item) => {
			if (item.id === current.id) return false;
			const itemErr = item.error_code.trim().toUpperCase();
			const itemFn = item.function_name.trim().toUpperCase();
			return (errorCode !== '' && itemErr === errorCode) || (fnName !== '' && itemFn === fnName);
		})
		.sort((a, b) => {
			const aExactErr = errorCode !== '' && a.error_code.trim().toUpperCase() === errorCode ? 0 : 1;
			const bExactErr = errorCode !== '' && b.error_code.trim().toUpperCase() === errorCode ? 0 : 1;
			if (aExactErr !== bExactErr) return aExactErr - bExactErr;
			return a.id - b.id;
		});
}

/**
 * Formats a troubleshooting guide and any sibling solutions for the same problem/error_code
 * into a structured Markdown + JSON block ready to paste into an AI assistant or MCP prompt.
 */
export function formatTroubleshootingAiContext(
	guide: TroubleshootingGuideDetail,
	relatedGuides: TroubleshootingGuideListItem[] = []
): string {
	const lines: string[] = [
		`# MCV2 Troubleshooting Guide: ${guide.title}`,
		'',
		`- **Category**: ${categoryLabel(guide.category)} (\`${guide.category}\`)`,
		`- **Severity**: ${severityLabel(guide.severity)} (\`${guide.severity}\`)`,
		`- **Device Scope**: ${guide.device_name || 'Global'}`,
		`- **Error Code / Tag**: ${guide.error_code || 'N/A'}`,
		`- **Function / Flow**: ${guide.function_name || 'N/A'}`,
		'',
		'## Observed Symptom / Log Signature',
		guide.symptom,
		'',
		'## Root Cause',
		guide.root_cause || 'Not specified.',
		'',
		'## Step-by-Step Resolution'
	];

	for (const step of guide.steps) {
		lines.push(`### Step ${step.position}: ${step.title}`);
		lines.push(step.instruction);
		if (step.script_snippet.trim()) {
			lines.push('```ini');
			lines.push(step.script_snippet.trim());
			lines.push('```');
		}
		if (step.verification.trim()) {
			lines.push(`- **Expected Verification**: ${step.verification.trim()}`);
		}
		lines.push('');
	}

	if (relatedGuides.length > 0) {
		lines.push(`## Alternative Solutions for This Problem / Error Code (${relatedGuides.length})`);
		lines.push(
			'> Note: This error code or function can manifest in multiple ways. If the steps above do not resolve the issue, inspect these sibling guides:'
		);
		for (const sibling of relatedGuides) {
			lines.push(
				`- **[Guide #${sibling.id}] ${sibling.title}** (${sibling.error_code || sibling.function_name}, Device: ${sibling.device_name || 'Global'}) — Symptom: ${sibling.symptom} | Root Cause: ${sibling.root_cause}`
			);
		}
		lines.push('');
	}

	return lines.join('\n').trim();
}
