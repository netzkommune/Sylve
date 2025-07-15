import { getFiles } from '$lib/api/system/file-explorer';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [files] = await Promise.all([cachedFetch('fx-files', async () => await getFiles(), 1)]);

	return {
		files
	};
}
