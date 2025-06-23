import type { Column, Row } from '$lib/types/components/tree-table';
import type { Iface } from '$lib/types/network/iface';
import type { CellComponent } from 'tabulator-tables';
import { getTranslation } from '../i18n';
import { generateNumberFromString } from '../numbers';
import { capitalizeFirstLetter } from '../string';
import { renderWithIcon } from '../table';

export function generateTableData(
	columns: Column[],
	interfaces: Iface[]
): {
	rows: Row[];
	columns: Column[];
} {
	const rows: Row[] = [];
	for (const iface of interfaces) {
		let isBridge = false;
		if (iface.groups) {
			if (iface.groups.includes('bridge')) {
				isBridge = true;
				iface.model = 'Bridge';
			}
		}

		// TODO: Skip sylve created VLANs for now
		if (iface.description.startsWith('svm-vlan')) {
			continue;
		}

		const row: Row = {
			id: generateNumberFromString(iface.ether),
			ether: iface.ether,
			name: iface.name,
			model: iface.model,
			description: iface.description,
			metric: iface.metric,
			mtu: iface.mtu,
			media: iface.media,
			isBridge: isBridge
		};

		rows.push(row);
	}

	return {
		rows,
		columns: columns
	};
}

export function getCleanIfaceData(iface: Iface): { [key: string | number]: any } {
	if (iface.groups) {
		if (iface.groups.includes('bridge')) {
			iface.model = 'Bridge';
			iface.name = `${iface.description} (${iface.name})`;
		}
	}

	const obj = {
		[capitalizeFirstLetter(getTranslation('common.name', 'Name'))]: iface.name,
		[capitalizeFirstLetter(getTranslation('common.description', 'Description'))]:
			iface.description || '-',
		[capitalizeFirstLetter(getTranslation('network.model', 'Model'))]: iface.model
			? iface.model
			: '-',
		[capitalizeFirstLetter(getTranslation('network.mac_address', 'MAC Address'))]:
			iface.ether || '-',
		[capitalizeFirstLetter(getTranslation('network.mtu', 'MTU'))]: iface.mtu,
		[capitalizeFirstLetter(getTranslation('network.metric', 'Metric'))]: iface.metric,
		[capitalizeFirstLetter(getTranslation('network.flags', 'Flags'))]: {
			[capitalizeFirstLetter(getTranslation('common.raw', 'Raw'))]: iface.flags.raw,
			[capitalizeFirstLetter(getTranslation('common.description', 'Description'))]:
				iface.flags.desc?.join(', ') || '-'
		},
		[capitalizeFirstLetter(getTranslation('network.enabled_capabilities', 'Enabled Capabilities'))]:
			{
				[capitalizeFirstLetter(getTranslation('common.raw', 'Raw'))]:
					iface.capabilities.enabled.raw,
				[capitalizeFirstLetter(getTranslation('common.description', 'Description'))]:
					iface.capabilities.enabled.desc?.join(', ') || '-'
			},
		[capitalizeFirstLetter(
			getTranslation('network.supported_capabilities', 'Supported Capabilities')
		)]: {
			[capitalizeFirstLetter(getTranslation('common.raw', 'Raw'))]:
				iface.capabilities.supported.raw,
			[capitalizeFirstLetter(getTranslation('common.description', 'Description'))]:
				iface.capabilities.supported.desc?.join(', ') || '-'
		}
	};

	if (iface.media !== null && iface.media !== undefined) {
		obj[capitalizeFirstLetter(getTranslation('network.media_options', 'Media Options'))] = {
			[capitalizeFirstLetter(getTranslation('common.status', 'Status'))]: iface.media.status,
			[capitalizeFirstLetter(getTranslation('network.type', 'Type'))]: iface.media.type,
			[capitalizeFirstLetter(getTranslation('network.sub_type', 'Sub Type'))]: iface.media.subtype,
			[capitalizeFirstLetter(getTranslation('common.mode', 'Mode'))]: iface.media.mode,
			[capitalizeFirstLetter(getTranslation('common.options', 'Options'))]: iface.media.options
				? iface.media.options?.join(', ') || '-'
				: '-'
		};
	}

	return obj;
}
