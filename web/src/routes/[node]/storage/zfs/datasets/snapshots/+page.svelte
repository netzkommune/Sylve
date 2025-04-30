<script lang="ts">
	import {
		createPeriodicSnapshot,
		createSnapshot,
		deleteSnapshot,
		getDatasets,
		getPeriodicSnapshots
	} from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import * as Command from '$lib/components/ui/command/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import ViewSnapshotJobs from '$lib/components/zfs/ViewSnapshotJobs.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Dataset, GroupedByPool, PeriodicSnapshot } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { cn } from '$lib/utils';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { handleError } from '$lib/utils/zfs/dataset/fs';
	import { generateTableData } from '$lib/utils/zfs/dataset/snapshot';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { tick } from 'svelte';
	import toast from 'svelte-french-toast';

	interface Data {
		pools: Zpool[];
		periodicSnapshots: PeriodicSnapshot[];
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
		},
		{
			queryKey: ['periodicSnapshots'],
			queryFn: async () => {
				return await getPeriodicSnapshots();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.periodicSnapshots
		}
	]);

	let activeRow: Row | null = $state(null);
	let grouped: GroupedByPool[] = $derived(groupByPool($results[0].data, $results[1].data));
	let pools: Zpool[] = $derived($results[0].data || []);
	let periodicSnapshots: PeriodicSnapshot[] = $derived($results[2].data || []);
	let datasets: Dataset[] = $derived.by(() => {
		return $results[1].data?.filter((dataset) => dataset.type !== 'snapshot') || [];
	});
	let tableData = $derived(generateTableData(grouped));
	let activePool: Zpool | null = $derived.by(() => {
		if (activeRow) {
			const poolGroup = grouped.find((pool) => pool.name === activeRow?.name);
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

	let activePeriodics: PeriodicSnapshot[] = $derived.by(() => {
		if (activePool) {
			for (const group of grouped) {
				if (group.name === activePool.name) {
					const fs = group.filesystems;
					const volumes = group.volumes;
					const guids = fs
						.map((fs) => ({
							guid: fs.properties.guid
						}))
						.concat(
							volumes.map((volume) => ({
								guid: volume.properties.guid
							}))
						);

					return periodicSnapshots.filter((snapshot) => {
						return guids.some((fs) => fs.guid === snapshot.guid);
					});
				}
			}
		}

		return [];
	});

	let query = $state('');
	let confirmModals = $state({
		active: '' as 'createSnapshot' | 'deleteSnapshot' | 'viewSnapshotJobs',
		parent: 'filesystem' as 'filesystem' | 'snapshot',
		deleteSnapshot: {
			open: false,
			recursive: false,
			data: '',
			title: ''
		},
		createSnapshot: {
			open: false,
			recursive: false,
			interval: 0,
			name: '',
			title: '',
			extraTitle: ''
		},
		viewSnapshotJobs: {
			open: false
		}
	});

	let comboBoxes = $state({
		pool: {
			open: false,
			value: '',
			data: pools.map((pool) => ({
				value: pool.name,
				label: pool.name
			}))
		},
		datasets: {
			open: false,
			value: '',
			data: [] as { value: string; label: string }[]
		},
		interval: {
			open: false,
			value: '0',
			data: [
				{ value: '0', label: 'None' },
				{ value: '60', label: 'Every Minute' },
				{ value: '3600', label: 'Every Hour' },
				{ value: '86400', label: 'Every Day' },
				{ value: '604800', label: 'Every Week' },
				{ value: '2419200', label: 'Every Month' },
				{ value: '29030400', label: 'Every Year' }
			]
		}
	});

	$effect(() => {
		const currentSelectedPool = comboBoxes.pool.value;
		if (currentSelectedPool) {
			comboBoxes.datasets.data = datasets
				.filter((dataset) => dataset.name.startsWith(currentSelectedPool))
				.map((dataset) => ({
					value: dataset.name,
					label: dataset.name
				}));
		} else {
			comboBoxes.datasets.data = [];
		}
	});

	$effect(() => {
		if (confirmModals.active === 'createSnapshot' && confirmModals.createSnapshot.open) {
			if (comboBoxes.pool.value && !comboBoxes.datasets.value) {
				confirmModals.createSnapshot.extraTitle = ` - ${comboBoxes.pool.value}`;
			} else if (comboBoxes.pool.value && comboBoxes.datasets.value) {
				confirmModals.createSnapshot.extraTitle = ` - ${comboBoxes.datasets.value}`;
			} else {
				confirmModals.createSnapshot.extraTitle = '';
			}
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

				activeDataset = null;
				activeRow = null;
			}
		}

		if (confirmModals.active === 'createSnapshot') {
			if (comboBoxes.datasets.value) {
				const dataset = datasets.find((dataset) => dataset.name === comboBoxes.datasets.value);
				if (dataset) {
					const interval = parseInt(comboBoxes.interval.value) || 0;

					if (interval === 0) {
						const response = await createSnapshot(
							dataset,
							confirmModals.createSnapshot.name,
							confirmModals.createSnapshot.recursive
						);

						if (response.error) {
							handleError(response);
							return;
						}
					} else if (interval > 0) {
						const response = await createPeriodicSnapshot(
							dataset,
							confirmModals.createSnapshot.name,
							confirmModals.createSnapshot.recursive,
							interval
						);

						console.log(response);
					}
				}
			} else {
				toast.error('Please select a dataset', {
					position: 'bottom-center'
				});
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
				confirmModals.deleteSnapshot.data = activeDataset?.name || '';
				confirmModals.deleteSnapshot.title = activeDataset?.name || '';
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Snapshot
		</Button>
	{/if}

	{#if type === 'view-periodics' && activePeriodics.length > 0}
		<Button
			on:click={async () => {
				confirmModals.active = 'viewSnapshotJobs';
				confirmModals.parent = 'snapshot';
				// confirmModals.deleteSnapshot.open = true;
				// confirmModals.deleteSnapshot.data = activeDataset?.name || '';
				// confirmModals.deleteSnapshot.title = activeDataset?.name || '';
				confirmModals.viewSnapshotJobs.open = true;
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="material-symbols:save-clock" class="mr-1 h-4 w-4" /> View Snapshot Jobs
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />

		<Button
			on:click={() => {
				confirmModals.active = 'createSnapshot';
				confirmModals.parent = 'snapshot';
				confirmModals.createSnapshot.open = true;
				confirmModals.createSnapshot.recursive = false;
				confirmModals.createSnapshot.interval = 0;
				confirmModals.createSnapshot.name = '';
				confirmModals.createSnapshot.title = '';
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

		{@render button('delete-snapshot')}
		{@render button('view-periodics')}
	</div>

	<TreeTable data={tableData} name={tableName} bind:parentActiveRow={activeRow} bind:query />
</div>

{#if confirmModals.active === 'createSnapshot'}
	<AlertDialog.Root
		bind:open={confirmModals[confirmModals.active].open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>
					<div class="flex items-center">
						<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-2 h-6 w-6" />
						Snapshot {confirmModals.createSnapshot.extraTitle}
					</div>
				</AlertDialog.Title>
			</AlertDialog.Header>

			<CustomValueInput
				label={capitalizeFirstLetter(getTranslation('common.name', 'Name')) +
					' | ' +
					capitalizeFirstLetter(getTranslation('common.prefix', 'Prefix'))}
				placeholder="after-upgrade"
				bind:value={confirmModals.createSnapshot.name}
				classes="flex-1 space-y-1"
			/>

			<div class="flex gap-4">
				<CustomComboBox
					bind:open={comboBoxes.pool.open}
					label="Pool"
					bind:value={comboBoxes.pool.value}
					data={comboBoxes.pool.data}
					classes="flex-1 space-y-1"
					placeholder="Select a pool"
				></CustomComboBox>

				<CustomComboBox
					bind:open={comboBoxes.datasets.open}
					label="Dataset"
					bind:value={comboBoxes.datasets.value}
					data={comboBoxes.datasets.data}
					classes="flex-1 space-y-1"
					placeholder="Select a dataset"
				></CustomComboBox>
			</div>

			<div class="flex-1 space-y-1">
				<CustomComboBox
					bind:open={comboBoxes.interval.open}
					label="Interval"
					bind:value={comboBoxes.interval.value}
					data={comboBoxes.interval.data}
					classes="flex-1 space-y-1"
					placeholder="Select an interval"
				></CustomComboBox>
			</div>

			<CustomCheckbox
				label="Recursive"
				bind:checked={confirmModals.createSnapshot.recursive}
				classes="flex items-center gap-2"
			></CustomCheckbox>

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
					}}
				>
					Delete
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}

<ViewSnapshotJobs
	bind:open={confirmModals.viewSnapshotJobs.open}
	{pools}
	{datasets}
	periodicSnapshots={activePeriodics}
></ViewSnapshotJobs>
