<script lang="ts">
	import type { HistoricalData } from '$lib/types/common';
	import { curveCatmullRom, line } from 'd3-shape';
	import { AreaChart, Axis, Tooltip } from 'layerchart';

	interface Props {
		data: HistoricalData[][];
		keys: {
			key: string;
			title: string;
			color: string;
		}[];
	}

	const { data, keys }: Props = $props();

	let flat: HistoricalData[] = $state(data.flat());
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
</script>

<div class="h-[300px] rounded border p-4">
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
				tickLength={0}
				tickLabelProps={{ class: 'fill-black dark:fill-white stroke-none' }}
				ticks={5}
			/>
			<Axis
				placement="left"
				grid={{ class: 'stroke-none' }}
				tickLength={0}
				tickLabelProps={{ class: 'fill-black dark:fill-white stroke-none' }}
				ticks={5}
			/>
		</svelte:fragment>
	</AreaChart>
</div>
