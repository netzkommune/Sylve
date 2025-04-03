<script lang="ts">
	import { listDisks } from '$lib/api/disk/disk';
	import { createPool, getPools } from '$lib/api/zfs/pool';
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
	import type { Disk, Partition } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { simplifyDisks, zpoolUseableDisks } from '$lib/utils/disk';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { flip } from 'svelte/animate';
	import { fade, slide } from 'svelte/transition';

	import { draggable, dropzone } from '$lib/utils/dnd';
	import humanFormat from 'human-format';
	import { untrack } from 'svelte';

	interface Data {
		disks: Disk[];
		pools: Zpool[];
	}

	interface VdevContainer {
		id: string;
		disks: Disk[];
		partitions: Partition[];
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

	let raidTypes = $state([
		{ value: 'stripe', label: 'Stripe', available: true },
		{ value: 'mirror', label: 'Mirror', available: false },
		{ value: 'raidz1', label: 'RAIDZ1', available: false },
		{ value: 'raidz2', label: 'RAIDZ2', available: false },
		{ value: 'raidz3', label: 'RAIDZ3', available: false }
	]);

	let modal = $state({
		open: false,
		name: 'test',
		vdevCount: 1,
		vdevContainers: [] as VdevContainer[],
		raidType: 'stripe',
		close: () => {
			modal.open = false;
			modal.vdevCount = 1;
			modal.vdevContainers = [];
		}
	});

	let disks = $derived($results[0].data as Disk[]);
	let pools = $results[1].data as Zpool[];
	let useableDisks = $derived(zpoolUseableDisks(disks, pools));
	let useablePartitions = $derived.by(() => {
		return Array.from(
			new Map(
				useableDisks
					.flatMap((disk) => disk.Partitions)
					.map((partition) => [partition.uuid, partition])
			).values()
		);
	});

	$effect(() => {
		modal.vdevCount = Math.max(1, Math.min(128, modal.vdevCount));
		untrack(() => {
			setRedundancyAvailability();
		});
	});

	let previousVdevCount = $state(modal.vdevCount);

	$effect(() => {
		if (modal.vdevCount < previousVdevCount) {
			const removedContainers = modal.vdevContainers.slice(modal.vdevCount);
			removedContainers.forEach((container) => {
				container.disks.forEach((disk) => {
					if (!useableDisks.some((ud) => ud.UUID === disk.UUID)) {
						useableDisks = [...useableDisks, disk];
					}
				});
				container.partitions.forEach((partition) => {
					if (!useableDisks.some((ud) => ud.Partitions.some((p) => p.name === partition.name))) {
						const parentDisk = disks.find((d) =>
							d.Partitions.some((p) => p.name === partition.name)
						);
						if (parentDisk) {
							useableDisks = [...useableDisks, { ...parentDisk }];
						}
					}
				});
			});
			modal.vdevContainers = modal.vdevContainers.slice(0, modal.vdevCount);
		}
		previousVdevCount = modal.vdevCount;
	});

	function setRedundancyAvailability() {
		const vdevLengths = modal.vdevContainers.map(
			(vdev) => vdev.disks.length + vdev.partitions.length
		);

		raidTypes = raidTypes.map((type) => {
			switch (type.value) {
				case 'stripe':
					return { ...type, available: true };
				case 'mirror':
					const allMirrors = vdevLengths.every((length) => length === 2) && vdevLengths.length > 0;
					return { ...type, available: allMirrors };
				case 'raidz1':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 3) && vdevLengths.length > 0
					};
				case 'raidz2':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 4) && vdevLengths.length > 0
					};
				case 'raidz3':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 5) && vdevLengths.length > 0
					};
				default:
					return type;
			}
		});

		if (!raidTypes.find((rt) => rt.value === modal.raidType)?.available) {
			modal.raidType = raidTypes.find((rt) => rt.available)?.value || 'stripe';
		}
	}

	function isDiskInVdev(diskId: string | undefined | string[]): boolean {
		if (!diskId) return false;

		if (typeof diskId === 'string') {
			return modal.vdevContainers.some((vdev) => {
				return vdev.disks.some((disk) => disk.UUID === diskId);
			});
		}

		if (Array.isArray(diskId)) {
			return modal.vdevContainers.some((vdev) => {
				return vdev.partitions.some((partition) => diskId.includes(partition.name));
			});
		}

		return false;
	}

	function handleDropToVdev(containerId: number, event: DragEvent) {
		const diskId = event.dataTransfer?.getData('application/disk');

		if (!modal.vdevContainers[containerId]) {
			modal.vdevContainers[containerId] = {
				id: `vdev-${containerId}`,
				disks: [],
				partitions: []
			};
		}

		const disk = disks.find((d) => d.UUID === diskId);

		if (disk) {
			const existingDisk = modal.vdevContainers[containerId].disks.find(
				(d) => d.UUID === disk.UUID
			);
			if (!existingDisk) {
				modal.vdevContainers[containerId].disks.push(disk);
				useableDisks = useableDisks.filter((ud) => ud.UUID !== disk.UUID);
			}
		}

		if (!disk) {
			const diskContainingPartition = disks.find((d) =>
				d.Partitions.some((p) => p.name === diskId)
			);

			if (diskContainingPartition) {
				const partition = diskContainingPartition.Partitions.find((p) => p.name === diskId);
				if (partition) {
					const existingPartition = modal.vdevContainers[containerId].partitions.find(
						(p) => p.name === partition.name
					);
					if (!existingPartition) {
						modal.vdevContainers[containerId].partitions.push(partition);
						useableDisks = useableDisks.filter(
							(ud) => !ud.Partitions.some((p) => p.name === partition.name)
						);
					}
				}
			}
		}

		setRedundancyAvailability();
		// console.log(modal.vdevContainers);
	}

	function vdevContains(id: number): boolean {
		const vdev = modal.vdevContainers[id];
		if (!vdev) return false;

		return vdev.disks.length > 0 || vdev.partitions.length > 0;
	}

	function removeFromVdev(id: number, diskId: string) {
		const vdev = modal.vdevContainers[id];
		if (!vdev) return;

		const diskIndex = vdev.disks.findIndex((d) => d.UUID === diskId);
		if (diskIndex !== -1) {
			const removedDisk = vdev.disks.splice(diskIndex, 1)[0];
			if (!useableDisks.some((ud) => ud.UUID === removedDisk.UUID)) {
				useableDisks = [...useableDisks, removedDisk];
			}
		}

		const partitionIndex = vdev.partitions.findIndex((p) => p.name === diskId);
		if (partitionIndex !== -1) {
			const removedPartition = vdev.partitions.splice(partitionIndex, 1)[0];
			const parentDisk = disks.find((d) =>
				d.Partitions.some((p) => p.name === removedPartition.name)
			);
			if (
				parentDisk &&
				!useableDisks.some((ud) => ud.Partitions.some((p) => p.name === removedPartition.name))
			) {
				useableDisks = [...useableDisks, { ...parentDisk }];
			}
		}

		setRedundancyAvailability();
	}

	function getVdevErrors(id: number): string {
		const vdev = modal.vdevContainers[id];
		const disks = vdev?.disks || [];
		const partitions = vdev?.partitions || [];
		const diskSizes = disks.map((disk) => disk.Size);
		const partSizes = partitions.map((partition) => partition.size);
		const allSizes = [...diskSizes, ...partSizes];

		const diskTypes = disks.map((disk) => disk.Type);
		for (let i = 0; i < diskTypes.length - 1; i++) {
			if (diskTypes[i] !== diskTypes[i + 1]) {
				return 'Disks within a VDEV should ideally be the same type';
			}
		}

		const partitionTypes = partitions.map((partition) => {
			const disk = useableDisks.find((d) => d.Partitions.some((p) => p.name === partition.name));
			return disk ? disk.Type : null;
		});

		for (let i = 0; i < partitionTypes.length - 1; i++) {
			if (partitionTypes[i] !== partitionTypes[i + 1]) {
				return 'Partitions within a VDEV should ideally be the same drive type';
			}
		}

		for (let i = 0; i < allSizes.length - 1; i++) {
			if (allSizes[i] !== allSizes[i + 1]) {
				if (partSizes.length === 0) {
					return 'Disks within a VDEV should ideally be the same size';
				} else if (diskSizes.length === 0) {
					return 'Partitions within a VDEV should ideally be the same size';
				} else {
					return 'Disks/Partitions within a VDEV should ideally be the same size';
				}
			}
		}

		return '';
	}

	async function makePool() {
		const create = {
			name: modal.name,
			vdevs: modal.vdevContainers.map((vdev) => ({
				name: vdev.id,
				devices: [
					...vdev.disks.map((disk) => disk.Device),
					...vdev.partitions.map((partition) => `/dev/${partition.name}`)
				]
			})),
			raidType:
				modal.raidType === 'stripe'
					? undefined
					: (modal.raidType as 'mirror' | 'raidz2' | 'raidz3' | 'raidz' | undefined),
			properties: {
				ashift: '12'
			},
			createForce: true
		};

		console.log(create);

		console.log(await createPool(create));
	}
</script>

{#snippet diskContainer(type: string)}
	<div id="{type.toLowerCase()}-container">
		<Label>{type}</Label>
		<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
			<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
				<div class="flex min-h-[80px] items-center justify-center gap-4">
					{#each useableDisks.filter((disk) => disk.Type === type && disk.Partitions.length === 0 && !isDiskInVdev(disk.UUID)) as disk (disk.UUID)}
						<div class="text-center" animate:flip={{ duration: 300 }}>
							<div class="cursor-move" use:draggable={disk.UUID ?? ''}>
								{#if type === 'HDD'}
									<Icon icon="mdi:harddisk" class="h-11 w-11 text-green-500" />
								{:else if type === 'SSD'}
									<Icon icon="icon-park-outline:ssd" class="h-11 w-11 text-blue-500" />
								{:else if type === 'NVMe'}
									<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
								{/if}
							</div>
							<div class="max-w-[64px] truncate text-xs">
								{disk.Device.replaceAll('/dev/', '')}
							</div>
							<div class="text-xs text-neutral-400">
								{humanFormat(disk.Size)}
							</div>
						</div>
					{/each}

					{#if useableDisks.filter((disk) => disk.Type === type).length === 0 || useableDisks.filter((disk) => disk.Type === type && disk.Partitions.length === 0 && !isDiskInVdev(disk.UUID)).length === 0}
						<div class="flex h-16 w-full items-center justify-center text-neutral-400">
							No available disks
						</div>
					{/if}
				</div>
			</ScrollArea>
		</div>
	</div>
{/snippet}

{#snippet partitionsContainer()}
	<div id="partitions-container">
		<Label>Partitions</Label>
		<div class="mt-1 rounded-lg bg-neutral-200 p-4 dark:bg-neutral-950">
			<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
				<div class="flex min-h-[80px] items-center justify-center gap-4">
					{#each useablePartitions.filter((partition) => !modal.vdevContainers
								.flatMap((vdev) => vdev.partitions)
								.some((p) => p.name === partition.name)) as partition (partition.name)}
						<div class="text-center" animate:flip={{ duration: 100 }}>
							<div class="cursor-move" use:draggable={partition.name}>
								<Icon
									icon="ant-design:partition-outlined"
									class="h-11 w-11 rotate-90 text-blue-500"
								/>
							</div>
							<div class="max-w-[64px] truncate text-xs">
								{partition.name}
							</div>
							<div class="text-xs text-neutral-400">
								{humanFormat(partition.size)}
							</div>
						</div>
					{/each}

					{#if useablePartitions.length === 0 || useablePartitions.filter((partition) => !modal.vdevContainers
									.flatMap((vdev) => vdev.partitions)
									.some((p) => p.name === partition.name)).length === 0}
						<div class="flex h-16 w-full items-center justify-center text-neutral-400">
							No available partitions
						</div>
					{/if}
				</div>
			</ScrollArea>
		</div>
	</div>
{/snippet}

{#snippet vdevContainer(id: number)}
	{#each modal.vdevContainers[id]?.disks || [] as disk (disk.UUID)}
		<div animate:flip={{ duration: 300 }} class="relative">
			{#if disk.Type === 'HDD'}
				<Icon icon="mdi:harddisk" class="h-11 w-11 text-green-500" />
			{:else if disk.Type === 'SSD'}
				<Icon icon="icon-park-outline:ssd" class="h-11 w-11 text-blue-500" />
			{:else if disk.Type === 'NVMe'}
				<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
			{/if}

			<div class="max-w-[48px] truncate text-center text-xs">
				{disk.Device.split('/').pop()}
			</div>

			<button
				class="absolute -right-1 -top-1 rounded-full bg-red-500 p-0.5 text-white hover:bg-red-600"
				onclick={() => removeFromVdev(id, disk.UUID as string)}
			>
				<Icon icon="mdi:close" class="h-3 w-3" />
			</button>
		</div>
	{/each}

	{#each modal.vdevContainers[id]?.partitions || [] as partition (partition.name)}
		<div animate:flip={{ duration: 300 }} class="relative">
			<Icon icon="ant-design:partition-outlined" class="h-11 w-11 rotate-90 text-blue-500" />

			<div class="max-w-[48px] truncate text-center text-xs">
				{partition.name.split('/').pop()}
			</div>

			<button
				class="absolute -right-1 -top-1 rounded-full bg-red-500 p-0.5 text-white hover:bg-red-600"
				onclick={() => removeFromVdev(id, partition.name)}
			>
				<Icon icon="mdi:close" class="h-3 w-3" />
			</button>
		</div>
	{/each}
{/snippet}

{#snippet vdevErrors(id: number)}
	{#if getVdevErrors(id) !== ''}
		<div class="absolute right-1 top-1 z-50 cursor-pointer text-yellow-700 hover:text-yellow-600">
			<Tooltip.Root>
				<Tooltip.Trigger><Icon icon="carbon:warning-filled" class="h-5 w-5" /></Tooltip.Trigger>
				<Tooltip.Content>
					<p>
						{@html getVdevErrors(id)}
					</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</div>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			on:click={() => (modal.open = !modal.open)}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
</div>

<Dialog.Root bind:open={modal.open} onOutsideClick={() => modal.close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
	>
		<div class="flex items-center justify-between px-4 py-3">
			<Dialog.Header class="p-0">
				<Dialog.Title>Create ZFS Pool</Dialog.Title>
			</Dialog.Header>

			<Dialog.Close
				class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
				onclick={() => modal.close()}
			>
				<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
			</Dialog.Close>
		</div>
		<Tabs.Root value="tab-devices" class="w-full overflow-hidden">
			<Tabs.List class="grid w-full grid-cols-2 p-0 px-4">
				<Tabs.Trigger value="tab-devices" class="border-b">Devices</Tabs.Trigger>
				<Tabs.Trigger value="tab-options" class="border-b">Options</Tabs.Trigger>
			</Tabs.List>
			<Tabs.Content class="mt-0" value="tab-devices">
				<Card.Root class="border-none pb-4">
					<Card.Content class="flex gap-4 p-4 !pb-0">
						<div class="flex-1 space-y-1">
							<Label for="name">Name</Label>
							<Input type="text" id="name" placeholder="name" bind:value={modal.name} />
						</div>
						<div class="flex-1 space-y-1">
							<Label for="vdev_count">Virtual Devices</Label>
							<Input
								type="number"
								id="vdev_count"
								placeholder="1"
								min={1}
								bind:value={modal.vdevCount}
							/>
						</div>
					</Card.Content>

					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div id="vdev-containers">
							<Label>VDEVs</Label>
							<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
								<div
									class="border-primary-foreground bg-primary-foreground mt-1 flex w-full items-center justify-center gap-7 overflow-hidden rounded-lg border-y p-4 pr-4"
								>
									{#each Array(modal.vdevCount) as _, i}
										<div class="relative flex flex-col">
											{@render vdevErrors(i)}

											<div
												class={`relative h-28 w-48 flex-shrink-0 overflow-auto rounded-lg bg-neutral-200 p-2 dark:bg-neutral-950 ${getVdevErrors(i) ? 'border border-yellow-700 ' : ''}`}
												use:dropzone={{
													on_dropzone: (_: unknown, event: DragEvent) => handleDropToVdev(i, event),
													dragover_class: 'droppable'
												}}
											>
												{#if !vdevContains(i)}
													<div
														class="flex h-full flex-col items-center justify-center gap-2 text-neutral-500"
													>
														<span>Drop disks here</span>
														<span class="dark:text-muted text-neutral-500">{i + 1}</span>
													</div>
												{:else}
													<div class="flex h-full flex-wrap items-center justify-center gap-2">
														{@render vdevContainer(i)}
													</div>
												{/if}
											</div>
										</div>
									{/each}
								</div></ScrollArea
							>
						</div>
					</Card.Content>

					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div id="disk-containers">
							<Label>Disks</Label>
							<div
								class="border-primary-foreground bg-primary-foreground mt-1 grid grid-cols-4 gap-6 overflow-hidden border-y p-4"
							>
								{@render diskContainer('HDD')}
								{@render diskContainer('SSD')}
								{@render diskContainer('NVMe')}
								{@render partitionsContainer()}
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tabs.Content>

			<Tabs.Content class="mt-0" value="tab-options">
				<Card.Root class="min-h-[20vh] border-none pb-6">
					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div transition:slide class="grid grid-cols-1 gap-4 md:grid-cols-2">
							<div class="h-full space-y-1">
								<Label class="w-24 whitespace-nowrap text-sm" for="raid">Redundancy</Label>
								<Select.Root
									selected={{
										label: raidTypes.find((rt) => rt.value === modal.raidType)?.label,
										value: raidTypes.find((rt) => rt.value === modal.raidType)?.value
									}}
									onSelectedChange={(value) => {
										modal.raidType = value?.value as string;
									}}
								>
									<Select.Trigger class="w-full">
										<Select.Value placeholder="Select Redundancy" />
									</Select.Trigger>
									<Select.Content class="max-h-36 overflow-y-auto">
										<Select.Group>
											{#each raidTypes as raidType}
												{#if raidType.available}
													<Select.Item value={raidType.value} label={raidType.label}
														>{raidType.label}</Select.Item
													>
												{/if}
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tabs.Content>
		</Tabs.Root>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-4 py-3">
			<div class="flex gap-2">
				<Button
					variant="outline"
					class="h-8 disabled:!pointer-events-auto"
					onclick={() => modal.close()}>Cancel</Button
				>
				<Button
					variant="default"
					class="h-8 bg-blue-700 text-white hover:bg-blue-600"
					onclick={() => makePool()}>Next</Button
				>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
