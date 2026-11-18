import { page } from 'vitest/browser';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import Devices from '../../src/routes/devices/+page.svelte';

const name = 'TEST DEVICE';
let deleteStatus: number;
let deleted: boolean;
let deleteRequests: string[];
let finishDelete: (() => void) | undefined;

beforeEach(() => {
	deleteStatus = 204;
	deleted = false;
	deleteRequests = [];
	finishDelete = undefined;
	vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
		if (String(input) === '/api/devices' && !init?.method) {
			return Response.json({
				items: [
					{ device_name: '', functions: 1, parameters: 2, criteria: 1, examples: 1 },
					...(deleted
						? []
						: [{ device_name: name, functions: 2, parameters: 3, criteria: 1, examples: 1 }])
				],
				total: deleted ? 1 : 2
			});
		}
		if (String(input) === '/api/devices/TEST%20DEVICE' && init?.method === 'DELETE') {
			deleteRequests.push(String(init.body));
			if (deleteStatus === 202) {
				await new Promise<void>((resolve) => {
					finishDelete = resolve;
				});
			}
			if (deleteStatus === 500) {
				return Response.json({ error: 'Deletion failed; try again' }, { status: 500 });
			}
			deleted = true;
			return new Response(null, { status: 204 });
		}
		throw new Error(`Unexpected request: ${init?.method ?? 'GET'} ${String(input)}`);
	});
});

afterEach(() => vi.restoreAllMocks());

it('requires exact text, resets after Cancel, and deletes a populated device', async () => {
	render(Devices);
	const remove = page.getByRole('button', { name: `Delete ${name}`, exact: true });
	await expect.element(remove).toBeVisible();
	await expect
		.element(page.getByRole('region', { name: 'Global definitions' }).getByRole('button'))
		.not.toBeInTheDocument();
	await remove.click();
	const dialog = page.getByRole('dialog', { name: `Delete ${name}?` });
	const confirm = dialog.getByRole('button', { name: 'Delete', exact: true });
	const input = dialog.getByRole('textbox', { name: `Type DELETE ${name} to confirm` });
	await expect.element(confirm).toBeDisabled();
	for (const text of ['DELETE', 'delete TEST DEVICE', 'DELETE OTHER', `DELETE ${name} `]) {
		await input.fill(text);
		await expect.element(confirm).toBeDisabled();
	}
	expect(deleteRequests).toEqual([]);
	await input.fill(`DELETE ${name}`);
	await expect.element(confirm).toBeEnabled();
	await dialog.getByRole('button', { name: 'Cancel' }).click();
	expect(deleteRequests).toEqual([]);
	await remove.click();
	await expect.element(input).toHaveValue('');
	await expect.element(confirm).toBeDisabled();
	await input.fill(`DELETE ${name}`);
	await confirm.click();
	await expect.element(dialog).not.toBeInTheDocument();
	await expect.element(page.getByRole('article', { name, exact: true })).not.toBeInTheDocument();
	expect(deleteRequests).toEqual([JSON.stringify({ confirmation: `DELETE ${name}` })]);
});

it('keeps API errors inline and permits retry without losing the device', async () => {
	deleteStatus = 500;
	render(Devices);
	await page.getByRole('button', { name: `Delete ${name}`, exact: true }).click();
	const dialog = page.getByRole('dialog', { name: `Delete ${name}?` });
	await dialog
		.getByRole('textbox', { name: `Type DELETE ${name} to confirm` })
		.fill(`DELETE ${name}`);
	await dialog.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect.element(dialog.getByRole('alert')).toHaveTextContent('Deletion failed; try again');
	expect(deleted).toBe(false);
	deleteStatus = 204;
	await dialog.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect.element(dialog).not.toBeInTheDocument();
	expect(deleteRequests).toHaveLength(2);
});

it('prevents duplicate submissions and dismissal while deletion is pending', async () => {
	deleteStatus = 202;
	render(Devices);
	await page.getByRole('button', { name: `Delete ${name}`, exact: true }).click();
	const dialog = page.getByRole('dialog', { name: `Delete ${name}?` });
	const input = dialog.getByRole('textbox', { name: `Type DELETE ${name} to confirm` });
	await input.fill(`DELETE ${name}`);
	await dialog.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect.element(dialog.getByRole('button', { name: 'Deleting...' })).toBeDisabled();
	await expect.element(dialog.getByRole('button', { name: 'Cancel' })).toBeDisabled();
	await expect.element(input).toBeDisabled();
	expect(deleteRequests).toHaveLength(1);
	finishDelete?.();
	await expect.element(dialog).not.toBeInTheDocument();
});
