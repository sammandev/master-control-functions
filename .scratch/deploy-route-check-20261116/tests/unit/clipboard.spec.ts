import { afterEach, expect, test, vi } from 'vitest';
import { toast } from 'svelte-sonner';
import { copyText } from '../../src/lib/clipboard';

vi.mock('svelte-sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

afterEach(() => {
	vi.unstubAllGlobals();
	vi.clearAllMocks();
});

function legacyClipboard() {
	const container = { appendChild: vi.fn() };
	class FocusTarget {
		focus = vi.fn();
		closest = vi.fn().mockReturnValue(container);
	}
	const trigger = new FocusTarget();
	const field = {
		value: '',
		readOnly: false,
		tabIndex: 0,
		style: { cssText: '' },
		focus: vi.fn(),
		select: vi.fn(),
		remove: vi.fn()
	};
	const command = vi.fn().mockReturnValue(true);
	vi.stubGlobal('HTMLElement', FocusTarget);
	vi.stubGlobal('navigator', {});
	vi.stubGlobal('document', {
		activeElement: trigger,
		body: { appendChild: vi.fn() },
		createElement: vi.fn().mockReturnValue(field),
		execCommand: command
	});
	return { container, trigger, field, command };
}

test('copies exact text through the modern API and shows success only after completion', async () => {
	let complete: (() => void) | undefined;
	const writeText = vi.fn(
		() =>
			new Promise<void>((resolve) => {
				complete = resolve;
			})
	);
	vi.stubGlobal('navigator', { clipboard: { writeText } });
	const pending = copyText(' DLL|TEST\r\nNEXT|NULL ');
	expect(toast.success).not.toHaveBeenCalled();
	complete?.();
	expect(await pending).toBe(true);
	expect(writeText).toHaveBeenCalledWith(' DLL|TEST\r\nNEXT|NULL ');
	expect(toast.success).toHaveBeenCalledWith('Copied to clipboard');
	expect(toast.error).not.toHaveBeenCalled();
});

test('copies without navigator.clipboard inside the active modal and restores focus', async () => {
	const { container, trigger, field, command } = legacyClipboard();
	expect(await copyText('BUILT|OPEN\nDLL|TX')).toBe(true);
	expect(field.value).toBe('BUILT|OPEN\nDLL|TX');
	expect(container.appendChild).toHaveBeenCalledWith(field);
	expect(trigger.closest).toHaveBeenCalledWith('[role="dialog"], [role="alertdialog"]');
	expect(field.select).toHaveBeenCalledOnce();
	expect(command).toHaveBeenCalledWith('copy');
	expect(field.remove).toHaveBeenCalledOnce();
	expect(trigger.focus).toHaveBeenCalledWith({ preventScroll: true });
	expect(toast.success).toHaveBeenCalledOnce();
});

test('falls back when modern clipboard permission is rejected', async () => {
	const { command } = legacyClipboard();
	vi.stubGlobal('navigator', {
		clipboard: { writeText: vi.fn().mockRejectedValue(new Error('denied')) }
	});
	expect(await copyText('SCRIPT')).toBe(true);
	expect(command).toHaveBeenCalledOnce();
	expect(toast.success).toHaveBeenCalledOnce();
});

test('does not report success when the fallback fails', async () => {
	const { command, field, trigger } = legacyClipboard();
	command.mockReturnValue(false);
	expect(await copyText('SCRIPT')).toBe(false);
	expect(toast.success).not.toHaveBeenCalled();
	expect(toast.error).toHaveBeenCalledWith('Copy failed. Select the text and copy it manually.');
	expect(field.remove).toHaveBeenCalledOnce();
	expect(trigger.focus).toHaveBeenCalledOnce();
});

test('cleans up the fallback even when the browser throws', async () => {
	const { command, field } = legacyClipboard();
	command.mockImplementation(() => {
		throw new Error('blocked');
	});
	expect(await copyText('SCRIPT')).toBe(false);
	expect(field.remove).toHaveBeenCalledOnce();
	expect(toast.error).toHaveBeenCalledOnce();
	expect(toast.success).not.toHaveBeenCalled();
});
