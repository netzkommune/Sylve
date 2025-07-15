import { listUsers } from '$lib/api/auth/local';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [users] = await Promise.all([
		cachedFetch('users', async () => await listUsers(), cacheDuration)
	]);

	return {
		users: users || []
	};
}
