import type { APIResponse } from '$lib/types/common';
import type { Column, Row } from '$lib/types/components/tree-table';
import type { Disk } from '$lib/types/disk/disk';
import type { Zpool } from '$lib/types/zfs/pool';
import { getTranslation } from '../i18n';
import { generateNumberFromString } from '../numbers';
import { renderWithIcon, sizeFormatter } from '../table';

export const raidTypeArr = [
	{
		value: 'stripe',
		label: getTranslation('zfs.pool.redundancy.stripe', 'Stripe'),
		available: true
	},
	{
		value: 'mirror',
		label: getTranslation('zfs.pool.redundancy.mirror', 'Mirror'),
		available: false
	},
	{
		value: 'raidz',
		label: getTranslation('zfs.pool.redundancy.raidz', 'RAIDZ'),
		available: false
	},
	{
		value: 'raidz2',
		label: getTranslation('zfs.pool.redundancy.raidz2', 'RAIDZ2'),
		available: false
	},
	{
		value: 'raidz3',
		label: getTranslation('zfs.pool.redundancy.raidz3', 'RAIDZ3'),
		available: false
	}
];

export function generateTableData(
	pools: Zpool[],
	disks: Disk[]
): {
	rows: Row[];
	columns: Column[];
} {
	let rows: Row[] = [];
	let columns: Column[] = [
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

				console.log(disks);

				if (isPool(pools, value)) {
					return renderWithIcon('bi:hdd-stack-fill', value);
				}

				if (value.match(/p\d+$/)) {
					return renderWithIcon('ant-design:partition-outlined', value);
				}

				if (value.startsWith('/dev/')) {
					const nameOnly = value.replace('/dev/', '');
					const disk = disks.find((disk) => disk.device === nameOnly);
					if (disk) {
						if (disk.type === 'HDD') {
							return renderWithIcon('mdi:harddisk', value);
						} else if (disk.type === 'SSD') {
							return renderWithIcon('icon-park-outline:ssd', value);
						} else if (disk.type === 'NVMe') {
							return renderWithIcon('bi:nvme', value, 'rotate-90');
						}
					}
				}

				return `<span class="whitespace-nowrap">${value}</span>`;
			}
		},
		{
			field: 'size',
			title: 'Size',
			formatter: sizeFormatter
		},
		{
			field: 'used',
			title: 'Used',
			formatter: sizeFormatter
		},
		{
			field: 'health',
			title: 'Health'
		},
		{
			field: 'redundancy',
			title: 'Redundancy'
		}
	];

	for (const pool of pools) {
		const poolRow = {
			id: generateNumberFromString(pool.name + '-pool'),
			name: pool.name,
			size: pool.size,
			used: pool.allocated,
			health: pool.health,
			redundancy: '',
			children: [] as Row[]
		};

		for (const vdev of pool.vdevs) {
			if (vdev.name.includes('mirror') || vdev.name.includes('raid') || vdev.devices.length > 1) {
				let redundancy = 'Stripe';
				let vdevLabel = vdev.name;

				if (vdev.name.startsWith('mirror')) {
					redundancy = 'Mirror';
					vdevLabel = vdev.name.replace(/mirror-?(\d+)/i, 'Mirror $1');
				} else if (vdev.name.startsWith('raidz')) {
					redundancy = 'RAIDZ ' + vdev.name.match(/raidz-?(\d+)/i)?.[1];
					vdevLabel = vdev.name.replace(/^raidz/i, 'RAIDZ');
				}

				const vdevRow = {
					id: generateNumberFromString(vdev.name),
					name: vdevLabel,
					size: vdev.alloc + vdev.free,
					used: vdev.alloc,
					health: vdev.health,
					redundancy: '-',
					children: [] as Row[]
				};

				for (const device of vdev.devices) {
					if (
						vdev.replacingDevices &&
						vdev.replacingDevices.some(
							(r) => r.oldDrive.name === device.name || r.newDrive.name === device.name
						)
					) {
						continue;
					}

					vdevRow.children.push({
						id: generateNumberFromString(device.name),
						name: device.name,
						size: device.size,
						used: '-',
						health: device.health,
						redundancy: '-',
						children: []
					});
				}

				if (vdev.replacingDevices && vdev.replacingDevices.length > 0) {
					for (const replacing of vdev.replacingDevices) {
						vdevRow.children.push({
							id: generateNumberFromString(replacing.oldDrive.name),
							name: `${replacing.oldDrive.name} [OLD]`,
							size: replacing.oldDrive.size,
							used: '-',
							health: `${replacing.oldDrive.health} (Being replaced)`,
							redundancy: '-',
							children: []
						});

						vdevRow.children.push({
							id: generateNumberFromString(replacing.newDrive.name),
							name: `${replacing.newDrive.name} [NEW]`,
							size: replacing.newDrive.size,
							used: '-',
							health: `${replacing.newDrive.health} (Replacement)`,
							redundancy: '-',
							children: []
						});
					}
				}

				poolRow.children.push(vdevRow);
				poolRow.redundancy = redundancy;
			} else {
				poolRow.children.push({
					id: generateNumberFromString(vdev.devices[0].name),
					name: vdev.devices[0].name,
					size: vdev.devices[0].size,
					used: '-',
					health: vdev.devices[0].health,
					redundancy: '-',
					children: []
				});
				poolRow.redundancy = 'Stripe';
			}
		}

		rows.push(poolRow);

		if (pool.spares && pool.spares.length > 0) {
			const sparesRow: Row = {
				id: generateNumberFromString(`${pool.name}-spares`),
				name: 'Spares',
				size:
					pool.spares.reduce((acc, spare) => acc + spare.size, 0) > 0
						? pool.spares.reduce((acc, spare) => acc + spare.size, 0)
						: '-',
				used: '-',
				health: '-',
				redundancy: '-',
				children: []
			};

			for (const spare of pool.spares) {
				sparesRow.children!.push({
					id: generateNumberFromString(spare.name),
					name: spare.name,
					size: spare.size,
					used: '-',
					health: spare.health,
					redundancy: '-',
					children: []
				});
			}

			poolRow.children!.push(sparesRow);
		}
	}

	// spares should be at the end of the pool
	rows = rows.map((row) => {
		if (row.children) {
			const sparesIndex = row.children.findIndex((child) => child.name === 'Spares');
			if (sparesIndex !== -1) {
				const sparesRow = row.children.splice(sparesIndex, 1)[0];
				row.children.push(sparesRow);
			}
		}
		return row;
	});

	return {
		rows,
		columns
	};
}

export function isPool(pools: Zpool[], name: string): boolean {
	return pools.some((pool) => pool.name === name);
}

export function isReplaceableDevice(pools: Zpool[], name: string): boolean {
	for (const pool of pools) {
		if (pool.vdevs.some((vdev) => vdev.name === name)) {
			return false; // False if we're striped
		}
	}

	return pools.some((pool) => {
		for (const vdev of pool.vdevs) {
			if (vdev.devices.some((device) => device.name === name)) {
				return true;
			}
		}
		return false;
	});
}

export function getPoolByDevice(pools: Zpool[], name: string): string {
	for (const pool of pools) {
		for (const vdev of pool.vdevs) {
			if (vdev.devices.some((device) => device.name === name)) {
				return pool.name;
			}
		}
	}

	return '';
}

export function parsePoolActionError(error: APIResponse): string {
	if (error.message && error.message === 'pool_create_failed') {
		if (error.error) {
			if (error.error.includes('mirror contains devices of different sizes')) {
				return getTranslation(
					'zfs.pool.errors.pool_create_failed_mirror_different_sizes',
					'Pool contains a mirror with devices of different sizes'
				);
			} else if (error.error.includes('raidz contains devices of different sizes')) {
				return getTranslation(
					'zfs.pool.errors.pool_create_failed_raidz_different_sizes',
					'Pool contains a raidz vdev with devices of different sizes'
				);
			}
		}
	}

	if (error.message && error.message === 'pool_delete_failed') {
		if (error.error) {
			if (error.error.includes('pool or dataset is busy')) {
				return getTranslation('zfs.pool.errors.pool_delete_failed_busy', 'Pool is busy');
			}
		}
	}

	return '';
}
