import { getDownloads } from '$lib/api/utilities/downloader';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 1000 * 60000;
	const [downloads] = await Promise.all([
		cachedFetch('downloads', async () => getDownloads(), cacheDuration)
	]);

	return {
		downloads
	};
}
