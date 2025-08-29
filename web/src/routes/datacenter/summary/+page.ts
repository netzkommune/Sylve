import { getNodes } from '$lib/api/cluster/cluster';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [nodes] = await Promise.all([
		cachedFetch('cluster-nodes', async () => getNodes(), cacheDuration)
	]);

	return {
		nodes: nodes
	};
}
