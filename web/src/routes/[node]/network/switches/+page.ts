import { createSwitch, deleteSwitch, getSwitches } from '$lib/api/network/switch';

export async function load() {
	// const cacheDuration = 1000 * 60000;
	// const [interfaces] = await Promise.all([
	//     cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration)
	// ]);
	// return {
	//     interfaces
	// };

	const switches = await getSwitches();
	console.log(switches);

	// const created = await createSwitch('test', 1280, 0, '', false, ['re0']);
	// console.log(created);

	// const deleted = await deleteSwitch(1);
	// console.log(deleted);
}
