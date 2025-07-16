<script lang="ts">
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import type { FileNode } from '$lib/types/system/file-explorer';
	import { getFileIcon } from '$lib/utils/icons';
	import { Copy, Download, Edit, Folder, FolderOpen, Scissors, Trash2 } from 'lucide-svelte';

	interface Props {
		items: FileNode[];
		onItemClick: (item: FileNode) => void;
		onItemSelect: (item: FileNode, event?: MouseEvent) => void;
		selectedItems: Set<string>;
		onItemDelete?: (item: FileNode) => void;
	}

	let { items, onItemClick, onItemSelect, selectedItems, onItemDelete }: Props = $props();
</script>

<div class="rounded-md">
	<div class="border-b p-3">
		<div class="text-muted-foreground text-sm font-medium">Name</div>
	</div>
	<div>
		{#each items as item}
			{@const itemName = item.id.split('/').pop() || item.id}
			{@const ItemIcon = item.type === 'folder' ? Folder : getFileIcon(itemName)}
			{@const isSelected = selectedItems.has(item.id)}
			<ContextMenu.Root>
				<ContextMenu.Trigger
					class="hover:bg-muted/50 group flex w-full cursor-pointer items-center justify-between border-b px-3 py-2 {isSelected
						? 'bg-muted'
						: ''}"
					ondblclick={() => onItemClick(item)}
					onclick={(e) => onItemSelect(item, e)}
					oncontextmenu={(e) => {
						if (!selectedItems.has(item.id)) {
							onItemSelect(item, e);
						}
					}}
				>
					<div class="flex items-center gap-3">
						<ItemIcon class="text-muted-foreground h-4 w-4" />
						<span class="text-sm">{itemName}</span>
					</div>
				</ContextMenu.Trigger>
				<ContextMenu.Content>
					{#if item.type === 'folder'}
						<ContextMenu.Item class="gap-2" onclick={() => onItemClick(item)}>
							<FolderOpen class="h-4 w-4" />
							Open
						</ContextMenu.Item>
					{:else}
						<ContextMenu.Item class="gap-2">
							<Download class="h-4 w-4" />
							Download
						</ContextMenu.Item>
					{/if}
					<ContextMenu.Item class="gap-2">
						<Copy class="h-4 w-4" />
						Copy
					</ContextMenu.Item>
					<ContextMenu.Item class="gap-2">
						<Scissors class="h-4 w-4" />
						Cut
					</ContextMenu.Item>
					<ContextMenu.Item class="gap-2">
						<Edit class="h-4 w-4" />
						Rename
					</ContextMenu.Item>
					<ContextMenu.Item class=" gap-2" onclick={() => onItemDelete?.(item)}>
						<Trash2 class="h-4 w-4" />
						Delete
					</ContextMenu.Item>
				</ContextMenu.Content>
			</ContextMenu.Root>
		{/each}
	</div>
</div>
