import type { Column, Row } from '$lib/types/components/tree-table';
import type { SwitchList } from '$lib/types/network/switch';
import type { CellComponent } from 'tabulator-tables';

export function generateTableData(
	switches: SwitchList | undefined,
	type: 'standard' = 'standard'
): {
	rows: Row[];
	columns: Column[];
} {
	const columns: Column[] = [
		{
			field: 'id',
			visible: false,
			title: 'ID'
		},
		{
			field: 'name',
			title: 'Name'
		},
		{
			field: 'ports',
			title: 'Ports',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				if (value && Array.isArray(value) && value.length > 0) {
					return value.map((port) => `<span>${port.name}</span>`).join(', ');
				}

				return '-';
			}
		},
		{
			field: 'mtu',
			title: 'MTU'
		},
		{
			field: 'vlan',
			title: 'VLAN'
		},
		{
			field: 'ipv4',
			title: 'IPv4'
		},
		{
			field: 'ipv6',
			title: 'IPv6'
		}
	];

	const rows: Row[] = [];

	if (switches && switches[type]) {
		for (const sw of switches[type]) {
			console.log(sw);
			rows.push({
				id: sw.id,
				name: sw.name,
				mtu: sw.mtu,
				vlan: sw.vlan || '-',
				ipv4: sw.address || '-',
				ipv6: sw.address6 || '-',
				ports: sw.ports
			});
		}
	}

	return {
		rows: rows,
		columns: columns
	};
}
