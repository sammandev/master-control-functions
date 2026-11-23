export function createLatestRequestGuard() {
	let current = 0;
	return {
		begin: () => ++current,
		isCurrent: (request: number) => request === current
	};
}
