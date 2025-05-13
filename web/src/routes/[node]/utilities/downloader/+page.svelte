<script lang="ts">
	import { getDownloads, startDownload } from '$lib/api/utilities/downloader';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Download } from '$lib/types/utilities/downloader';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { generateTableData } from '$lib/utils/utilities/downloader';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import toast from 'svelte-french-toast';
	import isMagnet from 'validator/lib/isMagnetURI';

	interface Data {
		downloads: Download[];
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['downloads'],
			queryFn: async () => {
				return await getDownloads();
			},
			refetchInterval: false,
			keepPreviousData: true,
			initialData: data.downloads,
			onSuccess: (data: Download[]) => {
				updateCache('downloads', data);
			}
		}
	]);

	let modalState = $state({
		isOpen: false,
		url: ''
	});

	let downloads = $derived($results[0].data as Download[]);
	let tableData = $derived(generateTableData(downloads));
	$inspect(tableData);
	let query: string = $state('');
	let activeRows: Row[] | null = $state(null);

	async function newDownload() {
		if (!modalState.url) {
			toast.error('Please enter a valid URL', { position: 'bottom-center' });
			return;
		}

		if (!isMagnet(modalState.url)) {
			toast.error('Please enter a valid magnet link', { position: 'bottom-center' });
			return;
		}

		const result = await startDownload(modalState.url);
		if (result) {
			modalState.isOpen = false;
			modalState.url = '';
			toast.success('Download started', { position: 'bottom-center' });
		} else {
			toast.error('Download failed', { position: 'bottom-center' });
		}
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />

		<Button onclick={() => (modalState.isOpen = true)} size="sm" class="h-6  ">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>
	</div>

	<Dialog.Root bind:open={modalState.isOpen} closeOnOutsideClick={false}>
		<Dialog.Content class="w-[80%] gap-0 overflow-hidden p-3 lg:max-w-xl">
			<div class="flex items-center justify-between py-1 pb-2">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						<div class="flex items-center gap-2">
							<Icon icon="mdi:download" class="text-primary h-5 w-5" />
							{capitalizeFirstLetter(getTranslation('common.download', 'Download'))}
						</div>
					</Dialog.Title>
				</Dialog.Header>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}
						on:click={() => {
							modalState.isOpen = false;
							modalState.url = '';
						}}
					>
						<Icon icon="radix-icons:reset" class="h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.reset', 'Reset'))}</span
						>
					</Button>
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						title={capitalizeFirstLetter(getTranslation('common.close', 'Close'))}
						on:click={() => {
							modalState.isOpen = false;
							modalState.url = '';
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

			<CustomValueInput
				label={capitalizeFirstLetter(getTranslation('common.url', 'URL'))}
				placeholder="magnet:?xt=urn:btih:7d5210a711291d7181d6e074ce5ebd56f3fedd60&dn=debian-12.10.0-amd64-netinst.iso&xl=663748608&tr=http%3A%2F%2Fbttracker.debian.org%3A6969%2Fannounce"
				bind:value={modalState.url}
				classes="flex-1 space-y-1"
			/>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-end gap-2 px-1 py-2">
					<Button onclick={newDownload} type="submit" size="sm"
						>{capitalizeFirstLetter(getTranslation('common.download', 'Download'))}</Button
					>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>

	<TreeTable
		data={tableData}
		name="tt-networkInterfaces"
		multipleSelect={false}
		bind:parentActiveRow={activeRows}
		bind:query
	/>
</div>
