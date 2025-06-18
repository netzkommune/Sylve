<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ChartContainer from '$lib/components/ui/chart/chart-container.svelte';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { curveNatural } from 'd3-shape';
	import { Area, AreaChart, ChartClipPath } from 'layerchart';

	import { type AreaChartElement } from '$lib/types/components/chart';

	interface Props {
		title: string;
		description?: string;
		elements: AreaChartElement[];
	}

	let { title, description = '', elements }: Props = $props();
	let config = $derived(
		elements.reduce((acc, element) => {
			acc[element.field] = { label: element.label, color: element.color };
			return acc;
		}, {} as Chart.ChartConfig)
	);

	let data = $derived.by(() => {
		let result: Array<{ date: Date } & { [key: string]: number | Date }> = [];
		for (const element of elements) {
			for (const data of element.data) {
				const existing = result.find(
					(item) => item.date instanceof Date && item.date.getTime() === data.date.getTime()
				);
				if (existing) {
					existing[element.field] = data.value;
				} else {
					const newData: { date: Date } & { [key: string]: number | Date } = { date: data.date };
					newData[element.field] = data.value;
					result.push(newData);
				}
			}
		}

		return result;
	});

	let series = $derived.by(() => {
		return elements.map((element) => ({
			key: element.field,
			label: element.label,
			color: element.color
		}));
	});
</script>

<Card.Root>
	<Card.Header class="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
		<div class="grid flex-1 gap-1 text-center sm:text-left">
			<Card.Title>{title}</Card.Title>
			{#if description}
				<Card.Description>{description}</Card.Description>
			{/if}
		</div>
	</Card.Header>

	<Card.Content>
		<ChartContainer {config} class="aspect-auto h-[250px] w-full">
			<AreaChart
				legend
				{data}
				x="date"
				{series}
				seriesLayout="stack"
				props={{
					area: {
						curve: curveNatural,
						'fill-opacity': 0.4,
						line: { class: 'stroke-1' },
						motion: 'tween'
					},
					xAxis: {
						ticks: 7,
						format: (v) => {
							return v.toLocaleDateString('en-US', {
								month: 'short',
								day: 'numeric'
							});
						}
					},
					yAxis: { format: () => '' }
				}}
			></AreaChart>
		</ChartContainer>
	</Card.Content>
</Card.Root>
