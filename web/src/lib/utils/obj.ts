export function deepDiff(
	obj1: Record<string, any>,
	obj2: Record<string, any>,
	path = ''
): { path: string; from: any; to: any }[] {
	const changes = [];

	for (const key of new Set([...Object.keys(obj1 || {}), ...Object.keys(obj2 || {})])) {
		const fullPath = path ? `${path}.${key}` : key;
		const val1 = obj1?.[key];
		const val2 = obj2?.[key];

		if (typeof val1 === 'object' && typeof val2 === 'object' && val1 && val2) {
			changes.push(...deepDiff(val1, val2, fullPath));
		} else if (val1 !== val2) {
			changes.push({ path: fullPath, from: val1, to: val2 });
		}
	}

	return changes;
}
