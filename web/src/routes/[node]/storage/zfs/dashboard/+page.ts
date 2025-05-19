import { getDatasets } from '$lib/api/zfs/datasets';
import { getPools, getPoolStats } from '$lib/api/zfs/pool';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 7 * 24 * 60 * 60;
	const [datasets, pools] = await Promise.all([
		cachedFetch('datasets', async () => await getDatasets(), cacheDuration),
		cachedFetch('pools', getPools, cacheDuration)
	]);

	console.log(await getPoolStats(1, 128));

	return {
		pools: pools,
		datasets: datasets
	};
}
