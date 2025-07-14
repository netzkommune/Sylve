<script lang="ts">
	import { deleteUser, listUsers } from '$lib/api/auth/local';
	import { handleAPIResponse } from '$lib/api/common';
	import Create from '$lib/components/custom/Authentication/Create.svelte';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import type { User } from '$lib/types/auth';
	import type { Column, Row } from '$lib/types/components/tree-table';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import { convertDbTime, getLastUsage } from '$lib/utils/time';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';
	import type { CellComponent } from 'tabulator-tables';

	interface Data {
		users: User[];
	}

	let { data }: { data: Data } = $props();

	function generateTableData(users: User[]): { rows: Row[]; columns: Column[] } {
		const columns: Column[] = [
			{ field: 'id', title: 'ID', visible: false },
			{ field: 'name', title: 'Name' },
			{
				field: 'email',
				title: 'E-Mail',
				formatter: (cell: CellComponent) => {
					const value = cell.getValue();
					return value ? value : '-';
				}
			},
			{
				field: 'lastUsage',
				title: 'Last Usage',
				formatter: (cell: CellComponent) => {
					const value = cell.getValue();
					return getLastUsage(value);
				}
			},
			{
				field: 'createdAt',
				title: 'Created At',
				formatter: (cell: CellComponent) => {
					const value = cell.getValue();
					return convertDbTime(value);
				}
			}
		];

		const rows: Row[] = users.map((user) => ({
			id: user.id,
			name: user.username,
			email: user.email,
			lastUsage: user.lastLoginTime ? convertDbTime(user.lastLoginTime) : 'Never',
			createdAt: user.createdAt ? convertDbTime(user.createdAt) : 'Never'
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

	let modals = $state({
		create: { open: false },
		delete: { open: false }
	});
</script>

{#snippet button(type: string)}
	{#if type === 'delete'}
		{#if activeRows && activeRows.length === 1}
			<Button
				onclick={() => {
					modals.delete.open = !modals.delete.open;
				}}
				size="sm"
				variant="outline"
				class="h-6.5"
			>
				<div class="flex items-center">
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					<span>Delete</span>
				</div>
			</Button>
		{/if}
	{/if}
{/snippet}

<div class="flex h-full flex-col overflow-hidden">
	<div class="flex h-10 w-full items-center gap-2 border-b p-2">
		<Search bind:query />

		<Button onclick={() => (modals.create.open = !modals.create.open)} size="sm" class="h-6">
			<div class="flex items-center">
				<Icon icon="gg:add" class="mr-1 h-4 w-4" />
				<span>New</span>
			</div>
		</Button>

		{@render button('delete')}
	</div>

	<TreeTable
		data={tableData}
		name={'tt-users'}
		bind:parentActiveRow={activeRows}
		multipleSelect={false}
		bind:query
	/>
</div>

{#if modals.create.open}
	<Create bind:open={modals.create.open} {users} />
{/if}

<AlertDialog
	bind:open={modals.delete.open}
	names={{
		parent: 'User',
		element: activeRow ? (activeRow.name as string) : ''
	}}
	actions={{
		onConfirm: async () => {
			const response = await deleteUser(activeRow.id as string);
			console.log(response);
			if (response.error) {
				handleAPIError(response);
				toast.error('Failed to delete user', {
					position: 'bottom-center'
				});
			} else {
				toast.success('User deleted', {
					position: 'bottom-center'
				});
			}

			modals.delete.open = false;
		},
		onCancel: () => {
			modals.delete.open = false;
		}
	}}
/>
