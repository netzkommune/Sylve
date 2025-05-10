<script lang="ts">
	import { getInterfaces } from '$lib/api/network/iface';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Iface } from '$lib/types/network/iface';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { generateTableData } from '$lib/utils/network/iface';
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
	let activeRow: Row | null = $state(null);
	let query: string = $state('');

	console.log($results[0].data as Iface[]);
</script>

{#snippet button(type: string)}
	{#if type === 'view'}
		<Button size="sm" class="h-6">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
	</div>

	<TreeTable
		data={tableData}
		name="tt-networkInterfaces"
		bind:parentActiveRow={activeRow}
		bind:query
	/>
</div>
