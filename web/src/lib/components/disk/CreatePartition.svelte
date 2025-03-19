<script lang="ts">
	import { createPartitions } from '$lib/api/disk/disk';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Slider } from '$lib/components/ui/slider/index.js';
	import * as Table from '$lib/components/ui/table';
	import type { Disk } from '$lib/types/disk/disk';
	import { handleAPIError } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import humanFormat from 'human-format';
	import { tick } from 'svelte';
	import toast from 'svelte-french-toast';

	interface Data {
		open: boolean;
		disk: Disk | null;
		onCancel: () => void;
	}

	let { open, disk, onCancel }: Data = $props();

	let newPartitions: { name: string; size: number }[] = $state([]);
	let remainingSpace = $state(0);
	let currentPartition = $state(0);

	$effect(() => {
		if (disk) {
			remainingSpace = calculateRemainingSpace(disk);
		}
	});

	function removePartition(index: number) {
		const removedPartition = newPartitions.splice(index, 1)[0];
		remainingSpace += removedPartition.size;
	}

	async function savePartitions() {
		if (disk) {
			const sizes = newPartitions.map((partition) => Math.floor(partition.size));
			const result = await createPartitions(disk.Device, sizes);
			if (result.status === 'success1') {
				let successMessage = '';
				if (sizes.length === 1) {
					successMessage = `${capitalizeFirstLetter(getTranslation('disk.partition', 'Partition'))}`;
				} else {
					successMessage = `${capitalizeFirstLetter(getTranslation('disk.partitions', 'Partitions'))}`;
				}

				successMessage += ` ${getTranslation('common.created', 'created')}`;

				toast.success(successMessage);
			} else {
				handleAPIError(result);
				let errorMessage =
					capitalizeFirstLetter(getTranslation('common.error', 'Error')) +
					getTranslation('common.creating', 'creating');

				if (sizes.length === 1) {
					errorMessage = `${capitalizeFirstLetter(getTranslation('disk.partition', 'Partition'))}`;
				} else {
					errorMessage = `${capitalizeFirstLetter(getTranslation('disk.partitions', 'Partitions'))}`;
				}
			}

			newPartitions = [];
		}
		onCancel();
	}

	async function addPartition() {
		if (currentPartition > 0) {
			newPartitions.push({
				name: `New Partition ${newPartitions.length + 1}`,
				size: currentPartition
			});
			remainingSpace -= currentPartition;
			currentPartition = 0;

			await tick();

			const table = document.getElementById('table-body');
			if (table) {
				table.scroll({
					top: table.scrollHeight,
					behavior: 'smooth'
				});
			}
		}
	}

	function close() {
		newPartitions = [];
		remainingSpace = 0;
		currentPartition = 0;
		onCancel();
	}

	function calculateRemainingSpace(disk: Disk) {
		if (!disk) return 0;
		const usedSpace =
			disk.Partitions && disk.Partitions.length > 0
				? disk.Partitions.reduce((total, partition) => total + partition.size, 0)
				: 0;

		let actual = disk.Size - usedSpace;

		if (actual > 500 * 1024 * 1024) {
			actual = actual - 500 * 1024 * 1024;
		}

		return actual;
	}
</script>

<Dialog.Root bind:open onOutsideClick={(e) => close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-hidden p-0 lg:max-w-3xl"
	>
		<div class="flex items-center justify-between p-4">
			<Dialog.Header class="p-0">
				<Dialog.Title>Create Partitions</Dialog.Title>
				<Dialog.Description></Dialog.Description>
			</Dialog.Header>

			<Dialog.Close
				class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
				onclick={() => close()}
			>
				<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
			</Dialog.Close>
		</div>

		<div class="max-h-[300px] overflow-y-auto px-4" id="table-body">
			<Table.Root>
				<Table.Header class="bg-background sticky top-0 z-10">
					<Table.Row>
						<Table.Head class="w-[200px]">Name</Table.Head>
						<Table.Head class="w-[150px] text-right">Size</Table.Head>
						<Table.Head class="w-[150px] text-right">Usage</Table.Head>
						<Table.Head class="w-[100px] text-right">Actions</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#if disk && disk.Partitions && disk.Partitions.length > 0}
						{#each disk.Partitions as partition}
							<Table.Row>
								<Table.Cell>{partition.name}</Table.Cell>
								<Table.Cell class="text-right">{humanFormat(partition.size)}</Table.Cell>
								<Table.Cell class="text-right">{partition.usage}</Table.Cell>
								<Table.Cell class="text-right">
									<span class="text-muted-foreground text-xs italic">Existing</span>
								</Table.Cell>
							</Table.Row>
						{/each}
					{/if}

					{#if newPartitions.length > 0}
						{#each newPartitions as partition, index}
							<Table.Row>
								<Table.Cell>{partition.name}</Table.Cell>
								<Table.Cell class="text-right">{humanFormat(partition.size)}</Table.Cell>
								<Table.Cell class="text-right">-</Table.Cell>
								<Table.Cell class="text-right">
									<Button variant="ghost" class="h-8" on:click={() => removePartition(index)}>
										<Icon icon="gg:trash" class="h-4 w-4" />
									</Button>
								</Table.Cell>
							</Table.Row>
						{/each}
					{/if}

					{#if (!disk || !disk.Partitions || disk.Partitions.length === 0) && newPartitions.length === 0}
						<Table.Row>
							<Table.Cell colspan={4} class="text-muted-foreground h-20 text-center">
								No partitions created yet
							</Table.Cell>
						</Table.Row>
					{/if}
				</Table.Body>
			</Table.Root>
		</div>

		<div class="space-y-2 border-t px-6 py-4">
			<div class="flex items-center gap-6">
				<div class="flex-1">
					<Slider
						value={[currentPartition]}
						max={remainingSpace}
						step={0.1}
						onValueChange={(e) => {
							currentPartition = e[0];
						}}
					/>
				</div>

				<div class={remainingSpace > 0 ? '' : 'cursor-not-allowed'}>
					<Button
						variant="outline"
						class="h-8 whitespace-nowrap"
						on:click={addPartition}
						disabled={currentPartition <= 0}
					>
						{#if remainingSpace > 0}
							Add Partition
						{:else}
							No space left
						{/if}
					</Button>
				</div>
			</div>
			<div class="flex justify-end">
				<span class="text-muted-foreground text-sm">
					Size: {humanFormat(currentPartition)}
				</span>
			</div>
		</div>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-6 py-4">
			<div class="flex gap-2">
				<div class="text-muted-foreground mt-2 text-sm">
					Remaining space: {humanFormat(remainingSpace)}
				</div>
				<Button variant="outline" class="h-8" on:click={() => close()}>Cancel</Button>
				{#if newPartitions.length > 0}
					<Button variant="outline" class="h-8" on:click={savePartitions}>Save Partitions</Button>
				{/if}
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
