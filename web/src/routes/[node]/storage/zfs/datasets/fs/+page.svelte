<script lang="ts">
	import {
		bulkDelete,
		deleteFileSystem,
		getDatasets,
		rollbackSnapshot
	} from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import CreateFS from '$lib/components/custom/ZFS/datasets/fs/Create.svelte';
	import EditFS from '$lib/components/custom/ZFS/datasets/fs/Edit.svelte';
	import CreateSnapshot from '$lib/components/custom/ZFS/datasets/snapshots/Create.svelte';
	import DeleteSnapshot from '$lib/components/custom/ZFS/datasets/snapshots/Delete.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { createFSProps, generateTableData, handleError } from '$lib/utils/zfs/dataset/fs';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
	}

	let { data }: { data: Data } = $props();
	let tableName = 'tt-zfsDatasets';

	const results = useQueries([
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.pools,
			onSuccess: (data: Zpool[]) => {
				updateCache('pools', data);
			}
		},
		{
			queryKey: ['zfs-filesystems'],
			queryFn: async () => {
				return await getDatasets('filesystem');
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.datasets,
			onSuccess: (data: Dataset[]) => {
				updateCache('zfs-filesystems', data);
			}
		}
	]);

	let pools: Zpool[] = $derived($results[0].data as Zpool[]);
	let datasets: Dataset[] = $derived($results[1].data as Dataset[]);
	let grouped = $derived(groupByPool(pools, datasets));
	let tableData = $derived(generateTableData(grouped));
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));

	let activeDataset: Dataset | null = $derived.by(() => {
		if (activeRow) {
			for (const dataset of grouped) {
				const filesystems = dataset.filesystems;
				const snapshots = dataset.snapshots;

				for (const fs of filesystems) {
					if (fs.name === activeRow.name) {
						return fs;
					}
				}

				for (const snap of snapshots) {
					if (snap.name === activeRow.name) {
						return snap;
					}
				}
			}
		}

		return null;
	});

	let activeDatasets: Dataset[] = $derived.by(() => {
		if (activeRows) {
			let datasets: Dataset[] = [];
			for (const row of activeRows) {
				for (const dataset of grouped) {
					const filesystems = dataset.filesystems;
					const snapshots = dataset.snapshots;

					for (const fs of filesystems) {
						if (fs.name === row.name) {
							datasets.push(fs);
						}
					}

					for (const snap of snapshots) {
						if (snap.name === row.name) {
							datasets.push(snap);
						}
					}
				}
			}
			return datasets;
		}

		return [];
	});

	let poolsSelected = $derived.by(() => {
		if (activeRows && activeRows.length > 0) {
			const filtered = activeRows.filter((row) => {
				return row.type === 'pool';
			});

			return filtered.length > 0;
		}

		return false;
	});

	let zfsProperties = $state(createFSProps);
	let query: string = $state('');

	let modals = $state({
		snapshot: {
			create: {
				open: false
			},
			rollback: {
				open: false
			},
			delete: {
				open: false
			}
		},
		fs: {
			create: {
				open: false
			},
			edit: {
				open: false
			},
			delete: {
				open: false
			}
		},
		bulk: {
			delete: {
				open: false,
				title: ''
			}
		}
	});
</script>

{#snippet button(type: string)}
	{#if activeRows && activeRows.length == 1}
		{#if type === 'rollback-snapshot' && activeDataset?.type === 'snapshot'}
			<Button
				onclick={async () => {
					if (activeDataset) {
						modals.snapshot.rollback.open = true;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:history" class="mr-1 h-4 w-4" />
					<span>Rollback To Snapshot</span>
				</div>
			</Button>
		{/if}

		{#if type === 'create-snapshot' && activeDataset?.type === 'filesystem'}
			<Button
				onclick={async () => {
					if (activeDataset) {
						modals.snapshot.create.open = true;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-1 h-4 w-4" />
					<span>Create Snapshot</span>
				</div>
			</Button>
		{/if}

		{#if type === 'delete-snapshot' && activeDataset?.type === 'snapshot'}
			<Button
				onclick={async () => {
					if (activeDataset) {
						modals.snapshot.delete.open = true;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					<span>Delete Snapshot</span>
				</div>
			</Button>
		{/if}

		{#if type === 'edit-filesystem' && activeDataset?.type === 'filesystem'}
			<Button
				onclick={async () => {
					if (activeDataset) {
						modals.fs.edit.open = true;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:pencil" class="mr-1 h-4 w-4" />
					<span>Edit Filesystem</span>
				</div>
			</Button>
		{/if}

		{#if type === 'delete-filesystem' && activeDataset?.type === 'filesystem' && activeDataset?.name.includes('/')}
			<Button
				onclick={async () => {
					if (activeDataset) {
						modals.fs.delete.open = true;
					}
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					<span>Delete Filesystem</span>
				</div>
			</Button>
		{/if}
	{:else if activeRows && activeRows.length > 1}
		{#if activeDatasets.length > 0 && !poolsSelected}
			{#if type === 'bulk-delete'}
				<Button
					onclick={async () => {
						let [snapLen, fsLen] = [0, 0];
						activeDatasets.forEach((dataset) => {
							if (dataset.type === 'snapshot') {
								snapLen++;
							} else if (dataset.type === 'filesystem') {
								fsLen++;
							}
						});

						let title = '';
						if (snapLen > 0 && fsLen > 0) {
							title = `${snapLen} snapshot${snapLen > 1 ? 's' : ''} and ${fsLen} filesystem${fsLen > 1 ? 's' : ''}`;
						} else if (snapLen > 0) {
							title = `${snapLen} snapshot${snapLen > 1 ? 's' : ''}`;
						} else if (fsLen > 0) {
							title = `${fsLen} filesystem${fsLen > 1 ? 's' : ''}`;
						}

						modals.bulk.delete.open = true;
						modals.bulk.delete.title = title;
					}}
					size="sm"
					variant="outline"
					class="h-6.5"
				>
					<div class="flex items-center">
						<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
						<span>Delete Datasets</span>
					</div>
				</Button>
			{/if}
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
		<Button
			onclick={() => {
				modals.fs.create.open = true;
			}}
			size="sm"
			class="h-6"
		>
			<div class="flex items-center">
				<Icon icon="gg:add" class="mr-1 h-4 w-4" />
				<span>New</span>
			</div>
		</Button>

		{@render button('create-snapshot')}
		{@render button('rollback-snapshot')}
		{@render button('delete-snapshot')}
		{@render button('edit-filesystem')}
		{@render button('delete-filesystem')}
		{@render button('bulk-delete')}
	</div>

	<TreeTable
		data={tableData}
		name={tableName}
		bind:parentActiveRow={activeRows}
		multipleSelect={true}
		bind:query
	/>
</div>

<!-- Create Snapshot -->
{#if modals.snapshot.create.open && activeDataset && activeDataset.type === 'filesystem'}
	<CreateSnapshot
		bind:open={modals.snapshot.create.open}
		dataset={activeDataset}
		recursion={true}
	/>
{/if}

<!-- Rollback to Snapshot -->
{#if modals.snapshot.rollback.open && activeDataset && activeDataset.type === 'snapshot'}
	<AlertDialogModal
		bind:open={modals.snapshot.rollback.open}
		customTitle={`Are you sure you want to rollback to the snapshot <b>${activeDataset.name}</b>? This action cannot be undone.`}
		actions={{
			onConfirm: async () => {
				if (activeDataset.properties.guid) {
					const response = await rollbackSnapshot(activeDataset.properties.guid);

					if (response.status === 'error') {
						handleAPIError(response);
						toast.success(`Rolled back to snapshot ${activeDataset.name}`, {
							position: 'bottom-center'
						});
					} else {
						toast.error(`Failed to rollback to snapshot ${activeDataset.name}`, {
							position: 'bottom-center'
						});
					}
				} else {
					toast.error('Snapshot GUID not found', {
						position: 'bottom-center'
					});
				}

				modals.snapshot.rollback.open = false;
			},
			onCancel: () => {
				modals.snapshot.rollback.open = false;
			}
		}}
	/>
{/if}

<!-- Delete Snapshot -->
{#if modals.snapshot.delete.open && activeDataset && activeDataset.type === 'snapshot'}
	<DeleteSnapshot bind:open={modals.snapshot.delete.open} dataset={activeDataset} />
{/if}

<!-- Delete FS -->
{#if modals.fs.delete.open && activeDataset && activeDataset.type === 'filesystem'}
	<AlertDialogModal
		bind:open={modals.fs.delete.open}
		names={{
			parent: 'filesystem',
			element: activeDataset.name
		}}
		actions={{
			onConfirm: async () => {
				if (activeDataset.properties.guid) {
					const response = await deleteFileSystem(activeDataset);

					if (response.status === 'success') {
						toast.success(`Deleted filesystem ${activeDataset.name}`, {
							position: 'bottom-center'
						});
					} else {
						handleAPIError(response);
						toast.error(`Failed to delete filesystem ${activeDataset.name}`, {
							position: 'bottom-center'
						});
					}
				} else {
					toast.error('Filesystem GUID not found', {
						position: 'bottom-center'
					});
				}

				modals.fs.delete.open = false;
			},
			onCancel: () => {
				modals.fs.delete.open = false;
			}
		}}
	/>
{/if}

<!-- Bulk delete -->
{#if modals.bulk.delete.open && activeDatasets.length > 0}
	<AlertDialogModal
		bind:open={modals.bulk.delete.open}
		customTitle={`Are you sure you want to delete ${modals.bulk.delete.title}? This action cannot be undone.`}
		actions={{
			onConfirm: async () => {
				const response = await bulkDelete(activeDatasets);
				if (response.status === 'success') {
					toast.success(`Deleted ${activeDatasets.length} datasets`, {
						position: 'bottom-center'
					});
				} else {
					handleAPIError(response);
					toast.error('Failed to delete datasets', {
						position: 'bottom-center'
					});
				}

				modals.bulk.delete.open = false;
			},
			onCancel: () => {
				modals.bulk.delete.open = false;
			}
		}}
	/>
{/if}

<!-- Create FS -->
{#if modals.fs.create.open}
	<CreateFS bind:open={modals.fs.create.open} {datasets} {grouped} />
{/if}

<!-- Edit FS -->
{#if modals.fs.edit.open && activeDataset && activeDataset.type === 'filesystem'}
	<EditFS bind:open={modals.fs.edit.open} dataset={activeDataset} />
{/if}
