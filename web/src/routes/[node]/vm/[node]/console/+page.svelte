<script lang="ts">
	import { page } from '$app/state';
	import { getVMs } from '$lib/api/vm/vm';
	import VncViewer from '$lib/components/custom/VNCViewer.svelte';
	import { onMount } from 'svelte';

	let vnc = $state({
		port: 0,
		password: ''
	});

	onMount(async () => {
		const vmList = await getVMs();
		const currentVmId = page.url.pathname.split('/')[3];
		const currentVm = vmList.find((vm) => vm.vmId === Number(currentVmId));

		if (currentVm && currentVm.vncPort) {
			vnc.port = currentVm.vncPort;
			vnc.password = currentVm.vncPassword || '';
		}
	});
</script>

{#if vnc.port !== 0}
	<VncViewer vncPort={vnc.port} vncPassword={vnc.password} />
{/if}
