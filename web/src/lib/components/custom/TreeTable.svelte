<script lang="ts">
	import type { Column, Row } from '$lib/types/components/tree-table';
	import Icon from '@iconify/svelte';
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
	}

	let { data, name, parentActiveRow = $bindable() }: Props = $props();
	let mouseOverRow = $state(false);

	function pruneEmptyChildren(rows: Row[]): Row[] {
		return rows.map((row) => {
			const hasValidChildren = Array.isArray(row.children) && row.children.length > 0;

			const cleanedRow: Row = {
				...row,
				...(hasValidChildren
					? {
							children: pruneEmptyChildren(row.children!)
						}
					: { children: undefined })
			};

			return cleanedRow;
		});
	}

	function syncRows(currentRows: RowComponent[], newRows: Row[], parentRow?: RowComponent) {
		for (const currentRow of currentRows) {
			const currentData = currentRow.getData();
			const newRow = newRows.find((r) => r.id === currentData.id);

			if (newRow) {
				for (const column of data.columns) {
					const cell = currentRow.getCells().find((c) => c.getField() === column.field);
					if (cell) {
						const newValue = newRow[column.field];
						const oldValue = cell.getValue();
						if (newValue !== oldValue) {
							cell.setValue(newValue);
						}
					}
				}

				const existingChildren = currentRow.getTreeChildren?.() ?? [];
				const newChildren = newRow.children ?? [];

				if (newChildren.length > 0 && existingChildren.length === 0 && table) {
					const pruned = pruneEmptyChildren([newRow])[0];
					table.updateData([pruned]);
				}

				syncRows(existingChildren, newChildren, currentRow);
			}
		}

		const currentRowIds = currentRows.map((r) => r.getData().id);
		const rowsToAdd = newRows.filter((r) => !currentRowIds.includes(r.id));

		if (rowsToAdd.length > 0 && table) {
			if (parentRow) {
				parentRow.getTreeChildren().forEach((child) => child.delete());
				parentRow.update({ children: pruneEmptyChildren(rowsToAdd) });
			} else {
				const currentData = table.getData();
				const newData = [...currentData, ...pruneEmptyChildren(rowsToAdd)];
				table.setData(newData);
			}
		}

		const newRowIds = newRows.map((r) => r.id);
		for (const row of currentRows) {
			const rowData = row.getData();
			if (!newRowIds.includes(rowData.id)) {
				if (rowData.id === parentActiveRow?.id) {
					parentActiveRow = null;
				}
				row.delete();
			}
		}
	}

	$effect(() => {
		if (data.rows.length > 0 && table) {
			untrack(() => {
				const rootRows = table?.getRows();
				if (rootRows) {
					syncRows(rootRows, pruneEmptyChildren(data.rows));
				}
			});
		}
	});

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

	onMount(() => {
		if (tableComponent) {
			table = new Tabulator(tableComponent, {
				layout: 'fitColumns',
				data: pruneEmptyChildren(data.rows),
				selectableRows: 1,
				dataTreeChildIndent: 16,
				dataTree: true,
				dataTreeChildField: 'children',
				columns: data.columns as ColumnDefinition[],
				dataTreeStartExpanded: false,
				persistence: {
					sort: true
				}
			});
		}

		table?.on('rowSelected', function (row: RowComponent) {
			selectParentActiveRow(row);
		});

		table?.on('rowDeselected', function (row: RowComponent) {
			parentActiveRow = null;
		});

		table?.on('dataTreeRowExpanded', function (row: RowComponent) {
			const selectedRows = table?.getSelectedRows();
			if (selectedRows) {
				for (const selectedRow of selectedRows) {
					selectedRow.deselect();
				}
			}
			row.select();
		});

		table?.on('dataTreeRowCollapsed', function (row: RowComponent) {
			const selectedRows = table?.getSelectedRows();
			if (selectedRows) {
				for (const selectedRow of selectedRows) {
					selectedRow.deselect();
				}
			}

			row.select();
		});

		table?.on('rowMouseMove', function (e, row) {
			mouseOverRow = true;
		});
	});
</script>

<div class="flex h-full flex-col">
	<div bind:this={tableComponent} class="flex-1 overflow-auto" id={name}></div>
</div>
