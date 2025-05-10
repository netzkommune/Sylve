<script lang="ts">
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { matchAny } from '$lib/utils/table';
	import { findRow, getAllRows, pruneEmptyChildren } from '$lib/utils/tree-table';
	import { onMount, untrack } from 'svelte';
	import {
		TabulatorFull as Tabulator,
		type ColumnDefinition,
		type RowComponent
	} from 'tabulator-tables';

	let tableComponent: HTMLDivElement | null = null;
	let table: Tabulator | null = $state(null);

	interface Props {
		data: {
			rows: Row[];
			columns: Column[];
		};
		name: string;
		parentActiveRow?: Row | null;
		query?: string;
	}

	let { data, name, parentActiveRow = $bindable(), query = $bindable() }: Props = $props();
	let tableInitialized = $state(false);
	let scroll = $state([0, 0]);

	function selectParentActiveRow(row: RowComponent) {
		const expandedRow = row.getData();
		for (const column of data.columns) {
			parentActiveRow = {
				...parentActiveRow,
				[`${column.field}`]: expandedRow[column.field],
				id: expandedRow.id ?? parentActiveRow?.id ?? 0
			};
		}
	}

	$effect(() => {
		if (data.rows) {
			untrack(async () => {
				if (query && query !== '') {
					return;
				}

				if (data.rows.length === 0) {
					if (data.rows.length === 0) {
						table?.clearData();
					}

					return;
				}

				const selectedIds = table?.getSelectedRows().map((row) => row.getData().id) || [];
				const treeExpands =
					getAllRows(table?.getRows() || []).map((row) => ({
						id: row.getData().id,
						expanded: row.isTreeExpanded()
					})) || [];

				await table?.replaceData(pruneEmptyChildren(data.rows));

				selectedIds.forEach((id) => {
					const row = findRow(table?.getRows() || [], id);
					if (row) {
						row.select();
						selectParentActiveRow(row);
					}
				});

				treeExpands?.forEach((treeExpand) => {
					if (treeExpand.expanded) {
						const row = findRow(table?.getRows() || [], treeExpand.id);
						if (row) {
							row.treeExpand();
						}
					} else {
						const row = findRow(table?.getRows() || [], treeExpand.id);
						if (row) {
							row.treeCollapse();
						}
					}
				});
			});
		}
	});

	onMount(() => {
		if (tableComponent) {
			table = new Tabulator(tableComponent, {
				data: pruneEmptyChildren(data.rows),
				reactiveData: true,
				columns: data.columns as ColumnDefinition[],
				layout: 'fitColumns',
				selectableRows: 1,
				dataTreeChildIndent: 16,
				dataTree: true,
				dataTreeChildField: 'children',
				dataTreeStartExpanded: true,
				persistenceID: name,
				paginationMode: 'local',
				persistence: {
					sort: true,
					page: true,
					filter: true
				},
				placeholder: 'No data available',
				pagination: true,
				paginationSize: 25,
				paginationCounter: 'pages'
			});
		}

		table?.on('rowSelected', function (row: RowComponent) {
			selectParentActiveRow(row);
		});

		table?.on('rowDeselected', function (row: RowComponent) {
			parentActiveRow = null;
		});

		table?.on('rowDblClick', function (event: UIEvent, row: RowComponent) {
			selectParentActiveRow(row);
		});

		table?.on('tableBuilt', function () {
			tableInitialized = true;
		});

		table?.on('scrollVertical', function (top) {
			scroll = [top, scroll[1]];
		});

		table?.on('scrollHorizontal', function (left) {
			scroll = [scroll[0], left];
		});

		table?.on('renderComplete', () => {
			const container = document.querySelector('.tabulator-tableholder') as HTMLDivElement;
			if (container) {
				container.scrollTop = scroll[0];
				container.scrollLeft = scroll[1];
			}
		});
	});

	function tableFilter(query: string) {
		if (table && tableInitialized) {
			if (query === '') {
				table.clearFilter(true);
				return;
			}
			table.setFilter(matchAny, { query: query });
		}
	}

	$effect(() => {
		tableFilter(query || '');
	});
</script>

<div bind:this={tableComponent} class="flex-1 cursor-pointer" id={name}></div>
