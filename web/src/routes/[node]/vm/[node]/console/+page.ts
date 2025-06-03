import { getVMs } from '$lib/api/vm/vm';

export async function load({ params }) {
	const vms = (await getVMs()) || [];
	const vm = vms.find((vm) => vm.vmId === Number(params.node));

	let port = 0;
	let password = '';

	if (vm) {
		port = vm.vncPort;
		password = vm.vncPassword;
	}

	return {
		port,
		password
	};
}
