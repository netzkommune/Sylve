import { listDisks } from '$lib/api/disk/disk';
import { editPool, getPools } from '$lib/api/zfs/pool';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [disks, pools] = await Promise.all([
		cachedFetch('disks', async () => await listDisks(), cacheDuration),
		cachedFetch('pools', getPools, cacheDuration)
	]);

	return {
		disks,
		pools
	};
}
