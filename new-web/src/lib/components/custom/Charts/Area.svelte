<script lang="ts">
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import type { AreaChartElement } from '$lib/types/components/chart';
	import Icon from '@iconify/svelte';
	import {
		CategoryScale,
		Chart,
		Filler,
		Legend,
		LinearScale,
		LineController,
		LineElement,
		PointElement,
		Title,
		Tooltip
	} from 'chart.js';
	import zoomPlugin from 'chartjs-plugin-zoom';
	import { format } from 'date-fns';
	import humanFormat from 'human-format';
	import { onDestroy, onMount } from 'svelte';

	let { title, description = '', elements, formatSize = false }: Props = $props();

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

	const labels = data.map((v) => [
		format(new Date(v.date), 'dd/MM/yyyy'),
		format(new Date(v.date), 'HH:mm')
	]);

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

	const datasets = series.map((s, i) => ({
		label: s.label,
		data: data.map((d) => Number(d[s.key])),
		borderColor: switchColor(s.color),
		backgroundColor: switchColor(s.color, 0.2),
		fill: i === 0 ? 'origin' : '-1',
		tension: 0.4,
		pointRadius: 0,
		pointHoverRadius: 4,
		order: s.label === 'CPU Usage' ? 2 : 1
	}));

	interface Props {
		title: string;
		description?: string;
		elements: AreaChartElement[];
		formatSize?: boolean;
	}

	let chart: Chart | undefined;
	let canvas: HTMLCanvasElement;

	Chart.register(
		LineController,
		LineElement,
		PointElement,
		LinearScale,
		CategoryScale,
		Title,
		Tooltip,
		Legend,
		Filler,
		zoomPlugin
	);

	onMount(() => {
		chart = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				transitions: {
					zoom: {
						animation: {
							duration: 1000,
							easing: 'easeOutCubic'
						}
					}
				},
				plugins: {
					legend: {
						position: 'top'
					},
					tooltip: {
						mode: 'index',
						intersect: false,
						callbacks: {
							title: (tooltipItems) => {
								const rawLabel = tooltipItems[0].label;
								const date = new Date(rawLabel);
								return [format(date, 'dd/MM/yyyy'), format(date, 'HH:mm:ss')];
							},
							label: (tooltipItem) => {
								const datasetLabel = tooltipItem.dataset.label || '';
								const value = Number(tooltipItem.raw);

								return `${datasetLabel}: ${
									formatSize ? humanFormat(value) : value.toLocaleString()
								}`;
							}
						}
					},
					zoom: {
						zoom: {
							wheel: { enabled: true },
							pinch: { enabled: true },
							mode: 'xy'
						},
						pan: {
							enabled: true,
							mode: 'xy'
						}
					}
				},

				scales: {
					x: {
						title: { color: '#ccc', display: true, text: 'Date' },
						ticks: {
							callback: function (value, index, ticks) {
								const labelValue = typeof value === 'number' ? value : Number(value);
								const date = new Date(this.getLabelForValue(labelValue));
								return [format(date, 'dd/MM/yyyy'), format(date, 'HH:mm')];
							}
						},
						grid: {
							color: '#333' // Optional: X-axis grid line color
						}
					},
					y: {
						beginAtZero: true,
						title: {
							color: '#ccc',
							display: true,
							text: 'Value'
						},
						ticks: {
							callback: function (value) {
								const numValue = Number(value);
								return formatSize ? humanFormat(numValue) : numValue.toLocaleString();
							}
						},
						grid: {
							color: '#333' // Optional: X-axis grid line color
						}
					}
				}
			}
		});
		setTimeout(() => chart?.resize(), 300);
	});

	onDestroy(() => {
		chart?.destroy();
	});
</script>

<Card.Root class="p-5">
	<Card.Header class="p-0">
		<Card.Title class="flex items-center justify-between gap-4">
			<div class="flex items-center gap-2">
				<Icon icon="solar:cpu-bold" class="h-5 w-5" />
				{title}
			</div>
			<div>
				<Button
					onclick={() => {
						chart?.resetZoom();
					}}
					variant="outline"
					size="sm"
					class="h-8"
				>
					<Icon icon="carbon:reset" class="h-4 w-4" />
					Reset zoom
				</Button>
			</div>
		</Card.Title>
		{#if description}
			<Card.Description>{description}</Card.Description>
		{/if}
	</Card.Header>

	<Card.Content class="h-full min-h-[300px] w-full p-0">
		<canvas bind:this={canvas} style="width: 100%; height: 100%; display: block;"></canvas>
	</Card.Content>
</Card.Root>
