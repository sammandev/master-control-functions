import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Class combiner for the UI wrappers (shadcn-svelte convention): merges
// conditional classes and dedupes conflicting Tailwind utilities so a
// caller-passed class (e.g. w-28) overrides the wrapper default (w-full).
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

const dialogMinorWords = new Set([
	'a',
	'an',
	'and',
	'as',
	'at',
	'but',
	'by',
	'for',
	'in',
	'of',
	'on',
	'or',
	'the',
	'to',
	'via',
	'vs'
]);

// Dialog headings use title case while preserving identifiers and acronyms.
export function dialogTitleCase(title: string): string {
	const parts = title.split(/(\s+)/);
	const wordIndexes = parts
		.map((part, index) => (/\p{L}|\p{N}/u.test(part) ? index : -1))
		.filter((index) => index >= 0);
	const firstWord = wordIndexes[0];
	const lastWord = wordIndexes.at(-1);

	return parts
		.map((part, index) => {
			const match = part.match(/^([^\p{L}\p{N}]*)([\p{L}\p{N}_'-]+)(.*)$/u);
			if (!match) return part;
			const [, prefix, word, suffix] = match;
			const lower = word.toLocaleLowerCase();
			const isIdentifier =
				word.includes('_') || /[A-Z].*[A-Z]/.test(word) || /\p{Ll}\p{Lu}/u.test(word);
			const isMinor = dialogMinorWords.has(lower);
			if (isIdentifier || (isMinor && index !== firstWord && index !== lastWord)) return part;
			return `${prefix}${lower.replace(/^\p{L}/u, (letter) => letter.toLocaleUpperCase())}${suffix}`;
		})
		.join('');
}

// shadcn-svelte utility types imported by registry components: strip
// bits-ui child/children props and attach the element-ref prop pattern.
export type WithoutChild<T> = T extends { child?: unknown } ? Omit<T, 'child'> : T;
export type WithoutChildren<T> = T extends { children?: unknown } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
