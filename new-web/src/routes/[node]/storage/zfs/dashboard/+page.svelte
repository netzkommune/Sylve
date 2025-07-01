<script lang="ts">
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { getPools, getPoolStats } from '$lib/api/zfs/pool';
	import BarChart from '$lib/components/custom/Charts/Bar.svelte';
	import LineGraph from '$lib/components/custom/Charts/LineGraph.svelte';
	import PieChart from '$lib/components/custom/Charts/Pie.svelte';
	import * as Card from '$lib/components/ui/card';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import type { PoolStatPointsResponse, Zpool } from '$lib/types/zfs/pool';
	import { updateCache } from '$lib/utils/http';
	import {
		getDatasetCompressionHist,
		getPoolStatsCombined,
		getPoolUsagePieData
	} from '$lib/utils/zfs/pool';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		pools: Zpool[];
		datasets: Dataset[];
		poolStats: PoolStatPointsResponse;
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
		},
		{
			queryKey: ['poolStats'],
			queryFn: async () => {
				return await getPoolStats(1, 128);
			},
			refetchInterval: 1000,
			keepPreviousData: false,
			initialData: Array.isArray(data.poolStats) ? data.poolStats[0] : data.poolStats,
			onSuccess: (data: PoolStatPointsResponse) => {
				updateCache('poolStats', data);
			}
		}
	]);

	let pools: Zpool[] = $derived($results[0].data as Zpool[]);
	let datasets: Dataset[] = $derived($results[1].data as Dataset[]);
	let poolStats: PoolStatPointsResponse = $derived($results[2].data as PoolStatPointsResponse);

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

	let poolState = $state({
		poolUsageOpen: false,
		poolUsageValue: pools[0]?.name || ''
	});

	let comboBoxData = $derived({
		poolUsage: pools.map((pool) => ({
			value: pool.name,
			label: pool.name
		}))
	});

	let pieCharts = $derived.by(() => {
		return {
			poolUsage: {
				data: getPoolUsagePieData(pools, poolState.poolUsageValue)
			}
		};
	});

	$inspect('pool usage', comboBoxData);
	$inspect('pie charts', pieCharts);
</script>

{#snippet card(type: string)}
	<Card.Root class="gap-2 p-5">
		<Card.Header class="p-0">
			<Card.Title class="p-0">
				<div class="flex items-center">
					<Icon icon={getCardDetails(type).icon} class="mr-2" />
					<span class="font-normal">{getCardDetails(type).title}</span>
				</div>
			</Card.Title>
		</Card.Header>
		<Card.Content class="p-0 pl-1">
			<p class="text-xl font-semibold">
				{counts[type as CardType]}
			</p>
		</Card.Content>
	</Card.Root>
{/snippet}

<div class="p-4">
	<div class="grid grid-cols-1 gap-4 md:grid-cols-3 lg:grid-cols-5">
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
									bind:open={poolState.poolUsageOpen}
									label=""
									bind:value={poolState.poolUsageValue}
									data={comboBoxData.poolUsage}
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
		{/if}
	</div>
</div>
