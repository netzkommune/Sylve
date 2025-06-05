import type { Column, Row } from '$lib/types/components/tree-table';
import type { SwitchList } from '$lib/types/network/switch';
import type { CellComponent } from 'tabulator-tables';
import { getTranslation } from '../i18n';
import { capitalizeFirstLetter } from '../string';
import { renderWithIcon } from '../table';

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
			title: getTranslation('common.ID', 'ID')
		},
		{
			field: 'name',
			title: capitalizeFirstLetter(getTranslation('common.name', 'Name')),
			formatter(cell: CellComponent) {
				const value = cell.getValue();
				const row = cell.getRow();
				const data = row.getData();
				const pSw = data.private || false;

				if (pSw) {
					return renderWithIcon('material-symbols-light:private-connectivity-outline', value);
				}

				return renderWithIcon('material-symbols:public', value);
			}
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
			title: getTranslation('network.MTU', 'MTU')
		},
		{
			field: 'vlan',
			title: getTranslation('network.VLAN', 'VLAN')
		},
		{
			field: 'ipv4',
			title: getTranslation('network.ipv4', 'IPv4'),
			formatter: (cell: CellComponent) => {
				const row = cell.getRow();
				const data = row.getData();
				const value = cell.getValue();
				if (value === '-' && data.dhcp) {
					return 'DHCP';
				}
			}
		},
		{
			field: 'ipv6',
			title: getTranslation('network.ipv6', 'IPv6')
		},
		{
			field: 'private',
			title: getTranslation('common.private', 'Private'),
			visible: false
		},
		{
			field: 'dhcp',
			title: getTranslation('network.DHCP', 'DHCP'),
			visible: false
		}
	];

	const rows: Row[] = [];

	if (switches && switches[type]) {
		for (const sw of switches[type]) {
			const portsOnly =
				sw.ports?.map((port) => {
					return port.name;
				}) || [];

			rows.push({
				id: sw.id,
				name: sw.name,
				mtu: sw.mtu,
				vlan: sw.vlan || '-',
				ipv4: sw.address || '-',
				ipv6: sw.address6 || '-',
				ports: sw.ports,
				private: sw.private,
				portsOnly: portsOnly,
				dhcp: sw.dhcp || false
			});
		}
	}

	return {
		rows: rows,
		columns: columns
	};
}
