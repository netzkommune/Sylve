<script lang="ts">
	import type { PieChartData } from '$lib/types/common';
	import { switchColor } from '$lib/utils/chart';
	import { ArcElement, Chart, Legend, PieController, Tooltip } from 'chart.js';
	import humanFormat from 'human-format';
	import { onDestroy, onMount } from 'svelte';

	interface Props {
		containerClass: string;
		data: PieChartData[];
		formatter?: 'size-formatter' | 'default';
	}

	const { containerClass, data: rawData, formatter = 'default' }: Props = $props();

	let data = $derived.by(() => {
		if (!rawData || rawData.length === 0) return [];
		return rawData.map((item, index) => ({
			...item,
			color: item.color || `chart-${(index % 5) + 1}`
		}));
	});

	let canvas: HTMLCanvasElement;
	let chart: Chart;

	Chart.register(ArcElement, PieController, Tooltip, Legend);

	onMount(() => {
		chart = new Chart(canvas, {
			type: 'pie',
			data: {
				labels: data.map((d) => d.label),
				datasets: [
					{
						data: data.map((d) => d.value),
						backgroundColor: data.map((d) => switchColor(d.color, 0.6)),
						borderColor: data.map((d) => switchColor(d.color, 1)),
						borderWidth: 1
					}
				]
			},
			options: {
				responsive: true,
				plugins: {
					legend: {
						position: 'top'
					},
					tooltip: {
						callbacks: {
							label: function (context) {
								const label = context.label || '';
								const value = context.raw as number;
								const displayValue = formatter === 'size-formatter' ? humanFormat(value) : value;
								return `${label}: ${displayValue}`;
							}
						}
					}
				}
			}
		});
	});

	$effect(() => {
		if (chart && data && data.length > 0) {
			chart.data.labels = data.map((d) => d.label);
			chart.data.datasets[0].data = data.map((d) => d.value);
			chart.data.datasets[0].backgroundColor = data.map((d) => switchColor(d.color));
			chart.update();
		}
	});

	onDestroy(() => {
		chart?.destroy();
	});
</script>

<div class={containerClass}>
	<canvas bind:this={canvas}></canvas>
</div>
