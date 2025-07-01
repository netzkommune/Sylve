<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';

	interface Props {
		open: boolean;
		names?: {
			parent: string;
			element: string;
		};
		actions: {
			onConfirm: () => void;
			onCancel: () => void;
		};
		customTitle?: string;
	}

	let { open = $bindable(), names, actions, customTitle }: Props = $props();
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content onInteractOutside={(e) => e.preventDefault()} class="p-5">
		<AlertDialog.Header>
			<AlertDialog.Title>Are you sure?</AlertDialog.Title>
			<AlertDialog.Description>
				{#if customTitle}
					{@html customTitle}
				{:else if names && names.parent && names.element}
					{'This action cannot be undone. This will permanently delete '}
					<span>{names.parent}</span>
					<span class="font-semibold">{names.element}</span>.
				{/if}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={actions.onCancel}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={actions.onConfirm}>Continue</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
