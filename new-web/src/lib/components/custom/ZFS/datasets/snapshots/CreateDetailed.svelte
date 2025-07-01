<script lang="ts">
	import { createPeriodicSnapshot, createSnapshot } from '$lib/api/zfs/datasets';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { APIResponse } from '$lib/types/common';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { handleAPIError } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		pools: Zpool[];
		datasets: Dataset[];
	}

	let { open = $bindable(), pools, datasets }: Props = $props();

	let options = {
		name: '',
		pool: {
			open: false,
			value: '',
			data: pools.map((pool) => ({
				label: pool.name,
				value: pool.name
			}))
		},
		datasets: {
			open: false,
			value: '',
			data: [] as { label: string; value: string }[]
		},
		interval: {
			open: false,
			value: '0',
			data: [
				{ value: '0', label: 'None' },
				{ value: '60', label: 'Every Minute' },
				{ value: '3600', label: 'Every Hour' },
				{ value: '86400', label: 'Every Day' },
				{ value: '604800', label: 'Every Week' },
				{ value: '2419200', label: 'Every Month' },
				{ value: '29030400', label: 'Every Year' }
			]
		},
		recursive: false
	};

	let properties = $state(options);

	$effect(() => {
		if (properties.pool.value) {
			const sets = datasets
				.filter((dataset) => dataset.name.startsWith(properties.pool.value))
				.map((dataset) => ({
					label: dataset.name,
					value: dataset.name
				}));

			if (JSON.stringify(sets) !== JSON.stringify(properties.datasets.data)) {
				properties.datasets.data = sets;
			}
		}
	});

	async function create() {
		const dataset = datasets.find((dataset) => dataset.name === properties.datasets.value);
		const pool = pools.find((pool) => pool.name === properties.pool.value);

		if (dataset) {
			const interval = parseInt(properties.interval.value) || 0;
			let response: APIResponse | null = null;

			if (interval === 0) {
				response = await createSnapshot(dataset, properties.name, properties.recursive);
			} else if (interval > 0) {
				response = await createPeriodicSnapshot(
					dataset,
					properties.name,
					properties.recursive,
					interval
				);
			}

			if (response?.error) {
				handleAPIError(response);
				toast.error('Failed to create snapshot', {
					position: 'bottom-center'
				});
				return;
			} else {
				toast.success(`Snapshot ${pool?.name}@${properties.name} created`, {
					position: 'bottom-center'
				});

				properties = options;
				open = false;
			}
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<div class="flex items-center justify-between">
			<Dialog.Header class="flex-1">
				<Dialog.Title>
					<div class="flex items-center">
						<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-2 h-6 w-6" />
						<span>Create Snapshot</span>
					</div>
				</Dialog.Title>
			</Dialog.Header>

			<div class="flex items-center gap-0.5">
				<Button
					size="sm"
					variant="ghost"
					class="h-8"
					title={'Reset'}
					onclick={() => {
						properties = options;
					}}
				>
					<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Reset'}</span>
				</Button>
				<Button
					size="sm"
					variant="ghost"
					class="h-8"
					title={'Close'}
					onclick={() => {
						properties = options;
						open = false;
					}}
				>
					<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Close'}</span>
				</Button>
			</div>
		</div>

		<CustomValueInput
			label={`${'Name'} | ${'Prefix'}`}
			placeholder="after-upgrade"
			bind:value={properties.name}
			classes="flex-1 space-y-1"
		/>

		<div class="flex gap-4">
			<CustomComboBox
				bind:open={properties.pool.open}
				label="Pool"
				bind:value={properties.pool.value}
				data={properties.pool.data}
				classes="flex-1 space-y-1"
				placeholder="Select a pool"
			></CustomComboBox>

			<CustomComboBox
				bind:open={properties.datasets.open}
				label="Dataset"
				bind:value={properties.datasets.value}
				data={properties.datasets.data}
				classes="flex-1 space-y-1"
				placeholder="Select a dataset"
			></CustomComboBox>
		</div>

		<div class="flex-1 space-y-1">
			<CustomComboBox
				bind:open={properties.interval.open}
				label="Interval"
				bind:value={properties.interval.value}
				data={properties.interval.data}
				classes="flex-1 space-y-1"
				placeholder="Select an interval"
				width="w-3/4"
			></CustomComboBox>
		</div>

		<CustomCheckbox
			label="Recursive"
			bind:checked={properties.recursive}
			classes="flex items-center gap-2"
		></CustomCheckbox>

		<Dialog.Footer>
			<Button
				size="sm"
				onclick={() => {
					create();
				}}>Create</Button
			>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
