<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { scaleUtc } from 'd3-scale';
	import { curveNatural } from 'd3-shape';
	import { AreaChart } from 'layerchart';

	import { type AreaChartElement } from '$lib/types/components/chart';
	import Icon from '@iconify/svelte';

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
		if (!elements?.length) return [];

		const THRESH = 60_000;
		const series = elements.map(({ field, data: pts }) => ({
			field,
			points: pts
				.map((p) => ({ t: new Date(p.date).getTime(), v: p.value }))
				.sort((a, b) => a.t - b.t)
		}));

		series.sort((a, b) => a.points.length - b.points.length);
		const [base, ...others] = series;

		const out = [];

		for (const { t: bt, v: bv } of base.points) {
			const rec = { date: new Date(bt), [base.field]: bv };
			let good = true;

			for (const { field, points } of others) {
				let bestDiff = Infinity,
					bestVal = null;
				for (const { t, v } of points) {
					const d = Math.abs(t - bt);
					if (d < bestDiff) {
						bestDiff = d;
						bestVal = v;
					}

					if (t - bt > bestDiff) break;
				}
				if (bestVal === null || bestDiff > THRESH) {
					good = false;
					break;
				}
				rec[field] = bestVal;
			}

			if (good) out.push(rec);
		}

		return out;
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
	<Card.Header class="flex items-center gap-2 space-y-0 border-b py-0 sm:flex-row">
		<div class="grid flex-1 gap-1 text-center sm:text-left">
			<Card.Title>
				<div class="flex items-center gap-2">
					<Icon icon="solar:cpu-bold" class="h-5 w-5" />
					{title}
				</div>
			</Card.Title>
			{#if description}
				<Card.Description>{description}</Card.Description>
			{/if}
		</div>
	</Card.Header>

	<Card.Content>
		<Chart.Container {config} class="h-48 w-full">
			<AreaChart
				legend
				{data}
				x="date"
				xScale={scaleUtc()}
				yPadding={[0, 25]}
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
						format: (v: Date) => {
							return (
								v.toLocaleDateString('en-US', {
									day: 'numeric',
									month: 'numeric',
									year: 'numeric'
								}) +
								'\n' +
								v.toLocaleTimeString('en-US', {
									hour: '2-digit',
									minute: '2-digit'
								})
							);
						}
					}
				}}
			>
				{#snippet tooltip()}
					<Chart.Tooltip
						indicator="dot"
						labelFormatter={(v: Date) => {
							return v.toLocaleDateString('en-US', {
								month: 'long',
								day: 'numeric',
								year: 'numeric',
								minute: '2-digit',
								hour: '2-digit',
								hour12: true
							});
						}}
					/>
				{/snippet}
			</AreaChart>
		</Chart.Container>
	</Card.Content>
</Card.Root>
