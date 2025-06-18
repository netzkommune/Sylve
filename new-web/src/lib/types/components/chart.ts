export interface AreaChartElement {
	field: string;
	label: string;
	color: string;
	data: Array<{
		date: Date;
		value: number;
	}>;
}
