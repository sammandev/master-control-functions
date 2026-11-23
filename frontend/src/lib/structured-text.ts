export type StructuredTextBlock =
	| { kind: 'paragraph'; text: string }
	| { kind: 'unordered-list'; items: string[] }
	| { kind: 'ordered-list'; items: string[] };

const bulletPattern = /^\s*[-*+]\s+(.+?)\s*$/;
const orderedPattern = /^\s*\d+[.)]\s+(.+?)\s*$/;
const listPrefixPattern = /^([\t ]*)([-*+]|(\d+)([.)]))[\t ]+/;
const emptyListMarkerPattern = /^(?:[-*+]|\d+[.)])?$/;

export function trimEmptyListItems(value: string): string {
	const lines = value.split('\n');
	while (lines.length && emptyListMarkerPattern.test(lines[lines.length - 1].trim())) {
		lines.pop();
	}
	return lines.join('\n').trim();
}

export function parseStructuredText(value: string): StructuredTextBlock[] {
	const blocks: StructuredTextBlock[] = [];
	let paragraph: string[] = [];
	let listKind: 'unordered-list' | 'ordered-list' | null = null;
	let listItems: string[] = [];

	const flushParagraph = () => {
		if (paragraph.length > 0) {
			blocks.push({ kind: 'paragraph', text: paragraph.join('\n') });
			paragraph = [];
		}
	};
	const flushList = () => {
		if (listKind && listItems.length > 0) {
			blocks.push({ kind: listKind, items: listItems });
			listItems = [];
		}
		listKind = null;
	};

	for (const line of value.replaceAll('\r\n', '\n').split('\n')) {
		if (!line.trim()) {
			flushParagraph();
			flushList();
			continue;
		}

		const bullet = line.match(bulletPattern);
		if (bullet) {
			flushParagraph();
			if (listKind !== 'unordered-list') flushList();
			listKind = 'unordered-list';
			listItems.push(bullet[1]);
			continue;
		}

		const ordered = line.match(orderedPattern);
		if (ordered) {
			flushParagraph();
			if (listKind !== 'ordered-list') flushList();
			listKind = 'ordered-list';
			listItems.push(ordered[1]);
			continue;
		}

		flushList();
		paragraph.push(line);
	}

	flushParagraph();
	flushList();
	return blocks;
}

export function getListEdit(
	value: string,
	caret: number,
	key: 'Enter' | 'Backspace'
): { start: number; end: number; text: string } | undefined {
	const start = value.slice(0, caret).lastIndexOf('\n') + 1;
	const nextBreak = value.indexOf('\n', caret);
	const end = nextBreak === -1 ? value.length : nextBreak;
	const line = value.slice(start, end);
	const prefix = listPrefixPattern.exec(line);
	if (!prefix || caret < start + prefix[0].length) return;

	if (!line.slice(prefix[0].length).trim()) {
		return { start, end, text: key === 'Enter' && start > 0 ? '\n' : '' };
	}
	if (key === 'Backspace') return;

	const [, indent, marker, number, delimiter] = prefix;
	const nextMarker = number ? `${BigInt(number) + 1n}${delimiter}` : marker;
	return { start: caret, end: caret, text: `\n${indent}${nextMarker} ` };
}
