<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import { createSwitch, deleteSwitch, getSwitches, updateSwitch } from '$lib/api/network/switch';
	import AlertDialog from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Iface } from '$lib/types/network/iface';
	import type { SwitchList } from '$lib/types/network/switch';
	import { isAPIResponse, updateCache } from '$lib/utils/http';
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
	import { toast } from 'svelte-sonner';

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
		active: '' as 'newSwitch' | 'editSwitch' | 'deleteSwitch',
		newSwitch: {
			open: false,
			name: '',
			mtu: '',
			vlan: '',
			address: '',
			address6: '',
			disableIPv6: false,
			private: false,
			ports: [] as string[],
			dhcp: false,
			slaac: false
		},
		editSwitch: {
			oldName: '',
			open: false,
			name: '',
			mtu: '',
			vlan: '',
			address: '',
			address6: '',
			disableIPv6: false,
			private: false,
			ports: [] as string[],
			dhcp: false,
			slaac: false
		},
		deleteSwitch: {
			open: false,
			name: '',
			id: 0
		}
	});

	let comboBoxes = $state({
		ports: {
			open: false,
			value: []
		}
	});

	async function confirmAction() {
		if (confirmModals.active === 'newSwitch' || confirmModals.active === 'editSwitch') {
			const activeModal = confirmModals[confirmModals.active];
			if (!isValidSwitchName(activeModal.name)) {
				toast.error(
					getTranslation('network.switch.errors.invalid_switch_name', 'Invalid switch name'),
					{
						position: 'bottom-center'
					}
				);

				return;
			}

			if (
				activeModal.mtu !== '' &&
				activeModal.mtu !== null &&
				!isValidMTU(parseInt(activeModal.mtu))
			) {
				toast.error(getTranslation('network.switch.errors.invalid_mtu', 'Invalid MTU'), {
					position: 'bottom-center'
				});

				return;
			}

			if (
				activeModal.vlan !== '' &&
				activeModal.vlan !== null &&
				!isValidVLAN(parseInt(activeModal.vlan))
			) {
				toast.error(getTranslation('network.switch.errors.invalid_vlan', 'Invalid VLAN'), {
					position: 'bottom-center'
				});

				return;
			}

			if (activeModal.address !== '' && !isValidIPv4(activeModal.address, true)) {
				toast.error(
					getTranslation('network.switch.errors.invalid_ipv4_cidr', 'Invalid IPv4 CIDR'),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			if (activeModal.address6 !== '' && !isValidIPv6(activeModal.address6, true)) {
				toast.error(
					getTranslation('network.switch.errors.invalid_ipv6_cidr', 'Invalid IPv6 CIDR'),
					{
						position: 'bottom-center'
					}
				);
				return;
			}

			if (confirmModals.active === 'newSwitch') {
				const created = await createSwitch(
					activeModal.name,
					parseInt(activeModal.mtu),
					parseInt(activeModal.vlan),
					activeModal.address,
					activeModal.address6,
					activeModal.private,
					activeModal.dhcp,
					comboBoxes.ports.value,
					activeModal.disableIPv6,
					activeModal.slaac
				);

				if (isAPIResponse(created) && created.status === 'success') {
					let message = `${capitalizeFirstLetter(getTranslation('network.switch.switch', 'Switch'))} ${confirmModals.newSwitch.name} ${getTranslation('common.created', 'created')}`;
					toast.success(message, {
						position: 'bottom-center'
					});
				} else {
					toast.error(
						getTranslation('network.switch.errors.create_switch', 'Error creating switch'),
						{
							position: 'bottom-center'
						}
					);
				}
			} else {
				const edited = await updateSwitch(
					activeRow?.id as number,
					parseInt(activeModal.mtu),
					parseInt(activeModal.vlan),
					activeModal.address,
					activeModal.address6,
					activeModal.private,
					comboBoxes.ports.value,
					activeModal.disableIPv6,
					activeModal.slaac,
					activeModal.dhcp
				);

				if (isAPIResponse(edited) && edited.status === 'success') {
					let message = `${capitalizeFirstLetter(getTranslation('network.switch.switch', 'Switch'))} ${confirmModals.editSwitch.name} ${getTranslation('common.updated', 'updated')}`;
					toast.success(message, {
						position: 'bottom-center'
					});
				} else {
					toast.error(
						getTranslation('network.switch.errors.update_switch', 'Error updating switch'),
						{
							position: 'bottom-center'
						}
					);
				}
			}

			resetModal(true);
		}
	}

	let tableData = $derived(generateTableData(switches, 'standard'));
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));

	function handleDelete() {
		if (activeRow && Object.keys(activeRow).length > 0) {
			confirmModals.active = 'deleteSwitch';
			confirmModals.deleteSwitch.open = true;
			confirmModals.deleteSwitch.name = activeRow.name;
			confirmModals.deleteSwitch.id = activeRow.id as number;
		}
	}

	function handleEdit() {
		if (activeRow && Object.keys(activeRow).length > 0) {
			confirmModals.active = 'editSwitch';
			confirmModals.editSwitch.open = true;
			confirmModals.editSwitch.oldName = activeRow.name;
			confirmModals.editSwitch.name = activeRow.name;
			confirmModals.editSwitch.mtu = activeRow.mtu as string;
			confirmModals.editSwitch.vlan = activeRow.vlan === '-' ? '' : (activeRow.vlan as string);
			confirmModals.editSwitch.address = activeRow.ipv4 === '-' ? '' : activeRow.ipv4;
			confirmModals.editSwitch.address6 = activeRow.ipv6 === '-' ? '' : activeRow.ipv6;
			confirmModals.editSwitch.disableIPv6 = (activeRow.disableIPv6 as boolean) || false;
			confirmModals.editSwitch.private = (activeRow.private as boolean) || false;
			confirmModals.editSwitch.dhcp = (activeRow.dhcp as boolean) || false;
			confirmModals.editSwitch.slaac = (activeRow.slaac as boolean) || false;

			comboBoxes.ports.value = activeRow.ports.map((port: { name: string }) => port.name);
		}
	}

	function resetModal(close: boolean = true) {
		if (close) {
			confirmModals.newSwitch.open = false;
			confirmModals.deleteSwitch.open = false;
			confirmModals.editSwitch.open = false;
		}

		confirmModals.newSwitch.name = '';
		confirmModals.newSwitch.mtu = '';
		confirmModals.newSwitch.vlan = '';
		confirmModals.newSwitch.address = '';
		confirmModals.newSwitch.address6 = '';
		confirmModals.newSwitch.disableIPv6 = false;
		confirmModals.newSwitch.private = false;
		confirmModals.newSwitch.dhcp = false;
		confirmModals.newSwitch.slaac = false;

		confirmModals.editSwitch.name = '';
		confirmModals.editSwitch.mtu = '';
		confirmModals.editSwitch.vlan = '';
		confirmModals.editSwitch.address = '';
		confirmModals.editSwitch.address6 = '';
		confirmModals.editSwitch.disableIPv6 = false;
		confirmModals.editSwitch.private = false;
		confirmModals.editSwitch.dhcp = false;
		confirmModals.editSwitch.slaac = false;

		comboBoxes.ports.value = [];
	}

	$effect(() => {
		if (confirmModals.newSwitch.slaac) {
			confirmModals.newSwitch.disableIPv6 = false;
		}

		if (confirmModals.editSwitch.slaac) {
			confirmModals.editSwitch.disableIPv6 = false;
		}
	});

	$effect(() => {
		if (confirmModals.newSwitch.disableIPv6) {
			confirmModals.newSwitch.slaac = false;
		}

		if (confirmModals.editSwitch.disableIPv6) {
			confirmModals.editSwitch.slaac = false;
		}
	});
</script>

{#snippet button(type: string)}
	{#if activeRow && Object.keys(activeRow).length > 0}
		{#if type === 'edit'}
			<Button
				onclick={handleEdit}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:pencil" class="mr-1 h-4 w-4" />
				{'Edit'}
			</Button>
		{:else if type === 'delete'}
			<Button
				onclick={handleDelete}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
				{'Delete'}
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
		<Button
			onclick={() => {
				confirmModals.active = 'newSwitch';
				confirmModals.newSwitch.open = true;
			}}
			size="sm"
			class="h-6"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{'New'}
		</Button>

		{@render button('edit')}
		{@render button('delete')}
	</div>

	<TreeTable
		name="tt-switches"
		data={tableData}
		bind:parentActiveRow={activeRows}
		multipleSelect={false}
	/>
</div>

{#if confirmModals.active === 'newSwitch' || confirmModals.active === 'editSwitch'}
	<Dialog.Root bind:open={confirmModals[confirmModals.active].open}>
		<Dialog.Content class="w-[90%] gap-4 p-5 lg:max-w-2xl">
			<div class="flex items-center justify-between px-1 py-1">
				<Dialog.Header>
					<Dialog.Title>
						<div class="flex items-center">
							<Icon icon="clarity:network-switch-line" class="mr-2 h-6 w-6" />
							{#if confirmModals.active === 'editSwitch'}
								{capitalizeFirstLetter(getTranslation('common.edit', 'Edit'))}
								{capitalizeFirstLetter(getTranslation('network.switch.switch', 'Switch'))}
								{'- ' + confirmModals.editSwitch.oldName}
							{:else}
								{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
								{capitalizeFirstLetter(getTranslation('network.switch.switch', 'Switch'))}
							{/if}
						</div>
					</Dialog.Title>
				</Dialog.Header>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
						onclick={() => resetModal(false)}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}</span
						>
					</Button>
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.close', 'Close'))}
						onclick={() => resetModal(true)}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

			{#if confirmModals.active === 'newSwitch'}
				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('common.name', 'Name'))}
					placeholder="public"
					bind:value={confirmModals[confirmModals.active].name}
					classes="flex-1 space-y-1.5"
				/>
			{/if}

			<div class="flex gap-4">
				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.mtu', 'MTU'))}
					placeholder="1280"
					bind:value={confirmModals[confirmModals.active].mtu}
					classes="flex-1 space-y-1.5"
					type="number"
				/>

				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.vlan', 'VLAN'))}
					placeholder="0"
					bind:value={confirmModals[confirmModals.active].vlan}
					classes="flex-1 space-y-1.5"
					type="number"
				/>
			</div>

			<div class="flex gap-4">
				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.ipv4', 'IPv4'))}
					placeholder="10.0.0.1/24"
					bind:value={confirmModals[confirmModals.active].address}
					classes="flex-1 space-y-1.5"
					disabled={confirmModals[confirmModals.active].dhcp ? true : false}
				/>

				<CustomValueInput
					label={capitalizeFirstLetter(getTranslation('network.ipv6', 'IPv6'))}
					placeholder="fdcb:cafe::beef/56"
					bind:value={confirmModals[confirmModals.active].address6}
					classes="flex-1 space-y-1.5"
					disabled={confirmModals[confirmModals.active].disableIPv6 ||
					confirmModals[confirmModals.active].slaac
						? true
						: false}
				/>
			</div>

			{#if confirmModals.active === 'newSwitch'}
				<CustomComboBox
					bind:open={comboBoxes.ports.open}
					label={capitalizeFirstLetter(getTranslation('network.ports', 'Ports'))}
					bind:value={comboBoxes.ports.value}
					data={generateComboboxOptions(useablePorts)}
					classes="flex-1 space-y-1"
					placeholder="Select ports"
					multiple={true}
					width="w-3/4"
				></CustomComboBox>
			{:else}
				<CustomComboBox
					bind:open={comboBoxes.ports.open}
					label={capitalizeFirstLetter(getTranslation('network.ports', 'Ports'))}
					bind:value={comboBoxes.ports.value}
					data={generateComboboxOptions(useablePorts, activeRow?.portsOnly)}
					classes="flex-1 space-y-1"
					placeholder="Select ports"
					multiple={true}
					width="w-3/4"
				></CustomComboBox>
			{/if}

			<div class="flex items-center gap-2">
				<CustomCheckbox
					label={capitalizeFirstLetter(getTranslation('common.private', 'Private'))}
					bind:checked={confirmModals[confirmModals.active].private}
					classes="flex items-center gap-2 mt-1"
				></CustomCheckbox>

				<CustomCheckbox
					label={capitalizeFirstLetter(getTranslation('common.dhcp', 'DHCP'))}
					bind:checked={confirmModals[confirmModals.active].dhcp}
					classes="flex items-center gap-2 mt-1"
				></CustomCheckbox>

				<CustomCheckbox
					label={capitalizeFirstLetter(getTranslation('common.slaac', 'SLAAC'))}
					bind:checked={confirmModals[confirmModals.active].slaac}
					classes="flex items-center gap-2 mt-1"
				></CustomCheckbox>

				<CustomCheckbox
					label={capitalizeFirstLetter(getTranslation('common.disable_ipv6', 'Disable IPV6'))}
					bind:checked={confirmModals[confirmModals.active].disableIPv6}
					classes="flex items-center gap-2 mt-1"
				></CustomCheckbox>
			</div>

			<Dialog.Footer class="flex justify-between gap-2 py-3">
				<div class="flex gap-2">
					{#if confirmModals.active === 'editSwitch'}
						<Button onclick={confirmAction} type="submit" size="sm"
							>{capitalizeFirstLetter(getTranslation('common.save', 'Save'))}</Button
						>
					{:else}
						<Button onclick={confirmAction} type="submit" size="sm"
							>{capitalizeFirstLetter(getTranslation('common.create', 'Create'))}</Button
						>
					{/if}
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}

<AlertDialog
	open={confirmModals.deleteSwitch.open}
	names={{ parent: 'switch', element: confirmModals.deleteSwitch.name }}
	actions={{
		onConfirm: async () => {
			const result = await deleteSwitch(confirmModals.deleteSwitch.id);
			if (isAPIResponse(result) && result.status === 'success') {
				let message = `${capitalizeFirstLetter(getTranslation('network.switch.switch', 'Switch'))} ${confirmModals.deleteSwitch.name} ${getTranslation('common.deleted', 'deleted')}`;
				toast.success(message, {
					position: 'bottom-center'
				});
			} else {
				if (result && result.error) {
					if (result.error === 'switch_in_use_by_vm') {
						toast.error(
							getTranslation(
								'network.switch.errors.switch_in_use_by_vm',
								'Switch is in use by a VM'
							),
							{
								position: 'bottom-center'
							}
						);
						return;
					}
				}

				toast.error(
					getTranslation('network.switch.errors.delete_switch', 'Error deleting switch'),
					{
						position: 'bottom-center'
					}
				);
			}

			activeRows = null;

			confirmModals.deleteSwitch.open = false;
			confirmModals.deleteSwitch.name = '';
			confirmModals.deleteSwitch.id = 0;
		},
		onCancel: () => {
			confirmModals.deleteSwitch.open = false;
			confirmModals.deleteSwitch.name = '';
			confirmModals.deleteSwitch.id = 0;
		}
	}}
></AlertDialog>
