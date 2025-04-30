<script lang="ts">
	import { deletePeriodicSnapshot } from '$lib/api/zfs/datasets';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Table from '$lib/components/ui/table';
	import type { Dataset, PeriodicSnapshot } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { dateToAgo } from '$lib/utils/time';
	import { getDatasetByGUID } from '$lib/utils/zfs/dataset/dataset';
	import Icon from '@iconify/svelte';

	interface Data {
		open: boolean;
		pools: Zpool[];
		datasets: Dataset[];
		periodicSnapshots: PeriodicSnapshot[];
	}

	let { open = $bindable(), pools, datasets, periodicSnapshots }: Data = $props();
	let shadowDeleted: number[] = $state([]);

	function close() {
		shadowDeleted = [];
		open = false;
	}

	function getDatasetName(guid: string) {
		const dataset = getDatasetByGUID(datasets, guid);
		if (dataset) {
			return dataset.name;
		}
		return '';
	}

	function intervalToString(interval: number) {
		switch (interval) {
			case 0:
				return 'None';
			case 60:
				return 'Every Minute';
			case 3600:
				return 'Every Hour';
			case 86400:
				return 'Every Day';
			case 604800:
				return 'Every Week';
			case 2419200:
				return 'Every Month';
			case 29030400:
				return 'Every Year';
			default:
				return `${interval} seconds`;
		}
	}

	async function saveJobs() {
		// console.log('Saving jobs:', shadowDeleted);
		try {
			for (const id of shadowDeleted) {
				const snapshot = periodicSnapshots.find((s) => s.id === id);
				if (snapshot) {
					console.log(await deletePeriodicSnapshot(snapshot.guid));
					shadowDeleted = shadowDeleted.filter((s) => s !== id);
				}
			}
		} catch (e) {}
	}
</script>

<Dialog.Root bind:open onOutsideClick={(e) => close()}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-hidden p-0 lg:max-w-3xl"
	>
		<div class="flex items-center justify-between p-4">
			<Dialog.Header class="p-0">
				<Dialog.Title>
					<div class="flex items-center gap-2">
						<Icon icon="material-symbols:save-clock" class="h-5 w-5" />
						<span>View Snapshot Jobs</span>
					</div>
				</Dialog.Title>
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
						<Table.Head class="w-[10px]">ID</Table.Head>
						<Table.Head class="w-[200px]">Dataset</Table.Head>
						<Table.Head class="w-[200px]">Prefix</Table.Head>
						<Table.Head class="w-[200px]">Interval</Table.Head>
						<Table.Head class="w-[200px]">Last Run</Table.Head>
						<Table.Head class="w-[200px]"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#if periodicSnapshots && periodicSnapshots.length > 0}
						{#each periodicSnapshots as snapshot, index}
							<Table.Row>
								<Table.Cell>{snapshot.id}</Table.Cell>
								<Table.Cell>{getDatasetName(snapshot.guid)}</Table.Cell>
								<Table.Cell>{snapshot.prefix}</Table.Cell>
								<Table.Cell>{intervalToString(snapshot.interval)}</Table.Cell>
								<Table.Cell title={snapshot.lastRunAt.toLocaleString()}
									>{dateToAgo(snapshot.lastRunAt)}</Table.Cell
								>

								{#if !shadowDeleted.includes(snapshot.id)}
									<Table.Cell>
										<Button
											variant="ghost"
											class="h-8"
											on:click={() => shadowDeleted.push(snapshot.id)}
										>
											<Icon icon="gg:trash" class="h-4 w-4" />
										</Button>
									</Table.Cell>
								{:else}
									<Table.Cell>
										<span class="text-muted-foreground text-xs italic">Deleted</span>
									</Table.Cell>
								{/if}
							</Table.Row>
						{/each}
					{:else}
						<Table.Row>
							<Table.Cell colspan={6} class="text-muted-foreground h-20 text-center">
								No snapshot jobs
							</Table.Cell>
						</Table.Row>
					{/if}
				</Table.Body>
			</Table.Root>
		</div>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-6 py-4">
			<div class="flex gap-2">
				<Button variant="outline" class="h-8" on:click={() => close()}>Cancel</Button>
				{#if shadowDeleted.length > 0}
					<Button variant="outline" class="h-8" on:click={saveJobs}>Save Snapshot Jobs</Button>
				{/if}
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
