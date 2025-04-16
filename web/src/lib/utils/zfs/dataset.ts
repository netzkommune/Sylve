import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import type { Zpool } from '$lib/types/zfs/pool';
import humanFormat from 'human-format';
import { generateNumberFromString } from '../numbers';
import { cleanChildren } from '../tree-table';

export const createFSProps = {
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
	]
};

export function groupByPool(
	pools: Zpool[] | undefined,
	datasets: Dataset[] | undefined
): GroupedByPool[] {
	if (!pools || !datasets) {
		return [];
	}

	const grouped = pools.map((pool) => {
		return {
			name: pool.name,
			filesystems: datasets.filter(
				(dataset) => dataset.name.startsWith(pool.name) && dataset.type === 'filesystem'
			),
			snapshots: datasets.filter(
				(dataset) => dataset.name.startsWith(pool.name) && dataset.type === 'snapshot'
			),
			volumes: []
		};
	});

	return grouped;
}

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
			field: 'used',
			title: 'Used'
		},
		{
			field: 'avail',
			title: 'Available'
		},
		{
			field: 'referenced',
			title: 'Referenced'
		},
		{
			field: 'mountpoint',
			title: 'Mount Point'
		}
	];

	for (const group of grouped) {
		const poolLevelFilesystem = group.filesystems.find(
			(fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name
		);

		const filesystemChildren = group.filesystems.filter(
			(fs) => fs.type === 'filesystem' && fs.name !== group.name
		);

		const snapshotChildren = group.snapshots;

		if (poolLevelFilesystem) {
			const poolSnapshots = snapshotChildren.filter((snapshot) =>
				snapshot.name.startsWith(group.name + '@')
			);

			const childFilesystemsWithSnapshots = filesystemChildren.map((filesystem: Dataset) => ({
				id: generateNumberFromString(filesystem.name) + 1,
				name: filesystem.name,
				used: humanFormat(filesystem.used),
				avail: humanFormat(filesystem.avail),
				referenced: humanFormat(filesystem.referenced),
				mountpoint: filesystem.mountpoint || '',
				children: snapshotChildren
					.filter((snapshot) => snapshot.name.startsWith(filesystem.name + '@'))
					.map((snapshot: Dataset) => ({
						id: generateNumberFromString(snapshot.name) + 2,
						name: snapshot.name,
						used: humanFormat(snapshot.used),
						avail: humanFormat(snapshot.avail),
						referenced: humanFormat(snapshot.referenced),
						mountpoint: snapshot.mountpoint || '',
						children: []
					}))
			}));

			rows.push({
				id: generateNumberFromString(group.name),
				name: group.name,
				used: humanFormat(poolLevelFilesystem.used),
				avail: humanFormat(poolLevelFilesystem.avail),
				referenced: humanFormat(poolLevelFilesystem.referenced),
				mountpoint: poolLevelFilesystem.mountpoint || '',
				children: [
					...poolSnapshots.map((snapshot: Dataset) => ({
						id: generateNumberFromString(snapshot.name) + 1,
						name: snapshot.name,
						used: humanFormat(snapshot.used),
						avail: humanFormat(snapshot.avail),
						referenced: humanFormat(snapshot.referenced),
						mountpoint: snapshot.mountpoint || '',
						children: [],
						isPoolSnapshot: true
					})),
					...childFilesystemsWithSnapshots
				].sort((a, b) => {
					const aIsPoolSnapshot = a.hasOwnProperty('isPoolSnapshot');
					const bIsPoolSnapshot = b.hasOwnProperty('isPoolSnapshot');
					if (aIsPoolSnapshot && !bIsPoolSnapshot) return -1;
					if (!aIsPoolSnapshot && bIsPoolSnapshot) return 1;
					return a.name.localeCompare(b.name);
				})
			});
		} else if (group.filesystems.length > 0) {
			rows.push(
				...group.filesystems
					.filter((fs) => fs.type === 'filesystem')
					.map((filesystem: Dataset) => ({
						id: generateNumberFromString(filesystem.name),
						name: filesystem.name,
						used: humanFormat(filesystem.used),
						avail: humanFormat(filesystem.avail),
						referenced: humanFormat(filesystem.referenced),
						mountpoint: filesystem.mountpoint || '',
						children: snapshotChildren
							.filter((snapshot) => snapshot.name.startsWith(filesystem.name + '@'))
							.map((snapshot: Dataset) => ({
								id: generateNumberFromString(snapshot.name) + 1,
								name: snapshot.name,
								used: humanFormat(snapshot.used),
								avail: humanFormat(snapshot.avail),
								referenced: humanFormat(snapshot.referenced),
								mountpoint: snapshot.mountpoint || '',
								children: []
							}))
							.sort((a, b) => a.name.localeCompare(b.name))
					}))
			);
		}
	}

	return {
		rows: rows.map(cleanChildren),
		columns
	};
}
