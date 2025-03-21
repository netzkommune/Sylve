<script lang="ts">
	import Button from '$lib/components/ui/button/button.svelte';
	import Icon from '@iconify/svelte';
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			onclick={() => console.log('New')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
</div>
