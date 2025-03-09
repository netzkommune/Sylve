<script lang="ts">
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ScrollArea from '$lib/components/ui/scroll-area/scroll-area.svelte';

	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import Icon from '@iconify/svelte';

	let isOpen = $state(false);
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			onclick={() => (isOpen = !isOpen)}
			size="sm"
			class="h-6 bg-muted-foreground/40 text-black   dark:bg-muted dark:text-white"
		>
			<Icon icon="tabler:edit" class="mr-2 h-4 w-4" />Edit
		</Button>
	</div>
	<Dialog.Root bind:open={isOpen} let:close>
		<Dialog.Content class="w-[80%] gap-0 overflow-hidden p-0 lg:max-w-3xl">
			<div class="flex items-center justify-between px-3 py-2">
				<Dialog.Header class="flex-1">
					<Dialog.Title>Notes</Dialog.Title>
				</Dialog.Header>
				<div class="flex items-center gap-0.5">
					<Button size="sm" variant="ghost" class="h-8">
						<Icon icon="radix-icons:reset" class="h-4 w-4" />
						<span class="sr-only">Reset</span>
					</Button>
					<Dialog.Close
						class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
					>
						<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
						<span class="sr-only">Close</span>
					</Dialog.Close>
				</div>
			</div>
			<ScrollArea orientation="vertical" class="h-96">
				<Textarea
					class="h-96 w-full"
					placeholder="You can use markdown for rich text formatting."
				/>
			</ScrollArea>

			<Dialog.Footer class="flex justify-end">
				<div class="flex w-full items-center justify-between gap-2 px-3 py-2">
					<Button
						type="submit"
						size="sm"
						class="h-8 bg-muted text-white hover:bg-muted-foreground/50">Help</Button
					>

					<div class="flex gap-2">
						<Dialog.Close>
							<Button variant="outline" class="h-8" on:click={close}>Cancel</Button></Dialog.Close
						>
						<Button
							type="submit"
							size="sm"
							class="h-8 w-16 bg-blue-600 text-white hover:bg-blue-500">Ok</Button
						>
					</div>
				</div>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
</div>
