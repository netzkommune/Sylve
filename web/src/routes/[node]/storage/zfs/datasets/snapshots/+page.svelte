<script lang="ts">
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { generateTableData } from '$lib/utils/zfs/dataset/snapshot';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
	}

	let { data }: { data: Data } = $props();
	let tableName = 'tt-zfsSnapshots';
	const results = useQueries([
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.pools
		},
		{
			queryKey: ['datasetList'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.datasets
		}
	]);

	let activeRow: Row | null = $state(null);
	let grouped: GroupedByPool[] = $derived(groupByPool($results[0].data, $results[1].data));
	let tableData = $derived(generateTableData(grouped));
	let activePool: Zpool | null = $derived.by(() => {
		if (activeRow) {
			const poolGroup = grouped.find((pool) => pool.name === activeRow?.name);
			console.log(poolGroup?.pool);
			return poolGroup ? poolGroup.pool : null;
		}
		return null;
	});

	let activeDataset: Dataset | null = $derived.by(() => {
		if (activeRow) {
			for (const poolGroup of grouped) {
				const snapshots = poolGroup.snapshots.filter(
					(snapshot) => snapshot.name === activeRow?.name
				);
				if (snapshots) {
					return snapshots.find((dataset) => dataset.name === activeRow?.name) || null;
				}
			}
		}

		return null;
	});

	let query = $state('');
</script>

{#snippet button(type: string)}
	{#if type === 'delete-snapshot' && activeDataset !== null}
		<Button
			on:click={async () => {}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Snapshot
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />

		<Button
			on:click={() => {
				// confirmModals.active = 'createFilesystem';
				// confirmModals.parent = 'filesystem';
				// confirmModals.createFilesystem.open = true;
				// confirmModals.createFilesystem.title = '';
			}}
			size="sm"
			class="h-6"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>

		<!-- {@render button('create-snapshot')}
		{@render button('rollback-snapshot')}
		{@render button('delete-snapshot')}
		{@render button('delete-filesystem')} -->
		<!-- {@render button('create-snapshot')} -->

		{@render button('create-snapshot')}
		{@render button('delete-snapshot')}
	</div>

	<TreeTable data={tableData} name={tableName} bind:parentActiveRow={activeRow} bind:query />
</div>
