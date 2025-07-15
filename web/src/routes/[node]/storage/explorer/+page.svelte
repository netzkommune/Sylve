<script lang="ts">
	import { getFiles } from '$lib/api/system/file-explorer';
	import { type FileNode } from '$lib/types/system/file-explorer';
	import { mode } from 'mode-watcher';
	//@ts-ignore next-line
	import { Filemanager, Willow, WillowDark } from 'wx-svelte-filemanager';

	interface Data {
		files: FileNode[];
	}

	let { data }: { data: Data } = $props();
	let api;

	let rawData = $state(data.files as FileNode[]);

	async function loadData(req: { id?: string }) {
		if (req && req.id) {
			const response = await getFiles(req.id);
			rawData = response;
		}
	}
</script>

<div class="h-full w-full">
	{#if mode.current === 'light'}
		<Willow>
			<Filemanager bind:this={api} data={rawData} onrequestdata={loadData} /></Willow
		>
	{:else}
		<WillowDark>
			<Filemanager bind:this={api} data={rawData} onrequestdata={loadData} /></WillowDark
		>
	{/if}
</div>

<style>
	:global(.wx-willow-theme) {
		--wx-theme-name: willow;
		--wx-color-primary: var(--primary) !important;
		--wx-fm-background: var(--background) !important;
		--wx-fm-select-background: var(--muted) !important;
		--wx-fm-segmented-background: var(--background) !important;
		--wx-fm-box-shadow: 0px 1px 2px rgba(44, 47, 60, 0.06), 0px 3px 10px rgba(44, 47, 60, 0.12);
		--wx-fm-tree: var(--background);
		--wx-fm-grid-border: 1px solid var(--border);
		--wx-fm-grid-header-color: #fafafb;
		--wx-fm-button-font-color: #9fa1ae;
		--wx-fm-toolbar-height: 56px;
	}

	:global(.wx-willow-dark-theme) {
		--wx-theme-name: willow-dark;
		color-scheme: dark;
		--wx-color-primary: var(--primary) !important;
		--wx-fm-background: var(--background) !important;
		--wx-fm-select-background: var(--muted) !important;
		--wx-fm-segmented-background: var(--background) !important;
		--wx-fm-box-shadow: none;
		--wx-fm-grid-border: 1px solid #384047;
		--wx-fm-grid-header-color: var(--background);
		--wx-fm-button-font-color: #9fa1ae;
		--wx-fm-toolbar-height: 56px;
		--wx-fm-select-color: rgb(80, 90, 100);
		--wx-table-select-background: rgba(33, 195, 255, 0.15);
		--wx-table-select-focus-background: rgba(139, 0, 0);
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-breadcrumbs .wx-item) {
		border: none !important;
	}

	:global(.wx-willow-dark-theme .wx-sidebar .wx-wrapper) {
		background-color: var(--sidebar) !important;
		border: 1px solid var(--sidebar) !important;
	}

	:global(.wx-willow-dark-theme .wx-sidebar .wx-wrapper .wx-button) {
		padding: 0 !important;
		border-radius: 0.3rem !important;
		height: 35px;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-breadcrumbs) {
		background-color: var(--background) !important;
	}
	:global(.wx-willow-dark-theme .wx-filemanager .wx-toolbar) {
		background-color: var(--sidebar) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-segmented-background) {
		background-color: var(--background) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-button) {
		background-color: var(--primary) !important;
		color: black !important;
	}

	:global(.wx-willow-dark-theme .wx-toolbar .wx-right) {
		background-color: var(--sidebar) !important;
	}

	:global(.wx-willow-dark-theme .wx-toolbar .wx-modes) {
		background-color: var(--sidebar) !important;
	}

	:global(.wx-willow-dark-theme .wx-toolbar .wx-segmented) {
		background-color: var(--sidebar) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-selected) {
		background-color: var(--muted) !important;
		color: var(--foreground) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-selected:hover) {
		background-color: var(--muted) !important;
		color: var(--foreground) !important;
	}

	:global(.wx-willow-dark-theme) {
		border-radius: 10px !important;
		overflow: hidden !important;
		border: var(--border) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-text) {
		background-color: var(--background) !important;
		border: 1pz solid var(--border) !important;
		border-radius: var(--radius) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-item) {
		background-color: var(--background) !important;
		border: 1px solid var(--border) !important;
		border-radius: var(--radius) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-item:hover) {
		background-color: var(--muted) !important;
		border: 1px solid var(--border) !important;
		border-radius: var(--radius) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-table-box .wx-row) {
		background-color: var(--background) !important;
		border: 1px solid var(--border) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-table-box .wx-row:hover) {
		background-color: var(--muted) !important;
	}

	:global(.wx-willow-dark-theme .wx-filemanager .wx-list) {
		background-color: var(--background) !important;
		border: var(--border) !important;
	}

	:global(.wx-willow-dark-theme .wx-menu .wx-item) {
		background-color: var(--muted) !important;
		border: var(--border) !important;
	}

	:global(.wx-willow-dark-theme .wx-menu .wx-item:hover) {
		background-color: var(--background) !important;
		border: var(--border) !important;
	}
</style>
