<script lang="ts">
	import { page } from '$app/state';
	import { getVMs } from '$lib/api/vm/vm';
	import Button from '$lib/components/ui/button/button.svelte';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import RFB from '@novnc/novnc/lib/rfb.js';
	import { onDestroy, onMount, tick } from 'svelte';

	interface Data {
		port: number;
		password: string;
	}

	let { data }: { data: Data } = $props();

	let status: string = $state('');
	let rfb: RFB | null = $state(null);
	let screen: HTMLDivElement;
	const options = {
		credentials: { password: data.password }
	};

	let vnc = $state({
		password: data.password,
		path: `/api/vnc/${encodeURIComponent(String(data.port))}`
	});

	function setStatus(newStatus: string) {
		let color = '';
		switch (newStatus) {
			case 'connected':
				color = 'bg-green-600';
				break;
			case 'disconnected':
				color = 'bg-red-600';
				break;
			case 'connecting':
				color = 'bg-yellow-600';
				break;
			default:
				color = 'bg-gray-600';
		}

		status = `
        <div class="flex items-center gap-2">
            <div class="h-4 w-4 rounded-full ${color}"></div>
            <span>${capitalizeFirstLetter(newStatus)}</span>
        </div>`;
	}

	function connectVNC() {
		rfb = new RFB(screen, vnc.path, options);

		rfb.addEventListener('connect', () => {
			setStatus('connected');
		});

		rfb.addEventListener('disconnect', () => {
			setStatus('disconnected');
		});
	}

	function disconnectVNC() {
		if (rfb) {
			rfb.disconnect();
			rfb = null;
		}
	}

	onMount(async () => {
		setStatus('connecting');
		connectVNC();
	});

	onDestroy(() => {
		disconnectVNC();
	});
</script>

<div class="flex h-full min-h-0 w-full flex-col">
	<div class="flex h-10 w-full items-center justify-between gap-2 border p-2">
		<div class="flex items-center gap-2">
			<p>{@html status}</p>

			<Button
				size="sm"
				variant="ghost"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
				title={'Reconnect'}
				onclick={() => {
					disconnectVNC();
					connectVNC();
				}}
			>
				<Icon icon="mdi:restart" class="pointer-events-none mr-2 h-4 w-4" />
				<span>Reconnect</span>
			</Button>
		</div>
	</div>

	<div id="screen" class="w-full flex-1 overflow-auto" bind:this={screen}></div>
</div>

<style>
	:global(#screen canvas) {
		width: 95% !important;
		height: 95% !important;
	}
</style>
