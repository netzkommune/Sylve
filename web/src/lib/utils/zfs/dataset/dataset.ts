import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
import type { Zpool } from '$lib/types/zfs/pool';

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
			pool: pool,
			filesystems: datasets.filter(
				(dataset) => dataset.name.startsWith(pool.name) && dataset.type === 'filesystem'
			),
			snapshots: datasets.filter(
				(dataset) => dataset.name.startsWith(pool.name) && dataset.type === 'snapshot'
			),
			volumes: datasets.filter(
				(dataset) => dataset.name.startsWith(pool.name) && dataset.type === 'volume'
			)
		};
	});

	return grouped;
}

export function getDatasetByGUID(
	datasets: Dataset[] | undefined,
	guid: string
): Dataset | undefined {
	if (!datasets) {
		return undefined;
	}

	const dataset = datasets.find((dataset) => dataset.guid === guid);
	return dataset;
}
