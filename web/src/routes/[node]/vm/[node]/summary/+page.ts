import { getVMDomain, getVMs } from '$lib/api/vm/vm';
import { cachedFetch } from '$lib/utils/http';

export async function load({ params }) {
	let vmId = params.node;

	const [vms, domain] = await Promise.all([
		cachedFetch('vm-list', async () => getVMs(), 7 * 24 * 60 * 60 * 1000),
		cachedFetch(`vm-domain-${vmId}`, async () => getVMDomain(Number(vmId)), 7 * 24 * 60 * 60 * 1000)
	]);

	return {
		vms: vms,
		domain: domain
	};
}
