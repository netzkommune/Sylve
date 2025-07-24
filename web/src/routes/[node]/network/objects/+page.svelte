<script lang="ts">
	import { getNetworkObjects } from '$lib/api/network/object';
	import CreateOrEdit from '$lib/components/custom/Network/Objects/CreateOrEdit.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import type { NetworkObject } from '$lib/types/network/object';
	import { updateCache } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		objects: NetworkObject[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['networkObjects'],
			queryFn: async () => {
				return await getNetworkObjects();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.objects,
			onSuccess: (data: NetworkObject[]) => {
				updateCache('networkObjects', data);
			}
		}
	]);

	let objects = $derived($results[0].data || []);
	let modals = $state({
		create: {
			open: false
		}
	});

	let columns: Column[] = $state([
		{
			field: 'id',
			title: 'ID',
			visible: false
		},
		{
			field: 'name',
			title: 'Name',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value || '-';
			}
		},
		{
			field: 'type',
			title: 'Type',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				switch (value) {
					case 'Host':
						return 'Host(s)';
					case 'Network':
						return 'Network(s)';
					case 'MAC':
						return 'MAC(s)';
					default:
						return value || '-';
				}
			}
		},
		{
			field: 'entries',
			title: 'Data',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				if (Array.isArray(value)) {
					if (typeof value === 'object') {
						if (value && value.length > 0) {
							return value.map((entry) => entry.value).join(', ');
						}
					}
				}

				return value;
			}
		},
		{
			field: 'updatedAt',
			title: 'Updated At',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value ? new Date(value).toLocaleString() : '-';
			}
		}
	]);

	const tableData: { rows: Row[]; columns: Column[] } = {
		columns,
		rows: objects.map((object) => {
			return {
				id: object.id,
				name: object.name,
				type: object.type,
				entries: object.entries,
				updatedAt: object.updatedAt
			};
		})
	};

	let activeRow: Row[] | null = $state(null);
	let query: string = $state('');
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
		<Button size="sm" class="h-6" onclick={() => (modals.create.open = !modals.create.open)}>
			<div class="flex items-center">
				<Icon icon="gg:add" class="mr-1 h-4 w-4" />
				<span>New</span>
			</div>
		</Button>
	</div>

	<TreeTable
		data={tableData}
		name="tt-network-objects"
		multipleSelect={false}
		bind:parentActiveRow={activeRow}
		bind:query
	/>
</div>

{#if modals.create.open}
	<CreateOrEdit bind:open={modals.create.open} />
{/if}
