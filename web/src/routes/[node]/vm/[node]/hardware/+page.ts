import { getRAMInfo } from '$lib/api/info/ram.js';
import { getVMDomain, getVMs } from '$lib/api/vm/vm';

export async function load({ params }) {
	const vms = (await getVMs()) || [];
	const ram = await getRAMInfo();
	const vm = vms.find((vm) => vm.vmId === Number(params.node));
	const domain = await getVMDomain(Number(vm?.vmId));

	return {
		vm,
		vms,
		ram,
		domain
	};
}
