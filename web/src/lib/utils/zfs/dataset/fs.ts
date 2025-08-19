import type { APIResponse } from '$lib/types/common';
import type { Column, Row } from '$lib/types/components/tree-table';
import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import { generateNumberFromString } from '$lib/utils/numbers';
import { capitalizeFirstLetter } from '$lib/utils/string';
import { renderWithIcon, sizeFormatter } from '$lib/utils/table';
import { cleanChildren } from '$lib/utils/tree-table';
import { toast } from 'svelte-sonner';

export const createFSProps = {
	atime: [
		{
			label: 'On',
			value: 'on'
		},
		{
			label: 'Off',
			value: 'off'
		}
	],
	checksum: [
		{
			label: 'On',
			value: 'on'
		},
		{
			label: 'Off',
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
			label: 'On',
			value: 'on'
		},
		{
			label: 'Off',
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
			label: 'Off',
			value: 'off'
		},
		{
			label: 'On',
			value: 'on'
		},
		{
			label: 'Verify',
			value: 'verify'
		}
	],
	encryption: [
		{
			label: 'Off',
			value: 'off'
		},
		{
			label: 'On',
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
	aclInherit: [
		{
			label: 'Discard',
			value: 'discard'
		},
		{
			label: 'No Allow',
			value: 'noallow'
		},
		{
			label: 'Restricted',
			value: 'restricted'
		},
		{
			label: 'Passthrough',
			value: 'passthrough'
		},
		{
			label: 'Passthrough-X',
			value: 'passthrough-x'
		}
	],
	aclMode: [
		{
			label: 'Discard',
			value: 'discard'
		},
		{
			label: 'Group Mask',
			value: 'groupmask'
		},
		{
			label: 'Passthrough',
			value: 'passthrough'
		},
		{
			label: 'Passthrough-X',
			value: 'passthrough-x'
		},
		{
			label: 'Restricted',
			value: 'restricted'
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
		},
		{
			field: 'type',
			title: 'Type',
			visible: false
		}
	];

	for (const group of grouped) {
		const poolLevelFilesystem = group.filesystems.find(
			(fs: Dataset) => fs.type === 'filesystem' && fs.name === group.name
		);

		const filesystemChildren = group.filesystems.filter(
			(fs) => fs.type === 'filesystem' && fs.name !== group.name
		);

		const snapshotChildren = group.snapshots.filter(
			(snapshot) => !snapshot.name.startsWith(group.name + '@')
		);

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
					})),
				type: filesystem.type
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
				}),
				type: 'pool'
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
		let value = '';

		if (error.error?.includes('snapshot')) {
			value = 'Snapshot already exists';
		} else {
			value = 'Filesystem already exists';
		}

		toast.error(value, {
			position: 'bottom-center'
		});
	}

	if (error.error?.includes('numeric value is too large')) {
		toast.error('Numeric value is too large', {
			position: 'bottom-center'
		});
	}

	if (error.error?.includes('invalid_encryption_key_length')) {
		toast.error('Invalid encryption key length', {
			position: 'bottom-center'
		});
	}

	if (error.error?.includes('pool or dataset is busy')) {
		toast.error('Pool or dataset is busy', {
			position: 'bottom-center'
		});
	}
}
