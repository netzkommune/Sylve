<script lang="ts">
	import type { Column, Row } from '$lib/types/components/tree-table';
	import Icon from '@iconify/svelte';
	import { onMount, untrack } from 'svelte';
	import { TabulatorFull as Tabulator, type RowComponent } from 'tabulator-tables';

	let tableComponent: HTMLDivElement | null = null;
	let table: Tabulator | null = $state(null);

	interface Props {
		data: {
			rows: Row[];
			columns: Column[];
		};
		name: string;
		hideId?: boolean;
		parentIcon?: string;
		itemIcon?: string;
		parentActiveRow?: Row | null;
	}

	let {
		data,
		name,
		parentIcon = undefined,
		itemIcon = undefined,
		hideId = true,
		parentActiveRow = $bindable()
	}: Props = $props();

	let columns = $derived.by(() => {
		return data.columns.map((column) => {
			return {
				title: column.label,
				field: column.key,
				visible: column.key !== 'id' || !hideId
			};
		});
	});

	$effect(() => {
		if (data.rows.length > 0) {
			untrack(() => {
				if (parentActiveRow === null) {
					if (table) {
						table?.updateOrAddData(data.rows);
					}
				}
			});
		}
	});

	function selectParentActiveRow(row: RowComponent) {
		const expandedRow = row.getData();
		for (const column of columns) {
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
				data: data.rows,
				selectableRows: 1,
				dataTreeChildIndent: 16,
				dataTree: true,
				dataTreeChildField: 'children',
				columns: columns,
				dataTreeStartExpanded: true
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
	});
</script>

<div class="flex h-full flex-col">
	<div bind:this={tableComponent} class="flex-1 overflow-auto"></div>
</div>
