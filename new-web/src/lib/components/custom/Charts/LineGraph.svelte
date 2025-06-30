<script lang="ts">
	import type { HistoricalData } from '$lib/types/common';
	import { formatValue, getDateFormatByInterval } from '$lib/utils/zfs/pool';
	import { curveCatmullRom } from 'd3-shape';
	import { format as dateFormat } from 'date-fns';
	import { toZonedTime } from 'date-fns-tz';
	import { AreaChart, Axis, Tooltip } from 'layerchart';

	interface Props {
		data: HistoricalData[][];
		keys: {
			key: string;
			title: string;
			color: string;
		}[];
		valueType: string;
		unformattedKeys?: string[];
		interval?: string;
	}

	const { data = $bindable([]), keys, valueType, unformattedKeys, interval }: Props = $props();

	let flat: HistoricalData[] = $derived.by(() => data.flat());
	let series = $derived.by(() => {
		let uniqueKeys = Array.from(
			new Set(flat.flatMap((d) => Object.keys(d).filter((k) => k !== 'date')))
		);

		return uniqueKeys.map((key, i) => ({
			key,
			label: keys.find((k) => k.key === key)?.title || key,
			color: keys.find((k) => k.key === key)?.color || 'pink'
		}));
	});

	const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

	const formatWithTimeZone = (date: Date | string, formatStr: string) => {
		const zonedDate = toZonedTime(new Date(date), timeZone);
		return dateFormat(zonedDate, formatStr);
	};

	let dateFormatString = $derived.by(() => getDateFormatByInterval(Number(interval), false));
</script>

<AreaChart
	props={{ area: { curve: curveCatmullRom } }}
	data={flat}
	x="date"
	{series}
	grid={true}
	legend
	tooltip={{ mode: 'quadtree' }}
	renderContext="svg"
>
	<svelte:fragment slot="axis">
		<Axis
			placement="bottom"
			grid={{ class: 'stroke-none' }}
			tickLength={15}
			tickLabelProps={{
				rotate: -30,
				class: 'fill-black dark:fill-white stroke-none'
			}}
			format={(d) => formatWithTimeZone(d, interval ? dateFormatString : 'hh:mm a')}
		/>
		<Axis
			placement="left"
			grid={{ class: 'stroke-none' }}
			tickLength={5}
			tickLabelProps={{ class: 'fill-black dark:fill-white stroke-none' }}
			format={(d) => String(formatValue(d, unformattedKeys, valueType))}
		/>
	</svelte:fragment>
	<svelte:fragment slot="tooltip">
		<Tooltip.Root class="bg-card border-border border p-2" let:data>
			<Tooltip.Header class="border-border border-b pb-1 text-sm">
				{formatWithTimeZone(data.date, interval ? dateFormatString : 'hh:mm a')}
			</Tooltip.Header>
			<Tooltip.List>
				{#each Object.entries(data).filter(([key]) => key !== 'date') as [key, value]}
					<Tooltip.Item
						label={series.find((s) => s.key === key)?.label || key}
						value={formatValue(Number(value), unformattedKeys, valueType)}
						color={series.find((s) => s.key === key)?.color || 'pink'}
					/>
				{/each}
			</Tooltip.List>
		</Tooltip.Root>
	</svelte:fragment>
</AreaChart>
