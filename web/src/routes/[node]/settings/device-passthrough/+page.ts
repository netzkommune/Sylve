import { addPPTDevice, getPCIDevices, getPPTDevices, removePPTDevice } from '$lib/api/system/pci';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = 1;
	const [pciDevices, pptDevices] = await Promise.all([
		cachedFetch('pciDevices', async () => await getPCIDevices(), cacheDuration),
		cachedFetch('pptDevices', async () => await getPPTDevices(), cacheDuration)
	]);

	console.log('Loaded PCI Devices:', pciDevices);
	console.log('Loaded PPT Devices:', pptDevices);

	// console.log(await addPPTDevice('3/0/0'));
	// await removePPTDevice('1');

	return {
		pciDevices,
		pptDevices
	};
}
