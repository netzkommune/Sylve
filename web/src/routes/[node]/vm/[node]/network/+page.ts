import { getInterfaces } from '$lib/api/network/iface';
import { getSwitches } from '$lib/api/network/switch';
import { getVMDomain, getVMs } from '$lib/api/vm/vm';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = 1000 * 60000;
	const vmId = params.node;

	const [vms, domain, interfaces, switches] = await Promise.all([
		cachedFetch('vm-list', async () => getVMs(), cacheDuration),
		cachedFetch(`vm-domain-${vmId}`, async () => getVMDomain(Number(vmId)), cacheDuration),
		cachedFetch('networkInterfaces', async () => await getInterfaces(), cacheDuration),
		cachedFetch('networkSwitches', async () => await getSwitches(), cacheDuration)
	]);

	const vm = vms.find((vm) => vm.vmId === Number(vmId));

	return {
		node: params.node,
		domain,
		interfaces,
		switches,
		vms,
		vm
	};
}
