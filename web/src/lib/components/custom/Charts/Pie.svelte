<script lang="ts">
	import type { PieChartData } from '$lib/types/common';
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

	const switchColor = (color: string, alpha: number = 1) => {
		const base = (val: string) => val.replace(')', ` / ${alpha})`);
		switch (color) {
			case 'chart1':
				return base('oklch(0.646 0.222 41.116)');
			case 'chart-2':
				return base('oklch(0.6 0.118 184.704)');
			case 'chart-3':
				return base('oklch(0.398 0.07 227.392)');
			case 'chart-4':
				return base('oklch(0.828 0.189 84.429)');
			case 'chart-5':
				return base('oklch(0.769 0.188 70.08)');
			default:
				return base('oklch(0.646 0.222 41.116)');
		}
	};

	onMount(() => {
		chart = new Chart(canvas, {
			type: 'pie',
			data: {
				labels: data.map((d) => d.label),
				datasets: [
					{
						data: data.map((d) => d.value),
						backgroundColor: data.map((d) => switchColor(d.color)),
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
