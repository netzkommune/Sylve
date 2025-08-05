<script lang="ts">
	import { newJail } from '$lib/api/jail/jail';
	import { getNetworkObjects } from '$lib/api/network/object';
	import { getSwitches } from '$lib/api/network/switch';
	import { getDownloads } from '$lib/api/utilities/downloader';
	import { getVMs } from '$lib/api/vm/vm';
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import type { CreateData } from '$lib/types/jail/jail';
	import type { NetworkObject } from '$lib/types/network/object';
	import type { SwitchList } from '$lib/types/network/switch';
	import type { Download } from '$lib/types/utilities/downloader';
	import type { VM } from '$lib/types/vm/vm';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import { handleAPIError } from '$lib/utils/http';
	import { isValidCreateData } from '$lib/utils/jail/jail';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';
	import Basic from './Basic.svelte';
	import Hardware from './Hardware.svelte';
	import Network from './Network.svelte';
	import Storage from './Storage.svelte';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();
	const tabs = [
		{ value: 'basic', label: 'Basic' },
		{ value: 'storage', label: 'Storage' },
		{ value: 'network', label: 'Network' },
		{ value: 'hardware', label: 'Hardware & Advanced' }
	];

	const results = useQueries([
		{
			queryKey: ['datasetList-svm'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: [],
			refetchOnMount: 'always'
		},
		{
			queryKey: ['downloads-svm'],
			queryFn: async () => {
				return await getDownloads();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: [],
			refetchOnMount: 'always'
		},
		{
			queryKey: ['networkSwitches-svm'],

			queryFn: async () => {
				return await getSwitches();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: {} as SwitchList,
			refetchOnMount: 'always'
		},
		{
			queryKey: ['vms-svm'],
			queryFn: async () => {
				return await getVMs();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: [],
			refetchOnMount: 'always'
		},
		{
			queryKey: ['network-objects-svm'],
			queryFn: async () => {
				return await getNetworkObjects();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: [],
			refetchOnMount: 'always'
		}
	]);

	let datasets: Dataset[] = $derived($results[0].data as Dataset[]);
	let downloads = $derived($results[1].data as Download[]);
	let networkSwitches: SwitchList = $derived($results[2].data as SwitchList);
	let networkObjects = $derived($results[4].data as NetworkObject[]);
	let vms: VM[] = $derived($results[3].data as VM[]);

	let filesystems: Dataset[] = $derived(
		datasets.filter((dataset) => dataset.type === 'filesystem')
	);

	let options = {
		name: '',
		id: 0,
		description: '',
		storage: {
			dataset: '',
			base: ''
		},
		network: {
			switch: 0,
			mac: 0,
			ipv4: 0,
			ipv4Gateway: 0,
			ipv6: 0,
			ipv6Gateway: 0,
			dhcp: false,
			slaac: false
		},
		hardware: {
			cpuCores: 1,
			ram: 0,
			startAtBoot: false,
			bootOrder: 0
		}
	};

	let modal: CreateData = $state(options);

	function resetModal() {
		modal = options;
	}

	async function create() {
		const data = $state.snapshot(modal);
		if (!isValidCreateData(data)) {
			return;
		} else {
			const response = await newJail(data);
			if (response.error) {
				handleAPIError(response);
				toast.error('Failed to create jail');
				return;
			}

			open = false;
			toast.success(`Jail ${data.name} created`);
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 flex h-[85vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform flex-col gap-0  overflow-auto p-5 transition-all duration-300 ease-in-out lg:h-[64vh] lg:max-w-2xl"
	>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex  justify-between gap-1 text-left">
				<div class="flex items-center gap-2">
					<Icon icon="hugeicons:prison" class="h-5 w-5 " />
					Create Jail
				</div>
				<div class="flex items-center gap-0.5">
					<Button size="sm" variant="link" class="h-4" onclick={() => resetModal()} title={'Reset'}>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">{'Reset'}</span>
					</Button>
					<Button
						size="sm"
						variant="link"
						class="h-4"
						onclick={() => (open = false)}
						title={'Close'}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">{'Close'}</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		<div class="mt-6 flex-1 overflow-y-auto">
			<Tabs.Root value="basic" class="w-full overflow-hidden">
				<Tabs.List class="grid w-full grid-cols-4 p-0 ">
					{#each tabs as { value, label }}
						<Tabs.Trigger class="border-b" {value}>{label}</Tabs.Trigger>
					{/each}
				</Tabs.List>

				{#each tabs as { value, label }}
					<Tabs.Content {value}>
						<div>
							{#if value === 'basic'}
								<Basic
									bind:name={modal.name}
									bind:id={modal.id}
									bind:description={modal.description}
								/>
							{:else if value === 'storage'}
								<Storage
									{filesystems}
									{downloads}
									bind:dataset={modal.storage.dataset}
									bind:base={modal.storage.base}
								/>
							{:else if value === 'network'}
								<Network
									bind:switch={modal.network.switch}
									bind:mac={modal.network.mac}
									bind:ipv4={modal.network.ipv4}
									bind:ipv4Gateway={modal.network.ipv4Gateway}
									bind:ipv6={modal.network.ipv6}
									bind:ipv6Gateway={modal.network.ipv6Gateway}
									bind:dhcp={modal.network.dhcp}
									bind:slaac={modal.network.slaac}
									switches={networkSwitches}
									{networkObjects}
								/>
							{:else if value === 'hardware'}
								<Hardware
									bind:cpuCores={modal.hardware.cpuCores}
									bind:ram={modal.hardware.ram}
									bind:startAtBoot={modal.hardware.startAtBoot}
									bind:bootOrder={modal.hardware.bootOrder}
								/>
							{/if}
						</div>
					</Tabs.Content>
				{/each}
			</Tabs.Root>
		</div>

		<Dialog.Footer>
			<div class="flex w-full justify-end md:flex-row">
				<Button size="sm" type="button" class="h-8" onclick={() => create()}>Create Jail</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
