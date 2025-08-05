import type { Column, Row } from '$lib/types/components/tree-table';
import type { NetworkObject } from '$lib/types/network/object';
import type { SwitchList } from '$lib/types/network/switch';
import type { CellComponent } from 'tabulator-tables';
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
			title: 'ID'
		},
		{
			field: 'name',
			title: 'Name',
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
			title: 'MTU'
		},
		{
			field: 'vlan',
			title: 'VLAN'
		},
		{
			field: 'ipv4',
			title: 'IPv4',
			formatter: (cell: CellComponent) => {
				const row = cell.getRow();
				const data = row.getData();
				const value = cell.getValue();

				if (value === '-' && data.dhcp) {
					return 'DHCP';
				}

				const addressObj = data.addressObj as NetworkObject;

				if (data.addressObj) {
					if (addressObj && addressObj.entries) {
						return addressObj.entries[0].value || '-';
					}
				}

				return value || '-';
			}
		},
		{
			field: 'ipv6',
			title: 'IPv6',
			formatter: (cell: CellComponent) => {
				const row = cell.getRow();
				const data = row.getData();
				const value = cell.getValue();

				if (value === '-' && data.slaac) {
					return 'SLAAC';
				}

				const addressObj = data.address6Obj as NetworkObject;

				if (data.address6Obj) {
					if (addressObj && addressObj.entries) {
						return addressObj.entries[0].value || '-';
					}
				}

				return value || '-';
			}
		},
		{
			field: 'private',
			title: 'Private',
			visible: false
		},
		{
			field: 'dhcp',
			title: 'DHCP',
			visible: false
		},
		{
			field: 'disableIPv6',
			title: 'Disable IPv6',
			visible: false
		},
		{
			field: 'slaac',
			title: 'SLAAC',
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
				addressObj: sw.addressObj || '-',
				address6Obj: sw.address6Obj || '-',
				ports: sw.ports,
				private: sw.private,
				portsOnly: portsOnly,
				dhcp: sw.dhcp || false,
				disableIPv6: sw.disableIPv6 || false,
				slaac: sw.slaac || false
			});
		}
	}

	return {
		rows: rows,
		columns: columns
	};
}
