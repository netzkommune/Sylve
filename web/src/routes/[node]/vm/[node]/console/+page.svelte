<script lang="ts">
	import type { VMDomain } from '$lib/types/vm/vm';
	import Icon from '@iconify/svelte';

	interface Data {
		port: number;
		password: string;
		domain: VMDomain;
		hash: string;
	}

	let { data }: { data: Data } = $props();
	let path = $derived(`/api/vnc/${encodeURIComponent(String(data.port))}?hash=${data.hash}`);

	let revealIframe = $state(false);

	if (data.domain && data.domain.status !== 'Shutoff') {
		setTimeout(() => {
			revealIframe = true;
		}, 1000);
	}
</script>

{#if data.domain && data.domain.status !== 'Shutoff'}
	<div class="relative flex h-full w-full flex-col">
		<iframe
			class="w-full flex-1 transition-opacity duration-500"
			class:opacity-0={!revealIframe}
			class:opacity-100={revealIframe}
			src={`/vnc/vnc.html?path=${path}&password=${data.password}`}
			title="VM Console"
		></iframe>

		{#if !revealIframe}
			<div class="absolute inset-0 z-10 flex items-center justify-center">
				<Icon icon="mdi:loading" class="text-primary h-10 w-10 animate-spin" />
			</div>
		{/if}
	</div>
{:else}
	<div
		class="text-primary dark:text-secondary flex h-full w-full flex-col items-center justify-center space-y-3 text-center text-base"
	>
		<Icon icon="mdi:server-off" class="dark:text-secondary text-primary h-14 w-14" />
		<div class="max-w-md">
			The VM is currently powered off.<br />
			Start the VM to access its console.
		</div>
	</div>
{/if}
