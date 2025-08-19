<script lang="ts">
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import type { Download } from '$lib/types/utilities/downloader';
	import type { Dataset } from '$lib/types/zfs/dataset';

	interface Props {
		filesystems: Dataset[];
		downloads: Download[];
		dataset: string;
		base: string;
	}

	let { filesystems, downloads, dataset = $bindable(), base = $bindable() }: Props = $props();

	let datasetOptions = $derived.by(() => {
		return filesystems
			.filter((fs) => fs.name.includes('/') && fs.used < 1024 * 1024)
			.map((fs) => ({
				label: fs.name,
				value: fs.guid || ''
			}));
	});

	let baseOptions = $derived.by(() => {
		return downloads
			.filter((download) => download.name.includes('txz'))
			.map((download) => ({
				label: download.name,
				value: download.uuid
			}));
	});

	let comboBoxes = $state({
		dataset: {
			open: false,
			options: [] as { label: string; value: string }[]
		},
		base: {
			open: false,
			options: [] as { label: string; value: string }[]
		}
	});
</script>

<div class="flex flex-col gap-4 p-4">
	<CustomComboBox
		bind:open={comboBoxes.dataset.open}
		label="Filesystem"
		bind:value={dataset}
		data={datasetOptions}
		classes="flex-1 space-y-1"
		placeholder="Select filesystem"
		triggerWidth="w-full "
		width="w-full"
	></CustomComboBox>

	<CustomComboBox
		bind:open={comboBoxes.base.open}
		label="Base"
		bind:value={base}
		data={baseOptions}
		classes="flex-1 space-y-1"
		placeholder="Select base"
		triggerWidth="w-full"
		width="w-full"
	></CustomComboBox>
</div>
