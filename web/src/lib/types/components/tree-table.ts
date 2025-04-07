export interface Row {
	id: number;
	[key: string]: any;
	children?: Row[];
}

export interface Column {
	key: string;
	label: string;
}

export type ExpandedRows = Record<number, boolean>;
