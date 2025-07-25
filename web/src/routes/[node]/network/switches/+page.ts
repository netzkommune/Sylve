import { getInterfaces } from '$lib/api/network/iface';
import { getNetworkObjects } from '$lib/api/network/object';
import { getSwitches } from '$lib/api/network/switch';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 1000 * 60000;
	const [interfaces, switches, networkObjects] = await Promise.all([
		cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration),
		cachedFetch('networkSwitches', async () => await getSwitches(), cacheDuration),
		cachedFetch('networkObjects', async () => await getNetworkObjects(), cacheDuration)
	]);

	return {
		interfaces,
		switches
	};
}
