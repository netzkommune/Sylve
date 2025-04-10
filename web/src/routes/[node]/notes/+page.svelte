<script lang="ts">
	import { createNote, deleteNote, updateNote } from '$lib/api/info/notes';
	import AlertDialog from '$lib/components/custom/AlertDialog.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import ScrollArea from '$lib/components/ui/scroll-area/scroll-area.svelte';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import type { APIResponse } from '$lib/types/common';
	import type { Note, Notes } from '$lib/types/info/notes';
	import { handleValidationErrors, isAPIResponse } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { marked } from 'marked';
	import toast from 'svelte-french-toast';

	interface Data {
		notes: Notes;
	}

	let { data }: { data: Data } = $props();

	let tableData = $state(data.notes);
	let modalState = $state({
		title: '',
		content: '',
		isOpen: false,
		isEditMode: false,
		isDeleteOpen: false
	});

	let selectedId: number | null = $state(null);
	let toRemove: Note | null = $state(null);

	function handleNote(note?: Note, editMode: boolean = true, reset: boolean = false) {
		if (reset) {
			modalState.title = '';
			modalState.content = '';
			selectedId = null;
			modalState.isEditMode = false;
			modalState.isOpen = false;
		} else {
			modalState.title = note?.title || '';
			modalState.content = note?.content || '';
			selectedId = note?.id || null;
			modalState.isEditMode = editMode;
			modalState.isOpen = true;
		}
	}

	async function saveNote() {
		if (!modalState.title.trim() || !modalState.content.trim()) return;

		const noteData = { title: modalState.title, content: modalState.content };
		let result: APIResponse | Note;

		if (modalState.isEditMode && selectedId !== null) {
			result = await updateNote(selectedId, noteData.title, noteData.content);
			if (isAPIResponse(result) && result.status === 'success') {
				tableData = [
					...tableData.map((note) => (note.id === selectedId ? { ...note, ...noteData } : note))
				];
			}
		} else {
			result = await createNote(noteData.title, noteData.content);
			if (!isAPIResponse(result)) {
				tableData = [...tableData, result];
			}
		}

		if (isAPIResponse(result) && result.status === 'error') {
			handleValidationErrors(result, 'notes');
		} else {
			handleNote(undefined, false, true);
		}
	}

	async function handleRemoveNote(id?: number, confirm: boolean = false) {
		if (!confirm) {
			selectedId = id as number;
			toRemove = tableData.find((note) => note.id === id) || null;
			modalState.isDeleteOpen = true;
		} else {
			const response = (await deleteNote(selectedId as number)) as APIResponse;
			if (response.status === 'success') {
				tableData = tableData.filter((note) => note.id !== selectedId);
			} else {
				toast.error(response.message || 'Failed to delete note', {
					position: 'bottom-center'
				});
			}
			modalState.isDeleteOpen = false;
		}
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button onclick={() => handleNote()} size="sm" class="h-6  ">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>

	<Dialog.Root bind:open={modalState.isOpen}>
		<Dialog.Content class="w-[80%] gap-0 overflow-hidden p-3 lg:max-w-3xl">
			<div class="flex items-center justify-between py-2">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						{modalState.isEditMode ? (selectedId ? 'Edit Note' : 'New Note') : 'View Note'}
					</Dialog.Title>
				</Dialog.Header>
				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						on:click={() => {
							modalState.title = '';
							modalState.content = '';
						}}
					>
						<Icon icon="radix-icons:reset" class="h-4 w-4" />
						<span class="sr-only">Reset</span>
					</Button>
					<Dialog.Close
						class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
					>
						<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
					</Dialog.Close>
				</div>
			</div>

			<div>
				<Label for="title">Title</Label>
				<div class="p-1">
					<Input
						type="text"
						id="title"
						bind:value={modalState.title}
						class="mt-2"
						placeholder="Enter title"
						disabled={!modalState.isEditMode}
					/>
				</div>
			</div>

			<div class="mt-2">
				<Label for="content">Content</Label>
				<ScrollArea orientation="vertical" class="h-96">
					{#if modalState.isEditMode}
						<div class="p-1">
							<Textarea
								bind:value={modalState.content}
								class="mt-2 h-[360px] w-full"
								placeholder="Write in Markdown format"
							/>
						</div>
					{:else}
						<div class="p-1">
							<article class="prose lg:prose-xl">
								{@html marked(modalState.content)}
							</article>
						</div>
					{/if}
				</ScrollArea>
			</div>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-between gap-2 px-3 py-2">
					<Button
						type="submit"
						size="sm"
						class="h-8 bg-muted text-white hover:bg-muted-foreground/50"
					>
						Help
					</Button>

					<div class="flex gap-2">
						<Dialog.Close>
							<Button
								variant="outline"
								class="h-8"
								on:click={() => handleNote(undefined, false, true)}>Cancel</Button
							>
						</Dialog.Close>
						{#if modalState.isEditMode}
							<Button
								onclick={saveNote}
								type="submit"
								size="sm"
								class="h-8 w-16 bg-blue-600 text-white hover:bg-blue-500"
							>
								Ok
							</Button>
						{/if}
					</div>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>

	<div class="flex h-full flex-col overflow-hidden">
		<Table.Root class="w-full table-fixed border-collapse">
			<Table.Header class="sticky top-0 z-[50] bg-background">
				<Table.Row>
					<Table.Head class="h-10 px-4 py-2">Title</Table.Head>
					<Table.Head class="h-10 px-4 py-2">Content</Table.Head>
					<Table.Head class="h-10 px-4 py-2"></Table.Head>
				</Table.Row>
			</Table.Header>

			<Table.Body class="flex-grow overflow-auto pb-32">
				{#each tableData as data}
					<Table.Row>
						<Table.Cell class="h-10 px-4 py-2">{data.title}</Table.Cell>
						<Table.Cell class="h-10 truncate px-4 py-2">
							{data.content}
						</Table.Cell>
						<Table.Cell class="flex h-10 gap-2 px-4 py-2">
							<Button
								size="sm"
								variant="ghost"
								class="h-8"
								on:click={() => handleNote(data, false)}
							>
								<Icon icon="mdi-light:eye" class="h-4 w-4" />
							</Button>
							<Button size="sm" variant="ghost" class="h-8" on:click={() => handleNote(data, true)}>
								<Icon icon="mingcute:edit-line" class="h-4 w-4" />
							</Button>
							<Button
								size="sm"
								variant="ghost"
								class="h-8"
								on:click={() => handleRemoveNote(data.id)}
							>
								<Icon icon="gg:trash" class="h-4 w-4" />
							</Button>
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>

	<AlertDialog
		open={modalState.isDeleteOpen}
		names={{ parent: 'note', element: toRemove?.title || '' }}
		actions={{
			onConfirm: () => {
				handleRemoveNote(undefined, true);
			},
			onCancel: () => {
				modalState.isDeleteOpen = false;
			}
		}}
	></AlertDialog>
</div>
