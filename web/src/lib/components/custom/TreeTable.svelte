<script lang="ts">
	import type { Disk } from '$lib/types/disk/disk';
	import Icon from '@iconify/svelte';
	import { TableHandler } from '@vincjo/datatables';
	import humanFormat from 'human-format';
	import { onMount } from 'svelte';

	let { data }: { data: Disk[] } = $props();

	const table = new TableHandler(data as Disk[]);
	type ExpandedRows = Record<number, boolean>;
	const expandedRows: ExpandedRows = $state({});
	let activeRow: string | null = $state(null); // Use string to support child rows uniquely

	function toggleChildren(index: number) {
		expandedRows[index] = !expandedRows[index];

		// If expanding, also make it active
		if (expandedRows[index]) {
			activeRow = index.toString();
		}
	}

	function handleRowClick(index: number) {
		// Toggle active state
		activeRow = activeRow === index.toString() ? null : index.toString();
	}

	function handleDoubleClick(index: number) {
		alert('hello world');
	}

	function isToggled(index: number) {
		return expandedRows[index] ?? false;
	}

	const keys = [
		'Device',
		'Type',
		'Usage',
		'Size',
		'GPT',
		'Model',
		'Serial',
		'S.M.A.R.T.',
		'Wearout'
	];

	let sortHandlers: Record<string, any> = {};

	keys.forEach((key) => {
		sortHandlers[key] = table.createSort(key as keyof Disk, {
			locales: 'en',
			options: { numeric: true, sensitivity: 'base' }
		});
	});
</script>

<div class="relative flex h-full w-full flex-col">
	<div class="flex-1">
		<div class="h-full overflow-y-auto">
			<table class="mb-10 w-full min-w-max border-collapse">
				<thead>
					<tr>
						{#each keys as key}
							<th
								class="h-8 w-48 cursor-pointer whitespace-nowrap border-b border-t px-3 text-left text-black dark:text-white"
								onclick={() => {
									sortHandlers[key].set();
								}}
							>
								<div class="flex">
									<span class="mr-1">{key}</span>
									{#if sortHandlers[key].field === key}
										<Icon
											icon={sortHandlers[key].direction === 'asc'
												? 'lucide:arrow-up'
												: 'lucide:arrow-down'}
											class="mt-1 h-4 w-4"
										/>
									{/if}
								</div>
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>
					{#each table.rows as row, index}
						<tr
							class={activeRow === index.toString() ? 'bg-muted-foreground/40 dark:bg-muted' : ''}
							onclick={(event: MouseEvent) => {
								// Ensure the click wasn't on the expand icon
								if (!(event.target as HTMLElement).closest('.toggle-icon')) {
									handleRowClick(index);
								}
							}}
							ondblclick={() => handleDoubleClick(index)}
						>
							{#each keys as key, keyIndex}
								{#if key === 'Device'}
									<td class="whitespace-nowrap px-3 py-1.5">
										<div class="flex items-center">
											<Icon
												icon={isToggled(index) ? 'lucide:minus-square' : 'lucide:plus-square'}
												class="toggle-icon mr-1.5 h-4 w-4 cursor-pointer"
												onclick={(event: MouseEvent) => {
													event.stopPropagation(); // Prevents row click event from firing
													toggleChildren(index);
												}}
											/>
											<Icon icon="mdi:harddisk" class="mr-1.5 h-4 w-4" />
											<span>{row.Device}</span>
										</div>
									</td>
								{:else if key === 'Size'}
									<td class="whitespace-nowrap px-3 py-1.5">{humanFormat(row.Size)}</td>
								{:else}
									<td class="whitespace-nowrap px-3 py-1.5">{row[key as keyof Disk]}</td>
								{/if}
							{/each}
						</tr>
						{#if expandedRows[index] && row.Partitions}
							{#each row.Partitions as child, childIndex}
								<tr
									class={activeRow === `${index}-${childIndex}`
										? 'bg-muted-foreground/40 dark:bg-muted'
										: ''}
									onclick={() => handleRowClick(`${index}-${childIndex}`)}
								>
									{#each keys as key, keyIndex}
										{#if key === 'Device'}
											<td class="whitespace-nowrap px-3 py-0">
												<div class="relative flex items-center">
													{#if row.Partitions.length > 1}
														<div
															class="bg-muted-foreground absolute left-1.5 top-0 h-full w-0.5"
															style="height: calc(100% + 0.8rem);"
															class:hidden={childIndex === row.Partitions.length - 1}
														></div>
													{:else}
														<div
															class="bg-muted-foreground absolute left-1.5 top-0 h-3 w-0.5"
														></div>
													{/if}
													<div class="relative left-1.5 top-0 mr-2 w-4">
														<div class="bg-muted-foreground h-0.5 w-4"></div>
													</div>
													{#if childIndex === row.Partitions.length - 1}
														<div class="absolute bottom-0 left-2 h-1/2 w-0.5 bg-transparent"></div>
													{/if}
													<Icon icon="mdi:harddisk" class="mr-1.5 h-4 w-4" />
													<span>{child.name}</span>
												</div>
											</td>
										{:else if key === 'Type'}
											<td class="whitespace-nowrap px-3 py-0">partition</td>
										{:else if key === 'Usage'}
											<td class="whitespace-nowrap px-3 py-0">{child.usage}</td>
										{:else if key === 'Size'}
											<td class="whitespace-nowrap px-3 py-0">{humanFormat(child.size)}</td>
										{:else if key === 'GPT'}
											<td class="whitespace-nowrap px-3 py-0">{row.GPT}</td>
										{:else}
											<td class="whitespace-nowrap px-3 py-0"></td>
										{/if}
									{/each}
								</tr>
							{/each}
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
