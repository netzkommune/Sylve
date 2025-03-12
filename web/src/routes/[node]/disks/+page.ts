import { listDisks } from '$lib/api/disk/disk';
import { simplifyDisks } from '$lib/utils/disk';

export async function load() {
	let disks = await simplifyDisks(await listDisks());
	disks = disks.filter((disk) => disk.Device !== '/dev/nda0');
	return {
		disks
	};
}
