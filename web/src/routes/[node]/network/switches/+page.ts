import { createSwitch, deleteSwitch, getSwitches, updateSwitch } from '$lib/api/network/switch';

export async function load() {
	// const cacheDuration = 1000 * 60000;
	// const [interfaces] = await Promise.all([
	//     cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration)
	// ]);
	// return {
	//     interfaces
	// };
	// const switches = await getSwitches();
	// console.log(switches);
	// const created = await createSwitch('offline', 1280, 0, '', false, ['em0']);
	// console.log(created);
	// const updated = await updateSwitch(4, 'public', 1500, 337, '', false, ['em0']);
	// console.log(updated);
	// const deleted = await deleteSwitch(4);
	// console.log(deleted);
}
