<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { store } from '$lib/stores/auth';
	import Icon from '@iconify/svelte';

	interface Props {
		isOpen: boolean;
		onClose: () => void;
		currentPath?: string;
		droppedFiles?: File[];
		onUploadComplete?: () => void;
	}

	let {
		isOpen = $bindable(false),
		onClose,
		currentPath = '/',
		droppedFiles = [],
		onUploadComplete
	}: Props = $props();

	import FilePond, { registerPlugin, supported } from 'svelte-filepond';

	// Import the Image EXIF Orientation and Image Preview plugins
	// Note: These need to be installed separately
	// `npm i filepond-plugin-image-preview filepond-plugin-image-exif-orientation --save`
	import { sha256 } from '$lib/utils/string';
	import FilePondPluginImageExifOrientation from 'filepond-plugin-image-exif-orientation';
	import FilePondPluginImagePreview from 'filepond-plugin-image-preview';
	import { onMount } from 'svelte';

	// Register the plugins
	registerPlugin(FilePondPluginImageExifOrientation, FilePondPluginImagePreview);

	// a reference to the component, used to call FilePond methods
	let pond;

	// pond.getFiles() will return the active files

	// the name to use for the internal file input
	let name = 'filepond';
	let hash = $state('');

	onMount(async () => {
		hash = await sha256($store, 1);
	});

	// handle filepond events
	function handleInit() {
		console.log('FilePond has initialised');

		// Add dropped files when FilePond is ready
		if (pond && droppedFiles.length > 0) {
			droppedFiles.forEach((file) => {
				pond.addFile(file);
			});
		}
	}

	function handleAddFile(err: any, fileItem: any) {
		console.log('A file has been added', fileItem);
	}

	function handleProcessFile(error: any, file: any) {
		if (error) {
			console.error('Upload failed:', error);
			return;
		}

		console.log('File uploaded successfully:', file);

		// Call the upload complete callback if provided
		if (onUploadComplete) {
			onUploadComplete();
		}
	}

	// Watch for changes in droppedFiles and add them to FilePond
	$effect(() => {
		if (pond && droppedFiles.length > 0 && isOpen) {
			// Clear existing files first
			pond.removeFiles();
			// Add new dropped files
			droppedFiles.forEach((file) => {
				pond.addFile(file);
			});
		}
	});
</script>

<Dialog.Root bind:open={isOpen}>
	<Dialog.Content
		onInteractOutside={onClose}
		class="fixed flex transform flex-col gap-2 overflow-auto p-5 transition-all duration-300 ease-in-out lg:max-w-md"
	>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex items-center justify-between text-left">
				<div class="flex items-center gap-2">
					<Icon icon="material-symbols:upload" class="h-6 w-6" />
					Upload File
				</div>
				<div class="flex items-center gap-0.5">
					<Button size="sm" variant="link" class="h-4" title="Close" onclick={onClose}>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Close</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>
		<div class="app mt-4">
			<FilePond
				bind:this={pond}
				{name}
				server={'/api/system/file-explorer/upload?path=' +
					encodeURIComponent(currentPath) +
					'&hash=' +
					hash}
				allowMultiple={true}
				oninit={handleInit}
				onaddfile={handleAddFile}
				onprocessfile={handleProcessFile}
				credits={false}
			/>
		</div>
	</Dialog.Content>
</Dialog.Root>
