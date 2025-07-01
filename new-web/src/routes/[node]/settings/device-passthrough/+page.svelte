<script lang="ts">
	import { addPPTDevice, getPCIDevices, getPPTDevices, removePPTDevice } from '$lib/api/system/pci';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import { type PCIDevice, type PPTDevice } from '$lib/types/system/pci';
	import { updateCache } from '$lib/utils/http';
	import { generateTableData } from '$lib/utils/system/pci';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';

	interface Data {
		pciDevices: PCIDevice[];
		pptDevices: PPTDevice[];
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['pciDevices'],
			queryFn: async () => {
				return (await getPCIDevices()) as PCIDevice[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.pciDevices,
			onSuccess: (data: PCIDevice[]) => {
				updateCache('pciDevices', data);
			}
		},
		{
			queryKey: ['pptDevices'],
			queryFn: async () => {
				return (await getPPTDevices()) as PPTDevice[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.pptDevices,
			onSuccess: (data: PPTDevice[]) => {
				updateCache('pptDevices', data);
			}
		}
	]);

	let pciDevices: PCIDevice[] = $derived($results[0].data as PCIDevice[]);
	let pptDevices: PPTDevice[] = $derived($results[1].data as PPTDevice[]);
	let tableData = $derived(generateTableData(pciDevices, pptDevices));
	let tableName: string = 'device-passthrough-tt';
	let query: string = $state('');
	let activeRow: Row[] | null = $state(null);

	let modalState = $state({
		isOpen: false,
		title: '',
		action: '',
		add: {
			domain: '',
			deviceId: ''
		},
		remove: {
			id: 0
		}
	});

	function addDevice(domain: string, deviceId: string) {
		const device = activeRow ? activeRow[0].device : '';
		const vendor = activeRow ? activeRow[0].vendor : '';

		modalState.isOpen = true;
		modalState.title = `Are you sure you want to pass through <b>${device}</b> by <b>${vendor}</b>? This will make it unavailable to the host.`;
		modalState.action = 'add';
		modalState.add.domain = domain;
		modalState.add.deviceId = deviceId;
	}

	function removeDevice(id: number) {
		const device = activeRow ? activeRow[0].device : '';
		const vendor = activeRow ? activeRow[0].vendor : '';
		modalState.isOpen = true;
		modalState.title = `Are you sure you want to remove passthrough for <b>${device}</b> by <b>${vendor}</b>? This will make it available to the host again.`;
		modalState.action = 'remove';
		modalState.remove.id = id;
	}
</script>

{#snippet button(type: string)}
	{#if activeRow !== null && activeRow.length === 1}
		{#if type === 'enable-passthrough' && !activeRow[0].name.startsWith('ppt')}
			<Button
				onclick={() =>
					activeRow && addDevice(activeRow[0].domain.toString(), activeRow[0].deviceId)}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<Icon icon="wpf:disconnected" class="mr-1 h-4 w-4" />
				Enable Passthrough
			</Button>
		{/if}

		{#if type === 'disable-passthrough' && activeRow[0].name.startsWith('ppt')}
			<Button
				onclick={() => activeRow && removeDevice(activeRow[0].pptId)}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<Icon icon="wpf:connected" class="mr-1 h-4 w-4" />
				Disable Passthrough
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />

		{@render button('enable-passthrough')}
		{@render button('disable-passthrough')}
	</div>

	<div class="flex h-full flex-col overflow-hidden">
		<TreeTable
			data={tableData}
			name={tableName}
			bind:parentActiveRow={activeRow}
			bind:query
			multipleSelect={false}
		/>
	</div>
</div>

<AlertDialog
	open={modalState.isOpen}
	names={{ parent: '', element: modalState?.title || '' }}
	customTitle={modalState.title}
	actions={{
		onConfirm: async () => {
			if (modalState.action === 'add') {
				const result = await addPPTDevice(modalState.add.domain, modalState.add.deviceId);
				if (result.status === 'success') {
					toast.success('Device added to passthrough', {
						position: 'bottom-center'
					});
				} else {
					toast.error('Failed to add device to passthrough', {
						position: 'bottom-center'
					});
				}

				modalState.isOpen = false;
			}

			if (modalState.action === 'remove') {
				const result = await removePPTDevice(modalState.remove.id.toString());

				if (result.status === 'success') {
					toast.success('Device removed from passthrough', {
						position: 'bottom-center'
					});
				} else {
					toast.error('Failed to remove device from passthrough', {
						position: 'bottom-center'
					});
				}

				modalState.isOpen = false;
			}
		},
		onCancel: () => {
			modalState.isOpen = false;
		}
	}}
></AlertDialog>
