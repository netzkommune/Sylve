<script lang="ts">
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { generateTableData, groupByPool } from '$lib/utils/zfs/dataset';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.pools
		},
		{
			queryKey: ['datasetList'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.datasets
		}
	]);

	let grouped = $derived(groupByPool($results[0].data, $results[1].data));
	let tableData = $derived(generateTableData(grouped));
	let activeRow: Row | null = $state(null);
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Button
			on:click={() => console.log('New dataset')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
	<div class="relative flex h-full w-full cursor-pointer flex-col">
		<div class="flex-1">
			<div class="h-full overflow-y-auto">
				<TreeTable
					data={tableData}
					name="tt-zfsDatasets"
					parentIcon={'carbon:partition-collection'}
					itemIcon={'eos-icons:file-system'}
					bind:parentActiveRow={activeRow}
				/>
			</div>
		</div>
	</div>
</div>
