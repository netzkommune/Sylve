import type { Column, Row } from '$lib/types/components/tree-table';
import type { Zpool } from '$lib/types/zfs/pool';
import humanFormat, { type ScaleLike } from 'human-format';

const options = {
	scale: 'binary' as ScaleLike, // Base 1024 — matches ZFS
	unit: 'B', // Adds the 'B' suffix like '20.6 GB'
	maxDecimals: 1 // Matches typical output from `zpool list`
};

export function generateTableData(pools: Zpool[]): {
	rows: Row[];
	columns: Column[];
} {
	let rows: Row[] = [];
	let columns: Column[] = [
		{
			key: 'name',
			label: 'Name'
		},
		{
			key: 'size',
			label: 'Size'
		},
		{
			key: 'used',
			label: 'Used'
		},
		{
			key: 'health',
			label: 'Health'
		},
		{
			key: 'redundancy',
			label: 'Redundancy'
		}
	];

	let id = 0;

	for (const pool of pools) {
		const poolRow = {
			id: id++,
			name: pool.name,
			size: humanFormat(pool.size, options),
			used: humanFormat(pool.allocated, options),
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
					redundancy = 'RAIDZ';
					vdevLabel = vdev.name.replace(/^raidz/i, 'RAIDZ');
				}

				const vdevRow = {
					id: id++,
					name: vdevLabel,
					size: humanFormat(vdev.alloc + vdev.free, options),
					used: humanFormat(vdev.alloc, options),
					health: vdev.health,
					redundancy: '-',
					children: [] as Row[]
				};

				// Add regular devices
				for (const device of vdev.devices) {
					// Skip devices that are part of a replacing operation
					if (
						vdev.replacingDevices &&
						vdev.replacingDevices.some(
							(r) => r.oldDrive.name === device.name || r.newDrive.name === device.name
						)
					) {
						continue;
					}

					vdevRow.children.push({
						id: id++,
						name: device.name,
						size: humanFormat(device.size, options),
						used: '-',
						health: device.health,
						redundancy: '-',
						children: []
					});
				}

				// Add replacing devices if they exist
				if (vdev.replacingDevices && vdev.replacingDevices.length > 0) {
					for (const replacing of vdev.replacingDevices) {
						// Add the old drive
						vdevRow.children.push({
							id: id++,
							name: `${replacing.oldDrive.name} [OLD]`,
							size: humanFormat(replacing.oldDrive.size, options),
							used: '-',
							health: `${replacing.oldDrive.health} (Being replaced)`,
							redundancy: '-',
							children: []
						});

						// Add the new drive
						vdevRow.children.push({
							id: id++,
							name: `${replacing.newDrive.name} [NEW]`,
							size: humanFormat(replacing.newDrive.size, options),
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
					id: id++,
					name: vdev.devices[0].name,
					size: humanFormat(vdev.devices[0].size, options),
					used: '-',
					health: vdev.devices[0].health,
					redundancy: '-',
					children: []
				});
				poolRow.redundancy = 'Stripe';
			}
		}

		rows.push(poolRow);
	}

	return {
		rows,
		columns
	};
}
