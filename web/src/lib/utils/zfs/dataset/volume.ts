import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import { generateNumberFromString } from '$lib/utils/numbers';
import { cleanChildren } from '$lib/utils/tree-table';
import humanFormat from 'human-format';

export const createVolProps = {
	atime: [
		{
			label: 'on',
			value: 'on'
		},
		{
			label: 'off',
			value: 'off'
		}
	],
	checksum: [
		{
			label: 'on',
			value: 'on'
		},
		{
			label: 'off',
			value: 'off'
		},
		{
			label: 'fletcher2',
			value: 'fletcher2'
		},
		{
			label: 'fletcher4',
			value: 'fletcher4'
		},
		{
			label: 'sha256',
			value: 'sha256'
		},
		{
			label: 'noparity',
			value: 'noparity'
		}
	],
	compression: [
		{
			label: 'on',
			value: 'on'
		},
		{
			label: 'off',
			value: 'off'
		},
		{
			label: 'gzip',
			value: 'gzip'
		},
		{
			label: 'lz4',
			value: 'lz4'
		},
		{
			label: 'lzjb',
			value: 'lzjb'
		},
		{
			label: 'zle',
			value: 'zle'
		},
		{
			label: 'zstd',
			value: 'zstd'
		},
		{
			label: 'zstd-fast',
			value: 'zstd-fast'
		}
	],
	dedup: [
		{
			label: 'off',
			value: 'off'
		},
		{
			label: 'on',
			value: 'on'
		},
		{
			label: 'Verify',
			value: 'verify'
		}
	],
	encryption: [
		{
			label: 'off',
			value: 'off'
		},
		{
			label: 'on',
			value: 'on'
		},
		{
			label: 'aes-128-ccm',
			value: 'aes-128-ccm'
		},
		{
			label: 'aes-192-ccm',
			value: 'aes-192-ccm'
		},
		{
			label: 'aes-256-ccm',
			value: 'aes-256-ccm'
		},
		{
			label: 'aes-128-gcm',
			value: 'aes-128-gcm'
		},
		{
			label: 'aes-192-gcm',
			value: 'aes-192-gcm'
		},
		{
			label: 'aes-256-gcm',
			value: 'aes-256-gcm'
		}
	],
	volblocksize: [
		{
			label: '512B (Legacy HDD sectors)',
			value: '512'
		},
		{
			label: '1K (1024B)',
			value: '1024'
		},
		{
			label: '2K (2048B)',
			value: '2048'
		},
		{
			label: '4K (4096B) - SSD/VMs',
			value: '4096'
		},
		{
			label: '8K (8192B)',
			value: '8192'
		},
		{
			label: '16K (16384B)',
			value: '16384'
		},
		{
			label: '32K (32768B) - Sequential workloads',
			value: '32768'
		},
		{
			label: '64K (65536B) - Large files',
			value: '65536'
		},
		{
			label: '128K (131072B) Media/backups',
			value: '131072'
		}
	]
};

export function generateTableData(grouped: GroupedByPool[]): { rows: Row[]; columns: Column[] } {
	const rows: Row[] = [];
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
			field: 'size',
			title: 'Size',
			formatter: (cell) => humanFormat(cell.getValue())
		},
		{
			field: 'referenced',
			title: 'Referenced',
			formatter: (cell) => {
				try {
					return humanFormat(cell.getValue());
				} catch (e) {
					return cell.getValue();
				}
			}
		},
		{
			field: 'guid',
			title: 'GUID',
			visible: false
		}
	];

	for (const group of grouped) {
		const poolRow: Row = {
			id: generateNumberFromString(group.name),
			name: group.name,
			size: 0,
			referenced: '-',
			guid: undefined,
			children: []
		};

		poolRow.size = group.pool?.size;

		const volumeChildren = group.volumes
			.filter((vol) => vol.name !== group.name)
			.map((vol) => {
				const volumeRow: Row = {
					id: generateNumberFromString(vol.name),
					name: vol.name,
					size: vol.volsize,
					referenced: vol.referenced,
					guid: vol.properties?.guid,
					children: []
				};

				const snapshots = group.snapshots.filter((snap) => snap.name.startsWith(vol.name + '@'));
				volumeRow.children?.push(
					...snapshots.map((snap) => ({
						id: generateNumberFromString(snap.name),
						name: snap.name.split('@')[1],
						size: snap.used,
						referenced: snap.referenced,
						guid: snap.properties?.guid,
						children: []
					}))
				);

				return volumeRow;
			});

		poolRow.children?.push(...volumeChildren);
		rows.push(poolRow);
	}

	return {
		rows: rows.filter((row) => row.children && row.children.length > 0).map(cleanChildren),
		columns
	};
}
