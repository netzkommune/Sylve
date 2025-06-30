<script lang="ts">
	import {
		bulkDeleteDownloads,
		deleteDownload,
		getDownloads,
		getSignedURL,
		startDownload
	} from '$lib/api/utilities/downloader';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { type APIResponse } from '$lib/types/common';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Download } from '$lib/types/utilities/downloader';
	import { handleValidationErrors, isAPIResponse, updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter, isDownloadURL } from '$lib/utils/string';
	import { generateTableData } from '$lib/utils/utilities/downloader';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';
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
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.downloads,
			onSuccess: (data: Download[]) => {
				updateCache('downloads', data);
			}
		}
	]);

	let modalState = $state({
		isOpen: false,
		isDelete: false,
		isBulkDelete: false,
		title: '',
		url: ''
	});

	let downloads = $derived($results[0].data as Download[]);
	let tableData = $derived(generateTableData(downloads));
	$inspect(tableData);
	let query: string = $state('');
	let activeRows: Row[] | null = $state(null);
	let onlyParentsSelected: boolean = $derived.by(() => {
		if (activeRows) {
			for (const row of activeRows) {
				if (row.type === '-') {
					return false;
				}
			}
		}

		return true;
	});

	let onlyChildSelected: boolean = $derived.by(() => {
		let hasParent = false;
		if (activeRows) {
			for (const row of activeRows) {
				if (row.type !== '-') {
					hasParent = true;
					break;
				}
			}
		}
		return !hasParent;
	});

	let httpDownloadSelected: boolean = $derived.by(() => {
		if (activeRows && activeRows.length === 1) {
			const row = activeRows[0];
			return row.type === 'http';
		}
		return false;
	});

	let isDownloadCompleted: boolean = $derived.by(() => {
		if (activeRows && activeRows.length === 1) {
			const row = activeRows[0];
			if (row.progress === '-') {
				const parent = downloads.find((d) => d.uuid === row.parentUUID);
				return parent ? parent.progress === 100 : false;
			} else if (row.progress === 100) {
				return true;
			}
		}
		return false;
	});

	async function newDownload() {
		if (!modalState.url) {
			toast.error('Please enter a valid URL', { position: 'bottom-center' });
			return;
		}

		if (!isMagnet(modalState.url) && !isDownloadURL(modalState.url)) {
			toast.error('Please enter a valid magnet or HTTP URL', { position: 'bottom-center' });
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

	async function handleDelete() {
		if (activeRows && activeRows.length == 1) {
			modalState.isDelete = true;
			modalState.title = activeRows[0].name;
		}

		if (activeRows && activeRows.length > 1) {
			for (const row of activeRows) {
				if (row.type !== '-') {
					modalState.isBulkDelete = false;
					modalState.title = '';
					return;
				}
			}
			modalState.isBulkDelete = true;
			modalState.title = `${activeRows.length} ${capitalizeFirstLetter(getTranslation('common.downloads', 'Downloads'))}`;
		}
	}

	async function handleDownload() {
		const row = activeRows ? activeRows[0] : null;
		if (row) {
			const result = await getSignedURL(row.name as string, (row.parentUUID as string) || row.uuid);
			if (isAPIResponse(result) && result.status === 'success') {
				const url = result.data as string;
				const link = document.createElement('a');
				link.href = url;
				link.download = row.name as string;
				document.body.appendChild(link);
				link.click();
			} else {
				handleValidationErrors(result as APIResponse, 'downloads');
			}
		}
	}
</script>

{#snippet button(type: string)}
	{#if type === 'delete' && onlyParentsSelected}
		{#if activeRows && activeRows.length >= 1}
			<Button
				onclick={handleDelete}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
				{#if activeRows.length > 1}
					{capitalizeFirstLetter(getTranslation('common.bulk_delete', 'Bulk Delete'))}
				{:else}
					{capitalizeFirstLetter(getTranslation('common.delete', 'Delete'))}
				{/if}
			</Button>
		{/if}
	{/if}

	{#if type === 'download' && onlyChildSelected && isDownloadCompleted}
		{#if activeRows && activeRows.length == 1}
			<Button
				onclick={handleDownload}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:download" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.download', 'Download'))}
			</Button>
		{/if}
	{/if}

	{#if type === 'download' && httpDownloadSelected && isDownloadCompleted}
		{#if activeRows && activeRows.length == 1}
			<Button
				onclick={handleDownload}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:download" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.download', 'Download'))}
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />

		<Button onclick={() => (modalState.isOpen = true)} size="sm" class="h-6  ">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>

		{@render button('delete')}
		{@render button('download')}
	</div>

	<Dialog.Root bind:open={modalState.isOpen}>
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
						onclick={() => {
							modalState.isOpen = true;
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
						onclick={() => {
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
				label={capitalizeFirstLetter(getTranslation('common.magnet', 'Magnet')) +
					' / ' +
					capitalizeFirstLetter(getTranslation('common.http', 'HTTP')) +
					' ' +
					capitalizeFirstLetter(getTranslation('common.url', 'URL'))}
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
		name="tt-downloader"
		multipleSelect={true}
		bind:parentActiveRow={activeRows}
		bind:query
	/>

	<AlertDialog
		open={modalState.isDelete}
		names={{ parent: 'download', element: modalState?.title || '' }}
		actions={{
			onConfirm: async () => {
				const id = activeRows ? activeRows[0]?.id : null;
				const result = await deleteDownload(id as number);
				if (isAPIResponse(result) && result.status === 'success') {
					modalState.isDelete = false;
					modalState.title = '';
					activeRows = null;
				} else {
					handleValidationErrors(result as APIResponse, 'downloads');
				}
			},
			onCancel: () => {
				modalState.isDelete = false;
				modalState.title = '';
			}
		}}
	></AlertDialog>

	<AlertDialog
		open={modalState.isBulkDelete}
		names={{ parent: 'download', element: modalState?.title || '' }}
		actions={{
			onConfirm: async () => {
				const ids = activeRows ? activeRows.map((row) => row.id) : [];
				const result = await bulkDeleteDownloads(ids as number[]);
				if (isAPIResponse(result) && result.status === 'success') {
					modalState.isBulkDelete = false;
					modalState.title = '';
					activeRows = null;
				} else {
					handleValidationErrors(result as APIResponse, 'downloads');
				}
			},
			onCancel: () => {
				modalState.isBulkDelete = false;
				modalState.title = '';
			}
		}}
	></AlertDialog>
</div>
