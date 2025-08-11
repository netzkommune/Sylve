<script lang="ts">
	import { getJails } from '$lib/api/jail/jail';
	import CPU from '$lib/components/custom/Jail/Hardware/CPU.svelte';
	import RAM from '$lib/components/custom/Jail/Hardware/RAM.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import type { RAMInfo } from '$lib/types/info/ram';
	import type { Jail } from '$lib/types/jail/jail';
	import { updateCache } from '$lib/utils/http';
	import { bytesToHumanReadable } from '$lib/utils/numbers';
	import { generateNanoId } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		jail: Jail;
		jails: Jail[];
		ram: RAMInfo;
	}

	let { data }: { data: Data } = $props();
	const results = useQueries([
		{
			queryKey: ['jail-list'],
			queryFn: async () => {
				return await getJails();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.jails,
			onSuccess: (data: Jail[]) => {
				updateCache('jail-list', data);
			}
		}
	]);

	let jails = $derived($results[0].data);
	let jail = $derived(jails?.find((j) => j.ctId === data.jail.ctId));

	$inspect(jail);

	let options = {
		ram: {
			value: jail?.memory,
			open: false
		},
		cpu: {
			value: jail?.cores,
			open: false
		}
	};

	let properties = $state(options);

	$effect(() => {
		if (jail) {
			properties.ram.value = jail.memory;
			properties.cpu.value = jail.cores;
		}
	});

	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
	let query = $state('');
	let table = $derived({
		columns: [
			{ title: 'Property', field: 'property' },
			{
				title: 'Value',
				field: 'value'
			}
		],
		rows: [
			{
				id: generateNanoId(`${properties.ram.value}-ram`),
				property: 'RAM',
				value: bytesToHumanReadable(properties.ram.value)
			},
			{
				id: generateNanoId(`${properties.cpu.value}-cpu`),
				property: 'CPU',
				value: properties.cpu.value
			}
		]
	});
</script>

{#snippet button(property: 'ram' | 'cpu', title: string)}
	<Button
		onclick={() => {
			properties[property].open = true;
		}}
		size="sm"
		variant="outline"
		class="h-6.5"
	>
		<div class="flex items-center">
			<Icon icon="mdi:pencil" class="mr-1 h-4 w-4" />
			<span>Edit {title}</span>
		</div>
	</Button>
{/snippet}

<div class="flex h-full w-full flex-col">
	{#if activeRows && activeRows?.length !== 0}
		<div class="flex h-10 w-full items-center gap-2 border-b p-2">
			{#if activeRow && activeRow.property === 'RAM'}
				{@render button('ram', 'RAM')}
			{/if}

			{#if activeRow && activeRow.property === 'CPU'}
				{@render button('cpu', 'CPU')}
			{/if}
		</div>
	{/if}

	<div class="flex h-full flex-col overflow-hidden">
		<TreeTable
			data={table}
			name={'jail-hardware-tt'}
			bind:parentActiveRow={activeRows}
			multipleSelect={false}
			bind:query
		/>
	</div>
</div>

{#if properties.ram.open}
	<RAM bind:open={properties.ram.open} ram={data.ram} {jail} />
{/if}

{#if properties.cpu.open}
	<CPU bind:open={properties.cpu.open} {jail} />
{/if}
