<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import { createSwitch, getSwitches } from '$lib/api/network/switch';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Iface } from '$lib/types/network/iface';
	import type { SwitchList } from '$lib/types/network/switch';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { generateComboboxOptions } from '$lib/utils/input';
	import { generateTableData } from '$lib/utils/network/switch';
	import { isValidMTU, isValidVLAN } from '$lib/utils/numbers';
	import {
		capitalizeFirstLetter,
		isValidIPv4,
		isValidIPv6,
		isValidSwitchName
	} from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import toast from 'svelte-french-toast';

	interface Data {
		interfaces: Iface[];
		switches: SwitchList;
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['networkInterfaces'],
			queryFn: async () => {
				return await getInterfaces();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.interfaces,
			onSuccess: (data: Iface[]) => {
				updateCache('networkInterfaces', data);
			}
		},
		{
			queryKey: ['networkSwitches'],
			queryFn: async () => {
				return await getSwitches();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.switches,
			onSuccess: (data: SwitchList) => {
				updateCache('networkSwitches', data);
			}
		}
	]);

	// console.log('Interfaces', $results[0].data);
	// console.log('Switches', $results[1].data);

	const interfaces = $derived($results[0].data);
	const switches = $derived($results[1].data);

	let query: string = $state('');
	let useablePorts = $derived.by(() => {
		let used: string[] = [];
		const available: string[] = [];

		if (switches) {
			if (switches.standard) {
				for (const sw of switches.standard) {
					if (sw.ports) {
						const ports = sw.ports.map((port) => port.name);
						used = [...used, ...ports];
					}
				}
			}
		}

		if (interfaces) {
			if (interfaces) {
				for (const iface of interfaces) {
					if (!used.includes(iface.name) && !iface.groups?.includes('bridge')) {
						available.push(iface.name);
					}
				}
			}
		}

		return available;
	});

	let confirmModals = $state({
		active: '' as 'newSwitch' | 'deleteSwitch',
		newSwitch: {
			open: false,
			name: '',
			mtu: '',
			vlan: '0',
			address: '',
			address6: '',
			private: false,
			ports: [] as string[]
		},
		deleteSwitch: {
			open: false,
			name: '',
			id: 0
		}
	});

	let comboBoxes = $state({
		switches: {
			open: false,
			value: []
		}
	});

	async function confirmAction() {
		if (confirmModals.active === 'newSwitch') {
			if (!isValidSwitchName(confirmModals.newSwitch.name)) {
				toast.error('Invalid switch name', {
					position: 'bottom-center'
				});

				return;
			}

			if (
				confirmModals.newSwitch.mtu !== '' &&
				!isValidMTU(parseInt(confirmModals.newSwitch.mtu))
			) {
				toast.error('Invalid MTU', {
					position: 'bottom-center'
				});

				return;
			}

			if (
				confirmModals.newSwitch.vlan !== '' &&
				!isValidVLAN(parseInt(confirmModals.newSwitch.vlan))
			) {
				toast.error('Invalid VLAN', {
					position: 'bottom-center'
				});

				return;
			}

			if (
				confirmModals.newSwitch.address !== '' &&
				!isValidIPv4(confirmModals.newSwitch.address, true)
			) {
				toast.error('Invalid IPv4 CIDR', {
					position: 'bottom-center'
				});
				return;
			}

			if (
				confirmModals.newSwitch.address6 !== '' &&
				!isValidIPv6(confirmModals.newSwitch.address6, true)
			) {
				toast.error('Invalid IPv6 CIDR', {
					position: 'bottom-center'
				});
				return;
			}

			const created = await createSwitch(
				confirmModals.newSwitch.name,
				parseInt(confirmModals.newSwitch.mtu),
				parseInt(confirmModals.newSwitch.vlan),
				confirmModals.newSwitch.address,
				confirmModals.newSwitch.address6,
				confirmModals.newSwitch.private,
				comboBoxes.switches.value
			);
		}
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
		<Button
			on:click={() => {
				confirmModals.active = 'newSwitch';
				confirmModals.newSwitch.open = true;
			}}
			size="sm"
			class="h-6"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>
	</div>

	<TreeTable name="tt-switches" data={generateTableData(switches, 'standard')} />
</div>

{#if confirmModals.active === 'newSwitch'}
	<Dialog.Root
		bind:open={confirmModals[confirmModals.active].open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<Dialog.Content>
			<div class="flex items-center justify-between px-1 py-3">
				<Dialog.Header>
					<Dialog.Title>
						<div class="flex items-center">
							<Icon icon="clarity:network-switch-line" class="mr-2 h-6 w-6" />
							New Switch
						</div>
					</Dialog.Title>
				</Dialog.Header>

				<Dialog.Close
					class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
					onclick={() => {
						confirmModals.newSwitch.open = false;
					}}
				>
					<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
				</Dialog.Close>
			</div>

			<CustomValueInput
				label={capitalizeFirstLetter(getTranslation('common.name', 'Name'))}
				placeholder="public"
				bind:value={confirmModals[confirmModals.active].name}
				classes="flex-1 space-y-1"
			/>

			<div class="flex gap-4">
				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.mtu', 'MTU'))}
					placeholder="1280"
					bind:value={confirmModals[confirmModals.active].mtu}
					classes="flex-1 space-y-1"
					type="number"
				/>

				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.vlan', 'VLAN'))}
					placeholder="0"
					bind:value={confirmModals[confirmModals.active].vlan}
					classes="flex-1 space-y-1"
					type="number"
				/>
			</div>

			<div class="flex gap-4">
				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.ipv4', 'IPv4'))}
					placeholder="10.0.0.1/24"
					bind:value={confirmModals[confirmModals.active].address}
					classes="flex-1 space-y-1"
				/>

				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.ipv6', 'IPv6'))}
					placeholder="fdcb:cafe::beef/56"
					bind:value={confirmModals[confirmModals.active].address6}
					classes="flex-1 space-y-1"
				/>
			</div>

			<CustomComboBox
				bind:open={comboBoxes.switches.open}
				label="Ports"
				bind:value={comboBoxes.switches.value}
				data={generateComboboxOptions(useablePorts)}
				classes="flex-1 space-y-1"
				placeholder="Select ports"
				multiple={true}
				width="w-3/4"
			></CustomComboBox>

			<CustomCheckbox
				label="Private"
				bind:checked={confirmModals[confirmModals.active].private}
				classes="flex items-center gap-2"
			></CustomCheckbox>

			<Dialog.Footer class="flex justify-between gap-2 py-3">
				<div class="flex gap-2">
					<Button
						variant="default"
						class="h-8 bg-blue-600 text-white hover:bg-blue-700"
						onclick={() => confirmAction()}
					>
						{capitalizeFirstLetter(getTranslation('common.create', 'Create'))}
					</Button>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}
