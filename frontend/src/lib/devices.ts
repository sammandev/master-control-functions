// Typed client for the device catalog. Device names and Function
// interfaces share one catalog (blank = Global); the reserved
// path segment "_global" addresses the blank device. HTTP plumbing from the
// shared client.
import { request } from './client';

export const GLOBAL_DEVICE_SEGMENT = '_global';

// URL segment for a device value: the blank device maps to the reserved
// segment; everything else is percent-encoded ("TEST PROGRAM" → TEST%20PROGRAM).
export function devicePath(deviceName: string): string {
	return deviceName === '' ? GLOBAL_DEVICE_SEGMENT : encodeURIComponent(deviceName);
}

// Link into the device browser: the route href for a device value.
export function deviceHref(deviceName: string): string {
	return `/devices/${devicePath(deviceName)}`;
}

// Shared styling for the app's accent links into (and inside) the device
// browser — deep links from the devices index, links panel, and dictionary
// badges, plus the function-name links on a device's Functions tab. Accent
// text, underline until hover, and a visible focus ring. The count-cell
// variant drops font-medium so the number keeps the table's weight.
const focusRingClass =
	'outline-none focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-ring';
export const deviceLinkClass = `font-medium text-primary underline hover:no-underline ${focusRingClass}`;
export const deviceCountLinkClass = `text-primary underline hover:no-underline ${focusRingClass}`;
// Badge-wrapped variant (dictionary pages): the badge carries the color,
// so the link only contributes the focus ring.
export const deviceBadgeLinkClass = `no-underline ${focusRingClass}`;

export type DeviceSource = 'BUILT-IN' | 'DLL';

export interface DeviceSummary {
	device_name: string;
	source?: DeviceSource | null;
	sources?: DeviceSource[];
	functions: number;
	parameters: number;
	criteria: number;
	examples: number;
}

export function deviceTypeLabel(device: Pick<DeviceSummary, 'source' | 'sources'>): string {
	return device.source ?? (device.sources?.length ? device.sources.join(' / ') : 'Unassigned');
}

export interface DeviceListResult {
	items: DeviceSummary[];
	total: number;
}

export async function listDevices(): Promise<DeviceListResult> {
	return request<DeviceListResult>('/api/devices');
}

// A registered Device is web-native (ADR-0005 precedent, ADR-0010): it
// carries registry metadata and is not exported.
export interface RegisteredDevice {
	device_name: string;
	source?: DeviceSource | null;
	created_at: string;
	updated_at: string;
}

export async function createDevice(
	deviceName: string,
	source?: DeviceSource
): Promise<RegisteredDevice> {
	return request<RegisteredDevice>('/api/devices', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ device_name: deviceName, source })
	});
}

export async function deleteDevice(deviceName: string, confirmation: string): Promise<void> {
	await request<void>(`/api/devices/${devicePath(deviceName)}`, {
		method: 'DELETE',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ confirmation })
	});
}

export async function renameDevice(deviceName: string, newName: string): Promise<RegisteredDevice> {
	return request<RegisteredDevice>(`/api/devices/${devicePath(deviceName)}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ device_name: newName })
	});
}

export interface DeviceFunctionRow {
	id: number;
	source: string;
	interface: string;
	name: string;
	description: string | null;
	convention: number | null;
	category: string | null;
	execution_role: string | null;
	execution_phase: string | null;
	updated_at: string;
}

export interface DeviceDefinitionRow {
	id: number;
	name: string;
	description: string;
	type: string;
	unit: string;
	default_value: string;
	allowed_values: string;
	cardinality: string;
	value_pattern: string;
	required: boolean;
	device_name: string;
}

export interface DeviceListPage<T> {
	items: T[];
	total: number;
	page: number;
	page_size: number;
}

export interface DeviceDefinitionsResult extends DeviceListPage<DeviceDefinitionRow> {
	// Global definitions the device's links resolve to under
	// exact-scope-then-global (ADR-0009); returned in full.
	global_items: DeviceDefinitionRow[];
	global_total: number;
}

export type DeviceDefinitionKind = 'parameters' | 'criteria';

export async function listDeviceFunctions(
	deviceName: string,
	page = 1,
	pageSize = 20
): Promise<DeviceListPage<DeviceFunctionRow>> {
	return request<DeviceListPage<DeviceFunctionRow>>(
		`/api/devices/${devicePath(deviceName)}/functions?page=${page}&page_size=${pageSize}`
	);
}

export async function listDeviceDefinitions(
	kind: DeviceDefinitionKind,
	deviceName: string,
	page = 1,
	pageSize = 20
): Promise<DeviceDefinitionsResult> {
	return request<DeviceDefinitionsResult>(
		`/api/devices/${devicePath(deviceName)}/${kind}?page=${page}&page_size=${pageSize}`
	);
}
