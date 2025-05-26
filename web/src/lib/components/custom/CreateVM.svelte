<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import Icon from '@iconify/svelte';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as RadioGroup from '$lib/components/ui/radio-group/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import * as Form from '$lib/components/ui/form/index.js';

	interface Props {
		open: boolean;
	}

	let { open = $bindable() }: Props = $props();

	let modal = $state({
		tabs: [
			{ value: 'vm_basic', label: 'Basic' },
			{ value: 'vm_storage', label: 'Storage' },
			{ value: 'vm_network', label: 'Network' },
			{ value: 'vm_hardware', label: 'Hardware' },
			{ value: 'vm_advanced', label: 'Advanced' }
		],
		basic: {
			vmName: '',
			vmId: '',
			description: '',
			bootIso: ''
		},
		storage: {
			storageType: 'zfs-volume',
			diskSize: '',
			storagePool: '',
			enableWriteBackCache: false,
			enableDiscard: false
		},
		network: {
			networkInterfaces: [] as string[],
			macAddress: '',
			VlanTag: '',
			enableFirewall: false
		},
		hardware: {
			cpuCores: '2',
			memorySize: '2048',
			cpuType: 'host',
			cpuFlags: '',
			pciDevices: [] as string[]
		},
		advanced: {
			vncPort: '',
			vncPassword: '',
			displayType: '',
			bootOrder: '',
			bios: '',
			advancedOptions: {
				startAtBoot: false,
				enableQemuGuestAgent: false,
				enableHotplug: false,
				enableNuma: false
			},
			startupOrder: ''
		}
	});

	let comboBoxes = $state({
		basic: {
			isoImage: {
				open: false,
				options: [
					{ label: 'Ubuntu 22.04', value: 'ubuntu-22.04.iso' },
					{ label: 'CentOS 8', value: 'centos-8.iso' },
					{ label: 'Windows 10', value: 'windows-10.iso' }
				]
			}
		},
		storage: {
			storageType: {
				options: [
					{
						label: 'ZFS Volume',
						value: 'zfs-volume',
						description: 'High performance with snapshots'
					},
					{
						label: 'Raw Disk',
						value: 'raw-disk',
						description: 'Direct disk access'
					},
					{
						label: 'QCOW2',
						value: 'qcow2',
						description: 'Compressed disk image'
					},
					{
						label: 'LVM',
						value: 'lvm',
						description: 'Logical volume manager'
					}
				]
			},
			storagePool: {
				open: false,
				options: [
					{ label: 'Local', value: 'local' },
					{ label: 'Local ZFS', value: 'local-zfs' },
					{ label: 'Shared Storage', value: 'shared storage' }
				]
			}
		},
		network: {
			networkInterfaces: {
				options: [
					{ label: 'vmbr0', value: 'vmbr0', description: 'Default bridge' },
					{ label: 'vmbr1', value: 'vmbr1', description: 'Internal network' },
					{ label: 'vmbr2', value: 'vmbr2', description: 'DMZ network' },
					{ label: 'NAT', value: 'NAT', description: 'Network address translation' }
				]
			}
		},
		hardware: {
			cpuCores: {
				open: false,
				options: [
					{ label: '1 Core', value: '1' },
					{ label: '2 Cores', value: '2' },
					{ label: '4 Cores', value: '4' },
					{ label: '6 Cores', value: '6' },
					{ label: '8 Cores', value: '8' },
					{ label: '12 Cores', value: '12' },
					{ label: '16 Cores', value: '16' },
					{ label: '24 Cores', value: '24' },
					{ label: '32 Cores', value: '32' }
				]
			},
			memorySize: {
				open: false,
				options: [
					{ label: '512 MB', value: '512' },
					{ label: '1 GB', value: '1024' },
					{ label: '2 GB', value: '2048' },
					{ label: '4 GB', value: '4096' },
					{ label: '8 GB', value: '8192' },
					{ label: '16 GB', value: '16384' },
					{ label: '32 GB', value: '32768' },
					{ label: '64 GB', value: '65536' }
				]
			},
			cpuType: {
				open: false,
				options: [
					{ label: 'Host', value: 'host' },
					{ label: 'kvm64', value: 'kvm64' },
					{ label: 'x86-64-v2', value: 'x86-64-v2' },
					{ label: 'x86-64-v3', value: 'x86-64-v3' }
				]
			},
			pciDevices: {
				options: [
					{ label: 'NVIDIA RTX 4090', value: 'gpu', description: 'PCI Address: 01:00.0' },
					{ label: 'AMD RX 7900 XTX', value: 'network-card', description: 'PCI Address: 02:00.0' },
					{
						label: 'Intel X710 10GbE',
						value: 'usb-controller',
						description: 'PCI Address: 03:00.0'
					},
					{
						label: 'USB 3.0 Controller',
						value: 'usb-controller',
						description: 'PCI Address: 04:00.0'
					},
					{
						label: 'Realtek ALC1220',
						value: 'sata-controller',
						description: 'PCI Address: 05:00.0'
					}
				]
			}
		},
		advanced: {
			bootIso: {
				open: false,
				options: [
					{ label: 'VNC', value: 'vnc' },
					{ label: 'SPICE', value: 'spice' },
					{ label: 'None Headless', value: 'noneHeadless' }
				]
			},
			bios: {
				open: false,
				options: [
					{ label: 'SeaBIOS', value: 'seaBios' },
					{ label: 'OVMF (UEFI)', value: 'ovmf(uefi)' }
				]
			}
		}
	});

	function addItem(field: string[], value: string) {
		if (!field.includes(value)) {
			field.push(value);
		}
	}

	function removeItem(field: string[], value: string) {
		const idx = field.indexOf(value);
		if (idx > -1) {
			field.splice(idx, 1);
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-5 transition-all duration-300 ease-in-out lg:max-w-[45%]"
	>
		<div class="flex items-center justify-between px-4 py-3">
			<Dialog.Header class="p-0">
				<Dialog.Title class="flex flex-col gap-1 text-left">
					<div class="flex items-center gap-2">
						<Icon icon="material-symbols:monitor-outline-rounded" class="h-5 w-5 " />
						Create Virtual Machine
					</div>
					<p class="text-sm text-muted-foreground">
						Configure your virtual machine with custom hardware and network settings
					</p>
				</Dialog.Title>
			</Dialog.Header>

			<div class="flex items-center gap-0.5">
				<Button
					size="sm"
					variant="ghost"
					class="h-8"
					title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
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
					onclick={() => (open = false)}
					title={capitalizeFirstLetter(getTranslation('common.close', 'Close'))}
				>
					<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
					<span class="sr-only"
						>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
					>
				</Button>
			</div>
		</div>

		<Tabs.Root value="vm_basic" class="w-full overflow-hidden">
			<Tabs.List class="grid w-full grid-cols-5 p-0 px-4">
				{#each modal.tabs as { value, label }}
					<Tabs.Trigger class="border-b" {value}>{label}</Tabs.Trigger>
				{/each}
			</Tabs.List>

			{#each modal.tabs as { value, label }}
				<Tabs.Content {value} class="">
					<div class="max-h-[65vh] overflow-y-auto">
						{#if value === 'vm_basic'}
							<div class="flex flex-col gap-4 p-4">
								<!-- <div class="flex flex-col gap-1">
								<div class="flex items-center gap-2">
									<Icon icon="material-symbols:settings-outline" class="h-5 w-5 " />
									Basic Configuration
								</div>
								<p class="text-sm text-muted-foreground">
									Set the fundamental properties of your virtual machine
								</p>
							</div> -->

								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="VM Name *"
										placeholder="my-virtual-machine"
										bind:value={modal.basic.vmName}
										classes="flex-1 space-y-1"
									/>
									<CustomValueInput
										label="VM ID"
										placeholder="auto-generated"
										bind:value={modal.basic.vmId}
										classes="flex-1 space-y-1"
									/>
								</div>

								<CustomValueInput
									label="Description"
									placeholder="Optional description for this virtual machine"
									type="textarea"
									textAreaCLasses="min-h-28"
									bind:value={modal.basic.description}
									classes="flex-1 space-y-1"
								/>

								<CustomComboBox
									bind:open={comboBoxes.basic.isoImage.open}
									label="Boot ISO Image *"
									bind:value={modal.basic.bootIso}
									data={comboBoxes.basic.isoImage.options}
									classes="flex-1 space-y-1"
									placeholder="Select an ISO file"
									width="w-[90%]"
								></CustomComboBox>
							</div>
						{:else if value === 'vm_storage'}
							<div class="flex flex-col gap-4 p-4">
								<!-- <div class="flex flex-col gap-1">
								<div class="flex items-center gap-2">
									<Icon icon="fluent:storage-16-regular" class="h-5 w-5 " />
									Storage Configuration
								</div>
								<p class="text-sm text-muted-foreground">
									Configure the storage backend and disk settings
								</p>
							</div> -->
								<p>Storage Type *</p>
								<RadioGroup.Root bind:value={modal.storage.storageType} class="border p-2">
									<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
										{#each comboBoxes.storage.storageType.options as option}
											<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
												<RadioGroup.Item value={option.value} id={option.value} />
												<Label for={option.value} class="flex flex-col gap-2">
													<p class=" cursor-pointer">{option.label}</p>
													<p class="cursor-text text-sm text-muted-foreground">
														{option.description}
													</p>
												</Label>
											</div>
										{/each}
									</ScrollArea>
								</RadioGroup.Root>

								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="Disk Size (GB) *"
										placeholder="32"
										type="number"
										bind:value={modal.storage.diskSize}
										classes="flex-1 space-y-1"
									/>
									<CustomComboBox
										bind:open={comboBoxes.storage.storagePool.open}
										label="Storage Pool"
										bind:value={modal.basic.bootIso}
										data={comboBoxes.storage.storagePool.options}
										classes="flex-1 space-y-1"
										placeholder="Select storage pool"
										width="w-[40%]"
									></CustomComboBox>
								</div>

								<CustomCheckbox
									label="Enable write-back cache"
									bind:checked={modal.storage.enableWriteBackCache}
									classes="flex items-center gap-2"
								></CustomCheckbox>

								<CustomCheckbox
									label="Enable discard/TRIM support"
									bind:checked={modal.storage.enableDiscard}
									classes="flex items-center gap-2"
								></CustomCheckbox>
							</div>
						{:else if value === 'vm_network'}
							<div class="flex flex-col gap-4 p-4">
								<!-- <div class="flex flex-col gap-1">
								<div class="flex items-center gap-2">
									<Icon icon="ph:network" class="h-5 w-5 " />
									Network Configuration
								</div>
								<p class="text-sm text-muted-foreground">
									Select network interfaces and configure networking
								</p>
							</div> -->
								<p>Network Interfaces</p>
								<div class="border p-2">
									<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
										<form action="/?/checkboxMultiple" method="POST" class="space-y-8">
											<Form.Fieldset
												form={modal.network.networkInterfaces}
												name="items"
												class="space-y-0"
											>
												<div class="space-y-2">
													{#each comboBoxes.network.networkInterfaces.options as item}
														{@const checked = modal.network.networkInterfaces.includes(item.value)}
														<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
															<Form.Control let:attrs>
																<Checkbox
																	{...attrs}
																	{checked}
																	onCheckedChange={(v) => {
																		if (v) {
																			addItem(modal.network.networkInterfaces, item.value);
																		} else {
																			removeItem(modal.network.networkInterfaces, item.value);
																		}
																	}}
																/>
																<div class="grid gap-1.5 leading-none">
																	<Label
																		for={item.value}
																		class="cursor-pointer text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
																	>
																		{item.label}
																	</Label>
																	{#if item.description}
																		<p class="text-sm text-muted-foreground">
																			{item.description}
																		</p>
																	{/if}
																</div>
															</Form.Control>
														</div>
													{/each}
													<Form.FieldErrors />
												</div>
											</Form.Fieldset>
										</form>
									</ScrollArea>
								</div>

								<hr />

								<p>Network Settings</p>
								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="MAC Address"
										placeholder="auto-generated"
										bind:value={modal.network.macAddress}
										classes="flex-1 space-y-1"
									/>
									<CustomValueInput
										label="VLAN Tag"
										placeholder="Optional"
										bind:value={modal.network.VlanTag}
										classes="flex-1 space-y-1"
									/>
								</div>

								<CustomCheckbox
									label="Enable firewall"
									bind:checked={modal.network.enableFirewall}
									classes="flex items-center gap-2"
								/>
							</div>
						{:else if value === 'vm_hardware'}
							<div class="flex flex-col gap-4 p-4">
								<!-- <div class="flex flex-col gap-1">
								<div class="flex items-center gap-2">
									<Icon icon="mingcute:chip-line" class="h-5 w-5 " />
									Hardware Configuration
								</div>
								<p class="text-sm text-muted-foreground">
									Configure CPU, memory, and PCI passthrough devices
								</p>
							</div> -->

								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomComboBox
										bind:open={comboBoxes.hardware.cpuCores.open}
										label="CPU Cores *"
										bind:value={modal.hardware.cpuCores}
										data={comboBoxes.hardware.cpuCores.options}
										classes="flex-1 space-y-1"
										placeholder="Select number of CPU cores"
										width="w-[84%] lg:w-[40%]"
									/>

									<CustomComboBox
										bind:open={comboBoxes.hardware.memorySize.open}
										label="Memory (MB) *"
										bind:value={modal.hardware.memorySize}
										data={comboBoxes.hardware.memorySize.options}
										classes="flex-1 space-y-1"
										placeholder="Select memory size"
										width="w-[84%] lg:w-[40%]"
									/>

									<CustomComboBox
										bind:open={comboBoxes.hardware.cpuType.open}
										label="CPU Type"
										bind:value={modal.hardware.cpuType}
										data={comboBoxes.hardware.cpuType.options}
										classes="flex-1 space-y-1"
										placeholder="Select CPU type"
										width="w-[84%] lg:w-[40%]"
									/>

									<CustomValueInput
										label="CPU Flags"
										placeholder="+vmx,+aes"
										bind:value={modal.hardware.cpuFlags}
										classes="flex-1 space-y-1"
									/>
								</div>

								<hr />

								<p>Network Interfaces</p>
								<div class="border p-2">
									<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
										<form action="/?/checkboxMultiple" method="POST" class="space-y-8">
											<Form.Fieldset
												form={modal.hardware.pciDevices}
												name="items"
												class="space-y-0"
											>
												<div class="space-y-2">
													{#each comboBoxes.hardware.pciDevices.options as item}
														{@const checked = modal.network.networkInterfaces.includes(item.value)}
														<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
															<Form.Control let:attrs>
																<Checkbox
																	{...attrs}
																	{checked}
																	onCheckedChange={(v) => {
																		if (v) {
																			addItem(modal.hardware.pciDevices, item.value);
																		} else {
																			removeItem(modal.hardware.pciDevices, item.value);
																		}
																	}}
																/>
																<div class="grid gap-1.5 leading-none">
																	<Label
																		for={item.value}
																		class="cursor-pointer text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
																	>
																		{item.label}
																	</Label>
																	{#if item.description}
																		<p class="text-sm text-muted-foreground">
																			{item.description}
																		</p>
																	{/if}
																</div>
															</Form.Control>
														</div>
													{/each}
													<Form.FieldErrors />
												</div>
											</Form.Fieldset>
										</form>
									</ScrollArea>
								</div>
							</div>
						{:else if value === 'vm_advanced'}
							<div class="flex flex-col gap-4 p-4">
								<!-- <div class="flex flex-col gap-1">
								<div class="flex items-center gap-2">
									<Icon icon="material-symbols:monitor-outline" class="h-5 w-5 " />
									Display & Advanced Settings
								</div>
								<p class="text-sm text-muted-foreground">
									Configure VNC, display options, and advanced VM settings
								</p>
							</div> -->

								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="VNC Port *"
										placeholder="5900"
										type="number"
										bind:value={modal.advanced.vncPort}
										classes="flex-1 space-y-1"
									/>
									<CustomValueInput
										label="VNC Password"
										placeholder="Optional"
										bind:value={modal.advanced.vncPassword}
										classes="flex-1 space-y-1"
									/>
								</div>

								<CustomComboBox
									bind:open={comboBoxes.advanced.bootIso.open}
									label="Display Type"
									bind:value={modal.advanced.displayType}
									data={comboBoxes.advanced.bootIso.options}
									classes="flex-1 space-y-1"
									placeholder="Select display type"
									width="w-[85%]"
								></CustomComboBox>

								<hr />

								<p>Boot Options</p>
								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="Boot Order"
										placeholder="cdrom, disk, network"
										bind:value={modal.advanced.bootOrder}
										classes="flex-1 space-y-1"
									/>
									<CustomComboBox
										bind:open={comboBoxes.basic.isoImage.open}
										label="BIOS"
										bind:value={modal.advanced.bios}
										data={comboBoxes.advanced.bios.options}
										classes="flex-1 space-y-1"
										placeholder="Select an ISO file"
										width="w-[84%] lg:w-[40%]"
									></CustomComboBox>
								</div>

								<p>Advanced Options</p>

								<CustomCheckbox
									label="Start at boot"
									bind:checked={modal.advanced.advancedOptions.startAtBoot}
									classes="flex items-center gap-2"
								></CustomCheckbox>

								<CustomCheckbox
									label="Enable QEMU Guest Agent"
									bind:checked={modal.advanced.advancedOptions.enableQemuGuestAgent}
									classes="flex items-center gap-2"
								></CustomCheckbox>

								<CustomCheckbox
									label="Enable Hotplug"
									bind:checked={modal.advanced.advancedOptions.enableHotplug}
									classes="flex items-center gap-2"
								></CustomCheckbox>

								<CustomCheckbox
									label="Enable NUMA"
									bind:checked={modal.advanced.advancedOptions.enableNuma}
									classes="flex items-center gap-2"
								></CustomCheckbox>

								<CustomValueInput
									label="Startup/Shutdown Order"
									placeholder="0"
									type="number"
									bind:value={modal.advanced.vncPassword}
									classes="flex-1 space-y-1"
								/>
							</div>
						{/if}
					</div>
				</Tabs.Content>
			{/each}
		</Tabs.Root>

		<Dialog.Footer>
			<div class="flex w-full justify-end px-3 py-3 md:flex-row">
				<Button size="sm" type="button" class="h-8">Create Virtual Machine</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
