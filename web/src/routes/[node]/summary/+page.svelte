<script lang="ts">
	import { getBasicInfo } from '$lib/api/info/basic';
	import { getCPUInfo } from '$lib/api/info/cpu';
	import { getRAMInfo, getSwapInfo } from '$lib/api/info/ram';
	import { getIODelay } from '$lib/api/zfs/pool';
	import LineGraph from '$lib/components/custom/LineGraph.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import type { HistoricalData } from '$lib/types/common';
	import type { BasicInfo } from '$lib/types/info/basic';
	import type { CPUInfo, CPUInfoHistorical } from '$lib/types/info/cpu';
	import type { RAMInfo } from '$lib/types/info/ram';
	import type { IODelay, IODelayHistorical } from '$lib/types/zfs/pool';
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
			initialData: data.basicInfo
		},
		{
			queryKey: ['cpuInfo'],
			queryFn: getCPUInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.cpuInfo
		},
		{
			queryKey: ['ramInfo'],
			queryFn: getRAMInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ramInfo
		},
		{
			queryKey: ['swapInfo'],
			queryFn: getSwapInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.swapInfo
		},
		{
			queryKey: ['ioDelay'],
			queryFn: getIODelay,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ioDelay
		},
		{
			queryKey: ['totalDiskUsage'],
			queryFn: getTotalDiskUsage,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.totalDiskUsage
		},
		{
			queryKey: ['cpuInfoHistorical'],
			queryFn: getCPUInfo,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.cpuInfoHistorical
		},
		{
			queryKey: ['ioDelayHistorical'],
			queryFn: getIODelay,
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.ioDelayHistorical
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

	let cpuHistoricalData: HistoricalData[] = $derived.by(() => {
		return cpuInfoHistorical.map((data) => ({
			date: new Date(data.createdAt),
			cpuUsage: data.usage
		}));
	});

	let ioDelayHistoricalData: HistoricalData[] = $derived.by(() => {
		return ioDelayHistorical.map((data) => ({
			date: new Date(data.createdAt),
			ioDelay: data.delay
		}));
	});
</script>

<div class="flex h-full w-full flex-col">
	<div class="min-h-0 flex-1">
		<ScrollArea orientation="both" class="h-full w-full">
			<div class="space-y-3 p-3">
				<Card.Root class="w-full">
					<Card.Header class="p-2 ">
						<Card.Description class="text-md ml-3 font-normal text-blue-600 dark:text-blue-500"
							>{basicInfo.hostname} (Started {secondsToHoursAgo(
								basicInfo.uptime
							)})</Card.Description
						>
					</Card.Header>
					<Card.Content class="p-2">
						<div class="ml-3 grid grid-cols-1 gap-4 md:grid-cols-2">
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="solar:cpu-bold" class="mr-1 h-5 w-5" />CPU Usage
									</p>
									<p>
										{floatToNDecimals(cpuInfo.usage, 2)}% of {cpuInfo.logicalCores} CPU(s)
									</p>
								</div>
								<Progress value={cpuInfo.usage || 0} max={100} class="h-2 w-[100%] " />
							</div>
							<div>
								<div class="flex w-full justify-between pb-1">
									<p class="inline-flex items-center">
										<Icon icon="ri:ram-fill" class="mr-1 h-5 w-5" />RAM Usage
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
										<Icon icon="bxs:server" class="mr-1 h-5 w-5" />Disk Usage
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
										<Icon icon="lets-icons:time-light" class="mr-1 h-5 w-5" />I/O Delay
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
										<Icon icon="ic:baseline-loop" class="mr-1 h-5 w-5" />SWAP usage
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
									<Table.Cell class="p-1.5 px-4">CPU(s)</Table.Cell>
									<Table.Cell class="p-1.5 px-4">{cpuInfo.logicalCores} x {cpuInfo.name}</Table.Cell
									>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4">Operating System</Table.Cell>
									<Table.Cell class="p-1.5 px-4">{basicInfo.os}</Table.Cell>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4">Load Average</Table.Cell>
									<Table.Cell class="p-1.5 px-4">{basicInfo.loadAverage}</Table.Cell>
								</Table.Row>
								<Table.Row>
									<Table.Cell class="p-1.5 px-4">Boot Mode</Table.Cell>
									<Table.Cell class="p-1.5 px-4">{basicInfo.bootMode}</Table.Cell>
								</Table.Row>

								<Table.Row>
									<Table.Cell class="p-1.5 px-4">Sylve Version</Table.Cell>
									<Table.Cell class="p-1.5 px-4">{basicInfo.sylveVersion}</Table.Cell>
								</Table.Row>
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>

				<Card.Root class="w-full">
					<Card.Header>
						<Card.Title>
							<div class="flex items-center justify-between space-x-2">
								<p>CPU Usage</p>
							</div>
						</Card.Title>
					</Card.Header>
					<Card.Content>
						<LineGraph
							data={[cpuHistoricalData, ioDelayHistoricalData]}
							keys={[
								{ key: 'cpuUsage', title: 'CPU Usage', color: 'hsl(0, 50%, 50%)' },
								{ key: 'ioDelay', title: 'I/O Delay', color: 'hsl(50, 50%, 50%)' }
							]}
						/>
					</Card.Content>
				</Card.Root>
			</div>
		</ScrollArea>
	</div>
</div>
