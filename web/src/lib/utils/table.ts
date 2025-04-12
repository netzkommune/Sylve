import type { RowComponent, Tabulator } from 'tabulator-tables';

export function deselectAllRows(table: Tabulator | null) {
	if (table) {
		table.getRows().forEach((row) => {
			console.log(row);
			row.deselect();
		});
	}
}

export function selectOneRow(table: Tabulator | null, row: RowComponent) {
	if (table) {
		deselectAllRows(table);
		row.select();
	}
}
