import { page } from 'vitest/browser';
import { describe, expect, it } from 'vitest';
import { render } from 'vitest-browser-svelte';
import DialogShellFixture from '../support/DialogShellFixture.svelte';

describe('shared dialog chrome', () => {
	it('keeps long dialog content out of the header and title-cases the heading', async () => {
		await render(DialogShellFixture, { kind: 'dialog' });
		const dialog = page.getByRole('dialog', { name: 'Example of Test' });
		await expect.element(dialog).toBeVisible();
		const dialogElement = (await dialog.element()) as HTMLElement;
		const headingElement = dialogElement.querySelector<HTMLElement>('.dialog-heading')!;
		const bodyElement = dialogElement.querySelector<HTMLElement>('.dialog-body')!;
		expect(bodyElement).not.toBeNull();
		const headingBox = headingElement.getBoundingClientRect();
		const bodyBox = bodyElement.getBoundingClientRect();
		expect(bodyBox.y).toBeGreaterThanOrEqual(headingBox.y + headingBox.height - 0.5);
	});

	it('keeps long alert content between its title and actions', async () => {
		await render(DialogShellFixture, { kind: 'alert' });
		const alert = page.getByRole('alertdialog', { name: 'Delete Test Value' });
		const title = alert.getByRole('heading', { name: 'Delete Test Value' });
		const confirm = alert.getByRole('button', { name: 'Confirm' });
		await expect.element(alert).toBeVisible();
		const alertElement = (await alert.element()) as HTMLElement;
		const titleElement = (await title.element()) as HTMLElement;
		const bodyElement = alertElement.querySelector<HTMLElement>('.alert-dialog-body')!;
		const confirmElement = (await confirm.element()) as HTMLElement;
		expect(bodyElement).not.toBeNull();
		const titleBox = titleElement.getBoundingClientRect();
		const bodyBox = bodyElement.getBoundingClientRect();
		const confirmBox = confirmElement.getBoundingClientRect();
		expect(bodyBox.y).toBeGreaterThanOrEqual(titleBox.y + titleBox.height - 0.5);
		expect(confirmBox.y).toBeGreaterThanOrEqual(bodyBox.y + bodyBox.height - 0.5);
	});
});
