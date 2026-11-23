export type StructuredTextBlock =
	| { kind: 'paragraph'; text: string }
	| { kind: 'unordered-list'; items: string[] }
	| { kind: 'ordered-list'; items: string[] };

const bulletPattern = /^\s*[-*+]\s+(.+?)\s*$/;
const orderedPattern = /^\s*\d+[.)]\s+(.+?)\s*$/;
const anyListPattern = /^(\s*)(?:[-*+]\s+|\d+[.)]\s+)/;

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

export interface ListMarkupResult {
	value: string;
	selectionStart: number;
	selectionEnd: number;
}

export function applyListMarkup(
	value: string,
	selectionStart: number,
	selectionEnd: number,
	kind: 'bullet' | 'ordered'
): ListMarkupResult {
	const lineStart = value.lastIndexOf('\n', Math.max(0, selectionStart - 1)) + 1;
	const effectiveEnd =
		selectionEnd > selectionStart && value[selectionEnd - 1] === '\n'
			? selectionEnd - 1
			: selectionEnd;
	const nextLineBreak = value.indexOf('\n', effectiveEnd);
	const lineEnd = nextLineBreak === -1 ? value.length : nextLineBreak;
	const lines = value.slice(lineStart, lineEnd).split('\n');
	const markerMatches = (line: string) =>
		kind === 'bullet' ? /^\s*[-*+]\s+/.test(line) : /^\s*\d+[.)]\s+/.test(line);
	const allMarked = lines.every((line) => !line.trim() || markerMatches(line));

	const transformed = lines.map((line, index) => {
		if (!line.trim()) return line;
		const indentation = line.match(/^\s*/)?.[0] ?? '';
		if (allMarked) {
			return line.replace(anyListPattern, '$1');
		}
		const content = line.replace(anyListPattern, '$1').slice(indentation.length);
		const marker = kind === 'bullet' ? '- ' : `${index + 1}. `;
		return `${indentation}${marker}${content}`;
	});
	const replacement = transformed.join('\n');
	return {
		value: value.slice(0, lineStart) + replacement + value.slice(lineEnd),
		selectionStart: lineStart,
		selectionEnd: lineStart + replacement.length
	};
}
