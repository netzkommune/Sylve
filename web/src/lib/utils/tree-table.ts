import type { Row } from '$lib/types/components/tree-table';

export function cleanChildren(row: Row): Row {
	let newRow = { ...row };

	if (Array.isArray(newRow.children)) {
		let cleanedChildren = newRow.children.map(cleanChildren).filter(Boolean);

		if (cleanedChildren.length > 0) {
			newRow.children = cleanedChildren;
		} else {
			delete newRow.children;
		}
	}

	return newRow;
}
