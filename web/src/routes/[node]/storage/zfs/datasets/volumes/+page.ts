import { getDownloads } from '$lib/api/utilities/downloader';
import { getDatasets } from '$lib/api/zfs/datasets';
import { getPools } from '$lib/api/zfs/pool';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;

	const [datasets, pools, downloads] = await Promise.all([
		cachedFetch('zfs-volumes', async () => await getDatasets('volume'), cacheDuration),
		cachedFetch('pools', getPools, cacheDuration),
		cachedFetch('downloads', async () => getDownloads(), cacheDuration)
	]);

	return {
		pools: pools,
		datasets: datasets,
		downloads: downloads
	};
}
