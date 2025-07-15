import { getInterfaces } from '$lib/api/network/iface';
import { getSambaConfig } from '$lib/api/samba/config';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [interfaces, sambaConfig] = await Promise.all([
		cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration),
		cachedFetch('samba-config', async () => await getSambaConfig(), cacheDuration)
	]);

	return {
		interfaces,
		sambaConfig
	};
}
