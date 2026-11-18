import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Class combiner for the UI wrappers (shadcn-svelte convention): merges
// conditional classes and dedupes conflicting Tailwind utilities so a
// caller-passed class (e.g. w-28) overrides the wrapper default (w-full).
export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

// shadcn-svelte utility types imported by registry components: strip
// bits-ui child/children props and attach the element-ref prop pattern.
export type WithoutChild<T> = T extends { child?: unknown } ? Omit<T, 'child'> : T;
export type WithoutChildren<T> = T extends { children?: unknown } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
