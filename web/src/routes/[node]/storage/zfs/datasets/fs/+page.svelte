<script lang="ts">
	import { createSnapshot, deleteSnapshot, getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Dataset, type GroupedByPool } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { generateTableData, groupByPool } from '$lib/utils/zfs/dataset';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { tick } from 'svelte';

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

	let grouped = $derived(groupByPool($results[0].data, $results[1].data));
	let tableData = $derived(generateTableData(grouped));
	let activeRow: Row | null = $state(null);
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

	$effect(() => {
		console.log(activeDataset);
	});

	let confirmModals = $state({
		active: '' as 'deleteSnapshot' | 'createSnapshot' | 'createFilesystem',
		parent: 'filesystem' as 'filesystem' | 'snapshot',
		deleteSnapshot: {
			open: false,
			data: '',
			title: ''
		},
		createSnapshot: {
			open: false,
			data: {
				name: '',
				recursive: false
			},
			title: ''
		},
		createFilesystem: {
			open: false,
			data: {
				name: '',
				properties: {
					parent: '',
					compression: 'on'
				}
			},
			title: ''
		}
	});

	async function confirmAction() {
		if (confirmModals.active === 'deleteSnapshot') {
			if (activeDataset) {
				await deleteSnapshot(activeDataset);
				activeRow = null;
			}
		}

		if (confirmModals.active === 'createSnapshot') {
			if (activeDataset) {
				await createSnapshot(
					activeDataset,
					confirmModals.createSnapshot.data.name,
					confirmModals.createSnapshot.data.recursive
				);
				activeRow = null;
			}
		}

		confirmModals[confirmModals.active].open = false;

		if (confirmModals.active === 'createSnapshot') {
			confirmModals.createSnapshot.data.name = '';
			confirmModals.createSnapshot.data.recursive = false;
		}

		if (confirmModals.active === 'deleteSnapshot') {
			confirmModals.deleteSnapshot.data = '';
			confirmModals.deleteSnapshot.title = '';
		}
	}
</script>

{#snippet button(type: string)}
	{#if type === 'delete-snapshot' && activeDataset?.type === 'snapshot'}
		<Button
			on:click={async () => {
				if (activeDataset) {
					confirmModals.active = 'deleteSnapshot';
					confirmModals.parent = 'snapshot';
					confirmModals.deleteSnapshot.open = true;
					confirmModals.deleteSnapshot.title = activeDataset.name;
				}
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Snapshot
		</Button>
	{/if}

	{#if type === 'create-snapshot' && activeDataset?.type === 'filesystem'}
		<Button
			on:click={async () => {
				if (activeDataset) {
					confirmModals.active = 'createSnapshot';
					confirmModals.parent = 'filesystem';
					confirmModals.createSnapshot.open = true;
					confirmModals.createSnapshot.title = activeDataset.name;
				}
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-1 h-4 w-4" /> Create Snapshot
		</Button>
	{/if}

	{#if type === 'create-filesystem' && activeDataset?.type === 'filesystem'}
		<Button
			on:click={async () => {
				if (activeDataset) {
					confirmModals.active = 'createFilesystem';
					confirmModals.parent = 'filesystem';
					confirmModals.createFilesystem.open = true;
					confirmModals.createFilesystem.title = activeDataset.name;
				}
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="material-symbols:files" class="mr-1 h-4 w-4" /> Create Filesystem
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Button on:click={() => console.log('New dataset')} size="sm" class="h-6">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>

		{@render button('delete-snapshot')}
		{@render button('create-snapshot')}
		{@render button('create-filesystem')}
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

{#if confirmModals.active == 'deleteSnapshot'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: 'snapshot',
			element: confirmModals.active ? confirmModals[confirmModals.active].title || '' : ''
		}}
		actions={{
			onConfirm: () => {
				if (confirmModals.active) {
					confirmAction();
				}
			},
			onCancel: () => {
				if (confirmModals.active) {
					confirmModals[confirmModals.active].open = false;
				}
			}
		}}
	></AlertDialogModal>
{/if}

{#if confirmModals.active === 'createSnapshot'}
	<AlertDialog.Root
		bind:open={confirmModals.createSnapshot.open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>
					<div class="flex items-center">
						<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-2 h-6 w-6" />
						Snapshot -
						{confirmModals.createSnapshot.data.name !== ''
							? `${confirmModals.createSnapshot.title}@${confirmModals.createSnapshot.data.name}`
							: `${confirmModals.createSnapshot.title}`}
					</div>
				</AlertDialog.Title>
			</AlertDialog.Header>

			<div class="flex-1 space-y-1">
				<Label for="name">Name</Label>
				<Input
					type="text"
					id="name"
					placeholder="before-upgrade"
					autocomplete="off"
					bind:value={confirmModals.createSnapshot.data.name}
				/>
			</div>

			<div class="flex items-center gap-2">
				<Checkbox id="createRecursive" aria-labelledby="createRecursive-label" />
				<Label
					id="createRecursive-label"
					for="createRecursive"
					class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
					onclick={() => {
						confirmModals.createSnapshot.data.recursive =
							!confirmModals.createSnapshot.data.recursive;
					}}
				>
					Recursive
				</Label>
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel
					onclick={() => {
						confirmModals.createSnapshot.open = false;
					}}>Cancel</AlertDialog.Cancel
				>
				<AlertDialog.Action
					onclick={() => {
						confirmAction();
					}}
				>
					Create
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}

{#if confirmModals.active === 'createFilesystem'}
	<AlertDialog.Root
		bind:open={confirmModals.createFilesystem.open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>
					<div class="flex items-center">
						<Icon icon="material-symbols:files" class="mr-2 h-6 w-6" />
						Create Filesystem
					</div>
				</AlertDialog.Title>
			</AlertDialog.Header>

			<div class="w-full">
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
					<div class="flex-1 space-y-1">
						<Label for="name">Name</Label>
						<Input
							type="text"
							id="name"
							placeholder="before-upgrade"
							autocomplete="off"
							bind:value={confirmModals.createFilesystem.data.name}
						/>
					</div>

					<div class="space-y-1 py-1">
						<div>
							<Label class="w-24 whitespace-nowrap text-sm">Parent</Label>
							<Select.Root
								selected={{
									label:
										confirmModals.createFilesystem.data.properties.parent || activeDataset?.name,
									value:
										confirmModals.createFilesystem.data.properties.parent || activeDataset?.name
								}}
								onSelectedChange={(value) => {
									confirmModals.createFilesystem.data.properties.parent = value?.value || '';
								}}
							>
								<Select.Trigger class="w-full">
									<Select.Value placeholder="Select Parent" />
								</Select.Trigger>

								<Select.Content class="max-h-36 overflow-y-auto">
									<Select.Group>
										{#each grouped as pool}
											{#each pool.filesystems as fs}
												<Select.Item value={fs.name} label={fs.name}>{fs.name}</Select.Item>
											{/each}
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</div>
					</div>
				</div>
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel
					onclick={() => {
						confirmModals.createFilesystem.open = false;
					}}>Cancel</AlertDialog.Cancel
				>
				<AlertDialog.Action
					onclick={() => {
						confirmAction();
					}}
				>
					Create
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
