<script lang="ts">
	import type { Column, ExpandedRows, Row } from '$lib/types/components/tree-table';
	import { onMount, untrack } from 'svelte';
	import { TabulatorFull as Tabulator } from 'tabulator-tables';

	let tableComponent: Tabulator;
	let table = null;

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
		function clean(row) {
			// Create a shallow copy of the row
			let newRow = { ...row };

			if (Array.isArray(newRow.children)) {
				// Recursively clean children
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
		table = new Tabulator(tableComponent, {
			layout: 'fitColumns',
			data: rows,
			reactiveData: true,
			dataTree: true,
			dataTreeChildField: 'children',
			columns: columns
		});
	});
</script>

<div bind:this={tableComponent}></div>
