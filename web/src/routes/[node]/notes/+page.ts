import { getNotes } from '$lib/api/info/notes';
import type { Note } from '$lib/types/info/notes';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [notes] = await Promise.all([cachedFetch('notes', async () => getNotes(), cacheDuration)]);

	return {
		notes: notes as Note[]
	};
}
