<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import { getSwitches } from '$lib/api/network/switch';
	import { getPCIDevices, getPPTDevices } from '$lib/api/system/pci';
	import { newVM } from '$lib/api/vm/vm';
	import { getDatasets } from '$lib/api/zfs/datasets';
	import { getPools } from '$lib/api/zfs/pool';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Form from '$lib/components/ui/form/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as RadioGroup from '$lib/components/ui/radio-group/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import type { Iface } from '$lib/types/network/iface';
	import type { SwitchList } from '$lib/types/network/switch';
	import type { PCIDevice, PPTDevice } from '$lib/types/system/pci';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter, isValidMACAddress } from '$lib/utils/string';
	import { getPCIDeviceId, getPPTDeviceId } from '$lib/utils/system/pci';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import humanFormat from 'human-format';
	import { untrack } from 'svelte';
	import toast from 'svelte-french-toast';

	interface Props {
		open: boolean;
	}

	const results = useQueries([
		{
			queryKey: ['poolList-svm'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: false
		},
		{
			queryKey: ['datasetList-svm'],
			queryFn: async () => {
				return await getDatasets();
			},
			refetchInterval: 1000,
			keepPreviousData: false
		},
		{
			queryKey: ['networkInterfaces-svm'],
			queryFn: async () => {
				return await getInterfaces();
			},
			refetchInterval: 1000,
			keepPreviousData: false
		},
		{
			queryKey: ['networkSwitches-svm'],

			queryFn: async () => {
				return await getSwitches();
			},
			refetchInterval: false,
			keepPreviousData: false
		},
		{
			queryKey: ['pciDevices-svm'],
			queryFn: async () => {
				return (await getPCIDevices()) as PCIDevice[];
			},
			refetchInterval: 1000,
			keepPreviousData: false
		},
		{
			queryKey: ['pptDevices-svm'],
			queryFn: async () => {
				return (await getPPTDevices()) as PPTDevice[];
			},
			refetchInterval: 1000,
			keepPreviousData: false
		}
	]);

	let pools: Zpool[] = $derived($results[0].data as Zpool[]);
	let datasets: Dataset[] = $derived($results[1].data as Dataset[]);
	let volumes: Dataset[] = $derived(datasets.filter((dataset) => dataset.type === 'volume'));
	let filesystems: Dataset[] = $derived(
		datasets.filter((dataset) => dataset.type === 'filesystem')
	);

	// let networkInterfaces: Iface[] = $derived($results[2].data as Iface[]);
	let networkSwitches: SwitchList = $derived($results[3].data as SwitchList);
	let pciDevices: PCIDevice[] = $derived($results[4].data as PCIDevice[]);
	let pptDevices: PPTDevice[] = $derived($results[5].data as PPTDevice[]);
	let passablePci: PCIDevice[] = $derived(
		pciDevices.filter((device) => device.name.startsWith('ppt'))
	);

	// $inspect(pools);

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
			zVol: '',
			diskSize: 4,
			zFilesystem: ''
		},
		network: {
			networkSwitch: 'none',
			networkInterfaces: [] as string[],
			macAddress: ''
		},
		hardware: {
			cpuSockets: '1',
			cpuCores: '2',
			cpuThreads: '2',
			memorySize: '2048',
			cpuType: 'host',
			pciDevices: [] as string[]
		},
		advanced: {
			vncPort: '',
			vncPassword: '',
			displayResolution: '',
			bootOrder: '',
			bios: '',
			advancedOptions: {
				startAtBoot: false
			},
			startupOrder: ''
		}
	});

	let comboBoxes = $state({
		basic: {},
		storage: {
			storageType: {
				options: [
					{
						label: 'ZFS Volume',
						value: 'zfs-volume',
						description: 'High performance with snapshots'
					},
					{
						label: 'Raw Image',
						value: 'raw-image',
						description: 'Simple raw disk image'
					},
					{
						label: 'No Storage',
						value: 'no-storage',
						description: 'No storage allocated'
					}
				],
				value: 'zfs-volume'
			},
			zfsFilesystem: {
				open: false,
				options: [] as { label: string; value: string }[],
				value: ''
			},
			zfsVolume: {
				open: false,
				options: [] as { label: string; value: string }[],
				value: ''
			},
			emulationType: {
				open: false,
				options: [
					{ label: 'VirtIO', value: 'virtio-blk', description: 'Recommended for performance' },
					{
						label: 'AHCI-HD',
						value: 'ahci-hd',
						description: 'AHCI controller attached to a SATA hard drive'
					},
					{
						label: 'NVMe',
						value: 'nvme',
						description: 'NVM Express (NVMe) controller'
					}
				],
				value: 'nvme'
			}
		},
		network: {
			networkSwitches: {
				options: [{ label: 'None', value: 'none', description: 'No network switch' }] as {
					label: string;
					value: string;
					description?: string;
				}[]
			}
		},
		hardware: {
			pciDevices: {
				options: [] as { label: string; value: string; description?: string }[]
			}
		},
		advanced: {
			displayResolution: {
				open: false,
				options: [
					{ label: '1024x768', value: '1024x768' },
					{ label: '1280x720', value: '1280x720' },
					{ label: '1920x1080', value: '1920x1080' },
					{ label: '2560x1440', value: '2560x1440' },
					{ label: '3840x2160', value: '3840x2160' }
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

	function setComboboxOptions(type: string) {
		if (type === 'volumes') {
			comboBoxes.storage.zfsVolume.options = volumes.map((volume) => ({
				label: `${volume.name} (${humanFormat(volume.volsize)})`,
				value: volume.properties.guid as string
			}));
		}

		if (type === 'filesystems') {
			comboBoxes.storage.zfsFilesystem.options = filesystems.map((fs) => ({
				label: fs.name,
				value: fs.properties.guid as string
			}));
		}

		if (type === 'networkSwitches') {
			if (networkSwitches.standard && networkSwitches.standard.length > 0) {
				for (const sw of networkSwitches.standard) {
					if (
						comboBoxes.network.networkSwitches.options.some((option) => option.value === sw.name)
					) {
						continue;
					}

					comboBoxes.network.networkSwitches.options.push({
						label: sw.name,
						value: sw.id.toString(),
						description: `${sw.ports?.map((port) => port.name).join(', ') || 'No ports available'}`
					});
				}

				comboBoxes.network.networkSwitches.options.sort((a, b) => {
					if (a.value === 'none') return 1;
					if (b.value === 'none') return -1;
					return a.label.localeCompare(b.label);
				});
			}
		}
	}

	// $effect(() => {
	// 	if (datasets && datasets.length > 0) {
	// 		setComboboxOptions('volumes');
	// 		setComboboxOptions('filesystems');
	// 	}
	// });

	// $effect(() => {
	// 	if (networkSwitches && networkSwitches.standard && networkSwitches.standard.length > 0) {
	// 		setComboboxOptions('networkSwitches');
	// 	}
	// });

	// $effect(() => {
	// 	if (pciDevices && pciDevices.length > 0) {
	// 		comboBoxes.hardware.pciDevices.options = passablePci.map((device) => ({
	// 			label: device.names.device,
	// 			value: getPPTDeviceId(device, pptDevices).toString(),
	// 			description: getPCIDeviceId(device)
	// 		}));
	// 	}
	// });

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

	async function createVM() {
		/* Basic */
		const name = modal.basic.vmName.trim();
		const id = modal.basic.vmId.trim();
		const description = modal.basic.description.trim();
		if (!name || !id) {
			toast.error('VM Name and ID are required', {
				position: 'bottom-center'
			});

			return;
		}

		// id should be a number
		if (isNaN(parseInt(id))) {
			toast.error('VM ID must be a valid number', {
				position: 'bottom-center'
			});
		}

		if (description.length > 512) {
			toast.error('Description cannot exceed 512 characters', {
				position: 'bottom-center'
			});
			return;
		}

		/* Storage */
		const storageType = modal.storage.storageType || 'none';
		const emulationType = comboBoxes.storage.emulationType.value || '';
		const storageSize = modal.storage.diskSize || 0;

		if (storageType === 'zfs-volume' && !comboBoxes.storage.zfsVolume.value) {
			toast.error('ZFS volume is required', {
				position: 'bottom-center'
			});
			return;
		}

		if (storageType === 'raw-image') {
			if (!storageSize) {
				toast.error('Disk size is required', {
					position: 'bottom-center'
				});
				return;
			}

			if (storageSize < 1 || storageSize === 0) {
				toast.error('Disk size must be at least 1 GB', {
					position: 'bottom-center'
				});
			}

			if (!comboBoxes.storage.zfsFilesystem.value) {
				toast.error('Filesystem dataset is required', {
					position: 'bottom-center'
				});
				return;
			}
		}

		if (emulationType === '') {
			toast.error('Emulation type is required', {
				position: 'bottom-center'
			});

			return;
		}

		/* Network */
		const networkSwitch = modal.network.networkSwitch;
		const macAddress = modal.network.macAddress.trim() || '';
		if (
			networkSwitch !== 'none' &&
			!isValidMACAddress(modal.network.macAddress) &&
			modal.network.macAddress
		) {
			toast.error('Invalid MAC address', {
				position: 'bottom-center'
			});
			return;
		}

		/* Hardware */
		const cpuSockets = parseInt(modal.hardware.cpuSockets);
		const cpuCores = parseInt(modal.hardware.cpuCores);
		const cpuThreads = parseInt(modal.hardware.cpuThreads);
		if (isNaN(cpuSockets) || cpuSockets < 1) {
			toast.error('CPU sockets must be at least 1', {
				position: 'bottom-center'
			});
			return;
		}

		if (isNaN(cpuCores) || cpuCores < 1) {
			toast.error('CPU cores must be at least 1', {
				position: 'bottom-center'
			});
			return;
		}

		if (isNaN(cpuThreads) || cpuThreads < 1) {
			toast.error('CPU threads must be at least 1', {
				position: 'bottom-center'
			});
			return;
		}

		const memorySize = parseInt(modal.hardware.memorySize);
		if (isNaN(memorySize) || memorySize < 512) {
			toast.error('Memory size must be at least 512 MB', {
				position: 'bottom-center'
			});
			return;
		}

		const pciDevices = modal.hardware.pciDevices.filter((device) => {
			return passablePci.some((pci) => getPPTDeviceId(pci, pptDevices) === parseInt(device));
		});

		/* Advanced */
		// const vncPort = parseInt(modal.advanced.vncPort);
		// if (isNaN(vncPort) || vncPort < 1000 || vncPort > 65535) {
		// 	toast.error('VNC port must be a valid number between 1000 and 65535', {
		// 		position: 'bottom-center'
		// 	});
		// 	return;
		// }

		// const vncPassword = modal.advanced.vncPassword.trim();
		// if (vncPassword.length > 64) {
		// 	toast.error('VNC password cannot exceed 64 characters', {
		// 		position: 'bottom-center'
		// 	});
		// 	return;
		// }

		// const displayResolution = modal.advanced.displayResolution;
		// if (!displayResolution) {
		// 	toast.error('Display resolution is required', {
		// 		position: 'bottom-center'
		// 	});
		// 	return;
		// }

		// let storageDataset = '';
		// if (storageType === 'raw-image') {
		// 	storageDataset = comboBoxes.storage.zfsFilesystem.value || '';
		// } else if (storageType === 'zfs-volume') {
		// 	storageDataset = comboBoxes.storage.zfsVolume.value || '';
		// }

		// const switchId = networkSwitch !== 'none' ? parseInt(networkSwitch) : 0;

		// await newVM(
		// 	name,
		// 	parseInt(id),
		// 	storageType,
		// 	storageDataset,
		// 	storageSize,
		// 	emulationType,
		// 	switchId,
		// 	macAddress,
		// 	cpuSockets,
		// 	cpuCores,
		// 	cpuThreads
		// );
	}

	// $effect(() => {
	// 	if (modal.storage.diskSize) {
	// 		if (modal.storage.diskSize < 1 || modal.storage.diskSize === 0) {
	// 			modal.storage.diskSize = 1;
	// 		}
	// 	}
	// });
</script>

<Dialog.Root bind:open closeOnOutsideClick={false}>
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
					<p class="text-muted-foreground text-sm">
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
								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="VM Name"
										placeholder="my-virtual-machine"
										bind:value={modal.basic.vmName}
										classes="flex-1 space-y-1"
									/>
									<CustomValueInput
										label="VM ID"
										placeholder="100"
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
							</div>
						{:else if value === 'vm_storage'}
							<div class="flex flex-col gap-4 p-4">
								<RadioGroup.Root bind:value={modal.storage.storageType} class="border p-2">
									<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
										{#each comboBoxes.storage.storageType.options as option}
											<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
												<RadioGroup.Item value={option.value} id={option.value} />
												<Label for={option.value} class="flex flex-col gap-2">
													<p class=" cursor-pointer">{option.label}</p>
													<p class="text-muted-foreground cursor-text text-sm">
														{option.description}
													</p>
												</Label>
											</div>
										{/each}
									</ScrollArea>
								</RadioGroup.Root>

								{#if modal.storage.storageType === 'zfs-volume'}
									<CustomComboBox
										bind:open={comboBoxes.storage.zfsVolume.open}
										label="ZFS Volume"
										bind:value={comboBoxes.storage.zfsVolume.value}
										data={comboBoxes.storage.zfsVolume.options}
										classes="flex-1 space-y-1"
										placeholder="Select ZFS volume"
										width="w-[40%]"
									></CustomComboBox>
								{/if}

								{#if modal.storage.storageType === 'raw-image'}
									<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
										<CustomValueInput
											label="Disk Size (GB)"
											placeholder="32"
											type="number"
											bind:value={modal.storage.diskSize}
											classes="flex-1 space-y-1"
										/>

										<CustomComboBox
											bind:open={comboBoxes.storage.zfsFilesystem.open}
											label="Filesystem Dataset"
											bind:value={comboBoxes.storage.zfsFilesystem.value}
											data={comboBoxes.storage.zfsFilesystem.options}
											classes="flex-1 space-y-1"
											placeholder="Select storage pool"
											width="w-[40%]"
										></CustomComboBox>
									</div>
								{/if}

								{#if modal.storage.storageType === 'zfs-volume' || modal.storage.storageType === 'raw-image'}
									<CustomComboBox
										bind:open={comboBoxes.storage.emulationType.open}
										label="Emulation Type"
										bind:value={comboBoxes.storage.emulationType.value}
										data={comboBoxes.storage.emulationType.options}
										classes="flex-1 space-y-1"
										placeholder="Select emulation type"
										width="w-[40%]"
									></CustomComboBox>
								{/if}
							</div>
						{:else if value === 'vm_network'}
							<div class="flex flex-col gap-4 p-4">
								<p>Network Switches</p>
								<RadioGroup.Root bind:value={modal.network.networkSwitch} class="border p-2">
									<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
										{#each comboBoxes.network.networkSwitches.options as option}
											<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
												<RadioGroup.Item value={option.value} id={option.value} />
												<Label for={option.value} class="flex flex-col gap-2">
													<p class=" cursor-pointer">{option.label}</p>
													<p class="text-muted-foreground cursor-text text-sm">
														{option.description}
													</p>
												</Label>
											</div>
										{/each}
									</ScrollArea>
								</RadioGroup.Root>

								{#if modal.network.networkSwitch !== 'none'}
									<CustomValueInput
										label="MAC Address"
										placeholder="56:49:fc:94:9b:4f"
										bind:value={modal.network.macAddress}
										classes="flex-1 space-y-1"
									/>
								{/if}
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
									<CustomValueInput
										label="CPU Sockets"
										placeholder="1"
										type="number"
										bind:value={modal.hardware.cpuSockets}
										classes="flex-1 space-y-1"
									/>

									<CustomValueInput
										label="CPU Cores"
										placeholder="1"
										type="number"
										bind:value={modal.hardware.cpuCores}
										classes="flex-1 space-y-1"
									/>

									<CustomValueInput
										label="CPU Threads"
										placeholder="2"
										type="number"
										bind:value={modal.hardware.cpuThreads}
										classes="flex-1 space-y-1"
									/>

									<CustomValueInput
										label="Memory Size (MB)"
										placeholder="2048"
										type="number"
										bind:value={modal.hardware.memorySize}
										classes="flex-1 space-y-1"
									/>
								</div>

								<p>PCI Passthrough</p>
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
																		<p class="text-muted-foreground text-sm">
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
								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomValueInput
										label="VNC Port"
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
									bind:open={comboBoxes.advanced.displayResolution.open}
									label="Display Resolution"
									bind:value={modal.advanced.displayResolution}
									data={comboBoxes.advanced.displayResolution.options}
									classes="flex-1 space-y-1"
									placeholder="Select display resolution"
									width="w-[85%]"
								></CustomComboBox>

								<p>Advanced Options</p>

								<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
									<CustomCheckbox
										label="Start at boot"
										bind:checked={modal.advanced.advancedOptions.startAtBoot}
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
							</div>
						{/if}
					</div>
				</Tabs.Content>
			{/each}
		</Tabs.Root>

		<Dialog.Footer>
			<div class="flex w-full justify-end px-3 py-3 md:flex-row">
				<Button size="sm" type="button" class="h-8" onclick={() => createVM()}
					>Create Virtual Machine</Button
				>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
