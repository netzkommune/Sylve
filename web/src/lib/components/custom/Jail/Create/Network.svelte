<script lang="ts">
	import * as RadioGroup from '$lib/components/ui/radio-group/index.js';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { NetworkObject } from '$lib/types/network/object';
	import type { SwitchList } from '$lib/types/network/switch';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { onMount } from 'svelte';
	import { isValidIPv4, isValidIPv6 } from '$lib/utils/string';
	import {
		generateIPOptions,
		generateMACOptions,
		generateNetworkOptions
	} from '$lib/utils/network/object';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';

	interface Props {
		switch: number;
		mac: number;
		ipv4: number;
		ipv4Gateway: number;
		ipv6: number;
		ipv6Gateway: number;
		dhcp: boolean;
		slaac: boolean;
		switches: SwitchList;
		networkObjects: NetworkObject[];
	}

	let {
		switch: nwSwitch = $bindable(),
		mac = $bindable(),
		ipv4 = $bindable(),
		ipv4Gateway = $bindable(),
		ipv6 = $bindable(),
		ipv6Gateway = $bindable(),
		dhcp = $bindable(),
		slaac = $bindable(),
		switches,
		networkObjects
	}: Props = $props();

	let usable = $derived({
		macs: networkObjects.filter(
			(obj) => obj.isUsed === false && obj.type === 'Mac' && obj.entries?.length === 1
		),
		ipv4: networkObjects.filter(
			(obj) =>
				obj.isUsed === false &&
				obj.type === 'Network' &&
				obj.entries?.length === 1 &&
				isValidIPv4(obj.entries[0].value, true)
		),
		ipv4Gateway: networkObjects.filter(
			(obj) => obj.type === 'Host' && obj.entries?.length === 1 && isValidIPv4(obj.entries[0].value)
		),
		ipv6: networkObjects.filter(
			(obj) =>
				obj.isUsed === false &&
				obj.type === 'Network' &&
				obj.entries?.length === 1 &&
				isValidIPv6(obj.entries[0].value, true)
		),
		ipv6Gateway: networkObjects.filter(
			(obj) => obj.type === 'Host' && obj.entries?.length === 1 && isValidIPv6(obj.entries[0].value)
		)
	});

	let comboBoxes = $state({
		mac: {
			open: false,
			value: '0'
		},
		ipv4: {
			open: false,
			value: '0'
		},
		ipv4Gateway: {
			open: false,
			value: '0'
		},
		ipv6: {
			open: false,
			value: '0'
		},
		ipv6Gateway: {
			open: false,
			value: '0'
		}
	});

	let checkBoxes = $state({
		dhcp: false,
		slaac: false
	});

	let swStr = $state('');

	onMount(() => {
		swStr = nwSwitch.toString();
	});

	$effect(() => {
		if (swStr) {
			nwSwitch = parseInt(swStr) || 0;
			if (nwSwitch === 0) {
				mac = 0;
				ipv4 = 0;
				ipv6 = 0;
				dhcp = false;
				slaac = false;
				checkBoxes.dhcp = false;
				checkBoxes.slaac = false;
			}
		}

		if (checkBoxes.dhcp) {
			comboBoxes.ipv4.value = '0';
			comboBoxes.ipv4Gateway.value = '0';
			dhcp = true;
		} else {
			if (comboBoxes.ipv4.value !== '0') {
				ipv4 = parseInt(comboBoxes.ipv4.value) || 0;
			}
		}

		if (checkBoxes.slaac) {
			comboBoxes.ipv6.value = '0';
			comboBoxes.ipv6Gateway.value = '0';
			slaac = true;
		} else {
			if (comboBoxes.ipv6.value !== '0') {
				ipv6 = parseInt(comboBoxes.ipv6.value) || 0;
			}
		}

		if (comboBoxes.mac.value !== '0') {
			mac = parseInt(comboBoxes.mac.value) || 0;
		} else {
			mac = 0;
		}
	});

	$effect(() => {
		if (comboBoxes.ipv4.value !== '0') {
			ipv4 = parseInt(comboBoxes.ipv4.value) || 0;
		} else {
			ipv4 = 0;
		}

		if (comboBoxes.ipv4Gateway.value !== '0') {
			ipv4Gateway = parseInt(comboBoxes.ipv4Gateway.value) || 0;
		} else {
			ipv4Gateway = 0;
		}

		if (comboBoxes.ipv6.value !== '0') {
			ipv6 = parseInt(comboBoxes.ipv6.value) || 0;
		} else {
			ipv6 = 0;
		}

		if (comboBoxes.ipv6Gateway.value !== '0') {
			ipv6Gateway = parseInt(comboBoxes.ipv6Gateway.value) || 0;
		} else {
			ipv6Gateway = 0;
		}
	});
</script>

{#snippet radioItem(id: number, name: string)}
	{@const i = `radio-${id}`}
	<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
		<RadioGroup.Item value={id.toString()} id={i} />
		<Label for={i} class="flex flex-col items-start gap-2">
			<p class="">{name}</p>
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
		<ScrollArea orientation="vertical" class="h-64 w-full max-w-full">
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
				bind:open={comboBoxes.ipv4.open}
				label="IPv4 Network"
				bind:value={comboBoxes.ipv4.value}
				data={generateNetworkOptions(usable.ipv4, 'IPv4')}
				classes="flex-1 space-y-1"
				placeholder="Select IPv4"
				width="w-full"
				disabled={usable.ipv4.length === 0 || checkBoxes.dhcp}
			></CustomComboBox>

			<CustomComboBox
				bind:open={comboBoxes.ipv4Gateway.open}
				label="IPv4 Gateway"
				bind:value={comboBoxes.ipv4Gateway.value}
				data={generateIPOptions(usable.ipv4Gateway, 'IPv4')}
				classes="flex-1 space-y-1"
				placeholder="Select IPv4"
				width="w-full"
				disabled={usable.ipv4Gateway.length === 0 || checkBoxes.dhcp}
			></CustomComboBox>

			<CustomComboBox
				bind:open={comboBoxes.ipv6.open}
				label="IPv6 Network"
				bind:value={comboBoxes.ipv6.value}
				data={generateNetworkOptions(usable.ipv6, 'IPv6')}
				classes="flex-1 space-y-1"
				placeholder="Select IPv6"
				width="w-full"
				disabled={usable.ipv6.length === 0 || checkBoxes.slaac}
			></CustomComboBox>

			<CustomComboBox
				bind:open={comboBoxes.ipv6Gateway.open}
				label="IPv6 Gateway"
				bind:value={comboBoxes.ipv6Gateway.value}
				data={generateIPOptions(usable.ipv6Gateway, 'IPv6')}
				classes="flex-1 space-y-1"
				placeholder="Select IPv6"
				width="w-full"
				disabled={usable.ipv6Gateway.length === 0 || checkBoxes.slaac}
			></CustomComboBox>

			<CustomComboBox
				bind:open={comboBoxes.mac.open}
				label="MAC Address"
				bind:value={comboBoxes.mac.value}
				data={generateMACOptions(usable.macs)}
				classes="flex-1 space-y-1"
				placeholder="Select MAC"
				width="w-full"
			></CustomComboBox>
		</div>

		<div class="mt-1 flex flex-row gap-4">
			<CustomCheckbox label="DHCP" bind:checked={checkBoxes.dhcp} classes="flex items-center gap-2"
			></CustomCheckbox>
			<CustomCheckbox
				label="SLAAC"
				bind:checked={checkBoxes.slaac}
				classes="flex items-center gap-2"
			></CustomCheckbox>
		</div>
	{/if}
</div>
