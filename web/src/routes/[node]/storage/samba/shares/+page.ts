import { listGroups } from '$lib/api/auth/groups';
import { getInterfaces } from '$lib/api/network/iface';
import { getSambaConfig } from '$lib/api/samba/config';
import { getSambaShares } from '$lib/api/samba/share';
import { getDatasets } from '$lib/api/zfs/datasets';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [datasets, shares, groups] = await Promise.all([
		cachedFetch('zfs-datasets', async () => await getDatasets(), cacheDuration),
		cachedFetch('samba-shares', async () => await getSambaShares(), cacheDuration),
		cachedFetch('groups', async () => await listGroups(), cacheDuration)
	]);

	return {
		datasets,
		shares,
		groups
	};
}
