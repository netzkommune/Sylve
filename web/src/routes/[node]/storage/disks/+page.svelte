<script lang="ts">
	import { destroyDisk, destroyPartition, initializeGPT, listDisks } from '$lib/api/disk/disk';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialog from '$lib/components/custom/AlertDialog.svelte';
	import KvTableModal from '$lib/components/custom/KVTableModal.svelte';
	import CreatePartition from '$lib/components/disk/CreatePartition.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as ContextMenu from '$lib/components/ui/context-menu';
	import { localStore } from '$lib/stores/localStore.svelte';
	import { type Disk, type Partition } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { diskSpaceAvailable, getGPTLabel, parseSMART, simplifyDisks } from '$lib/utils/disk';
	import { handleAPIError } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { TableHandler } from '@vincjo/datatables';
	import humanFormat from 'human-format';
	import { onMount, untrack } from 'svelte';
	import toast from 'svelte-french-toast';

	interface Data {
		disks: Disk[];
		pools: Zpool[];
	}

	type ExpandedRows = Record<number, boolean>;

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['diskList'],
			queryFn: async () => {
				return await simplifyDisks(await listDisks());
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.disks
		},
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.pools
		}
	]);

	const table = new TableHandler(data.disks);

	let disks = $derived($results[0].data as Disk[]);
	let pools = $results[1].data as Zpool[];

	$effect(() => {
		table.setRows($results[0].data as Disk[]);
	});

	onMount(() => {
		if (disks.length) {
			disks.forEach((_, index) => {
				expandedRows[index] = true;
			});
		}
	});

	let sortHandlers: Record<string, any> = {};
	let activeRow: string | null = $state(null);

	let wipeModal = $state({
		open: false,
		title: ''
	});

	let partitionModal: {
		open: boolean;
		disk: Disk | null;
	} = $state({
		open: false,
		disk: null
	});

	let smartModal = $state({
		open: false,
		title: '',
		KV: {},
		type: ''
	});

	const expandedRows: ExpandedRows = $state({});
	const keys = [
		'Device',
		'Type',
		'Usage',
		'Size',
		'GPT',
		'Model',
		'Serial',
		'S.M.A.R.T.',
		'Wearout'
	];

	let visibleColumns = localStore(
		'diskVisibleColumns',
		Object.fromEntries(keys.map((key) => [key, true]))
	);

	let openContextMenuId = $state<string | null>(null);

	function handleContextMenuOpen(id: string) {
		openContextMenuId = id;
	}

	function handleContextMenuClose() {
		openContextMenuId = null;
	}

	let activeDisk: Disk | null = $derived.by(() => {
		if (activeRow !== null) {
			return disks.find((disk) => disk.Device === activeRow) || null;
		}
		return null;
	});

	let activePartition: Partition | null = $derived.by(() => {
		if (activeRow !== null) {
			let [device, partitionIndex] = activeRow.split('-');
			let disk = disks.find((d) => d.Device === device);
			if (disk && partitionIndex !== undefined) {
				return disk.Partitions[parseInt(partitionIndex)] || null;
			}
		}
		return null;
	});

	function handleRowClick(device: string) {
		activeRow = activeRow === device ? null : device;
	}

	function toggleChildren(index: number) {
		expandedRows[index] = !expandedRows[index];
		if (expandedRows[index]) {
			activeRow = index.toString();
		}
	}

	function isToggled(index: number) {
		return expandedRows[index] ?? false;
	}

	keys.forEach((key) => {
		sortHandlers[key] = table.createSort(key as keyof Disk, {
			locales: 'en',
			options: { numeric: true, sensitivity: 'base' }
		});
	});

	async function diskAction(action: string) {
		if (action === 'smart') {
			if (activeDisk) {
				smartModal.open = false;
				smartModal.title = `${getTranslation('disk.smart', 'S.M.A.R.T')} Values (${activeDisk.Device})`;
				if (activeDisk.Type === 'NVMe') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					smartModal.open = true;
					smartModal.type = 'kv';
				} else if (activeDisk.Type === 'HDD' || activeDisk.Type === 'SSD') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					console.log(smartModal.KV);
					smartModal.open = true;
					smartModal.type = 'array';
				}
			}
		}

		if (action === 'wipe') {
			wipeModal.open = true;

			if (activePartition !== null) {
				wipeModal.title = `${getTranslation('common.this_action_cannot_be_undone', 'This action cannot be undone')}. ${getTranslation(
					'common.this_will_permanently',
					'This will permanently'
				)} <b>${getTranslation('common.delete', 'delete')}</b> ${getTranslation('disk.partition', 'disk')} <b>${activePartition.name}</b>.`;
			} else if (activeDisk !== null) {
				wipeModal.title = `${getTranslation('common.this_action_cannot_be_undone', 'This action cannot be undone')}. ${getTranslation(
					'common.this_will_permanently',
					'This will permanently'
				)} <b>${getTranslation('disk.wipe', 'wipe')}</b> ${getTranslation('disk.disk', 'disk')} <b>${activeDisk.Device}</b>.`;
			}
		}

		if (action === 'gpt') {
			if (activeDisk) {
				const response = await initializeGPT(activeDisk.Device);
				if (response.status === 'success') {
					toast.success(
						`${capitalizeFirstLetter(getTranslation('disk.disk', 'Disk'))} ${activeDisk.Device} ${getTranslation(
							'disk.gpt_initialized',
							'initialized with GPT'
						)}`
					);
				} else {
					handleAPIError(response);
				}
			}
		}

		if (action === 'partition') {
			partitionModal.open = true;
			partitionModal.disk = activeDisk;
		}
	}

	function toggleColumnVisibility(columnKey: string) {
		const wouldHideAll =
			Object.entries(visibleColumns.value).filter(([k, v]) => k !== columnKey && v).length === 0 &&
			visibleColumns.value[columnKey];

		if (!wouldHideAll) {
			visibleColumns.value[columnKey] = !visibleColumns.value[columnKey];
		}
	}

	function resetColumns() {
		visibleColumns.value = Object.fromEntries(keys.map((key) => [key, true]));
	}

	let buttonAbilities = $state({
		smart: {
			ability: false,
			reason: ''
		},
		gpt: {
			ability: false,
			reason: ''
		},
		wipe: {
			ability: false,
			reason: ''
		},
		createPartition: {
			ability: false,
			reason: ''
		}
	});

	$effect(() => {
		if (activeDisk) {
			untrack(() => {
				buttonAbilities.smart.ability = activeDisk['S.M.A.R.T.'] !== null;

				if (!buttonAbilities.smart.ability) {
					buttonAbilities.smart.reason = getTranslation(
						'disk.no_smart_data',
						'No S.M.A.R.T data available'
					);
				}

				buttonAbilities.gpt.ability = !activeDisk.GPT;
				if (!buttonAbilities.gpt.ability) {
					buttonAbilities.gpt.reason = getTranslation(
						'disk.gpt_already_initialized',
						'GPT already initialized'
					);
				}

				if (activeDisk.Usage === 'ZFS') {
					buttonAbilities.gpt.ability = false;
					buttonAbilities.gpt.reason = getTranslation(
						'disk.zfs_vdev',
						'ZFS Vdev does not require GPT'
					);
				}

				if (activeDisk.Usage === 'ZFS' || activeDisk.Usage === 'Unused') {
					buttonAbilities.wipe.ability = false;
					if (activeDisk.Usage === 'ZFS') {
						buttonAbilities.wipe.reason = getTranslation(
							'disk.zfs_vdev',
							'ZFS Vdev cannot be wiped'
						);
					} else if (activeDisk.Usage === 'Unused' && activeDisk.GPT) {
						buttonAbilities.wipe.ability = true;
					} else if (activeDisk.Usage === 'Unused' && !activeDisk.GPT) {
						buttonAbilities.wipe.reason = getTranslation('disk.no_gpt', 'GPT not initialized');
					}
				} else {
					buttonAbilities.wipe.ability = true;
				}

				buttonAbilities.createPartition.ability =
					activeDisk.GPT &&
					diskSpaceAvailable(activeDisk, 128 * 1024 * 1024) &&
					activeDisk.Usage !== 'ZFS';

				if (!buttonAbilities.createPartition.ability) {
					if (activeDisk.Usage === 'ZFS') {
						buttonAbilities.createPartition.reason = getTranslation(
							'disk.zfs_vdev',
							'ZFS Vdev cannot be partitioned'
						);
					} else if (!diskSpaceAvailable(activeDisk, 128 * 1024 * 1024)) {
						buttonAbilities.createPartition.reason = getTranslation(
							'disk.no_space',
							'No space available for partitioning'
						);
					} else if (!activeDisk.GPT) {
						buttonAbilities.createPartition.reason = getTranslation(
							'disk.no_gpt',
							'GPT not initialized'
						);
					}
				}
			});
		} else if (activePartition) {
			untrack(() => {
				buttonAbilities.gpt.ability = false;
				buttonAbilities.wipe.ability = true;
				buttonAbilities.createPartition.ability = false;
				buttonAbilities.smart.ability = false;
			});
		} else {
			untrack(() => {
				buttonAbilities.gpt.ability = false;
				buttonAbilities.wipe.ability = false;
				buttonAbilities.createPartition.ability = false;
				buttonAbilities.smart.ability = false;
			});
		}
	});
</script>

<div class="flex h-full flex-col overflow-hidden">
	<div class="inline-flex w-full gap-2 border-b px-3 py-2">
		<Button
			size="sm"
			class="h-8 bg-neutral-600 text-white hover:bg-neutral-700 disabled:!pointer-events-auto disabled:hover:bg-neutral-600"
			disabled={!buttonAbilities.smart.ability}
			onclick={() => diskAction('smart')}
		>
			Show S.M.A.R.T values
		</Button>
		<Button
			size="sm"
			class="h-8 bg-neutral-600 text-white hover:bg-neutral-700 disabled:!pointer-events-auto disabled:hover:bg-neutral-600"
			title={buttonAbilities.gpt.reason}
			disabled={!buttonAbilities.gpt.ability}
			onclick={() => diskAction('gpt')}
		>
			Initialize Disk with GPT
		</Button>

		{#if activeDisk}
			<Button
				size="sm"
				class="h-8 bg-neutral-600 text-white hover:bg-neutral-700 disabled:!pointer-events-auto disabled:hover:bg-neutral-600"
				title={buttonAbilities.wipe.reason}
				disabled={!buttonAbilities.wipe.ability}
				onclick={() => diskAction('wipe')}
			>
				Wipe Disk
			</Button>
		{/if}

		{#if activePartition}
			<Button
				size="sm"
				class="h-8 bg-neutral-600 text-white hover:bg-neutral-700"
				disabled={!buttonAbilities.wipe.ability}
				onclick={() => diskAction('wipe')}
			>
				Delete Partition
			</Button>
		{/if}

		<Button
			size="sm"
			class="{activeDisk === null
				? 'hidden'
				: ''} h-8 bg-neutral-600 text-white hover:bg-neutral-700 disabled:!pointer-events-auto disabled:hover:bg-neutral-600"
			title={buttonAbilities.createPartition.reason}
			onclick={() => diskAction('partition')}
			disabled={!buttonAbilities.createPartition.ability}
		>
			Create Partition
		</Button>
	</div>

	<KvTableModal
		titles={{
			main: smartModal.title,
			key: getTranslation('disk.attribute', 'Attribute'),
			value: getTranslation('disk.value', 'Value')
		}}
		open={smartModal.open}
		KV={smartModal.KV}
		type={smartModal.type}
		actions={{
			close: () => {
				smartModal.open = false;
			}
		}}
	></KvTableModal>

	<div class="relative flex h-full w-full cursor-pointer flex-col">
		<div class="flex-1">
			<div class="h-full overflow-y-auto">
				<table class="mb-10 w-full min-w-max border-collapse">
					<thead>
						<tr>
							{#each keys as key}
								{#if visibleColumns.value[key]}
									<th
										class="group h-8 w-48 whitespace-nowrap border border-neutral-300 px-3 text-left text-black dark:border-neutral-800 dark:text-white"
									>
										<ContextMenu.Root
											open={openContextMenuId === key}
											closeOnItemClick={false}
											onOpenChange={(open) =>
												open ? handleContextMenuOpen(key) : handleContextMenuClose()}
										>
											<ContextMenu.Trigger class="flex h-full w-full">
												<button
													class="relative flex w-full items-center"
													onclick={() => sortHandlers[key].set()}
												>
													<span>{key}</span>
													<Icon
														icon={sortHandlers[key].direction === 'asc'
															? 'lucide:sort-asc'
															: 'lucide:sort-desc'}
														class="ml-2 mt-1 h-4 w-4 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
													/>
												</button>
											</ContextMenu.Trigger>
											<ContextMenu.Content>
												<ContextMenu.Label>Toggle Columns</ContextMenu.Label>
												<ContextMenu.Separator />
												{#each keys as columnKey}
													<ContextMenu.CheckboxItem
														checked={visibleColumns.value[columnKey]}
														onCheckedChange={(e: boolean | 'indeterminate') => {
															toggleColumnVisibility(columnKey);
														}}
													>
														{columnKey}
													</ContextMenu.CheckboxItem>
												{/each}
												<ContextMenu.Separator />
												<ContextMenu.Item onclick={resetColumns}>Reset Columns</ContextMenu.Item>
											</ContextMenu.Content>
										</ContextMenu.Root>
									</th>
								{/if}
							{/each}
						</tr>
					</thead>

					<tbody>
						{#each table.rows as row, index}
							<tr
								class={activeRow === row.Device ? 'bg-muted-foreground/40 dark:bg-muted' : ''}
								onclick={(event: MouseEvent) => {
									if (!(event.target as HTMLElement).closest('.toggle-icon')) {
										handleRowClick(row.Device);
									}
								}}
							>
								{#each keys as key, keyIndex}
									{#if visibleColumns.value[key]}
										{#if key === 'Device'}
											<td class="whitespace-nowrap px-3 py-1.5">
												<div class="flex items-center">
													<Icon
														icon={isToggled(index) ? 'lucide:minus-square' : 'lucide:plus-square'}
														class="toggle-icon mr-1.5 h-4 w-4 cursor-pointer"
														onclick={(event: MouseEvent) => {
															event.stopPropagation();
															toggleChildren(index);
														}}
													/>
													<Icon icon="mdi:harddisk" class="mr-1.5 h-4 w-4" />
													<span>{row.Device}</span>
												</div>
											</td>
										{:else if key === 'GPT'}
											<td class="whitespace-nowrap px-3 py-1.5">
												{getGPTLabel(disks.filter((d) => d.Device === row.Device)[0], pools)}
											</td>
										{:else if key === 'Size'}
											<td class="whitespace-nowrap px-3 py-1.5">{humanFormat(row.Size)}</td>
										{:else}
											<td class="whitespace-nowrap px-3 py-1.5">{row[key as keyof Disk]}</td>
										{/if}
									{/if}
								{/each}
							</tr>
							{#if expandedRows[index] && row.Partitions}
								{#each row.Partitions as child, childIndex}
									<tr
										class={activeRow === `${row.Device}-${childIndex}`
											? 'bg-muted-foreground/40 dark:bg-muted'
											: ''}
										onclick={() => handleRowClick(`${row.Device}-${childIndex}`)}
									>
										{#each keys as key, _}
											{#if visibleColumns.value[key]}
												{#if key === 'Device'}
													<td class="whitespace-nowrap px-3 py-0">
														<div class="relative flex items-center">
															{#if row.Partitions.length > 1}
																<div
																	class="bg-muted-foreground absolute left-1.5 top-0 h-full w-0.5"
																	style="height: calc(100% + 0.8rem);"
																	class:hidden={childIndex === row.Partitions.length - 1}
																></div>
															{:else}
																<div
																	class="bg-muted-foreground absolute left-1.5 top-0 h-3 w-0.5"
																></div>
															{/if}
															<div class="relative left-1.5 top-0 mr-2 w-4">
																<div class="bg-muted-foreground h-0.5 w-4"></div>
															</div>
															{#if childIndex === row.Partitions.length - 1}
																<div
																	class="absolute bottom-0 left-2 h-1/2 w-0.5 bg-transparent"
																></div>
															{/if}
															<Icon icon="mdi:harddisk" class="mr-1.5 h-4 w-4" />
															<span>{child.name}</span>
														</div>
													</td>
												{:else if key === 'Type'}
													<td class="whitespace-nowrap px-3 py-0">partition</td>
												{:else if key === 'Usage'}
													<td class="whitespace-nowrap px-3 py-0">{child.usage}</td>
												{:else if key === 'Size'}
													<td class="whitespace-nowrap px-3 py-0">{humanFormat(child.size)}</td>
												{:else if key === 'GPT'}
													<td class="whitespace-nowrap px-3 py-0">{row.GPT ? 'Yes' : 'No'}</td>
												{:else}
													<td class="whitespace-nowrap px-3 py-0"></td>
												{/if}
											{/if}
										{/each}
									</tr>
								{/each}
							{/if}
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<AlertDialog
	open={wipeModal.open}
	names={{ parent: 'disks', element: wipeModal.title || '' }}
	actions={{
		onConfirm: async () => {
			if (activeDisk || activePartition) {
				if (activeDisk) {
					const result = await destroyDisk(activeDisk.Device);
					if (result.status === 'success') {
						toast.success(getTranslation('disk.full_wipe_success', 'Disk wiped successfully'));
					}
				} else if (activePartition) {
					const result = await destroyPartition(`/dev/${activePartition.name}`);
					if (result.status === 'success') {
						toast.success(getTranslation('disk.partition_wipe_success', 'Disk wiped successfully'));
					}
				}
			}
			wipeModal.title = '';
			wipeModal.open = false;
		},
		onCancel: () => {
			wipeModal.title = '';
			wipeModal.open = false;
		}
	}}
	customTitle={wipeModal.title}
></AlertDialog>

<CreatePartition
	open={partitionModal.open}
	disk={partitionModal.disk}
	onCancel={() => {
		partitionModal.open = false;
		partitionModal.disk = null;
	}}
/>
