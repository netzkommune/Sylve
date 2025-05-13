import type { CellComponent, EmptyCallback, FormatterParams } from 'tabulator-tables';

export interface Row {
	id: number | string;
	[key: string]: any;
	children?: Row[];
}

export interface Column {
	field: string;
	title: string;
	visible?: boolean;
	formatter?:
		| ((cell: CellComponent, formatterParams: FormatterParams, onRendered: EmptyCallback) => void)
		| string;
}

export type ExpandedRows = Record<number, boolean>;
