<script lang="ts">
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import type { FileNode } from '$lib/types/system/file-explorer';
	import Icon from '@iconify/svelte';
	import {
		Archive,
		Copy,
		Download,
		Edit,
		FileText,
		FolderOpen,
		ImageIcon,
		Music,
		Scissors,
		Trash2,
		Video
	} from 'lucide-svelte';

	interface Props {
		items: FileNode[];
		onItemClick: (item: FileNode) => void;
		onItemSelect: (item: FileNode, event?: MouseEvent) => void;
		selectedItems: Set<string>;
		onItemDelete?: (item: FileNode) => void;
	}

	let { items, onItemClick, onItemSelect, selectedItems, onItemDelete }: Props = $props();

	function getFileIcon(filename: string) {
		const ext = filename.split('.').pop()?.toLowerCase() || '';
		switch (ext) {
			case 'jpg':
			case 'jpeg':
			case 'png':
			case 'gif':
			case 'bmp':
			case 'svg':
				return ImageIcon;
			case 'mp4':
			case 'avi':
			case 'mkv':
			case 'mov':
			case 'wmv':
				return Video;
			case 'mp3':
			case 'wav':
			case 'flac':
			case 'ogg':
				return Music;
			case 'zip':
			case 'tar':
			case 'gz':
			case 'rar':
			case '7z':
				return Archive;
			case 'exe':
			case 'sh':
			case 'bin':
				return FileText;
			case 'pdf':
			case 'doc':
			case 'docx':
			case 'txt':
			case 'md':
			case 'html':
			case 'css':
			case 'js':
			case 'ts':
			case 'json':
			case 'xml':
			case 'cshrc':
			case 'profile':
			default:
				return FileText;
		}
	}
</script>

<div
	class="grid grid-cols-2 gap-4 p-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 2xl:grid-cols-10"
>
	{#each items as item}
		{@const itemName = item.id.split('/').pop() || item.id}
		{@const FileIcon = getFileIcon(itemName)}
		{@const isSelected = selectedItems.has(item.id)}

		<ContextMenu.Root>
			<ContextMenu.Trigger
				title={itemName}
				class="group relative flex w-full cursor-pointer flex-col items-center rounded-lg p-3 {isSelected
					? 'bg-muted border-secondary'
					: 'hover:bg-muted/50'}"
				ondblclick={() => onItemClick(item)}
				onclick={(e) => onItemSelect(item, e)}
				oncontextmenu={(e) => {
					if (!selectedItems.has(item.id)) {
						onItemSelect(item, e);
					}
				}}
			>
				{#if item.type === 'folder'}
					<Icon
						icon="material-symbols:folder-rounded"
						class="mb-2 h-12 w-12 flex-shrink-0 text-blue-400"
					/>
				{:else}
					<FileIcon class="mb-2 h-12 w-12 flex-shrink-0 text-blue-400" />
				{/if}
				<span
					class="line-clamp-2 w-full break-words px-1 text-center text-xs font-medium leading-tight"
					>{itemName}</span
				>
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
				<ContextMenu.Item
					class=" gap-2"
					onclick={() => {
						if (onItemDelete) {
							onItemDelete(item);
						}
					}}
				>
					<Trash2 class="h-4 w-4" />
					Delete
				</ContextMenu.Item>
			</ContextMenu.Content>
		</ContextMenu.Root>
	{/each}
</div>
