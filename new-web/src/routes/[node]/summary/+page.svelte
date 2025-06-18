<script lang="ts">
	import { getBasicInfo } from '$lib/api/info/basic';
	import { getCPUInfo } from '$lib/api/info/cpu';
	import { getRAMInfo, getSwapInfo } from '$lib/api/info/ram';
	import { getIODelay } from '$lib/api/zfs/pool';
	import AreaChart from '$lib/components/custom/Charts/Area.svelte';
	import LineGraph from '$lib/components/custom/LineGraph.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import type { HistoricalData } from '$lib/types/common';
	import type { BasicInfo } from '$lib/types/info/basic';
	import type { CPUInfo, CPUInfoHistorical } from '$lib/types/info/cpu';
	import type { RAMInfo } from '$lib/types/info/ram';
	import type { IODelay, IODelayHistorical } from '$lib/types/zfs/pool';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { bytesToHumanReadable, floatToNDecimals } from '$lib/utils/numbers';
	import { secondsToHoursAgo } from '$lib/utils/time';
	import { getTotalDiskUsage } from '$lib/utils/zfs';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		basicInfo: BasicInfo;
		cpuInfo: CPUInfo;
		cpuInfoHistorical: CPUInfoHistorical;
		ramInfo: RAMInfo;
		swapInfo: RAMInfo;
		ioDelay: IODelay;
		totalDiskUsage: number;
		ioDelayHistorical: IODelayHistorical;
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['basicInfo'],
			queryFn: getBasicInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.basicInfo,
			onSuccess: (data: BasicInfo) => {
				updateCache('basicInfo', data);
			}
		},
		{
			queryKey: ['cpuInfo'],
			queryFn: getCPUInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.cpuInfo,
			onSuccess: (data: CPUInfo | CPUInfoHistorical) => {
				updateCache('cpuInfo', data as CPUInfo);
			}
		},
		{
			queryKey: ['ramInfo'],
			queryFn: getRAMInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ramInfo,
			onSuccess: (data: RAMInfo) => {
				updateCache('ramInfo', data);
			}
		},
		{
			queryKey: ['swapInfo'],
			queryFn: getSwapInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.swapInfo,
			onSuccess: (data: RAMInfo) => {
				updateCache('swapInfo', data);
			}
		},
		{
			queryKey: ['ioDelay'],
			queryFn: getIODelay,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ioDelay,
			onSuccess: (data: IODelay | IODelayHistorical) => {
				updateCache('ioDelay', data as IODelay);
			}
		},
		{
			queryKey: ['totalDiskUsage'],
			queryFn: getTotalDiskUsage,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.totalDiskUsage,
			onSuccess: (data: number) => {
				updateCache('totalDiskUsage', data);
			}
		},
		{
			queryKey: ['cpuInfoHistorical'],
			queryFn: getCPUInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.cpuInfoHistorical,
			onSuccess: (data: CPUInfo | CPUInfoHistorical) => {
				updateCache('cpuInfoHistorical', data as CPUInfoHistorical);
			}
		},
		{
			queryKey: ['ioDelayHistorical'],
			queryFn: getIODelay,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ioDelayHistorical,
			onSuccess: (data: IODelay | IODelayHistorical) => {
				updateCache('ioDelayHistorical', data as IODelayHistorical);
			}
		}
	]);

	let basicInfo = $derived($results[0].data as BasicInfo);
	let cpuInfo = $derived($results[1].data as CPUInfo);
	let ramInfo = $derived($results[2].data as RAMInfo);
	let swapInfo = $derived($results[3].data as RAMInfo);
	let ioDelay = $derived($results[4].data as IODelay);
	let totalDiskUsage = $derived($results[5].data as number);
	let cpuInfoHistorical = $derived($results[6].data as CPUInfoHistorical);
	let ioDelayHistorical = $derived($results[7].data as IODelayHistorical);
	let chartElements = $derived.by(() => {
		return [
			{
				field: 'cpuUsage',
				label: getTranslation('summary.cpu_usage', 'CPU Usage'),
				color: 'var(--blue-500)',
				data: cpuInfoHistorical.map((data) => ({
					date: new Date(data.createdAt),
					value: Math.floor(data.usage)
				}))
			},
			{
				field: 'ioDelay',
				label: getTranslation('summary.io_delay', 'I/O Delay'),
				color: 'var(--red-500)',
				data: ioDelayHistorical.map((data) => ({
					date: new Date(data.createdAt),
					value: Math.floor(data.delay)
				}))
			}
		];
	});
</script>

<div class="flex h-full w-full flex-col">
	<div class="min-h-0 flex-1">
		<ScrollArea orientation="both" class="h-full w-full">
			<div class="space-y-3 p-3">
				<Card.Root class="w-full">
					<Card.Header class="-mb-3 -mt-3 p-0">
						<Card.Description class="text-md ml-3 font-normal text-blue-600 dark:text-blue-500"
							>{basicInfo.hostname} (Started {secondsToHoursAgo(
								basicInfo.uptime
							)})</Card.Description
						>
					</Card.Header>
					<Card.Content class="p-0">
						<div class="ml-3 grid grid-cols-1 gap-4 md:grid-cols-2">
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="solar:cpu-bold" class="mr-1 h-5 w-5" />
										{getTranslation('summary.cpu_usage', 'CPU Usage')}
									</p>
									<p>
										{floatToNDecimals(cpuInfo.usage, 2)}% {getTranslation('common.of', 'of')}
										{cpuInfo.logicalCores}
										{getTranslation('summary.CPU_s', 'CPU(s)')}
									</p>
								</div>
								<Progress value={cpuInfo.usage || 0} max={100} class="h-2 w-[100%] " />
							</div>
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="ri:ram-fill" class="mr-1 h-5 w-5" />
										{getTranslation('summary.ram_usage', 'RAM Usage')}
									</p>
									<p>
										{floatToNDecimals(ramInfo.usedPercent, 2)}% of {bytesToHumanReadable(
											ramInfo.total
										)}
									</p>
								</div>
								<Progress value={ramInfo.usedPercent || 0} max={100} class="h-2 w-[100%]" />
							</div>
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="bxs:server" class="mr-1 h-5 w-5" />{getTranslation(
											'summary.disk_usage',
											'Disk Usage'
										)}
									</p>
									<p>
										{floatToNDecimals(totalDiskUsage, 2)} %
									</p>
								</div>
								<Progress
									value={floatToNDecimals(totalDiskUsage, 2)}
									max={100}
									class="h-2 w-[100%]"
								/>
							</div>
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="lets-icons:time-light" class="mr-1 h-5 w-5" />
										{getTranslation('summary.io_delay', 'I/O Delay')}
									</p>
									<p>{floatToNDecimals(ioDelay.delay, 3) || 0} %</p>
								</div>
								<Progress
									value={floatToNDecimals(ioDelay.delay, 3) || 0}
									max={100}
									class="h-2 w-[100%]"
								/>
							</div>
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="ic:baseline-loop" class="mr-1 h-5 w-5" />{getTranslation(
											'summary.swap_usage',
											'Swap Usage'
										)}
									</p>
									<p>
										{floatToNDecimals(swapInfo.usedPercent, 2)}% of {bytesToHumanReadable(
											swapInfo.total
										)}
									</p>
								</div>
								<Progress value={swapInfo.usedPercent || 0} max={100} class="h-2 w-[100%]" />
							</div>
						</div>

						<Table.Root class="mt-5">
							<Table.Body>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4"
										>{getTranslation('summary.CPU_s', 'CPU(s)')}</Table.Cell
									>
									<Table.Cell class="p-1.5 px-4">{cpuInfo.logicalCores} x {cpuInfo.name}</Table.Cell
									>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4"
										>{getTranslation('summary.operating_system', 'Operating System')}</Table.Cell
									>
									<Table.Cell class="p-1.5 px-4">{basicInfo.os}</Table.Cell>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4"
										>{getTranslation('summary.load_average', 'Load Average')}</Table.Cell
									>
									<Table.Cell class="p-1.5 px-4">{basicInfo.loadAverage}</Table.Cell>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4"
										>{getTranslation('summary.boot_mode', 'Boot Mode')}</Table.Cell
									>
									<Table.Cell class="p-1.5 px-4">{basicInfo.bootMode}</Table.Cell>
								</Table.Row>

								<Table.Row>
									<Table.Cell class="p-1.5 px-4"
										>{getTranslation('summary.sylve_version', 'Sylve Version')}</Table.Cell
									>
									<Table.Cell class="p-1.5 px-4">{basicInfo.sylveVersion}</Table.Cell>
								</Table.Row>
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>

				<!-- <AreaChart title="CPU Usage" elements={chartElements} /> -->
				<!-- <Card.Root class="w-full">
					<Card.Header>
						<Card.Title>
							<div class="flex items-center space-x-2">
								<Icon icon="solar:cpu-bold" class="h-5 w-5" />
								<p>{getTranslation('summary.cpu_usage', 'CPU Usage')}</p>
							</div>
						</Card.Title>
					</Card.Header>
					<Card.Content class="h-[300px]">
						<LineGraph
							data={[cpuHistoricalData, ioDelayHistoricalData]}
							valueType="percentage"
							keys={[
								{
									key: 'cpuUsage',
									title: getTranslation('summary.cpu_usage', 'CPU Usage'),
									color: 'hsl(0, 50%, 50%)'
								},
								{
									key: 'ioDelay',
									title: getTranslation('summary.io_delay', 'I/O Delay'),
									color: 'hsl(50, 50%, 50%)'
								}
							]}
						/>
					</Card.Content>
				</Card.Root> -->
			</div>
		</ScrollArea>
	</div>
</div>
