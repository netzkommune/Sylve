import { getVMDomain, getVMs } from '$lib/api/vm/vm';
import { store as token } from '$lib/stores/auth';
import { sha256 } from '$lib/utils/string';
import { get } from 'svelte/store';

export async function load({ params }) {
	const vms = (await getVMs()) || [];
	const vm = vms.find((vm) => vm.vmId === Number(params.node));
	const domain = await getVMDomain(vm?.vmId || 0);

	let port = 0;
	let password = '';
	let hash = await sha256(get(token), 1);

	if (vm) {
		port = vm.vncPort;
		password = vm.vncPassword;
	}

	return {
		port,
		password,
		domain,
		hash
	};
}
