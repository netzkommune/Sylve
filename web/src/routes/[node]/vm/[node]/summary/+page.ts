import { getVMDomain, getVMs } from '$lib/api/vm/vm';
import { SEVEN_DAYS } from '$lib/utils.js';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	const cacheDuration = SEVEN_DAYS;
	const vmId = params.node;

	const [vms, domain] = await Promise.all([
		cachedFetch('vm-list', async () => getVMs(), cacheDuration),
		cachedFetch(`vm-domain-${vmId}`, async () => getVMDomain(Number(vmId)), cacheDuration)
	]);

	return {
		vms: vms,
		domain: domain
	};
}
