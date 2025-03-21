<script lang="ts">
	import Icon from '@iconify/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import CircleHelp from 'lucide-svelte/icons/circle-help';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as ContextMenu from '$lib/components/ui/context-menu';
	import { localStore } from '$lib/stores/localStore.svelte';
	import { TableHandler } from '@vincjo/datatables';
	import * as Select from '$lib/components/ui/select/index.js';
	import { slide } from 'svelte/transition';

	import MultiSelect from 'svelte-multiselect';

	const ui_libs = [
		{
			label: 'harddisk',
			disabled: false,
			preselected: false,
			defaultDisabledTitle: 'This is a disabled option'
		},
		{
			label: 'harddisk1',
			disabled: false,
			preselected: false
		},
		{
			label: 'ssd',
			disabled: false,
			preselected: false
		}
	];

	let selected = $state([]);

	const raid = [
		{ value: 'mirror', label: 'Mirror' },
		{ value: 'raidz1', label: 'RAIDZ1' },
		{ value: 'raidz2', label: 'RAIDZ2' },
		{ value: 'raidz3', label: 'RAIDZ3' }
	];

	const compression = [
		{ value: 'lz4', label: 'LZ4' },
		{ value: 'zstd', label: 'ZSTD' },
		{ value: 'gzip', label: 'GZIP' },
		{ value: 'zle', label: 'ZLE' }
	];

	const ashift = [
		{ value: 'lz4', label: 'LZ4' },
		{ value: 'zstd', label: 'ZSTD' },
		{ value: 'gzip', label: 'GZIP' },
		{ value: 'zle', label: 'ZLE' }
	];

	let modalIsOpen = $state(true);
	let advancedChecked = $state(false);

	interface ZfsData {
		Health: string;
		Percentage_Used: string;
		Total_Size: string;
	}

	const dummyData = [
		{
			Health: 'Healthy',
			Percentage_Used: '23%',
			Total_Size: '256 GB'
		},
		{
			Health: 'Warning',
			Percentage_Used: '75%',
			Total_Size: '512 GB'
		},
		{
			Health: 'Critical',
			Percentage_Used: '95%',
			Total_Size: '1 TB'
		}
	];

	const table = new TableHandler(dummyData);

	const keys = ['Health', 'Percentage_Used', 'Total_Size'];
	let sortHandlers: Record<string, any> = {};

	keys.forEach((key) => {
		sortHandlers[key] = table.createSort(key as keyof ZfsData, {
			locales: 'en',
			options: { numeric: true, sensitivity: 'base' }
		});
	});

	let visibleColumns = localStore(
		'zfsVisibleColumns',
		Object.fromEntries(keys.map((key) => [key, true]))
	);

	let openContextMenuId = $state<string | null>(null);

	function handleContextMenuOpen(id: string) {
		openContextMenuId = id;
	}

	function handleContextMenuClose() {
		openContextMenuId = null;
	}

	function toggleColumnVisibility(columnKey: string) {
		const wouldHideAll =
			Object.entries(visibleColumns.value).filter(([k, v]) => k !== columnKey && v).length === 0 &&
			visibleColumns.value[columnKey];

		if (!wouldHideAll) {
			visibleColumns.value[columnKey] = !visibleColumns.value[columnKey];
		}
	}

	function resetColumns() {
		visibleColumns.value = Object.fromEntries(keys.map((key) => [key, true]));
	}

	//modal keyValue field

	let pairs: { key: string; value: string }[] = $state([{ key: '', value: '' }]);

	function addPair() {
		pairs = [...pairs, { key: '', value: '' }];
	}

	function removePair(index: number) {
		pairs = pairs.filter((_, i) => i !== index);
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center border p-2">
		<Button
			onclick={() => (modalIsOpen = true)}
			size="sm"
			class="h-6 bg-muted-foreground/40 text-black dark:bg-muted dark:text-white"
		>
			<Icon icon="gg:add" class="mr-1 h-4 w-4" /> New
		</Button>
	</div>
	<div class="relative flex h-full w-full cursor-pointer flex-col">
		<div class="flex-1">
			<div class="h-full overflow-y-auto">
				<table class="mb-10 w-full min-w-max border-collapse">
					<thead>
						<tr>
							{#each keys as key}
								{#if visibleColumns.value[key]}
									<th
										class="group h-8 w-48 whitespace-nowrap border border-neutral-300 px-3 text-left text-black dark:border-neutral-800 dark:text-white"
									>
										<ContextMenu.Root
											open={openContextMenuId === key}
											closeOnItemClick={false}
											onOpenChange={(open) =>
												open ? handleContextMenuOpen(key) : handleContextMenuClose()}
										>
											<ContextMenu.Trigger class="flex h-full w-full">
												<button
													class="relative flex w-full items-center"
													onclick={() => sortHandlers[key].set()}
												>
													<span>{key}</span>
													<Icon
														icon={sortHandlers[key].direction === 'asc'
															? 'lucide:sort-asc'
															: 'lucide:sort-desc'}
														class="ml-2 mt-1 h-4 w-4 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
													/>
												</button>
											</ContextMenu.Trigger>
											<ContextMenu.Content>
												<ContextMenu.Label>Toggle Columns</ContextMenu.Label>
												<ContextMenu.Separator />
												{#each keys as columnKey}
													<ContextMenu.CheckboxItem
														checked={visibleColumns.value[columnKey]}
														onCheckedChange={(e: boolean | 'indeterminate') => {
															toggleColumnVisibility(columnKey);
														}}
													>
														{columnKey}
													</ContextMenu.CheckboxItem>
												{/each}
												<ContextMenu.Separator />
												<ContextMenu.Item onclick={resetColumns}>Reset Columns</ContextMenu.Item>
											</ContextMenu.Content>
										</ContextMenu.Root>
									</th>
								{/if}
							{/each}
						</tr>
					</thead>

					<tbody>
						{#each table.rows as row, index}
							<tr>
								{#each keys as key}
									{#if visibleColumns.value[key]}
										<td
											class="h-8 w-48 whitespace-nowrap border border-neutral-300 px-3 text-left text-black dark:border-neutral-800 dark:text-white"
										>
											{row[key]}
										</td>
									{/if}
								{/each}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<Dialog.Root bind:open={modalIsOpen}>
	<Dialog.Content class=" w-[80%] gap-0  p-0 lg:max-w-3xl">
		<div class="flex h-12 items-center justify-between border-b">
			<Dialog.Header class="flex justify-between">
				<Dialog.Title class="p-4 text-left">zfs</Dialog.Title>
			</Dialog.Header>
			<Dialog.Close
				class="mr-4 flex h-4 w-4 items-center justify-center rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
			>
				<Icon icon="lucide:x" class="h-4 w-4" />
				<span class="sr-only">Close</span>
			</Dialog.Close>
		</div>

		<div class="space-y-4 p-4">
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<div class="flex flex-col items-start gap-1">
					<Label class="w-24 whitespace-nowrap text-sm" for="terms">Pool Name:</Label>
					<Input class="h-8" type="text" id="name" placeholder="pool name" />
				</div>
				<div>
					<Label class="w-24 whitespace-nowrap text-sm" for="terms">vDevs:</Label>
					<div>
						<MultiSelect
							bind:selected
							options={ui_libs}
							placeholder="Select disks"
							ulOptionsClass="dark:bg-[#0c0a09] mt-1 border: 1px solid #f00 overflow-y-auto"
							liActiveOptionClass="bg-muted dark:text-white"
							on:change={(event) => console.log(event.detail)}
						/>
					</div>
				</div>
			</div>

			{#if selected.length > 0}
				<div transition:slide class=" flex items-center justify-center space-x-2">
					{#each selected as item}
						{#if item.label.includes('harddisk')}
							<div transition:slide class="flex items-center justify-center space-x-2">
								<Icon icon="mdi:harddisk" class="m-4 h-12 w-12 text-green-500" />
							</div>
						{:else}
							<div transition:slide class="flex items-center justify-center space-x-2">
								<Icon icon="clarity:ssd-solid" class="m-4 h-12 w-12 text-gray-500" />
							</div>
						{/if}
					{/each}
				</div>

				<div transition:slide class="grid grid-cols-1 gap-4 md:grid-cols-3">
					<div>
						<Label class="w-24 whitespace-nowrap text-sm" for="terms">RAID:</Label>
						<Select.Root portal={null}>
							<Select.Trigger class="w-full">
								<Select.Value placeholder="Select a RAID" />
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each raid as fruit}
										<Select.Item value={fruit.value} label={fruit.label}>{fruit.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
							<Select.Input name="favoriteFruit" />
						</Select.Root>
					</div>
					<div>
						<Label class="w-24 whitespace-nowrap text-sm" for="terms">Compression:</Label>
						<Select.Root portal={null}>
							<Select.Trigger class="w-full">
								<Select.Value placeholder="Select a Compression" />
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each compression as fruit}
										<Select.Item value={fruit.value} label={fruit.label}>{fruit.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
							<Select.Input name="favoriteFruit" />
						</Select.Root>
					</div>
					<div>
						<Label class="w-24 whitespace-nowrap text-sm" for="terms">ASHIFT:</Label>
						<Select.Root portal={null}>
							<Select.Trigger class="w-full">
								<Select.Value placeholder="Select a ASHIFT" />
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each ashift as fruit}
										<Select.Item value={fruit.value} label={fruit.label}>{fruit.label}</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
							<Select.Input name="favoriteFruit" />
						</Select.Root>
					</div>
				</div>
				<div transition:slide class="mt-2 flex items-center space-x-2 md:mt-0">
					<Label
						id="terms-label"
						for="terms"
						class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
					>
						Advanced
					</Label>
					<Checkbox id="terms" bind:checked={advancedChecked} aria-labelledby="terms-label" />
				</div>

				{#if advancedChecked}
					<div transition:slide class="max-h-[250px] space-y-2 overflow-y-auto">
						{#each pairs as pair, index}
							<div transition:slide class="flex items-center gap-4">
								<Input class="h-8" type="text" id="name" placeholder="key" bind:value={pair.key} />

								<Input
									class="h-8"
									type="text"
									id="name"
									placeholder="value"
									bind:value={pair.value}
								/>

								{#if pairs.length > 1}
									<button
										onclick={() => removePair(index)}
										class="rounded px-2 py-1 text-white hover:bg-muted"
									>
										<Icon icon="ic:twotone-remove" class="h-5 w-5" />
									</button>
								{/if}
							</div>
						{/each}
					</div>
					<div transition:slide class="flex justify-end">
						<button onclick={addPair} class=" rounded px-3 py-1 text-white hover:bg-muted">
							<Icon icon="icons8:plus" class="h-6 w-6" />
						</button>
					</div>
				{/if}
			{/if}
		</div>

		<Dialog.Footer class="h-12">
			<div class="flex w-full justify-end border-t px-3 py-3 md:flex-row">
				<div class="flex flex-col items-center gap-2 space-x-3 md:flex-row">
					<Button
						size="sm"
						type="button"
						class="h-7 w-full bg-blue-700 text-white hover:bg-blue-600"
					>
						Confirm
					</Button>
				</div>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	:global(div.multiselect > ul.options) {
		border: 1px solid #292524;
	}
	:global(div.multiselect) {
		@apply h-[32px] rounded border border-[#292524] p-1;
	}
</style>
