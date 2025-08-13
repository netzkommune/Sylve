<script lang="ts">
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import type { CPUInfo } from '$lib/types/info/cpu';
	import { getCache } from '$lib/utils/http';
	import humanFormat from 'human-format';

	interface Props {
		cpuCores: number;
		ram: number;
		startAtBoot: boolean;
		bootOrder: number;
		resourceLimits: boolean;
	}

	let {
		cpuCores = $bindable(),
		ram = $bindable(),
		startAtBoot = $bindable(),
		bootOrder = $bindable(),
		resourceLimits = $bindable()
	}: Props = $props();

	let cpuInfo: CPUInfo | null = $state(getCache('cpuInfo') || null);
	let humanSize = $state('1024 M');

	$effect(() => {
		if (!resourceLimits) {
			humanSize = '0 M';
			cpuCores = 0;
			return;
		} else {
			cpuCores = 1;
			humanSize = '1024 M';
		}

		if (cpuCores && cpuInfo) {
			if (cpuCores > cpuInfo.logicalCores) {
				cpuCores = cpuInfo.logicalCores - 1;
			}
		}

		try {
			const p = humanFormat.parse.raw(humanSize);
			ram = p.factor * p.value;
		} catch {
			ram = 1024;
		}
	});
</script>

<div class="flex flex-col gap-4 p-4">
	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<CustomValueInput
			label="CPU Cores"
			placeholder="1"
			type="number"
			bind:value={cpuCores}
			classes="flex-1 space-y-1.5"
			disabled={!resourceLimits}
		/>

		<CustomValueInput
			label="Memory Size"
			placeholder="10G"
			bind:value={humanSize}
			classes="flex-1 space-y-1.5"
			disabled={!resourceLimits}
		/>

		<CustomValueInput
			label="Boot Order"
			placeholder="1"
			type="number"
			bind:value={bootOrder}
			classes="flex-1 space-y-1.5"
		/>
	</div>

	<div class="flex flex-row gap-2">
		<CustomCheckbox
			label="Start On Boot"
			bind:checked={startAtBoot}
			classes="flex items-center gap-2"
		></CustomCheckbox>

		<CustomCheckbox
			label="Resource Limits"
			bind:checked={resourceLimits}
			classes="flex items-center gap-2"
		></CustomCheckbox>
	</div>
</div>
