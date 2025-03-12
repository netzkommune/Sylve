<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import ScrollArea from '$lib/components/ui/scroll-area/scroll-area.svelte';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import Icon from '@iconify/svelte';

	let isOpen = $state(false);
	let isEditMode = $state(false);
	let selectedId = $state<number | null>(null);
	let title = $state('');
	let content = $state('');
	let tableData = $state([]);

	// Save a new note or edit an existing one
	// function saveNote() {
	// 	if (title.trim() !== '' && content.trim() !== '') {
	// 		if (isEditMode && selectedId !== null) {
	// 			// Edit existing note
	// 			tableData = tableData.map((note) =>
	// 				note.id === selectedId ? { ...note, title, content } : note
	// 			);
	// 		} else {
	// 			// Add new note
	// 			tableData = [...tableData, { id: tableData.length + 1, title, content }];
	// 		}
	// 		resetDialog();
	// 	}
	// }

	// Open dialog for adding a new note
	function newNote() {
		title = '';
		content = '';
		selectedId = null;
		isEditMode = true;
		isOpen = true;
	}

	// Open dialog for viewing a note
	function viewNote(note) {
		title = note.title;
		content = note.content;
		selectedId = note.id;
		isEditMode = false;
		isOpen = true;
	}

	// Open dialog for editing a note
	function editNote(note) {
		title = note.title;
		content = note.content;
		selectedId = note.id;
		isEditMode = true;
		isOpen = true;
	}

	// Delete a note
	function deleteNote(id) {
		tableData = tableData.filter((note) => note.id !== id);
	}

	// Reset the dialog state
	function resetDialog() {
		title = '';
		content = '';
		selectedId = null;
		isEditMode = false;
		isOpen = false;
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			onclick={newNote}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>

	<Dialog.Root bind:open={isOpen} let:close>
		<Dialog.Content class="w-[80%] gap-0 overflow-hidden p-3 lg:max-w-3xl">
			<div class="flex items-center justify-between py-2">
				<Dialog.Header class="flex-1">
					<Dialog.Title>
						{isEditMode ? (selectedId ? 'Edit Note' : 'New Note') : 'View Note'}
					</Dialog.Title>
				</Dialog.Header>
				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="ghost"
						class="h-8"
						on:click={() => {
							title = '';
							content = '';
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
						bind:value={title}
						class="mt-2"
						placeholder="Enter title"
						disabled={!isEditMode}
					/>
				</div>
			</div>

			<div class="mt-2">
				<Label for="content">Content</Label>
				<ScrollArea orientation="vertical" class="h-96">
					{#if isEditMode}
						<div class="p-1">
							<Textarea
								bind:value={content}
								class="mt-2 h-[360px] w-full"
								placeholder="Write in Markdown format"
							/>
						</div>
					{:else}
						<div class="p-1">
							<Textarea value={content} class="h-[360px] w-full" disabled />
						</div>
					{/if}
				</ScrollArea>
			</div>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-between gap-2 px-3 py-2">
					<Button
						type="submit"
						size="sm"
						class="bg-muted hover:bg-muted-foreground/50 h-8 text-white"
					>
						Help
					</Button>

					<div class="flex gap-2">
						<Dialog.Close>
							<Button variant="outline" class="h-8" on:click={resetDialog}>Cancel</Button>
						</Dialog.Close>
						{#if isEditMode}
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

	<!-- Table -->
	<div class="flex h-full flex-col overflow-hidden">
		<Table.Root class="w-full table-fixed border-collapse">
			<Table.Header class="bg-background sticky top-0 z-[50]">
				<Table.Row>
					<Table.Head class="h-10 px-4 py-2">Title</Table.Head>
					<Table.Head class="h-10 px-4 py-2">Content</Table.Head>
					<Table.Head class="h-10 px-4 py-2">Actions</Table.Head>
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
							<Button size="sm" variant="ghost" class="h-8" on:click={() => viewNote(data)}>
								<Icon icon="mdi-light:eye" class="h-4 w-4" />
							</Button>
							<Button size="sm" variant="ghost" class="h-8" on:click={() => editNote(data)}>
								<Icon icon="mingcute:edit-line" class="h-4 w-4" />
							</Button>
							<Button size="sm" variant="ghost" class="h-8" on:click={() => deleteNote(data.id)}>
								<Icon icon="gg:trash" class="h-4 w-4" />
							</Button>
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
</div>
