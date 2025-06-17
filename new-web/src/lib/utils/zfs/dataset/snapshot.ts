import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import { generateNumberFromString } from '$lib/utils/numbers';
import { renderWithIcon, sizeFormatter } from '$lib/utils/table';

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
			title: 'Name',
			formatter: (cell) => {
				const value = cell.getValue();

				if (!value.includes('@') && !value.includes('/')) {
					return renderWithIcon('bi:hdd-stack-fill', value);
				}

				if (value.includes('/')) {
					const [pool, ...rest] = value.split('/');
					const name = rest.join('/');
					return renderWithIcon('carbon:ibm-cloud-vpc-block-storage-snapshots', name);
				}

				return renderWithIcon('carbon:ibm-cloud-vpc-block-storage-snapshots', value);
			}
		},
		{
			field: 'used',
			title: 'Used',
			formatter: sizeFormatter
		},
		{
			field: 'avail',
			title: 'Available',
			formatter: sizeFormatter
		},
		{
			field: 'referenced',
			title: 'Referenced',
			formatter: sizeFormatter
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

		if (!poolLevelFilesystem) continue;

		const poolSnapshots = group.snapshots.filter(
			(snapshot) =>
				snapshot.name.startsWith(group.name + '@') || snapshot.name.startsWith(group.name + '/')
		);

		const children = poolSnapshots.map((snapshot: Dataset) => ({
			id: generateNumberFromString(snapshot.name),
			name: snapshot.name,
			used: snapshot.used,
			avail: snapshot.avail,
			referenced: snapshot.referenced,
			mountpoint: snapshot.mountpoint || '',
			children: []
		}));

		rows.push({
			id: generateNumberFromString(group.name),
			name: group.name,
			used: poolLevelFilesystem.used,
			avail: poolLevelFilesystem.avail,
			referenced: poolLevelFilesystem.referenced,
			mountpoint: poolLevelFilesystem.mountpoint || '',
			children
		});
	}

	return {
		rows,
		columns
	};
}
