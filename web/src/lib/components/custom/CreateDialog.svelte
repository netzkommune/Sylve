<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import CircleHelp from 'lucide-svelte/icons/circle-help';
	import Icon from '@iconify/svelte';

	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';

	interface Props {
		title: string;
		tabs: { value: string; label: string }[];
		icon?: string;
		buttonText: string;
		buttonClass: string;
	}

	let { title, tabs, icon, buttonText, buttonClass }: Props = $props();

	let checked = $state(false);

	import { Input } from '$lib/components/ui/input/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	const fruits = [
		{ value: 'apple', label: 'Apple' },
		{ value: 'banana', label: 'Banana' },
		{ value: 'blueberry', label: 'Blueberry' },
		{ value: 'grapes', label: 'Grapes' },
		{ value: 'pineapple', label: 'Pineapple' }
	];
</script>

<Dialog.Root>
	<Dialog.Trigger class={buttonClass}>
		<Button size="sm" class="{buttonClass} w-full bg-blue-600 text-white hover:bg-blue-700">
			{#if icon}
				<Icon {icon} class="mr-1.5 h-5 w-5" />
			{/if}

			{buttonText}
		</Button>
	</Dialog.Trigger>

	<Dialog.Content class="w-[80%] gap-0 overflow-hidden p-0 lg:max-w-3xl">
		<div class="flex items-center justify-between">
			<Dialog.Header class="flex justify-between">
				<Dialog.Title class="p-4 text-left">{title}</Dialog.Title>
			</Dialog.Header>
			<Dialog.Close
				class="mr-4 flex h-5 w-5 items-center justify-center rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-0 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
			>
				<Icon icon="lucide:x" class="h-5 w-5" />
				<span class="sr-only">Close</span>
			</Dialog.Close>
		</div>

		<Tabs.Root value={tabs[0]?.value} class="mt-0 w-full overflow-hidden p-0">
			<ScrollArea orientation="horizontal" class="h-11 w-full max-w-full md:h-8">
				<Tabs.List class="flex items-center justify-start">
					{#each tabs as { value, label }}
						<Tabs.Trigger class="flex-shrink-0" {value}>{label}</Tabs.Trigger>
					{/each}
				</Tabs.List>
			</ScrollArea>
			{#each tabs as { value, label }}
				<Tabs.Content {value} class="mx-auto mt-1 w-[99%] border p-2 md:w-full md:p-3"
					><ScrollArea orientation="vertical" class="h-96 w-full  ">
						{#if value === 'vm_general'}
							<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
								<div class="flex flex-col items-start gap-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="terms">Node:</Label>
									<Select.Root>
										<Select.Trigger class="h-8 ">
											<Select.Value placeholder="Select a fruit" />
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												<Select.Label>Fruits</Select.Label>
												{#each fruits as fruit}
													<Select.Item value={fruit.value} label={fruit.label}
														>{fruit.label}</Select.Item
													>
												{/each}
											</Select.Group>
										</Select.Content>
										<Select.Input name="favoriteFruit" />
									</Select.Root>
								</div>

								<div class="flex flex-col items-start gap-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="terms">VM ID:</Label>
									<Input class="h-8" type="number" id="vm_id" placeholder="email" />
								</div>

								<div class="flex flex-col items-start gap-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="terms">Name:</Label>
									<Input class="h-8" type="text" id="name" placeholder="email" />
								</div>
								<div class="flex flex-col items-start gap-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="terms">Resource Pool:</Label>
									<Select.Root>
										<Select.Trigger class="h-8 ">
											<Select.Value placeholder="Select a fruit" />
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												<Select.Label>Fruits</Select.Label>
												{#each fruits as fruit}
													<Select.Item value={fruit.value} label={fruit.label}
														>{fruit.label}</Select.Item
													>
												{/each}
											</Select.Group>
										</Select.Content>
										<Select.Input name="favoriteFruit" />
									</Select.Root>
								</div>
							</div>
						{:else if value === 'ct_general'}
							<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms">Node</Label>
									<Select.Root>
										<Select.Trigger class="h-8 ">
											<Select.Value placeholder="Select a fruit" />
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												<Select.Label>Fruits</Select.Label>
												{#each fruits as fruit}
													<Select.Item value={fruit.value} label={fruit.label}
														>{fruit.label}</Select.Item
													>
												{/each}
											</Select.Group>
										</Select.Content>
										<Select.Input name="favoriteFruit" />
									</Select.Root>
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms">Resource Pool</Label>
									<Select.Root>
										<Select.Trigger class="h-8 ">
											<Select.Value placeholder="Select a fruit" />
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												<Select.Label>Fruits</Select.Label>
												{#each fruits as fruit}
													<Select.Item value={fruit.value} label={fruit.label}
														>{fruit.label}</Select.Item
													>
												{/each}
											</Select.Group>
										</Select.Content>
										<Select.Input name="favoriteFruit" />
									</Select.Root>
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms">CT ID</Label>
									<Input class="h-8" type="number" id="ct_id" placeholder="email" />
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms">Password</Label>
									<Input class="h-8" type="number" id="hostname" placeholder="email" />
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms">Hostname</Label>
									<Input class="h-8" type="number" id="hostname" placeholder="email" />
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms"
										>Confirm passowrd:</Label
									>
									<Input class="h-8" type="number" id="hostname" placeholder="email" />
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-44 whitespace-nowrap text-sm" for="terms"
										>Unprivileged container:</Label
									>
									<Checkbox id="terms" bind:checked aria-labelledby="terms-label" />
								</div>

								<div class="flex items-center gap-1">
									<Label class="w-44 whitespace-nowrap text-sm" for="terms">Nesting</Label>
									<Checkbox id="terms" bind:checked aria-labelledby="terms-label" />
								</div>
								<div class="flex items-center gap-1">
									<Label class="w-full whitespace-nowrap text-sm" for="terms"
										>SSH public key(s)</Label
									>
									<Textarea placeholder="Type your message here." />
								</div>
							</div>
						{/if}
					</ScrollArea></Tabs.Content
				>
			{/each}
		</Tabs.Root>

		<Dialog.Footer>
			<div class="flex w-full flex-col justify-between px-3 py-3 md:flex-row">
				<Button size="sm" class="h-8">
					<CircleHelp class="mr-2 h-4 w-4" />
					Help
				</Button>
				<div class="flex flex-col items-center gap-2 space-x-3 md:flex-row">
					<div class="mt-2 flex items-center space-x-2 md:mt-0">
						<Label
							id="terms-label"
							for="terms"
							class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
						>
							Advanced
						</Label>
						<Checkbox id="terms" bind:checked aria-labelledby="terms-label" />
					</div>
					<Button
						size="sm"
						type="button"
						class="disabled h-7 w-full bg-blue-600 text-white hover:bg-blue-700"
					>
						Back
					</Button>
					<Button
						size="sm"
						type="button"
						class="h-8 w-full bg-blue-600 text-white hover:bg-blue-700"
					>
						Next
					</Button>
				</div>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
