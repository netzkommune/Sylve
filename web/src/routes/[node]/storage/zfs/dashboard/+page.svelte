<script lang="ts">
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import BarChart from '$lib/components/custom/BarChart.svelte';
	import PieChart from '$lib/components/custom/PieChart.svelte';
	import * as Card from '$lib/components/ui/card';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { updateCache } from '$lib/utils/http';
	import { getDatasetCompressionHist, getPoolUsagePieData } from '$lib/utils/zfs/pool';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import humanFormat from 'human-format';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
	}

	type CardType = 'pools' | 'datasets' | 'file_systems' | 'volumes' | 'snapshots';

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.pools,
			onSuccess: (data: Zpool[]) => {
				updateCache('pools', data);
			}
		},
		{
			queryKey: ['datasetList'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: data.datasets,
			onSuccess: (data: Dataset[]) => {
				updateCache('datasets', data);
			}
		}
	]);

	let pools: Zpool[] = $derived($results[0].data as Zpool[]);
	let datasets: Dataset[] = $derived($results[1].data as Dataset[]);

	let filesystems: Dataset[] = $derived.by(() => {
		return datasets.filter((dataset) => dataset.type === 'filesystem');
	});

	let volumes: Dataset[] = $derived.by(() => {
		return datasets.filter((dataset) => dataset.type === 'volume');
	});

	let snapshots: Dataset[] = $derived.by(() => {
		return datasets.filter((dataset) => dataset.type === 'snapshot');
	});

	const counts = $derived({
		pools: pools.length,
		datasets: datasets.length,
		file_systems: filesystems.length,
		volumes: volumes.length,
		snapshots: snapshots.length
	});

	function getCardDetails(type: string) {
		switch (type) {
			case 'pools':
				return {
					title: 'Pools',
					icon: 'bi:hdd-stack-fill'
				};
			case 'datasets':
				return {
					title: 'Datasets',
					icon: 'material-symbols:dataset'
				};
			case 'file_systems':
				return {
					title: 'Filesystems',
					icon: 'eos-icons:file-system'
				};
			case 'volumes':
				return {
					title: 'Volumes',
					icon: 'carbon:volume-block-storage'
				};
			case 'snapshots':
				return {
					title: 'Snapshots',
					icon: 'carbon:ibm-cloud-vpc-block-storage-snapshots'
				};
			default:
				return {
					title: '',
					icon: ''
				};
		}
	}

	let comboBoxes = $state({
		poolUsage: {
			open: false,
			value: pools[0]?.name || '',
			data: pools.map((pool) => ({
				value: pool.name,
				label: pool.name
			}))
		},
		datasetCompression: {
			open: false,
			value: pools[0]?.name || '',
			data: pools.map((pool) => ({
				value: pool.name,
				label: pool.name
			}))
		}
	});

	let pieCharts = $derived.by(() => {
		return {
			poolUsage: {
				data: getPoolUsagePieData(pools, comboBoxes.poolUsage.value)
			}
		};
	});

	let histograms = $derived.by(() => {
		return {
			compression: {
				data: getDatasetCompressionHist(comboBoxes.datasetCompression.value, datasets)
			}
		};
	});

	// $inspect(histograms);
</script>

{#snippet card(type: string)}
	<Card.Root>
		<Card.Header>
			<Card.Title class="mb-[-10px]">
				<div class="flex items-center">
					<Icon icon={getCardDetails(type).icon} class="mr-2" />
					<span class="font-normal">{getCardDetails(type).title}</span>
				</div>
			</Card.Title>
		</Card.Header>
		<Card.Content class="mb-[-10px]">
			<p class="text-xl font-semibold">
				{counts[type as CardType]}
			</p>
		</Card.Content>
	</Card.Root>
{/snippet}

<div class="p-4">
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3 lg:grid-cols-4">
		{#each ['pools', 'datasets', 'file_systems', 'volumes', 'snapshots'] as type}
			<div>
				{@render card(type)}
			</div>
		{/each}
	</div>

	<div class="flex flex-wrap gap-4">
		{#if pools.length > 0}
			<div
				class="mt-3 flex h-[310px] min-h-[200px] w-[300px] min-w-[280px] resize flex-col overflow-auto"
			>
				<Card.Root class="flex flex-1 flex-col ">
					<Card.Header>
						<Card.Title class="mb-[-100px]">
							<div class="flex w-full items-center justify-between">
								<div class="flex items-center">
									<Icon icon="mdi:data-usage" class="mr-2" />
									<span class="text-sm font-bold md:text-lg xl:text-xl">Pool Usage</span>
								</div>
								<CustomComboBox
									bind:open={comboBoxes.poolUsage.open}
									label=""
									bind:value={comboBoxes.poolUsage.value}
									data={comboBoxes.poolUsage.data}
									classes=""
									placeholder="Select a pool"
									width="w-48"
									disallowEmpty={true}
								/>
							</div>
						</Card.Title>
					</Card.Header>

					<Card.Content class="flex-1 overflow-hidden">
						<div class="mt-4 flex h-full items-center justify-center">
							<PieChart
								containerClass="h-full w-full rounded"
								data={pieCharts.poolUsage.data}
								formatter={'size-formatter'}
							/>
						</div>
					</Card.Content>
				</Card.Root>
			</div>

			<div
				class="mt-3 flex h-[310px] min-h-[270px] w-[722px] min-w-[425px] resize flex-col overflow-auto"
			>
				<Card.Root class="flex flex-1 flex-col ">
					<Card.Header>
						<Card.Title class="mb-[-100px]">
							<div class="flex w-full items-center justify-between">
								<div class="flex items-center">
									<Icon icon="mdi:data-usage" class="mr-2" />
									<span class="text-sm font-bold md:text-lg xl:text-xl">Dataset Compression</span>
								</div>
								<CustomComboBox
									bind:open={comboBoxes.datasetCompression.open}
									label=""
									bind:value={comboBoxes.datasetCompression.value}
									data={comboBoxes.datasetCompression.data}
									classes=""
									placeholder="Select a pool"
									width="w-48"
									disallowEmpty={true}
								/>
							</div>
						</Card.Title>
					</Card.Header>

					<Card.Content class="flex-1 overflow-hidden">
						<div class="mt-4 flex h-full items-center justify-center">
							{#if histograms.compression.data.length === 0}
								<p class="text-sm font-semibold text-gray-500">No data available</p>
							{:else}
								<BarChart
									containerClass="h-full w-full rounded"
									data={histograms.compression.data}
									colors={{
										baseline: 'hsla(60, 50%, 50%, 0.5)',
										value: 'hsla(120, 50%, 50%, 0.5)'
									}}
								/>
							{/if}
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		{/if}
	</div>
</div>
