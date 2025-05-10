import type { Column, Row } from '$lib/types/components/tree-table';
import type { Iface } from '$lib/types/network/iface';
import type { CellComponent } from 'tabulator-tables';
import { generateNumberFromString } from '../numbers';

export function generateTableData(interfaces: Iface[]): {
	rows: Row[];
	columns: Column[];
} {
	const columns: Column[] = [
		{
			field: 'id',
			title: 'ID',
			visible: false
		},
		{
			field: 'name',
			title: 'Name'
		},
		{
			field: 'description',
			title: 'Description'
		},
		{
			field: 'ether',
			title: 'MAC Address',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value || '-';
			}
		},
		{
			field: 'metric',
			title: 'Metric'
		},
		{
			field: 'mtu',
			title: 'MTU'
		},
		{
			field: 'media',
			title: 'Status',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				const status = value?.status || '-';
				if (status === 'active') {
					return 'Active';
				}

				return status;
			}
		}
	];

	const rows: Row[] = [];
	for (const iface of interfaces) {
		const row: Row = {
			id: generateNumberFromString(iface.ether),
			ether: iface.ether,
			name: iface.name,
			description: iface.description,
			metric: iface.metric,
			mtu: iface.mtu,
			media: iface.media
		};

		rows.push(row);
	}

	return {
		rows,
		columns: columns
	};
}
