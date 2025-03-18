<script lang="ts">
	import KvTableModal from '$lib/components/custom/KVTableModal.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as ContextMenu from '$lib/components/ui/context-menu';
	import { localStore } from '$lib/stores/localStore.svelte';
	import { type Disk, type Partition } from '$lib/types/disk/disk';
	import { parseSMART } from '$lib/utils/disk';
	import { getTranslation } from '$lib/utils/i18n';
	import Icon from '@iconify/svelte';
	import { TableHandler } from '@vincjo/datatables';
	import humanFormat from 'human-format';
	import { onMount } from 'svelte';

	interface Data {
		disks: Disk[];
	}

	type ExpandedRows = Record<number, boolean>;

	let { data }: { data: Data } = $props();
	let disks = $state(data.disks);
	let sortHandlers: Record<string, any> = {};
	let activeRow: string | null = $state(null);

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

	const table = new TableHandler(data.disks);

	keys.forEach((key) => {
		sortHandlers[key] = table.createSort(key as keyof Disk, {
			locales: 'en',
			options: { numeric: true, sensitivity: 'base' }
		});
	});

	onMount(() => {
		if (disks.length) {
			disks.forEach((_, index) => {
				expandedRows[index] = true;
			});
		}
	});

	function diskAction(action: string) {
		if (action === 'smart') {
			if (activeDisk) {
				smartModal.title = `${getTranslation('disk.smart', 'S.M.A.R.T')} Values (${activeDisk.Device})`;
				if (activeDisk.Type === 'NVMe') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					smartModal.open = true;
					smartModal.type = 'kv';
				} else if (activeDisk.Type === 'HDD') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					smartModal.open = true;
					smartModal.type = 'array';
				}
			}
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
</script>

<div class="flex h-full flex-col overflow-hidden">
	<div class="inline-flex w-full gap-2 border-b px-3 py-2">
		<Button size="sm" class="h-8 bg-neutral-600 text-white hover:bg-neutral-700">Reload</Button>
		<Button
			size="sm"
			class="h-8 bg-neutral-600 text-white hover:bg-neutral-700"
			disabled={activeDisk === null}
			onclick={() => diskAction('smart')}
		>
			Show S.M.A.R.T values
		</Button>
		<Button
			size="sm"
			class="h-8 bg-neutral-600 text-white hover:bg-neutral-700"
			disabled={activeDisk === null ||
				!(activeDisk && activeDisk.Partitions.length < 1) ||
				(activeDisk && activeDisk.GPT)}
			onclick={() => diskAction('gpt')}
		>
			Initialize Disk with GPT
		</Button>
		<Button
			size="sm"
			class="h-8 bg-neutral-600 text-white hover:bg-neutral-700"
			disabled={activeDisk === null && activePartition === null}
			onclick={() => diskAction('wipe')}
		>
			Wipe {activePartition !== null ? 'Partition' : activeDisk !== null ? 'Disk' : ''}
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
												<div
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
												</div>
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
											<td class="whitespace-nowrap px-3 py-1.5">{row.GPT ? 'Yes' : 'No'}</td>
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
