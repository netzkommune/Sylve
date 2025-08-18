<script lang="ts">
	import { deleteSnapshot } from '$lib/api/zfs/datasets';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import type { Dataset } from '$lib/types/zfs/dataset';
	import { handleAPIError } from '$lib/utils/http';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		dataset: Dataset;
		askRecursive?: boolean;
	}

	let { open = $bindable(), dataset, askRecursive = true }: Props = $props();
	let recursive = $state(false);

	async function onCancel() {
		open = false;
	}

	async function onConfirm() {
		if (dataset.guid) {
			const response = await deleteSnapshot(dataset, recursive);

			if (response.status === 'success') {
				toast.success(`Deleted snapshot ${dataset.name}`, {
					position: 'bottom-center'
				});
			} else {
				handleAPIError(response);
				toast.error(`Failed to delete snapshot ${dataset.name}`, {
					position: 'bottom-center'
				});
			}
		} else {
			toast.error('Snapshot GUID not found', {
				position: 'bottom-center'
			});
		}

		open = false;
	}
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content onInteractOutside={(e) => e.preventDefault()}>
		<AlertDialog.Header>
			<AlertDialog.Title>Are you sure?</AlertDialog.Title>
			<AlertDialog.Description
				>This will delete the snapshot <b>{dataset.name}</b></AlertDialog.Description
			>
		</AlertDialog.Header>

		{#if askRecursive}
			<CustomCheckbox label="Recursive" bind:checked={recursive} classes="flex items-center gap-2"
			></CustomCheckbox>
		{/if}
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={onCancel}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={onConfirm}>Continue</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
