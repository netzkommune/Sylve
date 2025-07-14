<script lang="ts">
	import { listUsers } from '$lib/api/auth/local';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import type { User } from '$lib/types/auth';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { updateCache } from '$lib/utils/http';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		users: User[];
	}

	let { data }: { data: Data } = $props();

	function generateTableData(users: User[]): { rows: Row[]; columns: Column[] } {
		const columns: Column[] = [
			{ field: 'id', title: 'ID', visible: false },
			{ field: 'name', title: 'Name' }
		];

		const rows: Row[] = users.map((user) => ({
			id: user.id,
			name: user.username
		}));

		return { rows, columns };
	}

	const results = useQueries([
		{
			queryKey: ['users'],
			queryFn: async () => {
				return (await listUsers()) as User[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.users,
			onSuccess: (data: User[]) => {
				updateCache('users', data);
			}
		}
	]);

	let users: User[] = $derived($results[0].data as User[]);
	let tableData = $derived(generateTableData(users));
	let query: string = $state('');
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
</script>

<div class="flex h-full flex-col overflow-hidden">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />
	</div>

	<TreeTable
		data={tableData}
		name={'tt-users'}
		bind:parentActiveRow={activeRows}
		multipleSelect={false}
		bind:query
	/>
</div>
