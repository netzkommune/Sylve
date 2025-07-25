<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import { getSwitches } from '$lib/api/network/switch';
	import { detachNetwork } from '$lib/api/vm/network';
	import { getVMDomain, getVMs } from '$lib/api/vm/vm';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Network from '$lib/components/custom/VM/Hardware/Network.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import type { Iface } from '$lib/types/network/iface';
	import type { SwitchList } from '$lib/types/network/switch';
	import type { VM, VMDomain } from '$lib/types/vm/vm';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		vms: VM[];
		vm: VM;
		domain: VMDomain;
		interfaces: Iface[];
		switches: SwitchList;
		node: string;
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['networkInterfaces'],
			queryFn: async () => {
				return await getInterfaces();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.interfaces,
			onSuccess: (data: Iface[]) => {
				updateCache('networkInterfaces', data);
			}
		},
		{
			queryKey: ['networkSwitches'],
			queryFn: async () => {
				return await getSwitches();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.switches,
			onSuccess: (data: SwitchList) => {
				updateCache('networkSwitches', data);
			}
		},
		{
			queryKey: ['vms'],
			queryFn: async () => {
				return getVMs();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.vms,
			onSuccess: (data: VM[]) => {
				updateCache('vms', data);
			}
		},
		{
			queryKey: [`vm-domain-${data.vm.vmId}`],
			queryFn: async () => {
				return await getVMDomain(data.vm.vmId);
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.domain,
			onSuccess: (uData: VMDomain) => {
				updateCache(`vm-domain-${data.vm.vmId}`, uData);
			}
		}
	]);

	let interfaces = $derived($results[0].data || []);
	let switches = $derived($results[1].data || {});
	let vms = $derived($results[2].data || []);
	let vm = $derived(vms.find((vm) => vm.vmId === Number(data.node)));
	let domain = $derived(($results[3].data as VMDomain) || {});

	function generateTableData() {
		const rows: Row[] = [];
		const columns: Column[] = [
			{ field: 'id', title: 'ID', visible: false },
			{ field: 'name', title: 'Name' },
			{ field: 'mac', title: 'MAC Address' },
			{
				field: 'emulation',
				title: 'Emulation',
				formatter(cell: CellComponent, formatterParams, onRendered) {
					const value = cell.getValue();
					if (value === 'virtio') {
						return 'VirtIO';
					} else if (value === 'e1000') {
						return 'E1000';
					}

					return value;
				}
			}
		];

		if (vm?.networks) {
			for (const network of vm.networks) {
				const sw = switches.standard?.find((s) => s.id === network.switchId);
				const row: Row = {
					id: network.id,
					name: sw?.name || 'Unknown Switch',
					mac: network.mac,
					emulation: network.emulation || 'Unknown'
				};

				rows.push(row);
			}
		}

		return { rows, columns };
	}

	let table = $derived(generateTableData());
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
	let query = $state('');
	let usable = $derived.by(() => {
		return switches.standard?.filter((s) => {
			return !vm?.networks.some((n) => n.switchId === s.id);
		});
	});

	let options = {
		attach: {
			open: false
		},
		detach: {
			open: false,
			id: null as number | null,
			name: ''
		}
	};

	let properties = $state(options);
</script>

{#snippet button(type: string)}
	{#if domain && domain.status === 'Shutoff'}
		{#if type === 'detach' && activeRows && activeRows.length === 1}
			<Button
				onclick={() => {
					if (activeRows) {
						properties.detach.open = true;
						properties.detach.id = activeRows[0].id as number;
						properties.detach.name = activeRows[0].name as string;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="gg:remove" class="mr-1 h-4 w-4" />
					<span>Detach</span>
				</div>
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Button
			onclick={() => {
				if (vm) {
					if (usable?.length === 0) {
						toast.error('No available/unused switches to attach to', {
							position: 'bottom-center'
						});

						return;
					}

					properties.attach.open = true;
				}
			}}
			size="sm"
			class="h-6"
			title={domain && domain.status !== 'Shutoff' ? 'VM must be shut off to attach storage' : ''}
			disabled={domain && domain.status !== 'Shutoff'}
		>
			<div class="flex items-center">
				<Icon icon="gg:add" class="mr-1 h-4 w-4" />
				<span>New</span>
			</div>
		</Button>

		{@render button('detach')}
	</div>

	<div class="flex h-full flex-col overflow-hidden">
		<TreeTable
			data={table}
			name={'networks-tt'}
			bind:parentActiveRow={activeRows}
			multipleSelect={false}
			bind:query
		/>
	</div>
</div>

<AlertDialog
	open={properties.detach.open}
	customTitle={`This will detach the VM <b>${vm?.name}</b> from the switch <b>${properties.detach.name}</b>`}
	actions={{
		onConfirm: async () => {
			let response = await detachNetwork(vm?.vmId as number, properties.detach.id as number);
			if (response.status === 'error') {
				handleAPIError(response);
				toast.error('Failed to detach network', {
					position: 'bottom-center'
				});
			} else {
				toast.success('Network detached', {
					position: 'bottom-center'
				});
			}

			properties.detach.open = false;
			activeRows = null;
		},
		onCancel: () => {
			properties.detach.open = false;
			properties = options;
		}
	}}
/>

<Network bind:open={properties.attach.open} {switches} vm={vm ?? null} />
