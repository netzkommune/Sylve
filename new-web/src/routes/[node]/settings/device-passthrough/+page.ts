import { getPCIDevices, getPPTDevices } from '$lib/api/system/pci';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [pciDevices, pptDevices] = await Promise.all([
		cachedFetch('pciDevices', async () => await getPCIDevices(), cacheDuration),
		cachedFetch('pptDevices', async () => await getPPTDevices(), cacheDuration)
	]);

	return {
		pciDevices,
		pptDevices
	};
}
