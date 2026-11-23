export function parseAssignments(
	text: string
): { values: Record<string, string> } | { error: string } {
	const values: Record<string, string> = {};
	for (const rawLine of text.split('\n')) {
		const line = rawLine.trim();
		if (!line) continue;
		const eq = line.indexOf('=');
		if (eq <= 0) {
			return { error: `Use name=value on each line: "${line}"` };
		}
		values[line.slice(0, eq).trim()] = line.slice(eq + 1).trim();
	}
	return { values };
}
