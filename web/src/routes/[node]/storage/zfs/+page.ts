import { listDisks } from '$lib/api/disk/disk';
import { createPool, deletePool, getPools } from '$lib/api/zfs/pool';
import { simplifyDisks } from '$lib/utils/disk';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 3600 * 1000;
	const [disks, pools] = await Promise.all([
		cachedFetch('disks', async () => simplifyDisks(await listDisks()), cacheDuration),
		cachedFetch('pools', getPools, cacheDuration)
	]);

	/* 
    export async function createPool(
        name: string,
        vdevs: string[],
        raid: string,
        options: Record<string, string>
    ) {
        return await apiRequest('/zfs/pools', APIResponseSchema, 'POST', {
            name,
            vdevs,
            raid,
            options
        });
    }
    */

	// console.log(
	// 	await createPool({
	// 		name: 'test',
	// 		raidType: 'mirror',
	// 		vdevs: [
	// 			{
	// 				name: 'test',
	// 				devices: ['/dev/ada0p1', '/dev/ada1p1']
	// 			},
	// 			{
	// 				name: 'test2',
	// 				devices: ['/dev/ada0p2', '/dev/ada1p2']
	// 			}
	// 		],
	// 		properties: {
	// 			ashift: '12'
	// 		}
	// 	})
	// );

	// console.log(await deletePool('test'));

	return {
		disks,
		pools
	};
}
