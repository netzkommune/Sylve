<script lang="ts">
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as RadioGroup from '$lib/components/ui/radio-group/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { type Download } from '$lib/types/utilities/downloader';
	import type { Dataset } from '$lib/types/zfs/dataset';

	import humanFormat from 'human-format';

	interface Props {
		volumes: Dataset[];
		filesystems: Dataset[];
		downloads: Download[];
		type: string;
		guid: string;
		size: number;
		emulation: string;
		iso: string;
	}

	let {
		volumes,
		filesystems,
		downloads,
		type = $bindable(),
		guid = $bindable(),
		size = $bindable(),
		emulation = $bindable(),
		iso = $bindable()
	}: Props = $props();

	function details(type: string): [string, string] {
		switch (type) {
			case 'zvol':
				return ['ZFS Volume', 'Block devices managed by ZFS'];
			case 'raw':
				return ['Raw Disk', 'Disk images that can be used with any filesystem'];
			case 'none':
				return ['No Storage', 'No storage will be allocated for this virtual machine'];
			default:
				return ['', ''];
		}
	}

	let isos = $derived.by(() => {
		const options = [];
		if (downloads && downloads.length > 0) {
			for (const download of downloads) {
				if (download.progress === 100) {
					if (download.type === 'http') {
						if (download.name.endsWith('.iso')) {
							options.push({
								label: download.name,
								value: download.uuid
							});
						}
					} else if (download.type === 'torrent') {
						for (const file of download.files) {
							if (file.name.endsWith('.iso')) {
								options.push({
									label: file.name,
									value: download.uuid
								});
							}
						}
					}
				}
			}
		}

		options.push({
			label: 'None',
			value: 'None'
		});

		return options;
	});

	let comboBoxes = $state({
		volumes: {
			open: false,
			options: [] as { label: string; value: string }[]
		},
		filesystems: {
			open: false,
			options: [] as { label: string; value: string }[]
		},
		emulationType: {
			open: false,
			value: 'virtio',
			options: [
				{ label: 'VirtIO', value: 'virtio-blk' },
				{
					label: 'AHCI-HD',
					value: 'ahci-hd'
				},
				{
					label: 'NVMe',
					value: 'nvme'
				}
			]
		},
		isos: {
			open: false,
			options: isos || []
		}
	});

	$effect(() => {
		if (isos) {
			comboBoxes.isos.options = isos.map((iso) => ({
				label: iso.label || iso.value,
				value: iso.value || ''
			}));
		}
	});

	$effect(() => {
		if (volumes || filesystems) {
			comboBoxes.volumes.options = volumes
				.filter((v) => v.properties.volmode && v.properties.volmode === 'dev')
				.map((v) => ({
					label: v.name,
					value: v.properties.guid || ''
				}));

			comboBoxes.filesystems.options = filesystems.map((fs) => ({
				label: fs.name,
				value: fs.properties.guid || ''
			}));
		}
	});

	let humanSize = $state('1024 M');

	$effect(() => {
		if (humanSize) {
			try {
				const parsed = humanFormat.parse.raw(humanSize);
				console.log(parsed);
				size = parsed.factor * parsed.value;
			} catch (e) {
				size = 1024;
			}
		}
	});
</script>

{#snippet radioItem(type: string)}
	<div class="mb-2 flex items-center space-x-3 rounded-lg border p-4">
		<RadioGroup.Item value={type} id={type} />
		<Label for={type} class="flex flex-col gap-2">
			<p>{details(type)[0]}</p>
			<p class="text-muted-foreground text-sm">
				{details(type)[1]}
			</p>
		</Label>
	</div>
{/snippet}

{#snippet storageDetail(type: string)}
	{#if type === 'zvol'}
		<CustomComboBox
			bind:open={comboBoxes.volumes.open}
			label="ZFS Volume"
			bind:value={guid}
			data={comboBoxes.volumes.options}
			classes="flex-1 space-y-1"
			placeholder="Select ZFS volume"
			width="w-[70%]"
		></CustomComboBox>
	{/if}

	{#if type === 'raw'}
		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
			<CustomValueInput
				label="Disk Size"
				placeholder="10G"
				bind:value={humanSize}
				classes="flex-1 space-y-1"
			/>

			<CustomComboBox
				bind:open={comboBoxes.filesystems.open}
				label="Filesystem Dataset"
				bind:value={guid}
				data={comboBoxes.filesystems.options}
				classes="flex-1 space-y-1"
				placeholder="Select filesystem"
				width="w-[60%]"
			></CustomComboBox>
		</div>
	{/if}

	<CustomComboBox
		bind:open={comboBoxes.emulationType.open}
		label="Emulation Type"
		bind:value={emulation}
		data={comboBoxes.emulationType.options}
		classes="flex-1 space-y-1"
		placeholder="Select emulation type"
		width="w-[40%]"
	></CustomComboBox>
{/snippet}

<div class="flex flex-col gap-4 p-4">
	<RadioGroup.Root bind:value={type} class="border p-2">
		<ScrollArea orientation="vertical" class="h-60 w-full max-w-full">
			{#each ['zvol', 'raw', 'none'] as storageType}
				{@render radioItem(storageType)}
			{/each}
		</ScrollArea>
	</RadioGroup.Root>

	{#if type !== 'none'}
		{@render storageDetail(type)}
	{/if}

	<CustomComboBox
		bind:open={comboBoxes.isos.open}
		label="Installation Media"
		bind:value={iso}
		data={comboBoxes.isos.options}
		classes="flex-1 space-y-1"
		placeholder="Select installation media"
		width="w-[70%]"
	></CustomComboBox>
</div>
