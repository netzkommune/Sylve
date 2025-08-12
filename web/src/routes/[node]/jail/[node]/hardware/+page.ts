import { getRAMInfo } from '$lib/api/info/ram';
import { getJails } from '$lib/api/jail/jail';
import { SEVEN_DAYS } from '$lib/utils.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = SEVEN_DAYS;

	const [jails, ram] = await Promise.all([
		cachedFetch('jail-list', async () => getJails(), cacheDuration),
		cachedFetch('ramInfo', async () => await getRAMInfo(), cacheDuration)
	]);

	const jail = jails.find((jail) => jail.ctId === parseInt(params.node, 10));

	return {
		jails: jails,
		jail: jail,
		ram: ram
	};
}
