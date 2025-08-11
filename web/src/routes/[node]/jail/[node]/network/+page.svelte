<script lang="ts">
	import { page } from '$app/state';
	import { disinheritHostNetwork, getJails, inheritHostNetwork } from '$lib/api/jail/jail';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Jail, JailStat, JailState } from '$lib/types/jail/jail';
	import { updateCache } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		jails: Jail[];
		jailStates: JailState[];
		jail: Jail;
	}

	let { data }: { data: Data } = $props();
	const ctId = page.url.pathname.split('/')[3];
	const results = useQueries([
		{
			queryKey: ['jail-list'],
			queryFn: async () => {
				return await getJails();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.jails,
			onSuccess: (data: Jail[]) => {
				updateCache('jail-list', data);
			}
		}
	]);

	let jails = $derived($results[0].data || []);
	let jail = $derived.by(() => {
		if (jails.length > 0) {
			const found = jails.find((j) => j.ctId === parseInt(ctId));
			return found || data.jail;
		}

		return data.jail;
	});

	let inherited = $derived(jail.inheritIPv4 || jail.inheritIPv6);

	async function changeInheritance() {
		if (inherited) {
			const response = await disinheritHostNetwork(jail.ctId);
			console.log(response);
		} else {
			const response = await inheritHostNetwork(jail.ctId, true, true);
			console.log(response);
		}
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Button
			onclick={() => {
				changeInheritance();
			}}
			size="sm"
			variant="outline"
			class="h-6.5"
		>
			<div class="flex items-center">
				{#if inherited}
					<Icon icon="mdi:close-network" class="mr-1 h-4 w-4" />
					<span>Disinherit Network</span>
				{:else}
					<Icon icon="mdi:plus-network" class="mr-1 h-4 w-4" />
					<span>Inherit Network</span>
				{/if}
			</div>
		</Button>
	</div>
</div>
