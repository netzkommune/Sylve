<script lang="ts">
	import * as ContextMenu from '$lib/components/ui/context-menu';
	import { localStore } from '$lib/stores/localStore.svelte';
	import type { Column, ExpandedRows, Row } from '$lib/types/components/tree-table';
	import Icon from '@iconify/svelte';
	import { TableHandler } from '@vincjo/datatables';
	import { onMount, untrack } from 'svelte';
	import TableRow from '../ui/table/table-row.svelte';

	interface Props {
		data: {
			rows: Row[];
			columns: Column[];
		};
		name: string;
		itemIcon?: string;
	}

	let { data, name, itemIcon = undefined }: Props = $props();

	const table = new TableHandler(data.rows);

	let activeRow: string | null = $state(null);
	let expandedRows: ExpandedRows = $state({});
	let sortHandlers: Record<string, any> = $state({});
	let openContextMenuId = $state<string | null>(null);

	let visibleColumns = localStore(
		`${name}_visibleColumns`,
		data.columns.reduce((acc, col) => ({ ...acc, [col.key]: true }), {} as Record<string, boolean>)
	);

	onMount(() => {
		data.columns.forEach((column) => {
			sortHandlers[column.key] = table.createSort(column.key as keyof Row, {
				locales: 'en',
				options: { numeric: true, sensitivity: 'base' }
			});
		});

		if (data.rows.length) {
			data.rows.forEach((row) => {
				expandedRows[row.id as number] = true;
			});
		}
	});

	function handleContextMenuOpen(id: string) {
		openContextMenuId = id;
	}

	function handleContextMenuClose() {
		openContextMenuId = null;
	}

	function handleRowClick(id: string) {
		activeRow = activeRow === id ? null : id;
	}

	function toggleChildren(id: number) {
		expandedRows[id] = !expandedRows[id];
	}

	function isToggled(id: number) {
		return expandedRows[id] ?? false;
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
		visibleColumns.value = data.columns.reduce(
			(acc, column) => ({ ...acc, [column.key]: true }),
			{} as Record<string, boolean>
		);
	}

	function flattenTree(rows: Row[], level = 0, parentId: string | number | null = null): any[] {
		let result: any[] = [];

		rows.forEach((row, index) => {
			const isLast = index === rows.length - 1;
			const flatRow = {
				...row,
				__level: level,
				__isLast: isLast,
				__parentId: parentId,
				__hasChildren: row.children && row.children.length > 0
			};

			result.push(flatRow);

			if (row.children && row.children.length > 0 && expandedRows[row.id as number]) {
				result = result.concat(flattenTree(row.children, level + 1, row.id as number));
			}
		});

		return result;
	}

	let flattenedRows = $derived(flattenTree(data.rows));

	$effect(() => {
		untrack(() => {
			flattenedRows = flattenTree(table.rows as TableRow[]);
		});
	});
</script>

<div class="overflow-x-auto">
	<table class="mb-10 w-full min-w-max border-collapse">
		<thead>
			<tr>
				{#each data.columns as column}
					{#if visibleColumns.value[column.key]}
						<th
							class="group h-8 w-48 whitespace-nowrap border border-neutral-300 px-3 text-left text-black dark:border-neutral-800 dark:text-white"
						>
							<ContextMenu.Root
								open={openContextMenuId === column.key}
								closeOnItemClick={false}
								onOpenChange={(open) =>
									open ? handleContextMenuOpen(column.key) : handleContextMenuClose()}
							>
								<ContextMenu.Trigger class="flex h-full w-full">
									<button
										class="relative flex w-full items-center"
										onclick={() => sortHandlers[column.key]?.set()}
									>
										<span>{column.label}</span>
										<Icon
											icon={sortHandlers[column.key]?.direction === 'asc'
												? 'lucide:sort-asc'
												: 'lucide:sort-desc'}
											class="ml-2 mt-1 h-4 w-4 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
										/>
									</button>
								</ContextMenu.Trigger>
								<ContextMenu.Content>
									<ContextMenu.Label>Toggle Columns</ContextMenu.Label>
									<ContextMenu.Separator />
									{#each data.columns as colOption}
										<ContextMenu.CheckboxItem
											checked={visibleColumns.value[colOption.key]}
											onCheckedChange={() => {
												toggleColumnVisibility(colOption.key);
											}}
										>
											{colOption.label}
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
			{#if flattenedRows.length === 0}
				<tr>
					<td
						colspan={Object.values(visibleColumns.value).filter(Boolean).length || 1}
						class="h-8 px-3 py-1 text-center text-neutral-500 dark:text-neutral-400"
					>
						No data available.
					</td>
				</tr>
			{:else}
				{#each flattenedRows as row (row.id)}
					<tr
						class={activeRow === row.id ? 'bg-muted-foreground/40 dark:bg-muted' : ''}
						onclick={() => handleRowClick(row.id)}
					>
						{#each data.columns as column, colIndex}
							{#if visibleColumns.value[column.key]}
								{#if colIndex === 0}
									<td class="whitespace-nowrap px-3 py-0">
										<div class="flex items-center" style="padding-left: {row.__level * 1.5}rem;">
											{#if row.__level > 0}
												<div class="relative flex items-center">
													{#if !row.__isLast}
														<div
															class="bg-muted-foreground absolute bottom-0 left-2 h-full w-0.5"
															style="height: calc(100% + 0.8rem);"
														></div>
													{:else}
														<div
															class="bg-muted-foreground absolute bottom-0 left-2 h-3 w-0.5"
														></div>
													{/if}
													<div class="relative bottom-0 left-2 w-6">
														<div class="bg-muted-foreground h-0.5 w-6"></div>
													</div>
													{#if row.__isLast}
														<div class="absolute left-2 top-0 h-1/2 w-0.5 bg-transparent"></div>
													{/if}
												</div>
											{/if}

											{#if row.__hasChildren}
												<Icon
													icon={isToggled(row.id) ? 'lucide:minus-square' : 'lucide:plus-square'}
													class="toggle-icon mr-2 h-4 w-4 cursor-pointer"
													onclick={(event: MouseEvent) => {
														event.stopPropagation();
														toggleChildren(row.id);
													}}
												/>
											{:else}
												<span class="inline-block h-4 w-4"></span>
											{/if}

											{#if itemIcon}
												<Icon icon={itemIcon} class="mr-2 h-4 w-4" />
											{/if}

											<span>{row[column.key]}</span>
										</div>
									</td>
								{:else}
									<td class="whitespace-nowrap px-3 py-0">{row[column.key] ?? ''}</td>
								{/if}
							{/if}
						{/each}
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
</div>
