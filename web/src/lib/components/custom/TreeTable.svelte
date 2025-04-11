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
		parentIcon?: string;
		itemIcon?: string;
		parentActiveRow?: Row | null;
	}

	let {
		data,
		name,
		parentIcon = undefined,
		itemIcon = undefined,
		parentActiveRow = $bindable()
	}: Props = $props();

	let columns = $derived.by(() => {
		return data.columns.map((column) => {
			return {
				title: column.label,
				field: column.key
			};
		});
	});

	let rows = $derived.by(() => {
		function clean(row: Row) {
			let newRow = { ...row };

			if (Array.isArray(newRow.children)) {
				let cleanedChildren = newRow.children.map(clean).filter(Boolean);

				if (cleanedChildren.length > 0) {
					newRow.children = cleanedChildren;
				} else {
					delete newRow.children;
				}
			}

			return newRow;
		}

		return data.rows.map(clean);
	});

	$effect(() => {
		untrack(() => {
			if (table) {
				table.updateData(data.rows);
			}
		});
	});

	onMount(() => {
		if (tableComponent) {
			table = new Tabulator(tableComponent, {
				layout: 'fitColumns',
				data: rows,
				reactiveData: true,
				selectableRows: 1,
				dataTreeChildIndent: 16,
				dataTree: true,
				dataTreeChildField: 'children',
				columns: columns
			});
		}

		table?.on('rowSelected', function (row: RowComponent) {
			const selectedRow = row.getData();
			for (const column of columns) {
				parentActiveRow = {
					...parentActiveRow,
					[`${column.field}`]: selectedRow[column.field],
					id: selectedRow.id ?? parentActiveRow?.id ?? 0
				};
			}
		});

		table?.on('rowDeselected', function (row: RowComponent) {
			console.log('rowDeselected', row);
			parentActiveRow = null;
		});
	});
</script>

<div bind:this={tableComponent}></div>
