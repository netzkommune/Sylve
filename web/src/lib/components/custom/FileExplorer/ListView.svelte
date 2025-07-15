<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import type { FileNode } from '$lib/types/system/file-explorer';
	import { formatDistanceToNow, isToday, isYesterday } from 'date-fns';
	import {
		Archive,
		Copy,
		Download,
		Edit,
		FileText,
		Folder,
		FolderOpen,
		ImageIcon,
		MoreHorizontal,
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

	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function formatDate(dateInput: string | Date): string {
		const date = typeof dateInput === 'string' ? new Date(dateInput) : dateInput;

		if (isToday(date)) return 'Today';
		if (isYesterday(date)) return 'Yesterday';

		return formatDistanceToNow(date, { addSuffix: true });
	}
</script>

<div class="rounded-md border">
	<table class="w-full">
		<thead>
			<tr class=" border-b">
				<th class="text-muted-foreground p-3 text-left text-sm font-medium">Name</th>
				<th class="text-muted-foreground p-3 text-left text-sm font-medium">Size</th>
				<th class="text-muted-foreground p-3 text-left text-sm font-medium">Modified</th>
				<th class="w-12"></th>
			</tr>
		</thead>
		<tbody>
			{#each items as item, index}
				{@const itemName = item.id.split('/').pop() || item.id}
				{@const ItemIcon = item.type === 'folder' ? Folder : getFileIcon(itemName)}
				{@const isSelected = selectedItems.has(item.id)}
				<tr
					class="hover:bg-muted/50 group cursor-pointer border-b {isSelected ? 'bg-muted' : ''}"
					ondblclick={() => onItemClick(item)}
					onclick={(e) => onItemSelect(item, e)}
				>
					<td class="p-3">
						<div class="flex items-center gap-3">
							<ItemIcon class="text-muted-foreground h-4 w-4" />
							<span class="text-sm">{itemName}</span>
						</div>
					</td>
					<td class="p-3">
						<span class="text-muted-foreground text-sm">
							{item.type === 'file' ? formatFileSize(Number(item.size)) : '--'}
						</span>
					</td>
					<td class="p-3">
						<span class="text-muted-foreground text-sm">
							{item.type === 'file' ? formatDate(item.date) : '--'}
						</span>
					</td>
					<td class="p-3 text-right">
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button
									variant="ghost"
									size="sm"
									class="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
								>
									<MoreHorizontal class="h-4 w-4" />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content align="end">
								{#if item.type === 'folder'}
									<DropdownMenu.Item class="gap-2" onclick={() => onItemClick(item)}>
										<FolderOpen class="h-4 w-4" />
										Open
									</DropdownMenu.Item>
								{:else}
									<DropdownMenu.Item class="gap-2">
										<Download class="h-4 w-4" />
										Download
									</DropdownMenu.Item>
								{/if}
								<DropdownMenu.Item class="gap-2">
									<Copy class="h-4 w-4" />
									Copy
								</DropdownMenu.Item>
								<DropdownMenu.Item class="gap-2">
									<Scissors class="h-4 w-4" />
									Cut
								</DropdownMenu.Item>
								<DropdownMenu.Item class="gap-2">
									<Edit class="h-4 w-4" />
									Rename
								</DropdownMenu.Item>
								<DropdownMenu.Separator />
								<DropdownMenu.Item class="text-destructive gap-2">
									<Trash2 class="h-4 w-4" />
									Delete
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
