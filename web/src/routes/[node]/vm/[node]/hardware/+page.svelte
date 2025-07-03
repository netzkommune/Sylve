<script lang="ts">
	import { getVMs } from '$lib/api/vm/vm';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import CPU from '$lib/components/custom/VM/Hardware/CPU.svelte';
	import RAM from '$lib/components/custom/VM/Hardware/RAM.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Row } from '$lib/components/ui/table';
	import type { RAMInfo } from '$lib/types/info/ram';
	import type { VM, VMDomain } from '$lib/types/vm/vm';
	import { updateCache } from '$lib/utils/http';
	import { bytesToHumanReadable } from '$lib/utils/numbers';
	import { generateNanoId } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		vms: VM[];
		vm: VM;
		ram: RAMInfo;
		domain: VMDomain;
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['vm-list'],
			queryFn: async () => {
				return await getVMs();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.vms,
			onSuccess: (data: VM[]) => {
				updateCache('vm-list', data);
			}
		}
	]);

	let vms: VM[] = $derived($results[0].data ? $results[0].data : data.vms);
	let vm: VM | null = $derived(vms ? (vms.find((v: VM) => v.vmId === data.vm.vmId) ?? null) : null);

	let options = {
		cpu: {
			sockets: data.vm.cpuSockets,
			cores: data.vm.cpuCores,
			threads: data.vm.cpuThreads,
			pinning: data.vm.cpuPinning,
			vCPUs: data.vm.cpuSockets * data.vm.cpuCores * data.vm.cpuThreads,
			open: false
		},
		ram: {
			value: data.vm.ram,
			open: false
		}
	};

	let properties = $state(options);

	$effect(() => {
		if (vm) {
			properties.cpu.sockets = vm.cpuSockets;
			properties.cpu.cores = vm.cpuCores;
			properties.cpu.threads = vm.cpuThreads;
			properties.cpu.vCPUs = vm.cpuSockets * vm.cpuCores * vm.cpuThreads;
			properties.ram.value = vm.ram;
		}
	});

	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
	let query = $state('');

	let table = $derived({
		columns: [
			{ title: 'Property', field: 'property' },
			{ title: 'Value', field: 'value' }
		],
		rows: [
			{
				id: generateNanoId(`${properties.cpu.vCPUs}-vcpus`),
				property: 'vCPUs',
				value: properties.cpu.vCPUs
			},
			{
				id: generateNanoId(`${properties.ram.value}-ram`),
				property: 'RAM',
				value: bytesToHumanReadable(properties.ram.value)
			}
		]
	});
</script>

<div class="flex h-full w-full flex-col">
	{#if activeRows && activeRows?.length !== 0}
		<div class="flex h-10 w-full items-center gap-2 border-b p-2">
			{#if activeRow && activeRow.property === 'RAM'}
				<Button
					onclick={() => {
						properties.ram.open = true;
					}}
					size="sm"
					variant="outline"
					class="h-6.5"
					title={data.domain.status === 'Shutoff'
						? ''
						: 'RAM can only be edited when the VM is shut off'}
					disabled={data.domain.status !== 'Shutoff'}
				>
					<div class="flex items-center">
						<Icon icon="mdi:pencil" class="mr-1 h-4 w-4" />
						<span>Edit RAM</span>
					</div>
				</Button>
			{/if}

			{#if activeRow && activeRow.property === 'vCPUs'}
				<Button
					onclick={() => {
						properties.cpu.open = true;
					}}
					size="sm"
					variant="outline"
					class="h-6.5"
					title={data.domain.status === 'Shutoff'
						? ''
						: 'CPU can only be edited when the VM is shut off'}
					disabled={data.domain.status !== 'Shutoff'}
				>
					<div class="flex items-center">
						<Icon icon="mdi:pencil" class="mr-1 h-4 w-4" />
						<span>Edit CPU</span>
					</div>
				</Button>
			{/if}
		</div>
	{/if}

	<div class="flex h-full flex-col overflow-hidden">
		<TreeTable
			data={table}
			name={'hardware-tt'}
			bind:parentActiveRow={activeRows}
			multipleSelect={false}
			bind:query
		/>
	</div>
</div>

{#if properties.ram.open}
	<RAM bind:open={properties.ram.open} ram={data.ram} {vm} />
{/if}

{#if properties.cpu.open}
	<CPU bind:open={properties.cpu.open} {vm} {vms} />
{/if}
