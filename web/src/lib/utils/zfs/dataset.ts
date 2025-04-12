import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import type { Zpool } from '$lib/types/zfs/pool';
import humanFormat from 'human-format';
import { generateNumberFromString } from '../numbers';
import { cleanChildren } from '../tree-table';

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

export function generateTableData(grouped: GroupedByPool[]): {
	rows: Row[];
	columns: Column[];
} {
	const rows: Row[] = [];
	const columns: Column[] = [
		{
			key: 'id',
			label: 'ID'
		},
		{
			key: 'name',
			label: 'Name'
		},
		{
			key: 'used',
			label: 'Used'
		},
		{
			key: 'avail',
			label: 'Available'
		},
		{
			key: 'referenced',
			label: 'Referenced'
		},
		{
			key: 'mountpoint',
			label: 'Mount Point'
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
