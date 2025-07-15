<script lang="ts">
	import { getFiles } from '$lib/api/system/file-explorer';
	import GridView from '$lib/components/custom/FileExplorer/GridView.svelte';
	import ListView from '$lib/components/custom/FileExplorer/ListView.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Input } from '$lib/components/ui/input';
	import type { FileNode } from '$lib/types/system/file-explorer';
	import Icon from '@iconify/svelte';
	import { Grid3X3, List, Plus, Search, Upload } from 'lucide-svelte';

	interface Data {
		files: FileNode[];
	}

	let { data }: { data: Data } = $props();

	let viewMode = $state<'grid' | 'list'>('grid');
	let searchQuery = $state('');
	let currentPath = $state('/');
	let folderData = $state<{ [path: string]: FileNode[] }>({ '/': data.files });
	let selectedItems = $state<string[]>([]);

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
			const currentIndex = filteredItems.findIndex((i) => i.id === item.id);
			const lastSelectedIndex = filteredItems.findIndex(
				(i) => i.id === selectedItems[selectedItems.length - 1]
			);

			if (lastSelectedIndex !== -1) {
				const start = Math.min(currentIndex, lastSelectedIndex);
				const end = Math.max(currentIndex, lastSelectedIndex);
				const rangeIds = filteredItems.slice(start, end + 1).map((i) => i.id);
				selectedItems = [...new Set([...selectedItems, ...rangeIds])];
			}
		} else {
			selectedItems = isSelected && selectedItems.length === 1 ? [] : [item.id];
		}
	}

	$effect(() => {
		selectedItems = [];
	});

	function handleBreadcrumbClick(path: string) {
		currentPath = path;
	}

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
		if (folderData[folderId]) {
			return;
		}
		try {
			const response = await getFiles(folderId);
			folderData = { ...folderData, [folderId]: response };
		} catch (error) {
			console.error('Error loading folder data:', error);
			folderData = { ...folderData, [folderId]: [] };
		}
	}
</script>

<div class="flex h-full">
	<div class="flex flex-1 flex-col">
		<header class="flex shrink-0 items-center gap-2 border-b px-4">
			<Button
				variant="ghost"
				size="icon"
				class="cursor-pointer"
				onclick={handleBackClick}
				disabled={currentPath === '/'}
			>
				<Icon icon="tabler:arrow-left" class="pointer-events-none !h-6 !w-6" />
			</Button>

			<Breadcrumb.Root>
				<Breadcrumb.List>
					{#each breadcrumbItems as item, index}
						<Breadcrumb.Item>
							{#if item.isLast}
								<Breadcrumb.Page>{item.name}</Breadcrumb.Page>
							{:else}
								<Breadcrumb.Link
									href="#"
									onclick={(e: any) => {
										e.preventDefault();
										handleBreadcrumbClick(item.path);
									}}
								>
									{item.name}
								</Breadcrumb.Link>
							{/if}
						</Breadcrumb.Item>
						{#if !item.isLast}
							<Breadcrumb.Separator />
						{/if}
					{/each}
				</Breadcrumb.List>
			</Breadcrumb.Root>
		</header>

		<div class="flex flex-shrink-0 items-center justify-between gap-4 border-b px-4 py-2">
			<div class="flex items-center gap-2">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						><Button size="sm" class="gap-2"><Plus class="h-4 w-4" />Add New</Button
						></DropdownMenu.Trigger
					>
					<DropdownMenu.Content>
						<DropdownMenu.Group>
							<DropdownMenu.Item>Add New File</DropdownMenu.Item>
							<DropdownMenu.Item>Add New Foilder</DropdownMenu.Item>
							<DropdownMenu.Item>Upload File</DropdownMenu.Item>
						</DropdownMenu.Group>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</div>

			<div class="flex items-center gap-2">
				<div class="relative">
					<Search class="text-muted-foreground absolute left-2 top-2.5 h-4 w-4" />
					<Input placeholder="Search files..." bind:value={searchQuery} class="w-64 pl-8" />
				</div>
				<div class="flex items-center rounded-md border">
					<Button
						variant={viewMode === 'grid' ? 'default' : 'ghost'}
						size="sm"
						onclick={() => (viewMode = 'grid')}
						class="rounded-r-none"
					>
						<Grid3X3 class="h-4 w-4" />
					</Button>
					<Button
						variant={viewMode === 'list' ? 'default' : 'ghost'}
						size="sm"
						onclick={() => (viewMode = 'list')}
						class="rounded-l-none"
					>
						<List class="h-4 w-4" />
					</Button>
				</div>
			</div>
		</div>

		<div class="flex-1 overflow-y-auto">
			{#if viewMode === 'grid'}
				<GridView
					items={filteredItems}
					onItemClick={handleItemClick}
					onItemSelect={handleItemSelect}
					selectedItems={new Set(selectedItems)}
				/>
			{:else}
				<ListView
					items={filteredItems}
					onItemClick={handleItemClick}
					onItemSelect={handleItemSelect}
					selectedItems={new Set(selectedItems)}
				/>
			{/if}
		</div>

		<div class="bg-muted/30 flex flex-shrink-0 items-center justify-between border-t px-4 py-1">
			<div class="text-muted-foreground flex items-center gap-4 text-sm">
				<span>{filteredItems.length} items</span>
			</div>
			<div class="text-muted-foreground text-sm">
				{filteredItems.filter((item: any) => item.type === 'folder').length} folders,
				{filteredItems.filter((item: any) => item.type === 'file').length} files
			</div>
		</div>
	</div>
</div>
