import { toast } from 'svelte-sonner';

function copySelection(text: string): boolean {
	const previousFocus = document.activeElement;
	const field = document.createElement('textarea');
	field.value = text;
	field.readOnly = true;
	field.tabIndex = -1;
	field.style.cssText = 'position:fixed;top:0;left:0;width:1px;height:1px;opacity:0;';
	const container =
		previousFocus?.closest('[role="dialog"], [role="alertdialog"]') ?? document.body;
	try {
		container.appendChild(field);
		field.focus({ preventScroll: true });
		field.select();
		return typeof document.execCommand === 'function' && document.execCommand('copy');
	} finally {
		field.remove();
		if (previousFocus instanceof HTMLElement) previousFocus.focus({ preventScroll: true });
	}
}

export async function copyText(text: string): Promise<boolean> {
	let copied = false;
	try {
		if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
			await navigator.clipboard.writeText(text);
			copied = true;
		}
	} catch {
		copied = false;
	}
	if (!copied) {
		try {
			copied = typeof document !== 'undefined' && copySelection(text);
		} catch {
			copied = false;
		}
	}
	if (copied) toast.success('Copied to clipboard');
	else toast.error('Copy failed. Select the text and copy it manually.');
	return copied;
}
