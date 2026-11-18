import { expect, test } from 'vitest';
import { hello_world } from '#lib/paraglide/messages.js';
import { deLocalizeUrl, getTextDirection, locales, localizeUrl } from '#lib/paraglide/runtime.js';
import { reroute } from '../../src/hooks';

test.each(locales)('compiles interpolated messages for %s', (locale) => {
	expect(hello_world({ name: 'Catalog' }, { locale })).toBe(`Hello, Catalog from ${locale}!`);
	expect(getTextDirection(locale)).toBe('ltr');
});

test.each(locales)('preserves scoped routes, queries and fragments for %s', async (locale) => {
	const url = new URL(
		'https://example.test/devices/WIFI%20TEST?tab=functions&function=1#parameters'
	);
	const localized = localizeUrl(url, { locale });
	const prefix = locale === 'en' ? '' : `${locale}/`;

	expect(localized.pathname).toBe(`/${prefix}devices/WIFI%20TEST`);
	expect(localized.search).toBe(url.search);
	expect(localized.hash).toBe(url.hash);
	expect(deLocalizeUrl(localized).href).toBe(url.href);
	expect(await reroute({ url: localized, fetch })).toBe(url.pathname);
});
