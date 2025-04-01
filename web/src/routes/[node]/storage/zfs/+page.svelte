<script lang="ts">
	import { listDisks } from '$lib/api/disk/disk';
	import { getPools } from '$lib/api/zfs/pool';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import type { Disk } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { simplifyDisks } from '$lib/utils/disk';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { flip } from 'svelte/animate';
	import { slide } from 'svelte/transition';

	import { createEmptyArrayOfArrays } from '$lib/utils/arr';
	import { draggable, dropzone } from '$lib/utils/dnd';
	import { untrack } from 'svelte';

	interface Data {
		disks: Disk[];
		pools: Zpool[];
	}

	interface UnusedDisk {
		name: string;
		size: number;
		gpt: boolean;
		type: string;
	}

	interface DiskContainer {
		id: string;
		name: string;
		size: number;
		type: string;
	}

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

	let disks = $derived($results[0].data as Disk[]);
	let pools = $results[1].data as Zpool[];
	let advancedChecked: boolean = $state(false);

	$inspect('advancedChecked', advancedChecked);

	let useableDisks = $derived.by(() => {
		const unusedDisks: UnusedDisk[] = [];
		for (const disk of disks) {
			if (disk.Usage === 'Unused' && disk.GPT === false) {
				unusedDisks.push({
					name: disk.Device,
					size: disk.Size,
					gpt: disk.GPT,
					type: disk.Type === 'Unknown' ? 'HDD' : disk.Type
				});
			}

			if (disk.Usage === 'Partitions') {
				for (const partition of disk.Partitions) {
					for (const pool of pools) {
						let skip = false;

						for (const vdev of pool.vdevs) {
							if (vdev.name.includes(partition.name)) {
								skip = true;
								continue;
							}
						}

						if (partition.usage === 'ZFS' && !skip) {
							unusedDisks.push({
								name: `/dev/${partition.name}`,
								size: partition.size,
								gpt: disk.GPT,
								type: 'Partition'
							});
						}
					}
				}
			}
		}

		return unusedDisks;
	});

	let open: boolean = $state(false);
	let name: string = $state('');
	let vdevCount: number = $state(1);
	let createEnabled: boolean = $state(false);

	function mapAvailable(): DiskContainer[] {
		return useableDisks.map((disk) => ({
			id: disk.name,
			name: disk.name,
			size: disk.size,
			type: disk.type
		}));
	}

	function close() {
		open = false;
		name = '';
		vdevCount = 1;
	}

	let availableDisks: DiskContainer[] = $state(mapAvailable());
	let diskContainers: DiskContainer[][] | null = $state(null);

	$effect(() => {
		vdevCount = Math.max(1, Math.min(128, vdevCount));
		untrack(() => {
			console.log(useableDisks);
			if (diskContainers === null) {
				diskContainers = createEmptyArrayOfArrays(vdevCount);
			} else {
				// Create a new array with the new length first
				const newContainers = Array(vdevCount).fill([]);

				// Then copy over existing containers up to the minimum of old and new length
				const copyLength = Math.min(diskContainers.length, vdevCount);
				for (let i = 0; i < copyLength; i++) {
					newContainers[i] = [...diskContainers[i]]; // Create a new array reference
				}

				diskContainers = newContainers;
			}
		});
	});

	function handleDiskDrop(containerId: number, event: DragEvent) {
		event.preventDefault();

		const diskId = event.dataTransfer?.getData('application/disk');
		if (!diskId) return;

		let diskToMove: any = null;

		if (diskContainers) {
			// Try to find and remove the disk from a container
			const sourceContainerIndex = diskContainers.findIndex((container) =>
				container.some((disk) => disk.id === diskId)
			);

			if (sourceContainerIndex !== -1) {
				const container = diskContainers[sourceContainerIndex];
				diskToMove = container.find((disk) => disk.id === diskId) || null;

				// Remove it only if moving to a different container
				if (sourceContainerIndex !== containerId) {
					diskContainers[sourceContainerIndex] = container.filter((disk) => disk.id !== diskId);
				}
			} else {
				// Try to find and remove from available disks
				const diskIndex = availableDisks.findIndex((disk) => disk.id === diskId);
				if (diskIndex !== -1) {
					diskToMove = availableDisks[diskIndex];
					availableDisks.splice(diskIndex, 1); // Remove from availableDisks
				}
			}

			// Add to target container if found and it's not already there
			if (diskToMove && sourceContainerIndex !== containerId) {
				diskContainers[containerId].push(diskToMove);
				console.log(diskContainers);
			}

			// Enable create button if any container has disks
			createEnabled = diskContainers.some((container) => container.length > 0);
		}
	}

	// Function to handle removing a disk from a container
	function returnDiskToPool(containerId: number, diskId: string) {
		if (diskContainers) {
			const container = diskContainers[containerId];
			const diskIndex = container.findIndex((d) => d.id === diskId);
			if (diskIndex !== -1) {
				const disk = container[diskIndex];
				diskContainers[containerId] = container.filter((d) => d.id !== diskId);
				availableDisks = [...availableDisks, disk];
				createEnabled = diskContainers.some((container) => container.length > 0);
			}
		}
	}

	const raid = [
		{ value: 'mirror', label: 'Mirror' },
		{ value: 'raidz1', label: 'RAIDZ1' },
		{ value: 'raidz2', label: 'RAIDZ2' },
		{ value: 'raidz3', label: 'RAIDZ3' }
	];

	const compression = [
		{ value: 'lz4', label: 'LZ4' },
		{ value: 'zstd', label: 'ZSTD' },
		{ value: 'gzip', label: 'GZIP' },
		{ value: 'zle', label: 'ZLE' }
	];

	const ashift = [
		{ value: 'lz4', label: 'LZ4' },
		{ value: 'zstd', label: 'ZSTD' },
		{ value: 'gzip', label: 'GZIP' },
		{ value: 'zle', label: 'ZLE' }
	];

	let pairs: { key: string; value: string }[] = $state([{ key: '', value: '' }]);

	function addPair() {
		pairs = [...pairs, { key: '', value: '' }];
	}

	function removePair(index: number) {
		pairs = pairs.filter((_, i) => i !== index);
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			on:click={() => (open = !open)}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
</div>

<Dialog.Root bind:open onOutsideClick={() => close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
	>
		<div class="flex items-center justify-between px-4 py-3">
			<Dialog.Header class="p-0">
				<Dialog.Title>Create ZFS Pool</Dialog.Title>
			</Dialog.Header>

			<Dialog.Close
				class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
				onclick={() => close()}
			>
				<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
			</Dialog.Close>
		</div>
		<Tabs.Root value="tab-1" class="w-full overflow-hidden">
			<Tabs.List class="grid w-full grid-cols-2 p-0 px-4">
				<Tabs.Trigger value="tab-1" class="border-b">Devices</Tabs.Trigger>
				<Tabs.Trigger value="tab-2" class="border-b">Options</Tabs.Trigger>
			</Tabs.List>
			<Tabs.Content class="mt-0" value="tab-1">
				<Card.Root class="border-none pb-4">
					<Card.Content class="flex gap-4 p-4 !pb-0">
						<div class="flex-1 space-y-1">
							<Label for="name">Name</Label>
							<Input type="text" id="name" placeholder="name" bind:value={name} />
						</div>
						<div class="flex-1 space-y-1">
							<Label for="vdev_count">Virtual Devices</Label>
							<Input type="number" id="vdev_count" placeholder="1" min={1} bind:value={vdevCount} />
						</div>
					</Card.Content>
					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div>
							<Label for="vdev_count" class="">VDEV</Label>
							<div
								class="border-primary-foreground bg-primary-foreground mt-1 w-full overflow-hidden rounded-lg border-y p-4"
							>
								<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
									<div class="flex items-center justify-center gap-7 pr-4">
										{#each Array(vdevCount) as _, i}
											{#if diskContainers}
												{#if diskContainers[i]}
													<div class="relative flex flex-col">
														{#if diskContainers[i].length > 0}
															<div
																class="absolute right-1 top-1 z-50 cursor-pointer text-yellow-700 hover:text-yellow-600"
															>
																<Tooltip.Root>
																	<Tooltip.Trigger
																		><Icon
																			icon="carbon:warning-filled"
																			class="h-5 w-5"
																		/></Tooltip.Trigger
																	>
																	<Tooltip.Content>
																		<p>
																			Lorem Ipsum is simply dummy text of the printing and
																			typesetting industry. Lorem Ipsum has been the industry's
																		</p>
																	</Tooltip.Content>
																</Tooltip.Root>
															</div>
														{/if}
														<div
															class={`relative h-28 w-48 flex-shrink-0 overflow-auto rounded-lg bg-neutral-200 p-2 dark:bg-neutral-950
                                                            ${diskContainers[i].length > 0 ? 'border border-yellow-700 ' : ''}`}
															use:dropzone={{
																on_dropzone: (_: unknown, event: DragEvent) =>
																	handleDiskDrop(i, event),
																dragover_class: 'droppable'
															}}
														>
															{#if diskContainers[i].length === 0}
																<div
																	class="flex h-full items-center justify-center text-neutral-500"
																>
																	Drop disks here
																</div>
															{:else}
																<div
																	class="flex h-full flex-wrap items-center justify-center gap-2"
																>
																	{#each diskContainers[i] as disk (disk.id)}
																		<div animate:flip={{ duration: 300 }} class="relative">
																			<!-- {#if disk.type === 'NVMe'}
																			<Icon
																				icon="bi:nvme"
																				class="h-11 w-11 rotate-90 text-blue-500"
																			/>
																		{:else}
																			<Icon
																				icon={disk.type === 'SSD'
																					? 'icon-park-outline:ssd'
																					: 'mdi:harddisk'}
																				class="h-12 w-12 {disk.type === 'SSD'
																					? 'text-blue-500'
																					: 'text-green-500'}"
																			/>
																		{/if} -->
																			{#if disk.type === 'SSD'}
																				<Icon
																					icon="icon-park-outline:ssd"
																					class="h-11 w-11 text-blue-500"
																				/>
																			{:else if disk.type === 'NVMe'}
																				<Icon
																					icon="bi:nvme"
																					class="h-11 w-11 rotate-90 text-blue-500"
																				/>
																			{:else if disk.type === 'HDD'}
																				<Icon
																					icon="mdi:harddisk"
																					class="h-11 w-11 text-green-500"
																				/>
																			{:else if disk.type === 'Partition'}
																				<Icon
																					icon="ant-design:partition-outlined"
																					class="h-11 w-11 rotate-90 text-blue-500"
																				/>
																			{/if}
																			<div class="max-w-[48px] truncate text-center text-xs">
																				{disk.name.split('/').pop()}
																			</div>
																			<button
																				class="absolute -right-1 -top-1 rounded-full bg-red-500 p-0.5 text-white hover:bg-red-600"
																				onclick={() => returnDiskToPool(i, disk.id)}
																			>
																				<Icon icon="mdi:close" class="h-3 w-3" />
																			</button>
																		</div>
																	{/each}
																</div>
															{/if}
														</div>
														<p
															class="mt-2 text-center text-xs text-neutral-800 dark:text-neutral-300"
														>
															VDEV {i + 1}
														</p>
													</div>
												{/if}
											{/if}
										{/each}
									</div>
								</ScrollArea>
							</div>
						</div>

						<div>
							<Label for="vdev_count" class="">Disks</Label>
							<div
								class="border-primary-foreground bg-primary-foreground mt-1 grid grid-cols-4 gap-6 overflow-hidden border-y p-4"
							>
								<div class="">
									<Label class="">HDD</Label>
									<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
										<ScrollArea
											class="w-full whitespace-nowrap rounded-md"
											orientation="horizontal"
										>
											<div class="flex min-h-[80px] items-center justify-center gap-4">
												{#each availableDisks.filter((disk) => disk.type === 'HDD') as disk (disk.id)}
													<div class="text-center" animate:flip={{ duration: 300 }}>
														<div class="cursor-move" use:draggable={disk.id}>
															<Icon icon="mdi:harddisk" class="h-11 w-11 text-green-500" />
														</div>
														<div class="max-w-[64px] truncate text-xs">
															{disk.name.split('/').pop()}
														</div>
														<div class="text-xs text-neutral-400">
															{Math.round(disk.size / (1024 * 1024 * 1024))} GB
														</div>
													</div>
												{/each}

												{#if availableDisks.filter((disk) => disk.type === 'HDD').length === 0}
													<div
														class="flex h-16 w-full items-center justify-center text-neutral-400"
													>
														No available disks
													</div>
												{/if}
											</div>
										</ScrollArea>
									</div>
								</div>

								<div class="">
									<Label class="">SSD</Label>
									<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
										<ScrollArea
											class="w-full whitespace-nowrap rounded-md "
											orientation="horizontal"
										>
											<div class="flex min-h-[80px] items-center justify-center gap-4">
												{#each availableDisks.filter((disk) => disk.type === 'SSD') as disk (disk.id)}
													<div class="text-center" animate:flip={{ duration: 300 }}>
														<div class="cursor-move" use:draggable={disk.id}>
															<Icon icon="icon-park-outline:ssd" class="h-11 w-11 text-blue-500" />
														</div>
														<div class="max-w-[64px] truncate text-xs">
															{disk.name.split('/').pop()}
														</div>
														<div class="text-xs text-neutral-400">
															{Math.round(disk.size / (1024 * 1024 * 1024))} GB
														</div>
													</div>
												{/each}

												{#if availableDisks.filter((disk) => disk.type === 'SSD').length === 0}
													<div
														class="flex h-16 w-full items-center justify-center text-neutral-400"
													>
														No available disks
													</div>
												{/if}
											</div>
										</ScrollArea>
									</div>
								</div>

								<div class="">
									<Label class="">NVME</Label>
									<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
										<ScrollArea
											class="w-full whitespace-nowrap rounded-md "
											orientation="horizontal"
										>
											<div class="flex min-h-[80px] items-center justify-center gap-4">
												{#each availableDisks.filter((disk) => disk.type === 'NVMe') as disk (disk.id)}
													<div class="text-center" animate:flip={{ duration: 300 }}>
														<div class="cursor-move" use:draggable={disk.id}>
															<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
														</div>
														<div class="max-w-[64px] truncate text-xs">
															{disk.name.split('/').pop()}
														</div>
														<div class="text-xs text-neutral-400">
															{Math.round(disk.size / (1024 * 1024 * 1024))} GB
														</div>
													</div>
												{/each}

												{#if availableDisks.filter((disk) => disk.type === 'NVMe').length === 0}
													<div
														class="flex h-16 w-full items-center justify-center text-neutral-400"
													>
														No available disks
													</div>
												{/if}
											</div>
										</ScrollArea>
									</div>
								</div>

								<div class="">
									<Label class="">Partition</Label>
									<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
										<ScrollArea
											class="w-full whitespace-nowrap rounded-md "
											orientation="horizontal"
										>
											<div class="flex min-h-[80px] items-center justify-center gap-4">
												{#each availableDisks.filter((disk) => disk.type === 'Partition') as disk (disk.id)}
													<div class="text-center" animate:flip={{ duration: 300 }}>
														<div class="cursor-move" use:draggable={disk.id}>
															<Icon
																icon="ant-design:partition-outlined"
																class="h-11 w-11 rotate-90 text-blue-500"
															/>
														</div>
														<div class="max-w-[64px] truncate text-xs">
															{disk.name.split('/').pop()}
														</div>
														<div class="text-xs text-neutral-400">
															{Math.round(disk.size / (1024 * 1024 * 1024))} GB
														</div>
													</div>
												{/each}

												{#if availableDisks.filter((disk) => disk.type === 'Partition').length === 0}
													<div
														class="flex h-16 w-full items-center justify-center text-neutral-400"
													>
														No available partitions
													</div>
												{/if}
											</div>
										</ScrollArea>
									</div>
								</div>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tabs.Content>

			<Tabs.Content class="mt-0" value="tab-2">
				<Card.Root class="min-h-[20vh] border-none pb-6">
					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div transition:slide class="grid grid-cols-1 gap-4 md:grid-cols-3">
							<div class="h-full space-y-1">
								<Label class="w-24 whitespace-nowrap text-sm" for="raid">RAID:</Label>
								<Select.Root>
									<Select.Trigger class="w-full">
										<Select.Value placeholder="Select a RAID" />
									</Select.Trigger>
									<Select.Content class="max-h-36 overflow-y-auto">
										<Select.Group>
											{#each raid as fruit}
												<Select.Item value={fruit.value} label={fruit.label}
													>{fruit.label}</Select.Item
												>
											{/each}
										</Select.Group>
									</Select.Content>
									<Select.Input name="raid" />
								</Select.Root>
							</div>

							<div class="space-y-1">
								<Label class="w-24 whitespace-nowrap text-sm" for="compression">Compression:</Label>
								<Select.Root>
									<Select.Trigger class="w-full">
										<Select.Value placeholder="Select a Compression" />
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each compression as fruit}
												<Select.Item value={fruit.value} label={fruit.label}
													>{fruit.label}</Select.Item
												>
											{/each}
										</Select.Group>
									</Select.Content>
									<Select.Input name="compression" />
								</Select.Root>
							</div>

							<div class="space-y-1">
								<Label class="w-24 whitespace-nowrap text-sm" for="ashift">ASHIFT:</Label>
								<Select.Root>
									<Select.Trigger class="w-full">
										<Select.Value placeholder="Select a ASHIFT" />
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each ashift as fruit}
												<Select.Item value={fruit.value} label={fruit.label}
													>{fruit.label}</Select.Item
												>
											{/each}
										</Select.Group>
									</Select.Content>
									<Select.Input name="ashift" />
								</Select.Root>
							</div>
						</div>

						<div transition:slide class="mt-2 flex items-center space-x-2 md:mt-3">
							<Label
								id="advanced-label"
								for="advanced-checkbox"
								class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
							>
								Advanced
							</Label>
							<Checkbox
								id="advanced-checkbox"
								bind:checked={advancedChecked}
								aria-labelledby="advanced-label"
							/>
						</div>

						{#if advancedChecked}
							<div transition:slide class="max-h-[250px] space-y-2 overflow-y-auto">
								{#each pairs as pair, index}
									<div transition:slide class="flex items-center gap-4">
										<Input
											class="h-8"
											type="text"
											id="name"
											placeholder="key"
											bind:value={pair.key}
										/>

										<Input
											class="h-8"
											type="text"
											id="name"
											placeholder="value"
											bind:value={pair.value}
										/>

										{#if pairs.length > 1}
											<button
												onclick={() => removePair(index)}
												class="hover:bg-muted rounded px-2 py-1 text-white"
											>
												<Icon icon="ic:twotone-remove" class="h-6 w-6" />
											</button>
										{/if}
									</div>
								{/each}
							</div>
							<div transition:slide class="flex justify-end">
								<button onclick={addPair} class=" hover:bg-muted rounded px-2 py-1 text-white">
									<Icon icon="icons8:plus" class="h-6 w-6" />
								</button>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</Tabs.Content>
		</Tabs.Root>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-4 py-3">
			<div class="flex gap-2">
				<Button variant="outline" class="h-8 disabled:!pointer-events-auto" onclick={() => close()}
					>Cancel</Button
				>
				<Button variant="default" class="h-8 bg-blue-700 text-white hover:bg-blue-600">Next</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
