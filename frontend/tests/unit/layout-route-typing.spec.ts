import { readFileSync } from 'node:fs';
import ts from 'typescript';
import { expect, test } from 'vitest';

test('layout route comparisons type-check without optional demo routes', () => {
	const layout = readFileSync(new URL('../../src/routes/+layout.svelte', import.meta.url), 'utf8');
	const comparisons =
		layout.match(
			/page\.route\.id\s*===\s*['"][^'"]+['"]|isOptionalRoute\(page\.route\.id,\s*['"][^'"]+['"]\)/g
		) ?? [];
	const helper = layout.match(/function isOptionalRoute\([^)]*\): boolean\s*\{[^}]*\}/)?.[0] ?? '';
	for (const route of ['/demo/device-prototype', '/demo/functions-prototype']) {
		expect(comparisons.some((comparison) => comparison.includes(route))).toBe(true);
	}
	const source = `declare const page: { route: { id: '/' | '/devices' | '/functions' | null } };\n${helper}\n${comparisons.map((comparison) => `void (${comparison});`).join('\n')}`;
	const filename = 'optional-route-check.ts';
	const options: ts.CompilerOptions = { noEmit: true, skipLibCheck: true, strict: true, types: [] };
	const host = ts.createCompilerHost(options);
	const originalGetSourceFile = host.getSourceFile.bind(host);
	host.getSourceFile = (name, languageVersion, onError, shouldCreateNewSourceFile) =>
		name === filename
			? ts.createSourceFile(name, source, languageVersion)
			: originalGetSourceFile(name, languageVersion, onError, shouldCreateNewSourceFile);
	const program = ts.createProgram([filename], options, host);
	const errors = ts
		.getPreEmitDiagnostics(program)
		.filter((diagnostic) => diagnostic.file?.fileName === filename);
	expect(
		errors.map((diagnostic) => ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'))
	).toEqual([]);
});
