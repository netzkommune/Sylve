<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import KvTableModal from '$lib/components/custom/KVTableModal.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Iface } from '$lib/types/network/iface';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { generateTableData, getCleanIfaceData } from '$lib/utils/network/iface';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		interfaces: Iface[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['networkInterfaces'],
			queryFn: async () => {
				return await getInterfaces();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.interfaces,
			onSuccess: (data: Iface[]) => {
				updateCache('networkInterfaces', data);
			}
		}
	]);

	let tableData = $derived(generateTableData($results[0].data as Iface[]));
	let activeRow: Row[] | null = $state(null);
	let query: string = $state('');
	let viewModal = $state({
		title: '',
		key: getTranslation('disk.attribute', 'Attribute'),
		value: getTranslation('disk.value', 'Value'),
		open: false,
		KV: {},
		type: 'kv',
		actions: {
			close: () => {
				viewModal.open = false;
			}
		}
	});

	function viewInterface(iface: string) {
		const ifaceData = $results[0].data?.find((i: Iface) => i.name === iface);
		if (ifaceData) {
			viewModal.KV = getCleanIfaceData(ifaceData);
			viewModal.title = `${capitalizeFirstLetter(getTranslation('common.details', 'Details'))} - ${ifaceData.name}`;
			viewModal.open = true;
		}
	}
</script>

{#snippet button(type: string)}
	{#if type === 'view' && activeRow !== null && activeRow.length > 0}
		<Button
			onclick={() => activeRow !== null && viewInterface(activeRow[0]?.name)}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:eye" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.view', 'View'))}
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
		{@render button('view')}
	</div>

	<KvTableModal
		titles={{
			icon: 'carbon:network-interface',
			main: viewModal.title,
			key: viewModal.key,
			value: viewModal.value
		}}
		open={viewModal.open}
		KV={viewModal.KV}
		type={viewModal.type}
		actions={viewModal.actions}
	></KvTableModal>

	<TreeTable
		data={tableData}
		name="tt-networkInterfaces"
		multipleSelect={false}
		bind:parentActiveRow={activeRow}
		bind:query
	/>
</div>
