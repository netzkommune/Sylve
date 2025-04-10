import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import type { Zpool } from '$lib/types/zfs/pool';
import humanFormat from 'human-format';
import { generateNumberFromString } from '../numbers';

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
		rows.push({
			id: generateNumberFromString(group.name),
			name: group.name,
			used: humanFormat(
				group.filesystems
					.filter((fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name)
					.reduce((acc: number, fs: Dataset) => acc + fs.used, 0)
			),
			avail: humanFormat(
				group.filesystems
					.filter((fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name)
					.reduce((acc: number, fs: Dataset) => acc + fs.avail, 0)
			),
			referenced: humanFormat(
				group.filesystems
					.filter((fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name)
					.reduce((acc: number, fs: Dataset) => acc + fs.referenced, 0)
			),
			mountpoint: group.filesystems
				.filter((fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name)
				.map((fs: Dataset) => fs.mountpoint)
				.join(', '),
			children: [
				...group.filesystems
					.filter((f) => f.name !== group.name || group.filesystems.length < 2)
					.map((filesystem: Dataset) => {
						return {
							id: generateNumberFromString(filesystem.name) + 1,
							name: filesystem.name,
							used: humanFormat(filesystem.used),
							avail: humanFormat(filesystem.avail),
							referenced: humanFormat(filesystem.referenced),
							mountpoint: filesystem.mountpoint,
							children: []
						};
					})
			]
		});
	}

	return {
		rows,
		columns
	};
}
