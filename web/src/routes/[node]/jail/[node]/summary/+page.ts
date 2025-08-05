import { getJails } from '$lib/api/jail/jail';
import { SEVEN_DAYS } from '$lib/utils.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = SEVEN_DAYS;
	const vmId = params.node;

	const [jails] = await Promise.all([
		cachedFetch('jail-list', async () => getJails(), cacheDuration)
	]);

	return {
		jails: jails
	};
}
