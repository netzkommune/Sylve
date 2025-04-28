<script lang="ts">
	import { deleteSnapshot, getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { handleError } from '$lib/utils/zfs/dataset/fs';
	import { generateTableData } from '$lib/utils/zfs/dataset/snapshot';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import toast from 'svelte-french-toast';

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
	let confirmModals = $state({
		active: '' as 'deleteSnapshot',
		parent: 'filesystem' as 'filesystem' | 'snapshot',
		deleteSnapshot: {
			open: false,
			recursive: false,
			data: '',
			title: ''
		}
	});

	async function confirmAction() {
		if (confirmModals.active === 'deleteSnapshot') {
			if (activeDataset) {
				const response = await deleteSnapshot(
					activeDataset,
					confirmModals.deleteSnapshot.recursive
				);

				if (response.error) {
					handleError(response);
					return;
				}

				toast.success(
					`${capitalizeFirstLetter(getTranslation('common.snapshot', 'snapshot'))} ${activeDataset.name} ${getTranslation('common.deleted', 'deleted')}`,
					{
						position: 'bottom-center'
					}
				);

				activeRow = null;
			}
		}
	}
</script>

{#snippet button(type: string)}
	{#if type === 'delete-snapshot' && activeDataset !== null}
		<Button
			on:click={async () => {
				confirmModals.active = 'deleteSnapshot';
				confirmModals.parent = 'snapshot';
				confirmModals.deleteSnapshot.open = true;
				confirmModals.deleteSnapshot.data = activeDataset.name;
				confirmModals.deleteSnapshot.title = activeDataset.name;
			}}
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

{#if confirmModals.active === 'deleteSnapshot'}
	<AlertDialog.Root
		bind:open={confirmModals[confirmModals.active].open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>{getTranslation('are_you_sure', 'Are you sure?')}</AlertDialog.Title>
			</AlertDialog.Header>

			<div class="text-muted-foreground mb-2 text-sm">
				{getTranslation(
					'common.permanent_delete_msg',
					'This action cannot be undone. This will permanently delete'
				)}
				{confirmModals.parent}
				<span class="font-semibold">{confirmModals[confirmModals.active].title}</span>.
			</div>

			<div class="flex items-center gap-2">
				<Checkbox
					id="deleteRecursive"
					bind:checked={confirmModals[confirmModals.active].recursive}
					aria-labelledby="deleteRecursive-label"
				/>
				<Label
					id="deleteRecursive-label"
					for="deleteRecursive"
					class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
					onclick={() => {
						confirmModals[confirmModals.active].recursive =
							!confirmModals[confirmModals.active].recursive;
					}}
				>
					Recursive
				</Label>
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel
					on:click={() => {
						confirmModals[confirmModals.active].open = false;
					}}
				>
					Cancel
				</AlertDialog.Cancel>
				<AlertDialog.Action
					on:click={() => {
						confirmAction();
						// console.log('Delete action triggered');
					}}
				>
					Delete
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
