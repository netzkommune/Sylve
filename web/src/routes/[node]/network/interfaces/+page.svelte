<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import KvTableModal from '$lib/components/custom/KVTableModal.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { type Iface } from '$lib/types/network/iface';
	import { updateCache } from '$lib/utils/http';
	import { generateTableData, getCleanIfaceData } from '$lib/utils/network/iface';
	import { renderWithIcon } from '$lib/utils/table';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		interfaces: Iface[];
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
		}
	]);

	let columns: Column[] = $derived([
		{
			field: 'id',
			title: 'ID',
			visible: false
		},
		{
			field: 'name',
			title: 'Name',
			formatter(cell: CellComponent) {
				const value = cell.getValue();
				const row = cell.getRow();
				const data = row.getData();

				if (data.isBridge) {
					const name = data.description || value;
					return renderWithIcon('clarity:network-switch-line', name);
				}

				if (value === 'lo0') {
					return renderWithIcon('ic:baseline-loop', value);
				}

				return renderWithIcon('mdi:ethernet', value);
			}
		},
		{
			field: 'model',
			title: 'Model'
		},
		{
			field: 'description',
			title: 'Description',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				if (value) {
					return value;
				}

				return '-';
			}
		},
		{
			field: 'ether',
			title: 'MAC Address',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value || '-';
			}
		},
		{
			field: 'metric',
			title: 'Metric'
		},
		{
			field: 'mtu',
			title: 'MTU'
		},
		{
			field: 'media',
			title: 'Status',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				const status = value?.status || '-';
				if (status === 'active') {
					return 'Active';
				}

				return status;
			}
		},
		{
			field: 'isBridge',
			title: 'isBridge',
			visible: false
		}
	]);

	let tableData = $derived(generateTableData(columns, $results[0].data as Iface[]));
	let activeRow: Row[] | null = $state(null);
	let query: string = $state('');
	let viewModal = $state({
		title: '',
		key: 'Attribute',
		value: 'Value',
		open: false,
		KV: {},
		type: 'kv',
		actions: {
			close: () => {
				viewModal.open = false;
			}
		}
	});

	function viewInterface(iface: string) {
		const ifaceData = $results[0].data?.find((i: Iface) => i.name === iface);
		if (ifaceData) {
			viewModal.KV = getCleanIfaceData(ifaceData);
			viewModal.title = `Details - ${ifaceData.name}`;
			viewModal.open = true;
		}
	}
</script>

{#snippet button(type: string)}
	{#if type === 'view' && activeRow !== null && activeRow.length > 0}
		<Button
			onclick={() => activeRow !== null && viewInterface(activeRow[0]?.name)}
			size="sm"
			variant="outline"
			class="h-6.5"
		>
			<Icon icon="mdi:eye" class="mr-1 h-4 w-4" />
			{'View'}
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
		{@render button('view')}
	</div>

	<KvTableModal
		titles={{
			icon: 'carbon:network-interface',
			main: viewModal.title,
			key: viewModal.key,
			value: viewModal.value
		}}
		open={viewModal.open}
		KV={viewModal.KV}
		type={viewModal.type}
		actions={viewModal.actions}
	></KvTableModal>

	<TreeTable
		data={tableData}
		name="tt-networkInterfaces"
		multipleSelect={false}
		bind:parentActiveRow={activeRow}
		bind:query
	/>
</div>
