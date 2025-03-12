<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import { getTranslation } from '$lib/utils/i18n';

	interface Props {
		open: boolean;
		names: {
			parent: string;
			element: string;
		};
		actions: {
			onConfirm: () => void;
			onCancel: () => void;
		};
	}

	let { open, names, actions }: Props = $props();
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>{getTranslation('are_you_sure', 'Are you sure?')}</AlertDialog.Title>
			<AlertDialog.Description>
				{getTranslation(
					'common.permanent_delete_msg',
					'This action cannot be undone. This will permanently delete'
				)}
				{names.parent} <span class="font-semibold">{names.element}</span>.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={actions.onCancel}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={actions.onConfirm}>Continue</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
