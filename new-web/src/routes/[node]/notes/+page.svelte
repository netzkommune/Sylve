<script lang="ts">
	import { createNote, deleteNote, deleteNotes, getNotes, updateNote } from '$lib/api/info/notes';
	import AlertDialog from '$lib/components/custom/AlertDialog.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ScrollArea from '$lib/components/ui/scroll-area/scroll-area.svelte';
	import type { APIResponse } from '$lib/types/common';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Note } from '$lib/types/info/notes';
	import { handleValidationErrors, isAPIResponse, updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { generateTableData, markdownToTailwindHTML } from '$lib/utils/info/notes';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';

	interface Data {
		notes: Note[];
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['notes'],
			queryFn: async () => {
				return (await getNotes()) as Note[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.notes,
			onSuccess: (data: Note[]) => {
				updateCache('notes', data);
			}
		}
	]);

	let notes: Note[] = $derived($results[0].data as Note[]);
	let modalState = $state({
		title: '',
		content: '',
		isOpen: false,
		isEditMode: false,
		isDeleteOpen: false,
		isBulkDeleteOpen: false
	});

	let selectedId: number | null = $state(null);

	function handleNote(note?: Note, editMode: boolean = true, reset: boolean = false) {
		if (reset) {
			modalState.title = '';
			modalState.content = '';
			selectedId = null;
			modalState.isEditMode = false;
			modalState.isOpen = false;
			modalState.isDeleteOpen = false;
			modalState.isBulkDeleteOpen = false;
			activeRow = null;
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
		if (modalState.isEditMode && selectedId !== null) {
			const response = await updateNote(selectedId, modalState.title, modalState.content);
			if (response.status === 'success') {
				toast.success(
					`${capitalizeFirstLetter(getTranslation('notes.note', 'Note'))} ${modalState.title} ${getTranslation('common.updated', 'updated')}`,
					{ position: 'bottom-center' }
				);
				handleNote(undefined, false, true);
			} else {
				handleValidationErrors(response, 'notes');
			}
		} else {
			const response = await createNote(modalState.title, modalState.content);
			if ((response as Note).id) {
				toast.success(
					`${capitalizeFirstLetter(getTranslation('notes.note', 'Note'))} ${modalState.title} ${getTranslation('common.created', 'created')}`,
					{ position: 'bottom-center' }
				);
				handleNote(undefined, false, true);
			}

			if ((response as APIResponse).status) {
				handleValidationErrors(response as APIResponse, 'notes');
			}
		}
	}

	function viewNote(id: number | string | undefined) {
		const note = notes.find((note) => note.id === id);
		if (note) {
			modalState.title = note.title;
			modalState.content = note.content;
			modalState.isEditMode = false;
			modalState.isOpen = true;
		}
	}

	function handleDelete(id: number | string | undefined) {
		const note = notes.find((note) => note.id === id);
		if (note) {
			modalState.title = note.title;
			modalState.content = note.content;
			modalState.isEditMode = false;
			modalState.isDeleteOpen = true;
		}
	}

	function handleBulkDelete(ids: number[]) {
		const notesToDelete = notes.filter((note) => ids.includes(note.id));
		if (notesToDelete.length > 0) {
			modalState.title = `${notesToDelete.length} ${getTranslation('notes.notes', 'notes')}`;
			modalState.isBulkDeleteOpen = true;
		}
	}

	let tableName = 'tt-notes';
	let tableData = $derived(generateTableData(notes));
	let activeRow: Row[] | null = $state(null);
	let query: string = $state('');
</script>

{#snippet button(type: string)}
	{#if activeRow !== null && activeRow.length === 1}
		{#if type === 'view-note'}
			<Button
				onclick={() => activeRow && viewNote(activeRow[0]?.id)}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:eye" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.view', 'View'))}
			</Button>
		{/if}

		{#if type === 'delete-note'}
			<Button
				onclick={() => activeRow && handleDelete(activeRow[0]?.id)}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.delete', 'Delete'))}
			</Button>
		{/if}

		{#if type === 'edit-note'}
			<Button
				onclick={() => {
					const note = notes.find((note) => activeRow && note.id === activeRow[0]?.id);
					handleNote(note, true);
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:note-edit" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.edit', 'Edit'))}
			</Button>
		{/if}
	{/if}

	{#if activeRow !== null && activeRow.length > 1}
		{#if type === 'bulk-delete-note'}
			<Button
				onclick={() => {
					const ids = activeRow?.map((row) => row.id) || [];
					handleBulkDelete(ids as number[]);
				}}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="material-symbols:delete-sweep" class="mr-1 h-4 w-4" />
				{capitalizeFirstLetter(getTranslation('common.bulk', 'Bulk'))}
				{capitalizeFirstLetter(getTranslation('common.delete', 'Delete'))}
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />

		<Button onclick={() => handleNote()} size="sm" class="h-6  ">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>

		{@render button('view-note')}
		{@render button('edit-note')}
		{@render button('delete-note')}
		{@render button('bulk-delete-note')}
	</div>

	<Dialog.Root bind:open={modalState.isOpen}>
		<Dialog.Content class="w-[90%] gap-2 overflow-hidden p-5 lg:max-w-2xl">
			<div class="flex items-center justify-between">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						<div class="flex items-center gap-2">
							<Icon icon={modalState.isEditMode ? 'mdi:note-edit' : 'mdi:note'} class="h-5 w-5" />
							<span
								>{modalState.isEditMode
									? selectedId
										? capitalizeFirstLetter(getTranslation('common.edit', 'Edit'))
										: capitalizeFirstLetter(getTranslation('common.new', 'New'))
									: capitalizeFirstLetter(getTranslation('common.view', 'View'))}
								{capitalizeFirstLetter(getTranslation('notes.note', 'Note'))}</span
							>
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
							modalState.title = '';
							modalState.content = '';
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
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
							modalState.title = '';
							modalState.content = '';
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only"
							>{capitalizeFirstLetter(getTranslation('common.close', 'Close'))}</span
						>
					</Button>
				</div>
			</div>

			<CustomValueInput
				label={capitalizeFirstLetter(getTranslation('common.name', 'Name'))}
				placeholder="Post Upgrade Summary"
				bind:value={modalState.title}
				classes="flex-1 space-y-1"
			/>

			<div class="">
				<ScrollArea orientation="vertical" class="h-full">
					{#if modalState.isEditMode}
						<div>
							<CustomValueInput
								label={capitalizeFirstLetter(getTranslation('common.content', 'Content'))}
								placeholder="This is a note"
								bind:value={modalState.content}
								classes="flex-1 space-y-1 "
								type="textarea"
							/>
						</div>
					{:else}
						<div class="mt-2">
							<p
								class="mb-2 text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
							>
								{capitalizeFirstLetter(getTranslation('common.content', 'Content'))}
							</p>
							<article class="prose lg:prose-xl rounded-md border p-3">
								{@html markdownToTailwindHTML(modalState.content)}
							</article>
						</div>
					{/if}
				</ScrollArea>
			</div>
			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-end gap-2 px-1 py-2">
					{#if modalState.isEditMode}
						<Button onclick={saveNote} type="submit" size="sm"
							>{capitalizeFirstLetter(getTranslation('common.save', 'Save'))}</Button
						>
					{/if}
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>

	<div class="flex h-full flex-col overflow-hidden">
		<TreeTable data={tableData} name={tableName} bind:parentActiveRow={activeRow} bind:query />
	</div>

	<AlertDialog
		open={modalState.isDeleteOpen}
		names={{ parent: 'note', element: modalState?.title || '' }}
		actions={{
			onConfirm: async () => {
				const id = activeRow ? activeRow[0]?.id : null;
				const result = await deleteNote(id as number);
				if (isAPIResponse(result) && result.status === 'success') {
					handleNote(undefined, false, true);
				} else {
					handleValidationErrors(result as APIResponse, 'notes');
				}
			},
			onCancel: () => {
				modalState.isDeleteOpen = false;
			}
		}}
	></AlertDialog>

	<AlertDialog
		open={modalState.isBulkDeleteOpen}
		names={{ parent: '', element: modalState?.title || '' }}
		actions={{
			onConfirm: async () => {
				const ids = activeRow
					? activeRow.map((row) => (typeof row.id === 'number' ? row.id : parseInt(row.id)))
					: [];
				const result = await deleteNotes(ids);
				if (isAPIResponse(result) && result.status === 'success') {
					handleNote(undefined, false, true);
				} else {
					handleValidationErrors(result as APIResponse, 'notes');
				}
			},
			onCancel: () => {
				modalState.isBulkDeleteOpen = false;
			}
		}}
	></AlertDialog>
</div>
