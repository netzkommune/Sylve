<script lang="ts">
	import { getTokenHash } from '$lib/api/auth';
	import { handleAPIResponse } from '$lib/api/common';
	import {
		addFileOrFolder,
		copyOrMoveFileOrFolder,
		copyOrMoveFilesOrFolders,
		deleteFilesOrFolders,
		getFiles,
		renameFileOrFolder
	} from '$lib/api/system/file-explorer';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import Breadcrumb from '$lib/components/custom/FileExplorer/Breadcrumb.svelte';
	import CreateFileOrFolderModal from '$lib/components/custom/FileExplorer/CreateFileOrFolderModal.svelte';
	import GridView from '$lib/components/custom/FileExplorer/GridView.svelte';
	import ListView from '$lib/components/custom/FileExplorer/ListView.svelte';
	import RenameModal from '$lib/components/custom/FileExplorer/RenameModal.svelte';
	import Toolbar from '$lib/components/custom/FileExplorer/Toolbar.svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import { explorerCurrentPath } from '$lib/stores/basic';
	import type { FileNode } from '$lib/types/system/file-explorer';
	import { sortFileItems, type SortBy } from '$lib/utils/explorer';
	import { Clipboard, FileText, Folder, RotateCcw, UploadIcon } from 'lucide-svelte';
	import { get } from 'svelte/store';

	interface Data {
		files: FileNode[];
	}

	let { data }: { data: Data } = $props();

	let viewMode = $state<'grid' | 'list'>('grid');
	let searchQuery = $state('');
	let currentPath = $state(get(explorerCurrentPath));
	let folderData = $state<{ [path: string]: FileNode[] }>({ '/': data.files });
	let selectedItems = $state<string[]>([]);
	let sortBy = $state<SortBy>('name-asc');

	let copyFileOrFolder = $state({
		items: [] as string[],
		isCut: false
	});

	let modals = $state({
		create: {
			isOpen: false,
			isFolder: true,
			name: ''
		},
		delete: {
			isOpen: false,
			item: null as FileNode | null
		},
		rename: {
			isOpen: false,
			id: '',
			newName: ''
		}
	});

	function findItemsInPath(path: string) {
		return folderData[path] || [];
	}

	let currentItems = $derived(findItemsInPath(currentPath));

	let filteredItems = $derived(
		currentItems.filter((item) => {
			const itemName = item.id.split('/').pop() || item.id;
			return itemName.toLowerCase().includes(searchQuery.toLowerCase());
		})
	);

	let sortedItems = $derived(sortFileItems(filteredItems, sortBy));

	let breadcrumbItems = $derived.by(() => {
		const parts = currentPath.split('/').filter(Boolean);
		const items = [];

		items.push({ name: 'My Files', path: '/', isLast: parts.length === 0 });

		let currentBreadcrumbPath = '';
		for (let i = 0; i < parts.length; i++) {
			currentBreadcrumbPath += '/' + parts[i];
			items.push({
				name: parts[i],
				path: currentBreadcrumbPath,
				isLast: i === parts.length - 1
			});
		}
		return items;
	});

	async function handleItemClick(item: any) {
		if (item.type === 'folder') {
			searchQuery = '';
			currentPath = item.id;
			await loadFolderData(item.id);
		}
	}

	function handleItemSelect(item: FileNode, event?: MouseEvent) {
		const isSelected = selectedItems.includes(item.id);

		if (event?.ctrlKey || event?.metaKey) {
			selectedItems = isSelected
				? selectedItems.filter((id) => id !== item.id)
				: [...selectedItems, item.id];
		} else if (event?.shiftKey && selectedItems.length > 0) {
			const currentIndex = sortedItems.findIndex((i) => i.id === item.id);
			const lastSelectedIndex = sortedItems.findIndex(
				(i) => i.id === selectedItems[selectedItems.length - 1]
			);

			if (lastSelectedIndex !== -1) {
				const start = Math.min(currentIndex, lastSelectedIndex);
				const end = Math.max(currentIndex, lastSelectedIndex);
				const rangeIds = sortedItems.slice(start, end + 1).map((i) => i.id);
				selectedItems = [...new Set([...selectedItems, ...rangeIds])];
			}
		} else {
			selectedItems = isSelected && selectedItems.length === 1 ? [] : [item.id];
		}
	}

	$effect(() => {
		explorerCurrentPath.set(currentPath);
		selectedItems = [];
		if (currentPath !== '/' && !folderData[currentPath]) {
			loadFolderData(currentPath);
		}
	});

	function handleBackClick() {
		if (currentPath === '/') return;

		const parts = currentPath.split('/').filter(Boolean);
		if (parts.length > 1) {
			parts.pop();
			currentPath = '/' + parts.join('/');
		} else {
			currentPath = '/';
		}
	}

	async function loadFolderData(folderId: string) {
		try {
			const response = await getFiles(folderId);
			folderData = { [folderId]: response };
		} catch (error) {
			console.error('Error loading folder data:', error);
			folderData = { [folderId]: [] };
		}
	}

	async function createFileOrFolder() {
		let name = modals.create.name;
		let isFolder = modals.create.isFolder;
		const response = await addFileOrFolder(currentPath, name, isFolder);

		delete folderData[currentPath];
		await loadFolderData(currentPath);

		handleAPIResponse(response, {
			success: `${isFolder ? 'Folder' : 'File'} "${name}" created successfully`,
			error: `Failed to create ${isFolder ? 'folder' : 'file'} "${name}"`
		});

		modals.create.name = '';
	}

	async function handleDeleteFileOrFolder(item: FileNode) {
		modals.delete.item = item;
		modals.delete.isOpen = true;
	}

	async function refreshCurrentFolder() {
		delete folderData[currentPath];
		await loadFolderData(currentPath);
		selectedItems = [];
	}

	function handleEmptySpaceInteraction(e: MouseEvent) {
		const target = e.target as HTMLElement;

		const hasFileItemClasses =
			target.classList.contains('group') ||
			target.classList.contains('cursor-pointer') ||
			target.closest('.group.cursor-pointer') ||
			target.closest('[title]');

		const isContainerElement =
			target.classList.contains('grid-container') ||
			target.classList.contains('list-container') ||
			target.classList.contains('file-explorer-container') ||
			target.classList.contains('grid');

		if (!hasFileItemClasses && (isContainerElement || target === e.currentTarget)) {
			selectedItems = [];
		}
	}

	async function downloadFile(item: FileNode) {
		if (item.type !== 'file') return;

		const hash = await getTokenHash();
		const downloadUrl = `/api/system/file-explorer/download?id=${encodeURIComponent(item.id)}&hash=${hash}`;
		const filename = item.id.split('/').pop() || 'download';

		try {
			const link = Object.assign(document.createElement('a'), {
				href: downloadUrl,
				download: filename,
				style: 'display:none'
			});
			document.body.appendChild(link);
			link.click();
			link.remove();
		} catch (error) {
			console.error('Download failed:', error);
			window.open(downloadUrl, '_blank');
		}
	}

	async function handleCopyFileOrFolder(item: FileNode, isCut: boolean) {
		// Use selected items if multiple items are selected, otherwise use the single item
		const itemsToCopy = selectedItems.length > 0 ? selectedItems : [item.id];
		copyFileOrFolder.items = itemsToCopy;
		copyFileOrFolder.isCut = isCut;
	}

	async function pasteFileOrFolder() {
		if (!copyFileOrFolder.items || copyFileOrFolder.items.length === 0) return;

		const requestData: [string, string][] = copyFileOrFolder.items.map((itemId) => [
			itemId,
			currentPath
		]);

		await copyOrMoveFilesOrFolders(requestData, copyFileOrFolder.isCut);

		delete folderData[currentPath];
		await loadFolderData(currentPath);

		copyFileOrFolder.items = [];
		copyFileOrFolder.isCut = false;
	}

	async function handleRenameFileOrFolder(item: FileNode) {
		modals.rename.id = item.id;
		modals.rename.isOpen = true;
		let name = item.id.split('/').pop() || item.id;
		modals.rename.newName = name;
	}

	async function handleBreadcrumbNavigate(path: string) {
		currentPath = path;
		await loadFolderData(path);
	}

	async function rename() {
		if (!modals.rename.id || !modals.rename.newName) return;

		const response = await renameFileOrFolder(modals.rename.id, modals.rename.newName);
		delete folderData[currentPath];
		await loadFolderData(currentPath);
		handleAPIResponse(response, {
			success: 'Renamed successfully',
			error: response.error || 'Failed to rename'
		});
		modals.rename.isOpen = false;
		modals.rename.id = '';
		modals.rename.newName = '';
	}
</script>

<div class="flex h-full">
	<div class="flex flex-1 flex-col">
		<Breadcrumb
			onBackClick={handleBackClick}
			{currentPath}
			items={breadcrumbItems}
			onNavigate={handleBreadcrumbNavigate}
		/>

		<Toolbar
			{searchQuery}
			{sortBy}
			{viewMode}
			onSearchChange={(value) => (searchQuery = value)}
			onSortChange={(sort) => (sortBy = sort)}
			onViewModeChange={(mode) => (viewMode = mode)}
			onCreateFile={() => {
				modals.create.isFolder = false;
				modals.create.isOpen = true;
			}}
			onCreateFolder={() => {
				modals.create.isFolder = true;
				modals.create.isOpen = true;
			}}
			onUploadFile={() => {
				// TODO: Implement upload functionality
			}}
		/>

		<ContextMenu.Root>
			<ContextMenu.Trigger
				class="file-explorer-container flex-1 overflow-y-auto"
				onclick={handleEmptySpaceInteraction}
				oncontextmenu={handleEmptySpaceInteraction}
			>
				{#if viewMode === 'grid'}
					<div class="grid-container h-full w-full">
						<GridView
							items={sortedItems}
							onItemClick={handleItemClick}
							onItemSelect={handleItemSelect}
							selectedItems={new Set(selectedItems)}
							onItemDelete={handleDeleteFileOrFolder}
							onItemDownload={downloadFile}
							onItemCopy={handleCopyFileOrFolder}
							onItemRename={handleRenameFileOrFolder}
							isCopying={copyFileOrFolder.items.length > 0}
						/>
					</div>
				{:else}
					<div class="list-container h-full w-full">
						<ListView
							items={sortedItems}
							onItemClick={handleItemClick}
							onItemSelect={handleItemSelect}
							selectedItems={new Set(selectedItems)}
							onItemDelete={handleDeleteFileOrFolder}
							onItemDownload={downloadFile}
							onItemCopy={handleCopyFileOrFolder}
							onItemRename={handleRenameFileOrFolder}
							isCopying={copyFileOrFolder.items.length > 0}
						/>
					</div>
				{/if}
			</ContextMenu.Trigger>
			<ContextMenu.Content>
				<ContextMenu.Item class="gap-2" onclick={refreshCurrentFolder}>
					<RotateCcw />
					Refresh</ContextMenu.Item
				>
				{#if copyFileOrFolder.items.length > 0}
					<ContextMenu.Item class="gap-2" onclick={pasteFileOrFolder}>
						<Clipboard class="h-4 w-4" />
						Paste
					</ContextMenu.Item>
				{/if}
				<ContextMenu.Item
					class="gap-2"
					onclick={() => {
						modals.create.isFolder = false;
						modals.create.isOpen = true;
					}}
					><FileText />
					New File
				</ContextMenu.Item>
				<ContextMenu.Item
					class="gap-2"
					onclick={() => {
						modals.create.isFolder = true;
						modals.create.isOpen = true;
					}}
					><Folder />
					New Folder
				</ContextMenu.Item>
				<ContextMenu.Item class="gap-2">
					<UploadIcon />
					Upload File</ContextMenu.Item
				>
			</ContextMenu.Content>
		</ContextMenu.Root>

		<div class="bg-muted/30 flex flex-shrink-0 items-center justify-between border-t px-4 py-1">
			<div class="text-muted-foreground flex items-center gap-4 text-sm">
				<span>{sortedItems.length} items</span>
			</div>
			<div class="text-muted-foreground text-sm">
				{sortedItems.filter((item: any) => item.type === 'folder').length} folders,
				{sortedItems.filter((item: any) => item.type === 'file').length} files
			</div>
		</div>
	</div>
</div>

<CreateFileOrFolderModal
	bind:isOpen={modals.create.isOpen}
	bind:isFolder={modals.create.isFolder}
	bind:name={modals.create.name}
	onClose={() => {
		modals.create.isOpen = false;
		modals.create.isFolder = true;
	}}
	onReset={() => {
		modals.create.name = '';
	}}
	onCreate={() => {
		createFileOrFolder();
		modals.create.isOpen = false;
		modals.create.isFolder = true;
	}}
/>

<RenameModal
	bind:isOpen={modals.rename.isOpen}
	bind:newName={modals.rename.newName}
	onClose={() => {
		modals.rename.isOpen = false;
		modals.rename.id = '';
	}}
	onReset={() => {
		modals.rename.newName = modals.rename.id.split('/').pop() || '';
	}}
	onRename={() => {
		rename();
	}}
/>

<AlertDialog
	open={modals.delete.isOpen}
	names={{
		parent: modals.delete.item?.type === 'folder' ? 'folder' : 'file',
		element: modals.delete.item?.id.split('/').pop() || ''
	}}
	actions={{
		onConfirm: async () => {
			if (modals.delete.item) {
				const response = await deleteFilesOrFolders(selectedItems);

				delete folderData[currentPath];
				await loadFolderData(currentPath);
				handleAPIResponse(response, {
					success: `${modals.delete.item?.type === 'folder' ? 'Folder' : 'File'} ${modals.delete.item?.id.split('/').pop() || ''} deleted`,
					error: `Failed to delete ${modals.delete.item?.type === 'folder' ? 'folder' : 'file'}`
				});
			}
			modals.delete.isOpen = false;
			modals.delete.item = null;
		},
		onCancel: () => {
			modals.delete.isOpen = false;
			modals.delete.item = null;
		}
	}}
></AlertDialog>
