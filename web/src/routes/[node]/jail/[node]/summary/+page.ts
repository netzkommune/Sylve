import { getJails, getJailStates, getStats } from '$lib/api/jail/jail';
import { SEVEN_DAYS } from '$lib/utils.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = SEVEN_DAYS;
	const vmId = params.node;

	const [jails, jailStates] = await Promise.all([
		cachedFetch('jail-list', async () => getJails(), cacheDuration),
		cachedFetch('jail-states', async () => getJailStates(), cacheDuration),
		cachedFetch(`jail-stats-${vmId}`, async () => getStats(Number(vmId), 10), cacheDuration)
	]);

	const jail = jails.find((jail) => jail.ctId === parseInt(params.node, 10));

	return {
		jails: jails,
		jailStates: jailStates,
		jail: jail
	};
}
