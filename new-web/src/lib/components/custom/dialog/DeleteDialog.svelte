<script lang="ts">
	interface Props {
		open: boolean;
		onConfirm: () => void;
		onCancel: () => void;
		texts: {
			parent: string;
			child: string;
			error: string;
		};
	}
</script>
