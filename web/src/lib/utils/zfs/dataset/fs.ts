import type { APIResponse } from '$lib/types/common';
import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import { getTranslation } from '$lib/utils/i18n';
import { generateNumberFromString } from '$lib/utils/numbers';
import { capitalizeFirstLetter } from '$lib/utils/string';
import { renderWithIcon, sizeFormatter } from '$lib/utils/table';
import { cleanChildren } from '$lib/utils/tree-table';
import toast from 'svelte-french-toast';

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

				if (value.includes('@')) {
					const [, snapshot] = value.split('@');
					return renderWithIcon('carbon:ibm-cloud-vpc-block-storage-snapshots', snapshot);
				}

				if (value.includes('/')) {
					return renderWithIcon('material-symbols:files', value.substring(value.indexOf('/') + 1));
				}

				if (!value.includes('/') && !value.includes('@')) {
					return renderWithIcon('bi:hdd-stack-fill', value);
				}

				return `<span class="whitespace-nowrap">${value}</span>`;
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
				used: filesystem.used,
				avail: filesystem.avail,
				referenced: filesystem.referenced,
				mountpoint: filesystem.mountpoint || '',
				children: snapshotChildren
					.filter((snapshot) => snapshot.name.startsWith(filesystem.name + '@'))
					.map((snapshot: Dataset) => ({
						id: generateNumberFromString(snapshot.name) + 2,
						name: snapshot.name,
						used: snapshot.used,
						avail: snapshot.avail,
						referenced: snapshot.referenced,
						mountpoint: snapshot.mountpoint || '',
						children: []
					}))
			}));

			rows.push({
				id: generateNumberFromString(group.name),
				name: group.name,
				used: poolLevelFilesystem.used,
				avail: poolLevelFilesystem.avail,
				referenced: poolLevelFilesystem.referenced,
				mountpoint: poolLevelFilesystem.mountpoint || '',
				children: [
					...poolSnapshots.map((snapshot: Dataset) => ({
						id: generateNumberFromString(snapshot.name) + 1,
						name: snapshot.name,
						used: snapshot.used,
						avail: snapshot.avail,
						referenced: snapshot.referenced,
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
						used: filesystem.used,
						avail: filesystem.avail,
						referenced: filesystem.referenced,
						mountpoint: filesystem.mountpoint || '',
						children: snapshotChildren
							.filter((snapshot) => snapshot.name.startsWith(filesystem.name + '@'))
							.map((snapshot: Dataset) => ({
								id: generateNumberFromString(snapshot.name) + 1,
								name: snapshot.name,
								used: snapshot.used,
								avail: snapshot.avail,
								referenced: snapshot.referenced,
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

export function handleError(error: APIResponse): void {
	if (error.error?.includes('dataset already exists')) {
		let [key, value] = ['', ''];

		if (error.error?.includes('snapshot')) {
			key = 'zfs.datasets.snapshot_already_exists';
			value = 'snapshot already exists';
		} else {
			key = 'zfs.datasets.filesystem_already_exists';
			value = 'filesystem already exists';
		}

		toast.error(capitalizeFirstLetter(getTranslation(key, value)), {
			position: 'bottom-center'
		});
	}

	if (error.error?.includes('numeric value is too large')) {
		toast.error(
			capitalizeFirstLetter(
				getTranslation('zfs.datasets.numeric_value_too_large', 'Numeric value is too large')
			),
			{
				position: 'bottom-center'
			}
		);
	}

	if (error.error?.includes('invalid_encryption_key_length')) {
		toast.error(
			capitalizeFirstLetter(
				getTranslation(
					'zfs.datasets.invalid_encryption_key_length',
					'Invalid encryption key length'
				)
			),
			{
				position: 'bottom-center'
			}
		);
	}

	if (error.error?.includes('pool or dataset is busy')) {
		toast.error(
			capitalizeFirstLetter(
				getTranslation('zfs.datasets.pool_or_dataset_is_busy', 'Pool or dataset is busy')
			),
			{
				position: 'bottom-center'
			}
		);
	}
}
