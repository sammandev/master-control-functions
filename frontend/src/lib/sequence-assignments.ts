export function parseAssignments(
	text: string
): { values: Record<string, string> } | { error: string } {
	const values = new Map<string, string>();
	for (const rawLine of text.split('\n')) {
		const line = rawLine.trim();
		if (!line) continue;
		const eq = line.indexOf('=');
		if (eq <= 0) {
			return { error: `Use name=value on each line: "${line}"` };
		}
		const name = line.slice(0, eq).trim();
		if (values.has(name)) {
			return { error: `Duplicate assignment name: "${name}"` };
		}
		values.set(name, line.slice(eq + 1).trim());
	}
	return { values: Object.fromEntries(values) };
}
