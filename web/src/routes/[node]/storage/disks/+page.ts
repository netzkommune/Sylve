import { listDisks } from '$lib/api/disk/disk';
import { getPools } from '$lib/api/zfs/pool';
import { simplifyDisks } from '$lib/utils/disk';

export async function load() {
	let [disks, pools] = await Promise.all([simplifyDisks(await listDisks()), getPools()]);

	return {
		disks,
		pools
	};
}
