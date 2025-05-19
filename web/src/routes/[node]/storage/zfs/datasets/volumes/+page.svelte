<script lang="ts">
	import {
		bulkDelete,
		createSnapshot,
		createVolume,
		deleteSnapshot,
		deleteVolume,
		getDatasets
	} from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import Input from '$lib/components/ui/input/input.svelte';
	import Label from '$lib/components/ui/label/label.svelte';
	import * as Select from '$lib/components/ui/select/index.js';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import type { Dataset, GroupedByPool } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { getTranslation } from '$lib/utils/i18n';
	import { isValidSize } from '$lib/utils/numbers';
	import { capitalizeFirstLetter, generatePassword } from '$lib/utils/string';
	import { isValidPoolName } from '$lib/utils/zfs';
	import { groupByPool } from '$lib/utils/zfs/dataset/dataset';
	import { createVolProps, generateTableData, handleError } from '$lib/utils/zfs/dataset/volume';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import humanFormat from 'human-format';
	import toast from 'svelte-french-toast';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
	}

	let { data }: { data: Data } = $props();
	let tableName = 'tt-zfsVolumes';

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

	let grouped: GroupedByPool[] = $derived(groupByPool($results[0].data, $results[1].data));
	let table: {
		rows: Row[];
		columns: Column[];
	} = $derived(generateTableData(grouped));
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));

	let activePool: Zpool | null = $derived.by(() => {
		const pool = $results[0].data?.find((pool) => pool.name === activeRow?.name);
		return pool ?? null;
	});

	let activeVolume: Dataset | null = $derived.by(() => {
		if (activePool) return null;
		const volumes = $results[1].data?.filter((volume) => volume.type === 'volume');
		const volume = volumes?.find((volume) => volume.name.endsWith(activeRow?.name));
		return volume ?? null;
	});

	let isPoolSelected: boolean = $derived.by(() => {
		if (activeRows && activeRows.length > 0) {
			for (const row of activeRows) {
				if (row.guid === undefined) {
					return true;
				}
			}
		}

		return false;
	});

	let activeVolumes: Dataset[] = $derived.by(() => {
		if (activeRows && activeRows.length > 0) {
			const volumes = $results[1].data?.filter((volume) => volume.type === 'volume');
			return (
				volumes?.filter((volume) => activeRows?.some((row) => row.name.endsWith(volume.name))) ?? []
			);
		}

		return [];
	});

	let activeSnapshot: Dataset | null = $derived.by(() => {
		if (activePool) return null;
		const snapshots = $results[1].data?.filter((snapshot) => snapshot.type === 'snapshot');
		const snapshot = snapshots?.find((snapshot) => snapshot.name.endsWith(activeRow?.name));
		return snapshot ?? null;
	});

	type props = {
		checksum: string;
		compression: string;
		dedup: string;
		encryption: string;
		volblocksize: string;
	};

	let confirmModals = $state({
		active: '' as
			| 'createVolume'
			| 'deleteVolume'
			| 'deleteSnapshot'
			| 'createSnapshot'
			| 'deleteVolumes',
		parent: '',
		createVolume: {
			open: false,
			data: {
				name: '',
				properties: {
					parent: '',
					checksum: 'on',
					compression: 'on',
					dedup: 'off',
					encryption: 'off',
					encryptionKey: '',
					volblocksize: '16384',
					size: ''
				}
			},
			title: ''
		},
		deleteVolume: {
			open: false,
			data: '',
			title: ''
		},
		deleteSnapshot: {
			open: false,
			data: '',
			title: ''
		},
		createSnapshot: {
			open: false,
			data: {
				name: ''
			},
			title: ''
		},
		deleteVolumes: {
			open: false,
			data: '',
			title: ''
		}
	});

	let zfsProperties = $state(createVolProps);

	async function closeCreateVolumeModal() {
		confirmModals.createVolume.open = false;
		confirmModals.createVolume.data = {
			name: '',
			properties: {
				parent: '',
				checksum: 'on',
				compression: 'on',
				dedup: 'off',
				encryption: 'off',
				encryptionKey: '',
				volblocksize: '16384',
				size: ''
			}
		};
		confirmModals.createVolume.title = '';
	}

	async function confirmAction() {
		if (confirmModals.active === 'createVolume') {
			if (!isValidPoolName(confirmModals.createVolume.data.name)) {
				toast.error(
					capitalizeFirstLetter(
						getTranslation('zfs.datasets.invalid_volume_name', 'invalid volume name')
					),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			if (!confirmModals.createVolume.data.properties.parent) {
				toast.error(
					capitalizeFirstLetter(
						getTranslation('zfs.datasets.no_parent_selected', 'No parent selected')
					),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			if (confirmModals.createVolume.data.properties.encryption !== 'off') {
				if (confirmModals.createVolume.data.properties.encryptionKey === '') {
					toast.error(
						capitalizeFirstLetter(
							getTranslation('zfs.datasets.encryption_key_required', 'Encryption key is required')
						),
						{
							position: 'bottom-center'
						}
					);
					return;
				}
			}

			if (!isValidSize(confirmModals.createVolume.data.properties.size)) {
				toast.error(
					capitalizeFirstLetter(
						getTranslation('zfs.datasets.invalid_volume_size', 'Invalid volume size')
					),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			const parentSize = grouped.find(
				(group) => group.pool.name === confirmModals.createVolume.data.properties.parent
			)?.pool.free;

			if (!parentSize) {
				toast.error(
					capitalizeFirstLetter(
						getTranslation('zfs.datasets.parent_not_found', 'Parent not found')
					),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			if (humanFormat.parse(confirmModals.createVolume.data.properties.size) > parentSize) {
				toast.error(
					capitalizeFirstLetter(
						getTranslation(
							'zfs.datasets.vol_size_greater_than_available_space',
							'volume size is greater than available space'
						)
					),
					{
						position: 'bottom-center'
					}
				);
			}

			const response = await createVolume(
				confirmModals.createVolume.data.name,
				confirmModals.createVolume.data.properties.parent,
				confirmModals.createVolume.data.properties
			);

			if (response.error) {
				handleError(response);
				return;
			}

			let n = `${confirmModals.createVolume.data.properties.parent}/${confirmModals.createVolume.data.name}`;

			toast.success(
				`${capitalizeFirstLetter(getTranslation('common.volume', 'volume'))} ${n} ${capitalizeFirstLetter(getTranslation('common.created', 'created'))}`,
				{
					position: 'bottom-center'
				}
			);

			confirmModals.createVolume.open = false;
			confirmModals.createVolume.data.name = '';
			confirmModals.createVolume.data.properties.parent = '';
			confirmModals.createVolume.data.properties.size = '';
			confirmModals.createVolume.data.properties.encryptionKey = '';
			confirmModals.createVolume.data.properties.encryption = 'off';
			confirmModals.createVolume.data.properties.dedup = 'off';
			confirmModals.createVolume.data.properties.compression = 'on';
			confirmModals.createVolume.data.properties.checksum = 'on';
			confirmModals.createVolume.data.properties.volblocksize = '16384';
		}

		if (confirmModals.active === 'deleteVolume') {
			if (activeVolume) {
				const response = await deleteVolume(activeVolume);
				if (response.error) {
					handleError(response);
					return;
				}

				toast.success(
					`${capitalizeFirstLetter(getTranslation('common.volume', 'volume'))} ${activeVolume.name} ${capitalizeFirstLetter(getTranslation('common.deleted', 'deleted'))}`,
					{
						position: 'bottom-center'
					}
				);
			}
		}

		if (confirmModals.active === 'createSnapshot') {
			if (activeVolume) {
				const response = await createSnapshot(
					activeVolume,
					confirmModals.createSnapshot.data.name,
					false
				);

				if (response.error) {
					handleError(response);
					return;
				}

				activeRow = null;
			}
		}

		if (confirmModals.active === 'deleteSnapshot') {
			if (activeSnapshot) {
				const response = await deleteSnapshot(activeSnapshot);
				if (response.error) {
					handleError(response);
					return;
				}

				toast.success(
					`${capitalizeFirstLetter(getTranslation('common.snapshot', 'snapshot'))} ${activeSnapshot.name} ${capitalizeFirstLetter(getTranslation('common.deleted', 'deleted'))}`,
					{
						position: 'bottom-center'
					}
				);
			}
		}

		if (confirmModals.active === 'deleteVolumes') {
			if (activeVolumes.length > 0) {
				const response = await bulkDelete(activeVolumes);
				if (response.error) {
					handleError(response);
					return;
				}

				toast.success(
					`${capitalizeFirstLetter(getTranslation('common.volumes', 'volumes'))} ${activeVolumes
						.map((volume) => volume.name)
						.join(', ')} ${capitalizeFirstLetter(getTranslation('common.deleted', 'deleted'))}`,
					{
						position: 'bottom-center'
					}
				);
			}
		}
	}

	let query: string = $state('');
</script>

{#snippet button(type: string)}
	{#if activeRows && activeRows.length == 1}
		{#if type === 'create-snapshot' && activeVolume?.type === 'volume'}
			<Button
				on:click={async () => {
					if (activeVolume) {
						confirmModals.active = 'createSnapshot';
						confirmModals.parent = 'volume';
						confirmModals.createSnapshot.open = true;
						confirmModals.createSnapshot.title = activeVolume.name;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-1 h-4 w-4" /> Create Snapshot
			</Button>
		{/if}

		{#if type === 'delete-snapshot' && activeSnapshot?.type === 'snapshot'}
			<Button
				on:click={async () => {
					if (activeSnapshot) {
						confirmModals.active = 'deleteSnapshot';
						confirmModals.parent = 'snapshot';
						confirmModals.deleteSnapshot.open = true;
						confirmModals.deleteSnapshot.title = activeSnapshot.name;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Snapshot
			</Button>
		{/if}

		{#if type === 'delete-volume' && activeVolume?.type === 'volume'}
			<Button
				on:click={async () => {
					if (activeRow) {
						confirmModals.active = 'deleteVolume';
						confirmModals.parent = 'volume';
						confirmModals.deleteVolume.open = true;
						confirmModals.deleteVolume.data = activeRow.name;
						confirmModals.deleteVolume.title = activeRow.name;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Volume
			</Button>
		{/if}
	{:else if activeRows && activeRows.length > 1}
		{#if activeVolumes.length > 0 && type === 'delete-volumes' && !isPoolSelected}
			<Button
				on:click={async () => {
					if (activeRow) {
						confirmModals.active = 'deleteVolumes';
						confirmModals.parent = 'volume';
						confirmModals.deleteVolumes.open = true;
						confirmModals.deleteVolumes.title = `${activeVolumes.length} volumes`;
					}
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" /> Delete Volumes
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
		<Button
			on:click={() => {
				confirmModals.active = 'createVolume';
				confirmModals.createVolume.open = true;
				confirmModals.createVolume.title = '';
			}}
			size="sm"
			class="h-6"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>

		{@render button('create-snapshot')}
		{@render button('delete-snapshot')}
		{@render button('delete-volume')}
		{@render button('delete-volumes')}
	</div>

	<TreeTable
		data={table}
		name={tableName}
		bind:parentActiveRow={activeRows}
		bind:query
		multipleSelect={true}
	/>
</div>

{#snippet simpleSlect(prop: keyof props, label: string, placeholder: string)}
	<div class="space-y-1">
		<Label class="w-24 whitespace-nowrap text-sm">{label}</Label>
		<Select.Root
			selected={{
				label:
					zfsProperties[prop].find(
						(option) => option.value === confirmModals.createVolume.data.properties[prop]
					)?.label || confirmModals.createVolume.data.properties[prop],
				value: confirmModals.createVolume.data.properties[prop]
			}}
			onSelectedChange={(value) => {
				confirmModals.createVolume.data.properties[prop] = value?.value || '';
			}}
		>
			<Select.Trigger class="w-full">
				<Select.Value {placeholder} />
			</Select.Trigger>

			<Select.Content class="max-h-36 overflow-y-auto">
				<Select.Group>
					{#each zfsProperties[prop] as option}
						<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Group>
			</Select.Content>
		</Select.Root>
	</div>
{/snippet}

{#if confirmModals.active === 'createVolume'}
	<Dialog.Root
		bind:open={confirmModals.createVolume.open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<Dialog.Content
			class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
		>
			<div class="flex items-center justify-between">
				<Dialog.Header class="flex justify-between p-4">
					<Dialog.Title class="flex items-center text-left">
						<Icon icon="carbon:volume-block-storage" class="mr-2 h-5 w-5" />Create Volume</Dialog.Title
					>
				</Dialog.Header>
				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
						onclick={() => {
							confirmModals.createVolume.data = {
								name: '',
								properties: {
									parent: '',
									checksum: 'on',
									compression: 'on',
									dedup: 'off',
									encryption: 'off',
									encryptionKey: '',
									volblocksize: '16384',
									size: ''
								}
							};
							confirmModals.createVolume.title = '';
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
						onclick={() => closeCreateVolumeModal()}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

			<div class="w-full p-4">
				<div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
					<div class="space-y-1">
						<Label for="name">Name</Label>
						<Input
							type="text"
							id="name"
							placeholder="volume"
							autocomplete="off"
							bind:value={confirmModals.createVolume.data.name}
						/>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Size</Label>
						<Input
							type="text"
							class="w-full text-left"
							min="0"
							bind:value={confirmModals.createVolume.data.properties.size}
							placeholder="128M"
						/>
					</div>

					<div class="space-y-1">
						<Label class="w-24 whitespace-nowrap text-sm">Parent</Label>
						<Select.Root
							selected={{
								label: confirmModals.createVolume.data.properties.parent || activePool?.name,
								value: confirmModals.createVolume.data.properties.parent || activePool?.name
							}}
							onSelectedChange={(value) => {
								confirmModals.createVolume.data.properties.parent = value?.value || '';
							}}
						>
							<Select.Trigger class="w-full">
								<Select.Value placeholder="Select Parent" />
							</Select.Trigger>

							<Select.Content class="max-h-36 overflow-y-auto">
								<Select.Group>
									{#each grouped as group}
										<Select.Item value={group.pool.name} label={group.pool.name}
											>{group.pool.name}</Select.Item
										>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>

					{@render simpleSlect('volblocksize', 'Block Size', 'Select block size')}
					{@render simpleSlect('checksum', 'Checksum', 'Select checksum algorithm')}
					{@render simpleSlect('compression', 'Compression', 'Select compression type')}
					{@render simpleSlect('dedup', 'Deduplication', 'Select deduplication mode')}
					{@render simpleSlect('encryption', 'Encryption', 'Select encryption')}

					{#if confirmModals.createVolume.data.properties.encryption !== 'off'}
						<div class="space-y-1">
							<Label class="w-24 whitespace-nowrap text-sm">Passphrase</Label>
							<div class="flex w-full max-w-sm items-center space-x-2">
								<Input
									type="password"
									id="d-passphrase"
									placeholder="Enter or generate passphrase"
									class="w-full"
									autocomplete="off"
									bind:value={confirmModals.createVolume.data.properties.encryptionKey}
									showPasswordOnFocus={true}
								/>

								<Button
									onclick={() => {
										confirmModals.createVolume.data.properties.encryptionKey = generatePassword();
									}}
								>
									<Icon
										icon="fad:random-2dice"
										class="h-6 w-6"
										onclick={() => {
											confirmModals.createVolume.data.properties.encryptionKey = generatePassword();
										}}
									/>
								</Button>
							</div>
						</div>
					{/if}
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

{#if confirmModals.active == 'deleteVolume'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: 'volume',
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

{#if confirmModals.active == 'deleteVolumes'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: '',
			element: confirmModals[confirmModals.active].title
		}}
		actions={{
			onConfirm: () => {
				if (confirmModals.active) {
					confirmAction();
				}
			},
			onCancel: () => {
				if (confirmModals.active) {
					console.log(confirmModals[confirmModals.active]);
					confirmModals[confirmModals.active].open = false;
				}
			}
		}}
	></AlertDialogModal>
{/if}

{#if confirmModals.active === 'createSnapshot'}
	<Dialog.Root
		bind:open={confirmModals.createSnapshot.open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<Dialog.Content class="p-5">
			<div class="flex items-center justify-between">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						<div class="flex items-center">
							<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-2 h-6 w-6" />
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
							confirmModals.createSnapshot.title = '';
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
							confirmModals.createSnapshot = {
								open: false,
								data: {
									name: ''
								},
								title: ''
							};
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

{#if confirmModals.active == 'deleteSnapshot'}
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
