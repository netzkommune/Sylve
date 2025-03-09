import { listDisks, simplifyDisks } from '$lib/api/disk/disk';

export async function load() {
	const disks = await simplifyDisks(await listDisks());
	return {
		disks
	};
}
