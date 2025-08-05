import { getSambaShares } from '$lib/api/samba/share';
import { getDatasets } from '$lib/api/zfs/datasets';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [datasets, shares] = await Promise.all([
		cachedFetch('datasets', async () => await getDatasets(), cacheDuration),
		cachedFetch('samba-shares', async () => await getSambaShares(), cacheDuration)
	]);

	return {
		datasets,
		shares
	};
}
