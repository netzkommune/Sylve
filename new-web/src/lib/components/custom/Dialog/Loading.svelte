<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import Icon from '@iconify/svelte';

	interface Props {
		open: boolean;
		title: string;
		description: string;
		iconColor?: string;
	}

	let { open = $bindable(), iconColor = 'text-red-500', title, description }: Props = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="sm:max-w-[425px]"
		onInteractOutside={(e) => e.preventDefault()}
		onEscapeKeydown={(e) => e.preventDefault()}
	>
		<Dialog.Header class="flex flex-col items-center justify-center text-center">
			<Dialog.Title class="mb-2 text-lg font-semibold">{title}</Dialog.Title>
			<Icon icon="mdi:loading" class={`mb-4 animate-spin text-4xl ${iconColor}`} />
			<Dialog.Description class="text-muted-foreground mt-3 text-sm">
				{@html description}
			</Dialog.Description>
		</Dialog.Header>
	</Dialog.Content>
</Dialog.Root>
