<script lang="ts">
	import type { PieChartData, SeriesDataWithBaseline } from '$lib/types/common';
	import humanFormat from 'human-format';
	import { BarChart, Tooltip } from 'layerchart';

	type Colors = {
		baseline: string;
		value: string;
	};

	interface Data {
		containerClass: string;
		data: SeriesDataWithBaseline[];
		// formatter?: 'size-formatter' | 'default';
		colors: Colors;
	}

	const { containerClass, data, colors }: Data = $props();

	$inspect(data);
</script>

<div class={containerClass}>
	<div class="h-[300px] rounded border p-4">
		<BarChart
			{data}
			x="name"
			series={[
				{ key: 'baseline', color: colors.baseline, props: { insets: { x: 8 } } },
				{
					key: 'value',
					color: colors.value,
					props: { insets: { x: 8 } }
				}
			]}
			renderContext={'svg'}
		></BarChart>
	</div>
</div>
