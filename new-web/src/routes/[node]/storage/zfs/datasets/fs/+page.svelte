<script lang="ts">
	import {
		bulkDelete,
		createFileSystem,
		createSnapshot,
		deleteFileSystem,
		deleteSnapshot,
		getDatasets,
		rollbackSnapshot
	} from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { isValidSize } from '$lib/utils/numbers';
	import { capitalizeFirstLetter, generatePassword } from '$lib/utils/string';
	import { isValidPoolName } from '$lib/utils/zfs';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { createFSProps, generateTableData, handleError } from '$lib/utils/zfs/dataset/fs';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import humanFormat, { type ParsedInfo, type ScaleLike } from 'human-format';
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
			queryKey: ['datasetList'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.datasets,
			onSuccess: (data: Dataset[]) => {
				updateCache('datasets', data);
			}
		}
	]);

	let grouped = $derived(groupByPool($results[0].data, $results[1].data));
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

	$effect(() => {
		if (activeDataset && confirmModals.active === 'createFilesystem') {
			confirmModals.createFilesystem.data.properties.parent = activeDataset.name;
		}
	});

	let confirmModals = $state({
		active: '' as
			| 'deleteSnapshot'
			| 'createSnapshot'
			| 'rollbackSnapshot'
			| 'createFilesystem'
			| 'deleteFilesystem'
			| 'bulkDeleteDatasets',
		parent: 'filesystem' as 'filesystem' | 'snapshot' | 'dataset',
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
		rollbackSnapshot: {
			open: false,
			data: {
				name: ''
			},
			title: ''
		},
		createFilesystem: {
			open: false,
			data: {
				name: '',
				properties: {
					parent: '',
					atime: 'on',
					checksum: 'on',
					compression: 'on',
					dedup: 'off',
					encryption: 'off',
					encryptionKey: '',
					quota: ''
				}
			},
			title: ''
		},
		deleteFilesystem: {
			open: false,
			data: '',
			title: ''
		},
		bulkDeleteDatasets: {
			open: false,
			data: '',
			title: ''
		}
	});

	async function closeCreateFsModal() {
		confirmModals.createFilesystem.open = false;
		confirmModals.createFilesystem.data = {
			name: '',
			properties: {
				parent: '',
				atime: 'on',
				checksum: 'on',
				compression: 'on',
				dedup: 'off',
				encryption: 'off',
				encryptionKey: '',
				quota: ''
			}
		};
		confirmModals.createFilesystem.title = '';
	}

	async function confirmAction() {
		if (confirmModals.active === 'deleteSnapshot') {
			if (activeDataset) {
				const response = await deleteSnapshot(activeDataset);

				if (response.error) {
					handleError(response);
					return;
				}

				activeRow = null;
			}
		}

		if (confirmModals.active === 'createSnapshot') {
			if (activeDataset) {
				const response = await createSnapshot(
					activeDataset,
					confirmModals.createSnapshot.data.name,
					confirmModals.createSnapshot.data.recursive
				);

				if (response.error) {
					handleError(response);
					return;
				}

				activeRow = null;
			}
		}

		if (confirmModals.active === 'rollbackSnapshot') {
			if (activeDataset) {
				await rollbackSnapshot(activeDataset.properties.guid || '');
				activeRow = null;
			}
		}

		if (confirmModals.active === 'createFilesystem') {
			if (!isValidPoolName(confirmModals.createFilesystem.data.name)) {
				toast.error('Invalid name', {
					position: 'bottom-center'
				});
				return;
			}

			if (!confirmModals.createFilesystem.data.properties.parent) {
				toast.error('No parent selected', {
					position: 'bottom-center'
				});
				return;
			}

			if (confirmModals.createFilesystem.data.properties.encryption !== 'off') {
				if (confirmModals.createFilesystem.data.properties.encryptionKey === '') {
					toast.error('Encryption key is required', {
						position: 'bottom-center'
					});
					return;
				}
			}

			if (confirmModals.createFilesystem.data.properties.quota !== '') {
				if (!isValidSize(confirmModals.createFilesystem.data.properties.quota)) {
					toast.error('Invalid quota size', {
						position: 'bottom-center'
					});
					return;
				}
			}

			const response = await createFileSystem(
				confirmModals.createFilesystem.data.name,
				confirmModals.createFilesystem.data.properties.parent,
				confirmModals.createFilesystem.data.properties
			);

			if (response.error) {
				handleError(response);
				return;
			}

			let n = `${confirmModals.createFilesystem.data.properties.parent}/${confirmModals.createFilesystem.data.name}`;

			toast.success(
				`${capitalizeFirstLetter(getTranslation('common.filesystem', 'filesystem'))} ${n} ${capitalizeFirstLetter(getTranslation('common.created', 'created'))}`,
				{
					position: 'bottom-center'
				}
			);

			activeRow = null;
		}

		if (confirmModals.active === 'deleteFilesystem') {
			if (activeDataset) {
				const response = await deleteFileSystem(activeDataset);

				if (response.error) {
					handleError(response);
					return;
				}

				activeRow = null;
			}
		}

		if (confirmModals.active === 'bulkDeleteDatasets') {
			if (activeDatasets.length > 0) {
				const response = await bulkDelete(activeDatasets);
				if (response.error) {
					handleError(response);
					return;
				}

				activeRows = null;
			}
		}

		confirmModals[confirmModals.active].open = false;

		if (confirmModals.active === 'createSnapshot') {
			confirmModals.createSnapshot.data.name = '';
			confirmModals.createSnapshot.data.recursive = false;
		}

		if (
			confirmModals.active === 'deleteSnapshot' ||
			confirmModals.active === 'deleteFilesystem' ||
			confirmModals.active === 'bulkDeleteDatasets'
		) {
			confirmModals[confirmModals.active].data = '';
			confirmModals[confirmModals.active].title = '';
		}

		if (confirmModals.active === 'rollbackSnapshot') {
			confirmModals.rollbackSnapshot.data.name = '';
			confirmModals.rollbackSnapshot.title = '';
		}

		if (confirmModals.active === 'createFilesystem') {
			confirmModals.createFilesystem.data.name = '';
			confirmModals.createFilesystem.data.properties = {
				parent: '',
				atime: 'on',
				checksum: 'on',
				compression: 'on',
				dedup: 'off',
				encryption: 'off',
				encryptionKey: '',
				quota: ''
			};
		}
	}

	let remainingSpace = $state(0);
	let currentPartition = $state(0);
	let currentPartitionInput = $derived(confirmModals.createFilesystem.data.properties.quota);

	$effect(() => {
		if (currentPartitionInput === '') {
			currentPartition = 0;
		} else {
			let parsed: ParsedInfo<ScaleLike> | null = null;

			try {
				parsed = humanFormat.parse.raw(currentPartitionInput);
			} catch (e) {
				parsed = { factor: 1, value: 0, prefix: 'B' };
				currentPartitionInput = '1B';
			}

			if (parsed) {
				untrack(() => {
					currentPartition = parsed.factor * parsed.value;
					if (currentPartition > remainingSpace) {
						currentPartition = remainingSpace;
						currentPartitionInput = humanFormat(remainingSpace);
					}
				});
			}
		}
	});

	let zfsProperties = $state(createFSProps);
	let query: string = $state('');
</script>

{#snippet button(type: string)}
	{#if activeRows && activeRows.length == 1}
		{#if type === 'rollback-snapshot' && activeDataset?.type === 'snapshot'}
			<Button
				onclick={async () => {
					if (activeDataset) {
						confirmModals.active = 'rollbackSnapshot';
						confirmModals.parent = 'snapshot';
						confirmModals.rollbackSnapshot.open = true;
						confirmModals.rollbackSnapshot.data.name = activeDataset.name;
						confirmModals.rollbackSnapshot.title = activeDataset.name;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="eos-icons:snapshot-rollback" class="mr-1 h-4 w-4" /> Rollback To Snapshot
			</Button>
		{/if}

		{#if type === 'delete-snapshot' && activeDataset?.type === 'snapshot'}
			<Button
				onclick={async () => {
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
				onclick={async () => {
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

		{#if type === 'delete-filesystem' && activeDataset?.type === 'filesystem' && activeDataset?.name.includes('/')}
			<Button
				onclick={async () => {
					if (activeDataset) {
						confirmModals.active = 'deleteFilesystem';
						confirmModals.parent = 'filesystem';
						confirmModals.deleteFilesystem.open = true;
						confirmModals.deleteFilesystem.title = activeDataset.name;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Filesystem
			</Button>
		{/if}
	{:else if activeRows && activeRows.length > 1}
		{#if activeDatasets.length > 0 && !poolsSelected}
			{#if type === 'bulk-delete'}
				<Button
					onclick={async () => {
						confirmModals.active = 'bulkDeleteDatasets';
						confirmModals.parent = 'dataset';
						confirmModals.bulkDeleteDatasets.open = true;

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
							title = `${snapLen} snapshots and ${fsLen} filesystems`;
						} else if (snapLen > 0) {
							title = `${snapLen} snapshots`;
						} else if (fsLen > 0) {
							title = `${fsLen} filesystems`;
						}

						confirmModals.bulkDeleteDatasets.title = title;
					}}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
				>
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Datasets
				</Button>
			{/if}
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
		<Button
			onclick={() => {
				confirmModals.active = 'createFilesystem';
				confirmModals.parent = 'filesystem';
				confirmModals.createFilesystem.open = true;
				confirmModals.createFilesystem.title = '';
			}}
			size="sm"
			class="h-6"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>

		{@render button('create-snapshot')}
		{@render button('rollback-snapshot')}
		{@render button('delete-snapshot')}
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

{#if confirmModals.active == 'deleteSnapshot' || confirmModals.active == 'deleteFilesystem'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: confirmModals.parent,
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

{#if confirmModals.active === 'bulkDeleteDatasets'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: '',
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

{#if confirmModals.active === 'rollbackSnapshot'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		customTitle={`Are you sure you would like to rollback to <b>${confirmModals[confirmModals.active].data.name}</b>?`}
		names={{
			parent: confirmModals.parent,
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
	<Dialog.Root bind:open={confirmModals.createSnapshot.open}>
		<Dialog.Content class="p-5">
			<div class="flex items-center justify-between">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						<div class="flex items-center gap-2">
							<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class=" h-6 w-6" />
							Snapshot -
							{confirmModals.createSnapshot.data.name !== ''
								? `${confirmModals.createSnapshot.title}@${confirmModals.createSnapshot.data.name}`
								: `${confirmModals.createSnapshot.title}`}
						</div>
					</Dialog.Title>
				</Dialog.Header>
				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
						onclick={() => {
							confirmModals.createSnapshot.data.name = '';
							confirmModals.createSnapshot.data.recursive = false;
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}</span
						>
					</Button>
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.close', 'Close'))}
						onclick={() => {
							console.log('close');

							confirmModals.createSnapshot.open = false;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

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
				<Checkbox
					id="createRecursive"
					bind:checked={confirmModals.createSnapshot.data.recursive}
					aria-labelledby="createRecursive-label"
				/>
				<Label
					id="createRecursive-label"
					for="createRecursive"
					class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
				>
					Recursive
				</Label>
			</div>

			<Dialog.Footer>
				<Button
					size="sm"
					onclick={() => {
						confirmAction();
					}}>Create</Button
				>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}

{#if confirmModals.active === 'createFilesystem'}
	<Dialog.Root bind:open={confirmModals.createFilesystem.open}>
		<Dialog.Content
			class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-5 transition-all duration-300 ease-in-out lg:max-w-[70%]"
		>
			<div class="flex items-center justify-between pb-3">
				<Dialog.Header class="flex justify-between">
					<Dialog.Title class="flex items-center text-left">
						<Icon icon="material-symbols:files" class="mr-2 h-5 w-5" />Create Filesystem</Dialog.Title
					>
				</Dialog.Header>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
						onclick={() => {
							confirmModals.createFilesystem.data.name = '';
							confirmModals.createFilesystem.data.properties = {
								parent: '',
								atime: 'on',
								checksum: 'on',
								compression: 'on',
								dedup: 'off',
								encryption: 'off',
								encryptionKey: '',
								quota: ''
							};
							confirmModals.createFilesystem.title = '';
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}</span
						>
					</Button>
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.close', 'Close'))}
						onclick={() => {
							closeCreateFsModal();
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

			<div class="w-full">
				<div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
					<div class="space-y-1">
						<Label for="name" class="w-24 whitespace-nowrap text-sm">Name</Label>
						<Input
							type="text"
							id="name"
							placeholder="before-upgrade"
							autocomplete="off"
							bind:value={confirmModals.createFilesystem.data.name}
						/>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Parent</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.parent}
						>
							<Select.Trigger class="w-full">
								{confirmModals.createFilesystem.data.properties.parent ||
									(activeDataset ? activeDataset.name : 'Select Parent')}
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

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">ATime</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.atime}
						>
							<Select.Trigger class="w-full">
								{confirmModals.createFilesystem.data.properties.atime === 'on'
									? 'Yes'
									: confirmModals.createFilesystem.data.properties.atime === 'off'
										? 'No'
										: 'Select ATime'}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each zfsProperties.atime as option}
										<Select.Item value={option.value} label={option.label}
											>{option.label}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Checkum</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.checksum}
						>
							<Select.Trigger class="w-full">
								{confirmModals.createFilesystem.data.properties.checksum}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each zfsProperties.checksum as option}
										<Select.Item value={option.value} label={option.label}
											>{option.label}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Compression</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.compression}
						>
							<Select.Trigger class="w-full">
								{confirmModals.createFilesystem.data.properties.compression}
							</Select.Trigger>

							<Select.Content>
								<Select.Group>
									{#each zfsProperties.compression as option}
										<Select.Item value={option.value} label={option.label}
											>{option.label}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Deduplication</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.dedup}
						>
							<Select.Trigger class="w-full">
								{#if confirmModals.createFilesystem.data.properties.dedup === 'off'}
									No
								{:else if confirmModals.createFilesystem.data.properties.dedup === 'on'}
									Yes
								{:else}
									{confirmModals.createFilesystem.data.properties.dedup}
								{/if}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each zfsProperties.dedup as option}
										<Select.Item value={option.value} label={option.label}
											>{option.label}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Encryption</Label>
						<Select.Root
							type="single"
							bind:value={confirmModals.createFilesystem.data.properties.encryption}
						>
							<Select.Trigger class="w-full">
								{confirmModals.createFilesystem.data.properties.encryption === 'off'
									? 'No'
									: confirmModals.createFilesystem.data.properties.encryption === 'on'
										? 'Yes'
										: confirmModals.createFilesystem.data.properties.encryption}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each zfsProperties.encryption as option}
										<Select.Item value={option.value} label={option.label}
											>{option.label}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					{#if confirmModals.createFilesystem.data.properties.encryption !== 'off'}
						<div class="space-y-1">
							<Label class="w-24 whitespace-nowrap text-sm">Passphrase</Label>
							<div class="flex w-full max-w-sm items-center space-x-2">
								<Input
									type="password"
									id="d-passphrase"
									placeholder="Enter or generate passphrase"
									class="w-full"
									autocomplete="off"
									bind:value={confirmModals.createFilesystem.data.properties.encryptionKey}
									showPasswordOnFocus={true}
								/>

								<Button
									onclick={() => {
										confirmModals.createFilesystem.data.properties.encryptionKey =
											generatePassword();
									}}
								>
									<Icon
										icon="fad:random-2dice"
										class="h-6 w-6"
										onclick={() => {
											confirmModals.createFilesystem.data.properties.encryptionKey =
												generatePassword();
										}}
									/>
								</Button>
							</div>
						</div>
					{/if}

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Quota</Label>
						<Input
							type="text"
							class="w-full text-left"
							min="0"
							max={remainingSpace}
							bind:value={confirmModals.createFilesystem.data.properties.quota}
							placeholder="256M (Empty for no quota)"
						/>
					</div>
				</div>
			</div>

			<Dialog.Footer>
				<div class="flex items-center justify-end space-x-4 p-4">
					<Button
						size="sm"
						type="button"
						class="h-8 w-full "
						onclick={() => {
							confirmAction();
						}}
					>
						Create
					</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}
