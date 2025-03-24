<script lang="ts">
	import { listDisks } from '$lib/api/disk/disk';
	import { getPools } from '$lib/api/zfs/pool';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Disk } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { simplifyDisks } from '$lib/utils/disk';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

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

	$effect(() => {
		console.log('unused disks', useableDisks);
	});

	let open: boolean = $state(false);
	let name: string = $state('');
	let vdevCount: number = $state(1);
	let createEnabled: boolean = $state(false);

	$effect(() => {
		vdevCount = Math.max(1, Math.min(128, vdevCount));
	});

	function close() {
		open = false;
		name = '';
		vdevCount = 1;
	}
</script>

<Button on:click={() => (open = !open)}>Toggle Dialog</Button>

<Dialog.Root bind:open onOutsideClick={() => close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-3xl"
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

		<div class="flex items-center gap-4 p-4">
			<div class="flex-1">
				<Label for="name">Name</Label>
				<Input type="text" id="name" placeholder="tank" bind:value={name} />
			</div>
			<div class="flex-1">
				<Label for="vdev_count">Virtual Devices</Label>
				<Input type="number" id="vdev_count" placeholder="1" min={1} bind:value={vdevCount} />
			</div>
		</div>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-6 py-4">
			<div class="flex gap-2">
				<Button variant="outline" class="h-8 disabled:!pointer-events-auto" on:click={() => close()}
					>Cancel</Button
				>
				<Button
					variant="outline"
					class="h-8 disabled:!pointer-events-auto"
					on:click={() => close()}
					disabled={createEnabled !== true}>Create</Button
				>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
