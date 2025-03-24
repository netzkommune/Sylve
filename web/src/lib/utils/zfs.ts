/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { getPools } from '$lib/api/zfs/pool';
import type { Zpool } from '$lib/types/zfs/pool';

export async function getTotalDiskUsage(): Promise<number> {
	try {
		const pools = await getPools();
		const total = pools.reduce((acc, pool) => acc + pool.size, 0);
		const used = pools.reduce((acc, pool) => acc + pool.allocated, 0);

		if (total === 0) {
			return 0.0;
		}

		return (used / total) * 100;
	} catch (e) {
		return 0.0;
	}
}

export function isDeviceVdev(device: string, pools: Zpool[]): boolean {
	if (pools.length === 0) {
		return false;
	}

	for (const pool of pools) {
		for (const vdev of pool.vdevs) {
			if (vdev.name === device || vdev.name === `/dev/${device}`) {
				return true;
			}
		}
	}

	return false;
}

export function getHealthHelpers(health: string): { icon: string; color: string; text: string } {
	switch (health) {
		case 'ONLINE':
			return {
				icon: 'carbon:checkmark-filled',
				color: 'text-green-600 dark:text-green-500',
				text: 'Online'
			};
		case 'DEGRADED':
			return {
				icon: 'carbon:warning-filled',
				color: 'text-yellow-600 dark:text-yellow-500',
				text: 'Degraded'
			};
		case 'FAULTED':
			return {
				icon: 'carbon:close-filled',
				color: 'text-red-600 dark:text-red-500',
				text: 'Faulted'
			};
		case 'OFFLINE':
			return {
				icon: 'material-symbols:offline-pin-off',
				color: 'text-red-600 dark:text-red-500',
				text: 'Offline'
			};
		case 'UNAVAIL':
			return {
				icon: 'carbon:warning-alt-filled',
				color: 'text-yellow-600 dark:text-yellow-500',
				text: 'Unavailable'
			};
		case 'REMOVED':
			return {
				icon: 'carbon:warning-alt-filled',
				color: 'text-yellow-600 dark:text-yellow-500',
				text: 'Removed'
			};
		default:
			return {
				icon: 'carbon:warning-alt-filled',
				color: 'text-yellow-600 dark:text-yellow-500',
				text: 'Unknown'
			};
	}
}
