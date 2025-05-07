<script lang="ts">
	import type { PieChartData } from '$lib/types/common';
	import { PieChart } from 'layerchart';

	interface Data {
		containerClass: string;
		data: PieChartData[];
	}

	const { containerClass, data }: Data = $props();

	$inspect(data);
</script>

<div class={containerClass}>
	<PieChart {data} key="label" value="value" c="color" renderContext="svg" />
</div>
