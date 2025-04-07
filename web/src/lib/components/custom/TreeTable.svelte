<script lang="ts">
	import * as ContextMenu from '$lib/components/ui/context-menu';
	import { localStore } from '$lib/stores/localStore.svelte';
	import Icon from '@iconify/svelte';
	import { TableHandler } from '@vincjo/datatables';

	interface ZFSPools {
		Id: number;
		Name: string;
		Size: string;
		Health: string;
		Redundancy: string;
		Children?: ZFSPools[];
	}

	interface Props {
		data: ZFSPools[];
		keys: string[];
	}

	let { data, keys }: Props = $props();

	const table = new TableHandler(data);

	let activeRow: string | null = $state(null);
	type ExpandedRows = Record<number, boolean>;

	const expandedRows: ExpandedRows = $state({});

	function toggleChildren(index: number) {
		expandedRows[index] = !expandedRows[index];
		if (expandedRows[index]) {
			activeRow = index.toString();
		}
	}

	function isToggled(index: number) {
		return expandedRows[index] ?? false;
	}

	let visibleColumns = localStore(
		'zfsVisibleColumns',
		Object.fromEntries(keys.map((key) => [key, true]))
	);

	let openContextMenuId = $state<string | null>(null);
	let sortHandlers: Record<
		string,
		{
			direction: 'asc' | 'desc' | null;
			set: () => void;
		}
	> = Object.fromEntries(
		keys.map((key) => [
			key,
			{
				direction: null,
				set: () => {
					// Implement sorting logic here
					const currentHandler = sortHandlers[key];
					currentHandler.direction =
						currentHandler.direction === null
							? 'asc'
							: currentHandler.direction === 'asc'
								? 'desc'
								: null;
				}
			}
		])
	);

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
</script>

{#snippet treeRow(
	item: ZFSPools,
	level = 0,
	isLast = false,
	hasChildNodes = false,
	ancestorsIsLast: boolean[] = []
)}
	<tr>
		<td class="h-8 whitespace-nowrap px-3 py-1 text-left text-black dark:text-white">
			<div class="relative flex items-center">
				{#if level > 0}
					<!-- Indentation + vertical guides -->
					{#each Array(level) as _, i}
						<div class="relative flex h-full w-5 items-center justify-center">
							<!-- Vertical line (except for current branch line) -->
							{#if i < level - 1}
								{#if !ancestorsIsLast[i]}
									<div
										class="bg-muted-foreground absolute -top-5 left-1.5 h-full w-0.5"
										style="height: calc(100% + 2rem);"
									></div>
								{/if}
							{:else}
								<!-- Branch line: vertical (if not last), and horizontal always -->
								{#if !isLast}
									<div
										class="bg-muted-foreground absolute -top-5 left-1.5 h-full w-0.5"
										style="height: calc(100% + 2rem);"
									></div>
								{/if}
								<div
									class="bg-muted-foreground absolute -top-5 left-1.5 h-full w-0.5"
									style="height: calc(100% + 1.2rem);"
								></div>
								<!-- Horizontal connector -->
								<div class="bg-muted-foreground relative left-1.5 h-0.5 w-4"></div>
							{/if}
						</div>
					{/each}
				{/if}
				<!-- Expand/collapse toggle icon if there are children -->
				{#if hasChildNodes}
					<Icon
						icon={isToggled(item.Id) ? 'lucide:minus-square' : 'lucide:plus-square'}
						class="toggle-icon mr-1.5 h-4 w-4 cursor-pointer"
						onclick={(event: MouseEvent) => {
							event.stopPropagation();
							toggleChildren(item.Id);
						}}
					/>
				{/if}
				<Icon icon="mdi:harddisk" class="mr-1.5 h-4 w-4" />
				<span class="truncate">{item.Name}</span>
			</div>
		</td>
		<td class="h-8 whitespace-nowrap px-3 py-1 text-left text-black dark:text-white">
			{item.Size}
		</td>
		<td class="h-8 whitespace-nowrap px-3 py-1 text-left text-black dark:text-white">
			{item.Health}
		</td>
		<td class="h-8 whitespace-nowrap px-3 py-1 text-left text-black dark:text-white">
			{item.Redundancy}
		</td>
	</tr>
{/snippet}

<table class="mb-10 w-full min-w-max border-collapse">
	<thead>
		<tr>
			{#each keys as key}
				{#if visibleColumns.value[key]}
					<!-- <th
						class="group h-8 w-48 whitespace-nowrap border border-neutral-300 px-3 text-left text-black dark:border-neutral-800 dark:text-white"
					>
						<span>{key}</span>
					</th> -->
					<th>
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
		{#each data as pool (pool.Id)}
			<!-- Render root pool row -->

			{@render treeRow(pool, 0, false, pool.Children && pool.Children.length > 0, [])}
			{#if isToggled(pool.Id)}
				{#if pool.Children && pool.Children.length > 0}
					{#each pool.Children as child, i (child.Id)}
						<!-- checking pool.children has own children -->
						{#if child.Children && child.Children.length > 0}
							<!-- if Child has its own children -->
							{@render treeRow(child, 1, i === pool.Children.length - 1, true, [
								i === pool.Children.length - 1
							])}
							{#if isToggled(child.Id)}
								{#each child.Children as grandchild, j (grandchild.Id)}
									{@render treeRow(grandchild, 2, j === child.Children.length - 1, false, [
										i === pool.Children.length - 1,
										j === child.Children.length - 1
									])}
								{/each}
							{/if}
						{:else}
							<!-- Regular child without parent reference -->
							{@render treeRow(
								child,
								1,
								i === pool.Children.length - 1,
								child.Children && child.Children.length > 0,
								[i === pool.Children.length - 1]
							)}
						{/if}
					{/each}
				{/if}
			{/if}
		{/each}
	</tbody>
</table>
