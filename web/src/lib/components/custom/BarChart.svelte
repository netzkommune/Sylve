<script lang="ts">
	import type { PieChartData, SeriesDataWithBaseline } from '$lib/types/common';
	import { scaleBand, scaleLinear } from 'd3-scale';
	import { format } from 'date-fns';
	import humanFormat from 'human-format';
	import { Axis, BarChart, Highlight, Tooltip } from 'layerchart';

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
</script>

<div class={containerClass}>
	<BarChart
		{data}
		x="name"
		cDomain={[]}
		cRange={[colors.baseline, colors.value]}
		series={[
			{ key: 'baseline', color: colors.baseline, props: { insets: { x: 8 } } },
			{
				key: 'value',
				color: colors.value,
				props: { insets: { x: 8 } }
			}
		]}
		xScale={scaleBand().padding(0.2)}
		yScale={scaleLinear()}
		renderContext={'svg'}
		padding={{ bottom: 50, left: 20 }}
	>
		<svelte:fragment slot="axis">
			<Axis
				placement="left"
				labelProps={{ class: 'fill-green-500 stroke-none' }}
				tickLabelProps={{
					class: 'fill-neutral-300 dark:fill-neutral-200 stroke-none'
				}}
				format={(d) => humanFormat(d)}
				rule={{
					class: 'stroke-border dark:stroke-border'
				}}
				ticks={2}
			/>
			<Axis
				placement="bottom"
				tickLength={5}
				ticks={1}
				rule={{
					class: 'stroke-border dark:stroke-border '
				}}
				tickLabelProps={{
					rotate: 337,
					textAnchor: 'end',
					class: 'fill-neutral-300 dark:fill-neutral-200 stroke-none'
				}}
			/>
		</svelte:fragment>
		<svelte:fragment slot="highlight">
			<Highlight
				area={{
					class: 'fill-neutral-900/50 dark:fill-neutral-900/60'
				}}
			/>
		</svelte:fragment>
		<svelte:fragment slot="tooltip">
			<Tooltip.Root class="bg-secondary" let:data>
				<Tooltip.Header class={'border-b border-neutral-200 dark:border-neutral-700'}
					>{data.name}</Tooltip.Header
				>
				<Tooltip.List>
					<Tooltip.Item
						label="baseline"
						value={humanFormat(data.baseline)}
						color={colors.baseline}
					/>
					<Tooltip.Item label="value" value={humanFormat(data.value)} color={colors.value} />
				</Tooltip.List>
			</Tooltip.Root>
		</svelte:fragment>
	</BarChart>
</div>
