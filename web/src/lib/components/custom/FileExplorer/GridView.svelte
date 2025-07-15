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
	}

	let { items, onItemClick, onItemSelect, selectedItems }: Props = $props();

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

<div class="grid grid-cols-2 gap-4 p-2 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-10">
	{#each items as item}
		{@const itemName = item.id.split('/').pop() || item.id}
		{@const FileIcon = getFileIcon(itemName)}
		{@const isSelected = selectedItems.has(item.id)}

		<ContextMenu.Root>
			<ContextMenu.Trigger
				class="group relative flex cursor-pointer flex-col items-center rounded-lg p-4 {isSelected
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
					<Icon icon="material-symbols:folder-rounded" class="h-12 w-12 text-blue-400" />
				{:else}
					<FileIcon class="h-12 w-12 text-blue-400" />
				{/if}
				<span class="line-clamp-2 text-center text-sm font-medium">{itemName}</span>
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
				<ContextMenu.Separator />
				<ContextMenu.Item class="text-destructive gap-2">
					<Trash2 class="h-4 w-4" />
					Delete
				</ContextMenu.Item>
			</ContextMenu.Content>
		</ContextMenu.Root>
	{/each}
</div>
