<script lang="ts">
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { generateComboboxOptions } from '$lib/utils/input';
	import Icon from '@iconify/svelte';
	import type { CellComponent } from 'tabulator-tables';

	let modals = $state({
		open: false
	});

	let comboBoxes = $state({
		type: {
			open: false,
			value: '',
			types: ['Host', 'Network', 'Port', 'Country']
		},
		list: {
			open: false,
			value: '',
			types: ['List1', 'List2', 'List3']
		}
	});

	let confirmModals = $state({
		name: '',
		type: ''
	});

	let columns: Column[] = $state([
		{
			field: 'id',
			title: 'ID',
			visible: false
		},
		{
			field: 'name',
			title: 'Name',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value || '-';
			}
		},
		{
			field: 'type',
			title: 'Type',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value?.toUpperCase?.() || '-';
			}
		},
		{
			field: 'data',
			title: 'Data',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return typeof value === 'object' ? JSON.stringify(value) : value || '-';
			}
		},
		{
			field: 'updatedAt',
			title: 'Updated At',
			formatter: (cell: CellComponent) => {
				const value = cell.getValue();
				return value ? new Date(value).toLocaleString() : '-';
			}
		}
	]);

	const tableData: { rows: Row[]; columns: Column[] } = {
		columns,
		rows: [
			{
				id: '1',
				name: 'Dummy Object A',
				type: 'host',
				data: 'Dummy data a',
				updatedAt: new Date().toISOString()
			},
			{
				id: '2',
				name: 'Dummy Object B',
				type: 'port',
				data: 'Dummy data b',
				updatedAt: new Date(Date.now() - 1000 * 60 * 60).toISOString(),
				children: [
					{
						id: '2-1',
						name: 'Child Object B1',
						type: 'host',
						data: 'Dummy data b1',
						updatedAt: new Date(Date.now() - 1000 * 60 * 10).toISOString()
					}
				]
			},
			{
				id: '3',
				name: 'Dummy Object C',
				type: 'network',
				data: 'Dummy data c',
				updatedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString()
			}
		]
	};

	let activeRow: Row[] | null = $state(null);
	let query: string = $state('');

	function resetModal() {
		confirmModals.name = '';
		comboBoxes.type.value = '';
		comboBoxes.list.value = '';
		comboBoxes.type.open = false;
		comboBoxes.list.open = false;
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Button size="sm" class="h-6" onclick={() => (modals.open = !modals.open)}>Open</Button>
	</div>

	<TreeTable
		data={tableData}
		name="tt-network-objects"
		multipleSelect={false}
		bind:parentActiveRow={activeRow}
		bind:query
	/>
</div>

<Dialog.Root bind:open={modals.open}>
	<Dialog.Content>
		<div class="flex items-center justify-between">
			<Dialog.Header>
				<Dialog.Title>
					<div class="flex items-center">Object</div>
				</Dialog.Title>
			</Dialog.Header>

			<div class="flex items-center gap-0.5">
				<Button size="sm" variant="link" class="h-4" title={'Reset'} onclick={resetModal}>
					<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Reset'}</span>
				</Button>
				<Button
					size="sm"
					variant="link"
					class="h-4"
					title={'Close'}
					onclick={() => (modals.open = false)}
				>
					<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
					<span class="sr-only">{'Close'}</span>
				</Button>
			</div>
		</div>
		<div class="flex gap-4">
			<CustomValueInput
				label={'Name'}
				placeholder="Object Name"
				bind:value={confirmModals.name}
				classes="flex-1 space-y-1.5"
				type="text"
			/>

			<CustomComboBox
				bind:open={comboBoxes.type.open}
				label={'Type'}
				bind:value={comboBoxes.type.value}
				data={generateComboboxOptions(comboBoxes.type.types)}
				classes="flex-1 space-y-1"
				placeholder="Select type"
				width="w-3/4"
			></CustomComboBox>
		</div>

		{#if comboBoxes.type.value !== ''}
			<div class="flex gap-4">
				{#if comboBoxes.type.value === 'Host' || comboBoxes.type.value === 'Network'}
					<CustomValueInput
						placeholder={`Enter ${comboBoxes.type.value.toLowerCase()} name`}
						bind:value={confirmModals.type}
						classes="flex-1 space-y-1.5"
						type="text"
					/>
				{:else if comboBoxes.type.value === 'Port'}
					<CustomValueInput
						placeholder="Port Number"
						bind:value={confirmModals.type}
						classes="flex-1 space-y-1.5"
						type="number"
					/>
				{:else if comboBoxes.type.value === 'Country'}
					<CustomComboBox
						bind:open={comboBoxes.type.open}
						label={'Option 1'}
						bind:value={comboBoxes.type.value}
						data={generateComboboxOptions(comboBoxes.type.types)}
						classes="flex-1 space-y-1"
						placeholder="Select type"
						width="w-3/4"
					></CustomComboBox>
					<CustomComboBox
						bind:open={comboBoxes.type.open}
						label={'option 1'}
						bind:value={comboBoxes.type.value}
						data={generateComboboxOptions(comboBoxes.type.types)}
						classes="flex-1 space-y-1"
						placeholder="Select type"
						width="w-3/4"
					></CustomComboBox>
				{/if}
			</div>
		{/if}

		<div class="flex gap-4">
			<CustomComboBox
				bind:open={comboBoxes.list.open}
				label={'List'}
				bind:value={comboBoxes.list.value}
				data={generateComboboxOptions(comboBoxes.list.types)}
				classes="flex-1 space-y-1"
				placeholder="Select type"
				width="w-3/4"
			></CustomComboBox>
		</div>
	</Dialog.Content>
</Dialog.Root>
