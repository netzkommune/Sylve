import { getJails, getJailStates } from '$lib/api/jail/jail';
import { getInterfaces } from '$lib/api/network/iface';
import { getNetworkObjects } from '$lib/api/network/object';
import { getSwitches } from '$lib/api/network/switch';
import { SEVEN_DAYS } from '$lib/utils.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = SEVEN_DAYS;
	const ctId = params.node;

	const [jails, jailStates, interfaces, switches, networkObjects] = await Promise.all([
		cachedFetch('jail-list', async () => getJails(), cacheDuration),
		cachedFetch('jail-states', async () => getJailStates(), cacheDuration),
		cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration),
		cachedFetch('networkSwitches', async () => await getSwitches(), cacheDuration),
		cachedFetch('networkObjects', async () => await getNetworkObjects(), cacheDuration)
	]);

	const jail = jails.find((jail) => jail.ctId === parseInt(ctId, 10));

	return {
		jails: jails,
		jailStates: jailStates,
		jail: jail,
		interfaces,
		switches,
		networkObjects
	};
}
