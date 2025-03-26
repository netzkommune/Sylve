<script lang="ts">
	import { listDisks } from '$lib/api/disk/disk';
	import { getPools } from '$lib/api/zfs/pool';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Disk } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { simplifyDisks } from '$lib/utils/disk';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { flip } from 'svelte/animate';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { slide } from 'svelte/transition';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';

	import { dropzone, draggable } from '$lib/utils/dnd';

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

	let useableDisks = $derived.by(() => {
		const unusedDisks: UnusedDisk[] = [];
		for (const disk of disks) {
			if (disk.Usage === 'Unused' && disk.GPT === false) {
				unusedDisks.push({
					name: disk.Device,
					size: disk.Size,
					gpt: disk.GPT,
					type: disk.Type
				});
			}

			if (disk.Usage === 'Partitions') {
				for (const partition of disk.Partitions) {
					for (const pool of pools) {
						for (const vdev of pool.vdevs) {
							if (
								vdev.name !== `/dev/${partition.name}` &&
								vdev.name !== partition.name &&
								partition.usage === 'ZFS'
							) {
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
		}

		return unusedDisks;
	});

	let open: boolean = $state(false);
	let name: string = $state('');
	let vdevCount: number = $state(21);
	let createEnabled: boolean = $state(false);

	$effect(() => {
		vdevCount = Math.max(1, Math.min(128, vdevCount));
	});

	function close() {
		open = false;
		name = '';
		vdevCount = 1;
	}

	// Convert useable disks to a format with IDs for drag and drop
	let availableDisks = $state(
		useableDisks.map((disk) => ({
			id: disk.name,
			name: disk.name,
			size: disk.size,
			type: disk.type
		}))
	);

	// Add dummy disks for testing if needed
	if (availableDisks.length === 0) {
		availableDisks = [
			// HDD disks
			{ id: 'disk1', name: '/dev/sda', size: 500 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk2', name: '/dev/sdb', size: 1000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk3', name: '/dev/sdc', size: 2000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk4', name: '/dev/sdd', size: 4000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk5', name: '/dev/sde', size: 8000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk6', name: '/dev/sdf', size: 10000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk7', name: '/dev/sdg', size: 12000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk8', name: '/dev/sdh', size: 14000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk9', name: '/dev/sdi', size: 16000 * 1024 * 1024 * 1024, type: 'HDD' },
			{ id: 'disk10', name: '/dev/sdj', size: 18000 * 1024 * 1024 * 1024, type: 'HDD' },
			// SSD disks
			{ id: 'ssd1', name: '/dev/ssd0n1', size: 250 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd2', name: '/dev/ssd1n1', size: 500 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd3', name: '/dev/ssd2n1', size: 1000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd4', name: '/dev/ssd3n1', size: 2000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd5', name: '/dev/ssd4n1', size: 4000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd6', name: '/dev/ssd5n1', size: 8000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd7', name: '/dev/ssd6n1', size: 10000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd8', name: '/dev/ssd7n1', size: 12000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd9', name: '/dev/ssd8n1', size: 14000 * 1024 * 1024 * 1024, type: 'SSD' },
			{ id: 'ssd10', name: '/dev/ssd9n1', size: 16000 * 1024 * 1024 * 1024, type: 'SSD' },

			{ id: 'nvm1', name: '/dev/nvme0n1', size: 250 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm2', name: '/dev/nvme1n1', size: 500 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm3', name: '/dev/nvme2n1', size: 1000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm4', name: '/dev/nvme3n1', size: 2000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm5', name: '/dev/nvme4n1', size: 4000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm6', name: '/dev/nvme5n1', size: 8000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm7', name: '/dev/nvme6n1', size: 10000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm8', name: '/dev/nvme7n1', size: 12000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm9', name: '/dev/nvme8n1', size: 14000 * 1024 * 1024 * 1024, type: 'NVMe' },
			{ id: 'nvm10', name: '/dev/nvme9n1', size: 16000 * 1024 * 1024 * 1024, type: 'NVMe' }
		];
	}

	// Create containers for the disks
	let diskContainers: { id: string; name: string; size: number; type: string }[][] = $state(
		Array(vdevCount)
			.fill(null)
			.map((_, i) => [])
	);

	function handleDiskDrop(containerId: number, event: DragEvent) {
		// event.preventDefault();
		console.log('Dropped disk in container', containerId);

		const diskId = event.dataTransfer ? event.dataTransfer.getData('application/disk') : null;

		// Check if the disk is already in a container
		let foundInContainer = -1;
		let diskToMove = null;

		// Look in containers first
		for (let i = 0; i < diskContainers.length; i++) {
			const diskIndex = diskContainers[i].findIndex((d) => d.id === diskId);
			if (diskIndex !== -1) {
				diskToMove = diskContainers[i][diskIndex];
				foundInContainer = i;
				break;
			}
		}

		// Look in available disks if not found in containers
		if (foundInContainer === -1) {
			const diskIndex = availableDisks.findIndex((d) => d.id === diskId);
			if (diskIndex !== -1) {
				diskToMove = availableDisks[diskIndex];
				availableDisks = availableDisks.filter((d) => d.id !== diskId);
			}
		} else if (foundInContainer !== containerId) {
			// Remove from original container if it's a different one
			diskContainers[foundInContainer] = diskContainers[foundInContainer].filter(
				(d) => d.id !== diskId
			);
		}

		// Add to target container if we found the disk
		if (diskToMove && foundInContainer !== containerId) {
			diskContainers[containerId] = [...diskContainers[containerId], diskToMove];
		}

		// Enable create button if any disks are assigned
		createEnabled = diskContainers.some((container) => container.length > 0);
	}

	// Function to handle removing a disk from a container
	function returnDiskToPool(containerId: number, diskId: string) {
		const container = diskContainers[containerId];
		const diskIndex = container.findIndex((d) => d.id === diskId);

		if (diskIndex !== -1) {
			const disk = container[diskIndex];
			diskContainers[containerId] = container.filter((d) => d.id !== diskId);
			availableDisks = [...availableDisks, disk];

			// Update create button state
			createEnabled = diskContainers.some((container) => container.length > 0);
		}
	}

	// Update containers when vdevCount changes
	$effect(() => {
		const oldContainers = [...diskContainers];
		diskContainers = Array(vdevCount)
			.fill(null)
			.map((_, i) => {
				return oldContainers[i] || [];
			});
	});

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
			class="h-6 bg-muted-foreground/40 text-black dark:bg-muted dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
</div>

<Dialog.Root bind:open onOutsideClick={() => close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
	>
		<div class="flex items-center justify-between p-4">
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
			<Tabs.List class="grid w-full grid-cols-2">
				<Tabs.Trigger value="tab-1">Tab 1</Tabs.Trigger>
				<Tabs.Trigger value="tab-2">Tab 2</Tabs.Trigger>
			</Tabs.List>
			<Tabs.Content value="tab-1">
				<Card.Root class="pb-6">
					<Card.Content class="flex gap-6 !pb-0">
						<div class="flex-1 space-y-1">
							<Label for="name">Name</Label>
							<Input type="text" id="name" placeholder="name" bind:value={name} />
						</div>
						<div class="flex-1 space-y-1">
							<Label for="vdev_count">Virtual Devices</Label>
							<Input type="number" id="vdev_count" placeholder="1" min={1} bind:value={vdevCount} />
						</div>
					</Card.Content>
					<Card.Content class="flex flex-col gap-6 !pb-0">
						<Label for="vdev_count" class="">VDEV</Label>

						<!-- Disk Containers Section -->
						<div
							class="w-full overflow-hidden border-y border-primary-foreground bg-primary-foreground p-4"
						>
							<ScrollArea class="w-full whitespace-nowrap rounded-md " orientation="horizontal">
								<div class="flex justify-center gap-7 pr-4">
									{#each Array(vdevCount) as _, i}
										<div class="flex flex-col">
											<div
												class="h-28 w-48 flex-shrink-0 overflow-auto rounded-lg border border-neutral-300 bg-neutral-200 p-2 dark:border-neutral-800 dark:bg-neutral-950"
												use:dropzone={{
													on_dropzone: (_: unknown, event: DragEvent) => handleDiskDrop(i, event)
												}}
											>
												{#if diskContainers[i].length === 0}
													<div class="flex h-full items-center justify-center text-neutral-500">
														Drop disks here
													</div>
												{:else}
													<div class="flex flex-wrap items-center justify-center gap-2">
														{#each diskContainers[i] as disk (disk.id)}
															<div animate:flip={{ duration: 300 }} class="relative">
																{#if disk.type === 'NVMe'}
																	<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
																{:else}
																	<Icon
																		icon={disk.type === 'SSD'
																			? 'icon-park-outline:ssd'
																			: 'mdi:harddisk'}
																		class="h-12 w-12 {disk.type === 'SSD'
																			? 'text-blue-500'
																			: 'text-green-500'}"
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
											<p class="mt-2 text-center text-xs text-neutral-100">VDEV {i + 1}</p>
										</div>
									{/each}
								</div>
							</ScrollArea>
						</div>

						<Label for="vdev_count" class="">Disks</Label>
						<div
							class="grid grid-cols-3 gap-6 overflow-hidden border-y border-primary-foreground bg-primary-foreground p-4"
						>
							<div class="">
								<label class="">HDD</label>
								<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
									<ScrollArea
										class="mt-1 w-full whitespace-nowrap rounded-md"
										orientation="horizontal"
									>
										<div class="flex justify-center gap-4 pr-4">
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

											{#if availableDisks.length === 0}
												<div class="flex h-16 w-full items-center justify-center text-neutral-400">
													No available disks
												</div>
											{/if}
										</div>
									</ScrollArea>
								</div>
							</div>

							<div class="">
								<label class="">SSD</label>
								<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
									<ScrollArea
										class="mt-1 w-full whitespace-nowrap rounded-md "
										orientation="horizontal"
									>
										<div class="flex justify-center gap-4 pr-4">
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
												<div class="flex h-16 w-full items-center justify-center text-neutral-400">
													No available disks
												</div>
											{/if}
										</div>
									</ScrollArea>
								</div>
							</div>

							<div class="">
								<label class="">NVME</label>
								<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
									<ScrollArea
										class="mt-1 w-full whitespace-nowrap rounded-md "
										orientation="horizontal"
									>
										<div class="flex justify-center gap-4 pr-4">
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
												<div class="flex h-16 w-full items-center justify-center text-neutral-400">
													No available disks
												</div>
											{/if}
										</div>
									</ScrollArea>
								</div>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tabs.Content>
			<Tabs.Content value="tab-2">
				<Card.Root>
					<Card.Content>
						<div transition:slide class="grid grid-cols-1 gap-4 md:grid-cols-3">
							<div>
								<Label class="w-24 whitespace-nowrap text-sm" for="terms">RAID:</Label>
								<Select.Root portal={null}>
									<Select.Trigger class="w-full">
										<Select.Value placeholder="Select a RAID" />
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each raid as fruit}
												<Select.Item value={fruit.value} label={fruit.label}
													>{fruit.label}</Select.Item
												>
											{/each}
										</Select.Group>
									</Select.Content>
									<Select.Input name="favoriteFruit" />
								</Select.Root>
							</div>
							<div>
								<Label class="w-24 whitespace-nowrap text-sm" for="terms">Compression:</Label>
								<Select.Root portal={null}>
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
									<Select.Input name="favoriteFruit" />
								</Select.Root>
							</div>
							<div>
								<Label class="w-24 whitespace-nowrap text-sm" for="terms">ASHIFT:</Label>
								<Select.Root portal={null}>
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
									<Select.Input name="favoriteFruit" />
								</Select.Root>
							</div>
						</div>
						<div transition:slide class="mt-2 flex items-center space-x-2 md:mt-0">
							<Label
								id="terms-label"
								for="terms"
								class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
							>
								Advanced
							</Label>
							<Checkbox id="terms" bind:checked={advancedChecked} aria-labelledby="terms-label" />
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
												class="rounded px-2 py-1 text-white hover:bg-muted"
											>
												<Icon icon="ic:twotone-remove" class="h-5 w-5" />
											</button>
										{/if}
									</div>
								{/each}
							</div>
							<div transition:slide class="flex justify-end">
								<button onclick={addPair} class=" rounded px-3 py-1 text-white hover:bg-muted">
									<Icon icon="icons8:plus" class="h-6 w-6" />
								</button>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</Tabs.Content>
		</Tabs.Root>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-6 py-4">
			<div class="flex gap-2">
				<Button variant="outline" class="h-8 disabled:!pointer-events-auto" onclick={() => close()}
					>Cancel</Button
				>
				<Button variant="default" class="h-8 bg-blue-700 text-white hover:bg-blue-600">Next</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
