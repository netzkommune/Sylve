<script lang="ts">
	import { flashVolume } from '$lib/api/zfs/datasets';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Download } from '$lib/types/utilities/downloader';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import { sleep } from '$lib/utils';
	import { handleAPIError } from '$lib/utils/http';
	import { getISOs } from '$lib/utils/utilities/downloader';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		dataset: Dataset;
		downloads: Download[];
	}

	let { open = $bindable(), dataset, downloads }: Props = $props();
	let options = {
		select: {
			open: false,
			uuid: '',
			data: getISOs(downloads, true)
		},
		loading: false
	};

	let properties = $state(options);
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="p-5"
		onInteractOutside={(e) => e.preventDefault()}
		onEscapeKeydown={(e) => e.preventDefault()}
	>
		<div class="flex items-center justify-between">
			<Dialog.Header class="flex-1">
				<Dialog.Title>
					<div class="flex items-center">
						<Icon icon="mdi:usb-flash-drive-outline" class="mr-2 h-6 w-6" />
						Flash File to {dataset.name}
					</div>
				</Dialog.Title>
			</Dialog.Header>

			<div class="flex items-center gap-0.5">
				<Button
					size="sm"
					variant="ghost"
					class="h-8"
					title={'Reset'}
					onclick={() => {
						properties = options;
					}}
				>
					<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">Reset</span>
				</Button>
				<Button
					size="sm"
					variant="ghost"
					class="h-8"
					title={'Close'}
					onclick={() => {
						properties = options;
						open = false;
					}}
				>
					<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">Close</span>
				</Button>
			</div>
		</div>

		<div class="flex-1 space-y-1">
			<CustomComboBox
				bind:open={properties.select.open}
				label="Select File"
				bind:value={properties.select.uuid}
				data={properties.select.data}
				classes="flex-1 space-y-1.5"
				placeholder="File"
				triggerWidth="w-full"
				width="w-full"
			></CustomComboBox>
		</div>

		<Dialog.Footer class="flex justify-end">
			<div class="flex w-full items-center justify-end gap-2 px-1 py-2">
				<Button
					onclick={async () => {
						properties.loading = true;
						await sleep(1000);

						const response = await flashVolume(
							dataset.properties.guid || '',
							properties.select.uuid
						);

						if (response.status === 'error') {
							handleAPIError(response);
							toast.error('Error flashing volume', {
								position: 'bottom-center'
							});
						} else {
							toast.success(`${'Volume ' + dataset.name + ' flashed'}`, {
								position: 'bottom-center'
							});
						}

						properties = options;
						open = false;
					}}
					type="submit"
					size="sm"
					disabled={!properties.select.uuid || properties.loading}
				>
					{#if properties.loading}
						<Icon icon="mdi:loading" class="h-4 w-4 animate-spin" />
					{:else}
						<span>Flash</span>
					{/if}
				</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
