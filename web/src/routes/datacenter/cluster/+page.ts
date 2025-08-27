import { getDetails } from '$lib/api/cluster/cluster.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = 1000;

	const [cluster] = await Promise.all([
		cachedFetch('cluster-info', async () => getDetails(), cacheDuration)
	]);

	return {
		cluster
	};
}
