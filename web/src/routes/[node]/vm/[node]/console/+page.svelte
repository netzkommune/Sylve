<script lang="ts">
	import { page } from '$app/state';
	import { getVMs } from '$lib/api/vm/vm';
	import { onMount } from 'svelte';
	import NoVNC from 'svelte-vnc';

	let vnc = $state({
		password: '',
		host: window.location.hostname,
		port: 0 as number | null,
		path: ''
	});

	onMount(async () => {
		const vmList = await getVMs();
		const currentVmId = page.url.pathname.split('/')[3];
		const currentVm = vmList.find((vm) => vm.vmId === Number(currentVmId));

		if (currentVm && currentVm.vncPort) {
			vnc.port = window.location.port ? Number(window.location.port) : null;
			vnc.password = currentVm.vncPassword || '';
			vnc.path = `api/vnc/${encodeURIComponent(String(currentVm.vncPort))}`;
		}
	});
</script>

{#if vnc.port !== 0}
	<NoVNC
		class="h-full w-full"
		host={vnc.host}
		port={vnc.port}
		path={vnc.path}
		password={vnc.password}
		clearLocalStorage={true}
		reconnect={true}
		shared={false}
	/>
{/if}
