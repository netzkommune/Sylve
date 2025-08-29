<script lang="ts">
	import Icon from '@iconify/svelte';
</script>

<div
	class="text-primary dark:text-secondary flex h-full w-full flex-col items-center justify-center space-y-3 text-center text-base"
>
	<Icon icon="mdi:server-off" class="dark:text-secondary text-primary h-14 w-14" />
	<div class="max-w-md">
		<span class="block whitespace-pre-line text-lg">
			This node can’t be reached right now. Check the network connection and confirm the node is
			powered on.
		</span>
	</div>
</div>
