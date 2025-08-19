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
	import { cronToHuman } from '$lib/utils/time';
	import Icon from '@iconify/svelte';
	import { interval } from 'date-fns';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		pools: Zpool[];
		datasets: Dataset[];
		reload?: boolean;
	}

	let { open = $bindable(), pools, datasets, reload = $bindable() }: Props = $props();

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
			type: 'none' as 'none' | 'minutes' | 'cronExpr',
			open: false,
			value: 'none',
			data: [
				{ value: 'none', label: 'None' },
				{ value: 'minutes', label: 'Simple' },
				{ value: 'cronExpr', label: 'Cron Expression' }
			],
			values: {
				cron: '',
				interval: {
					open: false,
					data: [
						{ value: '60', label: 'Every Minute' },
						{ value: '3600', label: 'Every Hour' },
						{ value: '86400', label: 'Every Day' },
						{ value: '604800', label: 'Every Week' },
						{ value: '2419200', label: 'Every Month' },
						{ value: '29030400', label: 'Every Year' }
					],
					value: ''
				}
			}
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
			const intervalType = properties.interval.value;
			let response: APIResponse | null = null;

			if (intervalType === 'none') {
				response = await createSnapshot(dataset, properties.name, properties.recursive);
			} else if (intervalType === 'minutes') {
				const intervalValue = parseInt(properties.interval.values.interval.value) || 0;
				response = await createPeriodicSnapshot(
					dataset,
					properties.name,
					properties.recursive,
					intervalValue,
					''
				);
			} else if (intervalType === 'cronExpr') {
				response = await createPeriodicSnapshot(
					dataset,
					properties.name,
					properties.recursive,
					0,
					properties.interval.values.cron
				);
			}

			reload = true;

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
	<Dialog.Content class="w-3/4 p-5">
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex justify-between">
				<div class="flex items-center">
					<Icon icon="carbon:ibm-cloud-vpc-block-storage-snapshots" class="mr-2 h-6 w-6" />
					<span>Create Snapshot</span>
				</div>
				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="link"
						class="h-4"
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
						variant="link"
						class="h-4"
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
			</Dialog.Title>
		</Dialog.Header>

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

		<div class="w-full space-y-4">
			<CustomComboBox
				bind:open={properties.interval.open}
				label="Interval"
				bind:value={properties.interval.value}
				data={properties.interval.data}
				classes="w-full space-y-1"
				placeholder="Select an interval"
				width="w-full"
			/>

			{#if properties.interval.value === 'cronExpr'}
				<CustomValueInput
					label={`
  <span class="text-sm font-medium text-gray-200">
    Cron Expression${
			cronToHuman(properties.interval.values.cron)
				? `&nbsp;<span class="text-green-300 font-semibold">(${cronToHuman(properties.interval.values.cron)})</span>`
				: ''
		}
  </span>
`}
					labelHTML={true}
					placeholder="0 0 * * *"
					bind:value={properties.interval.values.cron}
					classes="w-full space-y-1"
				/>
			{/if}

			{#if properties.interval.value === 'minutes'}
				<CustomComboBox
					bind:open={properties.interval.values.interval.open}
					label="Interval"
					bind:value={properties.interval.values.interval.value}
					data={properties.interval.values.interval.data}
					classes="w-full space-y-1"
					placeholder="Select an interval"
					width="w-full"
				/>
			{/if}
		</div>

		<CustomCheckbox
			label="Recursive"
			bind:checked={properties.recursive}
			classes="flex items-center gap-2"
		></CustomCheckbox>

		<Dialog.Footer>
			<Button
				size="sm"
				class="w-full lg:w-28"
				onclick={() => {
					create();
				}}>Create</Button
			>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
