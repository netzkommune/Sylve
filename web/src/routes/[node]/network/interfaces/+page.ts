import { getInterfaces } from '$lib/api/network/iface';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 1000 * 60000;
	const [interfaces] = await Promise.all([
		cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration)
	]);

	return {
		interfaces
	};
}
