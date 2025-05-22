import { getDatasets } from '$lib/api/zfs/datasets';
import { getPools, getPoolStats } from '$lib/api/zfs/pool';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
    const cacheDuration = 1;
    const [datasets, pools, poolStats] = await Promise.all([
        cachedFetch('datasets', async () => await getDatasets(), cacheDuration),
        cachedFetch('pools', getPools, cacheDuration),
        cachedFetch('pool-stats', async () => await getPoolStats(1, 128), cacheDuration),
    ]);

    return {
        pools: pools,
        datasets: datasets,
        poolStats: poolStats

    };
}
