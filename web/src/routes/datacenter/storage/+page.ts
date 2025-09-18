import { getNotes } from '$lib/api/cluster/notes';
import { getStorages } from '$lib/api/cluster/storage';
import type { Note } from '$lib/types/info/notes';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [storages] = await Promise.all([
		cachedFetch('cluster-storages', async () => getStorages(), cacheDuration)
	]);

	return {
		storages: storages
	};
}
