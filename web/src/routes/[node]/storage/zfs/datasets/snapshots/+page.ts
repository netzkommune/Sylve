import { getDatasets, getPeriodicSnapshots } from '$lib/api/zfs/datasets';
import { getPools } from '$lib/api/zfs/pool';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [datasets, pools, periodicSnapshots] = await Promise.all([
		cachedFetch('datasets', async () => await getDatasets(), cacheDuration),
		cachedFetch('pools', getPools, cacheDuration),
		cachedFetch('periodicSnapshots', async () => await getPeriodicSnapshots(), cacheDuration)
	]);

	return {
		pools: pools,
		periodicSnapshots: periodicSnapshots,
		datasets: datasets
	};
}
