<script lang="ts">
	import { createNetworkObject } from '$lib/api/network/object';
	import Button from '$lib/components/ui/button/button.svelte';
	import ComboBoxBindable from '$lib/components/ui/custom-input/combobox-bindable.svelte';
	import ComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { handleAPIError } from '$lib/utils/http';
	import { generateComboboxOptions } from '$lib/utils/input';
	import { isValidIPv4, isValidIPv6, isValidMACAddress } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();
	let options = {
		name: '',
		type: {
			combobox: {
				open: false,
				value: '',
				options: generateComboboxOptions(['Host(s)', 'Network(s)', 'MAC(s)'])
			}
		},
		hosts: {
			combobox: {
				open: false,
				value: [] as string[],
				options: [] as { label: string; value: string }[]
			}
		},
		networks: {
			combobox: {
				open: false,
				value: [] as string[],
				options: [] as { label: string; value: string }[]
			}
		},
		macs: {
			combobox: {
				open: false,
				value: [] as string[],
				options: [] as { label: string; value: string }[]
			}
		}
	};

	let properties = $state(options);

	async function create() {
		let error = '';

		if (properties.name === '') {
			error = 'Name is required';
		}

		if (properties.type.combobox.value === '') {
			error = 'Type is required';
		}

		if (
			properties.type.combobox.value === 'Host(s)' &&
			properties.hosts.combobox.value.length === 0
		) {
			error = 'At least one host must be selected';
		} else if (
			properties.type.combobox.value === 'Network(s)' &&
			properties.networks.combobox.value.length === 0
		) {
			error = 'At least one network must be selected';
		} else if (
			properties.type.combobox.value === 'MAC(s)' &&
			properties.macs.combobox.value.length === 0
		) {
			error = 'At least one MAC must be selected';
		}

		let values = [] as string[];

		if (properties.type.combobox.value === 'Host(s)') {
			const hosts = Array.from(new Set(properties.hosts.combobox.value));
			properties.hosts.combobox.value = hosts;

			let hasIPv4 = false;
			let hasIPv6 = false;

			for (const host of hosts) {
				if (isValidIPv4(host)) {
					hasIPv4 = true;
				} else if (isValidIPv6(host)) {
					hasIPv6 = true;
				} else {
					error = `Invalid host IP: ${host}`;
					break;
				}
			}

			if (!error && hasIPv4 && hasIPv6) {
				error = 'Cannot mix IPv4 and IPv6 addresses';
			}

			values = hosts;
		}

		if (properties.type.combobox.value === 'Network(s)') {
			const networks = Array.from(new Set(properties.networks.combobox.value));
			properties.networks.combobox.value = networks;

			let hasIPv4 = false;
			let hasIPv6 = false;

			for (const net of networks) {
				if (isValidIPv4(net, true)) {
					hasIPv4 = true;
				} else if (isValidIPv6(net, true)) {
					hasIPv6 = true;
				} else {
					error = `Invalid network CIDR: ${net}`;
					break;
				}
			}

			if (!error && hasIPv4 && hasIPv6) {
				error = 'Cannot mix IPv4 and IPv6 networks';
			}

			values = networks;
		}

		if (properties.type.combobox.value === 'MAC(s)') {
			const macs = Array.from(new Set(properties.macs.combobox.value));
			properties.macs.combobox.value = macs;

			for (const mac of macs) {
				if (!isValidMACAddress(mac)) {
					error = `Invalid MAC address: ${mac}`;
					break;
				}
			}

			values = macs;
		}

		if (error) {
			toast.error(error, {
				position: 'bottom-center'
			});
			return;
		}

		let oType = '';

		switch (properties.type.combobox.value) {
			case 'Host(s)':
				oType = 'Host';
				break;
			case 'Network(s)':
				oType = 'Network';
				break;
			case 'MAC(s)':
				oType = 'MAC';
				break;
			default:
				oType = properties.type.combobox.value;
		}

		const response = await createNetworkObject(properties.name, oType, values);
		if (response.error) {
			handleAPIError(response);
			toast.error('Failed to create network object', {
				position: 'bottom-center'
			});

			return;
		} else {
			toast.success('Created object', {
				position: 'bottom-center'
			});

			open = false;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<div class="flex items-center justify-between">
			<Dialog.Header>
				<Dialog.Title>
					<div class="flex items-center">
						<Icon icon="clarity:objects-solid" class="mr-2 h-6 w-6" />
						<span class="text-lg font-semibold">Create New Object</span>
					</div>
				</Dialog.Title>
			</Dialog.Header>

			<div class="flex items-center gap-0.5">
				<Button
					size="sm"
					variant="link"
					class="h-4"
					title={'Reset'}
					onclick={() => (properties = options)}
				>
					<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Reset'}</span>
				</Button>
				<Button size="sm" variant="link" class="h-4" title={'Close'} onclick={() => (open = false)}>
					<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Close'}</span>
				</Button>
			</div>
		</div>

		<div class="flex gap-4">
			<CustomValueInput
				label={'Name'}
				placeholder="Windows"
				bind:value={properties.name}
				classes="flex-1 space-y-1.5"
				type="text"
			/>

			<ComboBox
				bind:open={properties.type.combobox.open}
				label={'Type'}
				bind:value={properties.type.combobox.value}
				data={properties.type.combobox.options}
				classes="flex-1 space-y-1"
				placeholder="Select type"
				width="w-3/4"
			></ComboBox>
		</div>

		{#if properties.type.combobox.value !== ''}
			<div class="flex gap-4 overflow-auto">
				{#if properties.type.combobox.value === 'Host(s)' || properties.type.combobox.value === 'Network(s)' || properties.type.combobox.value === 'MAC(s)'}
					{#if properties.type.combobox.value === 'Host(s)'}
						<ComboBoxBindable
							bind:open={properties.hosts.combobox.open}
							label={'Hosts'}
							bind:value={properties.hosts.combobox.value}
							data={properties.hosts.combobox.options}
							classes="flex-1 space-y-1"
							placeholder="Select hosts"
							width="w-3/4"
							multiple={true}
						></ComboBoxBindable>
					{:else if properties.type.combobox.value === 'Network(s)'}
						<ComboBoxBindable
							bind:open={properties.networks.combobox.open}
							label={'Networks'}
							bind:value={properties.networks.combobox.value}
							data={properties.networks.combobox.options}
							classes="flex-1 space-y-1"
							placeholder="Select networks"
							width="w-3/4"
							multiple={true}
						></ComboBoxBindable>
					{:else if properties.type.combobox.value === 'MAC(s)'}
						<ComboBoxBindable
							bind:open={properties.macs.combobox.open}
							label={'MACs'}
							bind:value={properties.macs.combobox.value}
							data={properties.macs.combobox.options}
							classes="flex-1 space-y-1"
							placeholder="Select MACs"
							width="w-3/4"
							multiple={true}
						></ComboBoxBindable>
					{/if}
				{/if}
			</div>
		{/if}

		<Dialog.Footer class="flex justify-end">
			<div class="flex w-full items-center justify-end gap-2">
				<Button onclick={create} type="submit" size="sm">Create</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
