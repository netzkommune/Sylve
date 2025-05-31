<script lang="ts">
	import * as RadioGroup from '$lib/components/ui/radio-group/index.js';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import type { SwitchList } from '$lib/types/network/switch';
	import { Label } from '$lib/components/ui/label/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import { onMount } from 'svelte';

	interface Props {
		switch: number;
		mac: string;
		emulation: string;
		switches: SwitchList;
	}

	let {
		switch: nwSwitch = $bindable(),
		mac = $bindable(),
		emulation = $bindable(),
		switches
	}: Props = $props();

	let comboBoxes = $state({
		emulation: {
			open: false,
			value: 'virtio',
			options: [
				{ label: 'VirtIO', value: 'virtio' },
				{ label: 'E1000', value: 'e1000' }
			]
		}
	});

	let swStr = $state('');

	onMount(() => {
		swStr = nwSwitch.toString();
	});

	$effect(() => {
		if (swStr) {
			nwSwitch = parseInt(swStr) || 0;
		}
	});
</script>

{#snippet radioItem(id: number, name: string)}
	{@const i = `radio-${id}`}
	<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
		<RadioGroup.Item value={id.toString()} id={i} />
		<Label for={i} class="flex flex-col gap-2">
			<p>{name}</p>
			<p class="text-muted-foreground text-sm">
				{name === 'None'
					? 'No network switch will be allocated now, you can add it later'
					: 'Standard switch'}
			</p>
		</Label>
	</div>
{/snippet}

<div class="flex flex-col gap-4 p-4">
	<RadioGroup.Root bind:value={swStr} class="border p-2">
		<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
			{#if switches && switches.standard}
				{#each switches.standard ?? [] as sw}
					{@render radioItem(sw.id, sw.name)}
				{/each}
			{/if}
			{@render radioItem(0, 'None')}
		</ScrollArea>
	</RadioGroup.Root>

	{#if swStr !== '0'}
		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
			<CustomComboBox
				bind:open={comboBoxes.emulation.open}
				label="Emulation Type"
				bind:value={emulation}
				data={comboBoxes.emulation.options}
				classes="flex-1 space-y-1"
				placeholder="Select emulation type"
				width="w-[40%]"
			></CustomComboBox>

			<CustomValueInput
				label="MAC Address"
				placeholder="56:49:fc:94:9b:4f"
				bind:value={mac}
				classes="flex-1 space-y-1"
			/>
		</div>
	{/if}
</div>
