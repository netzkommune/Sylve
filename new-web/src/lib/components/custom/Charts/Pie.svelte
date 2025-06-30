<script lang="ts">
	import type { PieChartData } from '$lib/types/common';
	import humanFormat from 'human-format';
	import { PieChart, Tooltip } from 'layerchart';

	interface Data {
		containerClass: string;
		data: PieChartData[];
		formatter?: 'size-formatter' | 'default';
	}

	const { containerClass, data, formatter = 'default' }: Data = $props();
</script>

<div class={containerClass}>
	<PieChart
		{data}
		key="label"
		value="value"
		cRange={data.map((d) => d.color)}
		renderContext="svg"
		legend
		><svelte:fragment slot="tooltip">
			<Tooltip.Root let:data class="bg-secondary">
				<Tooltip.List>
					<Tooltip.Item
						label={data.label}
						value={formatter === 'size-formatter' ? humanFormat(data.value) : data.value}
					/>
				</Tooltip.List>
			</Tooltip.Root>
		</svelte:fragment>
	</PieChart>
</div>
