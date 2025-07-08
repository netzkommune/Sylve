<script lang="ts">
	import { getAuditRecords } from '$lib/api/info/audit';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import type { AuditRecord } from '$lib/types/info/audit';
	import { convertDbTime } from '$lib/utils/time';
	import { useQueries } from '@sveltestack/svelte-query';

	const results = useQueries([
		{
			queryKey: ['auditRecord'],
			queryFn: async () => {
				return await getAuditRecords();
			},
			refetchInterval: 1000,
			keepPreviousData: true
		}
	]);

	let data = $derived($results[0].data as AuditRecord[]);

	const pathToActionMap: Record<string, string> = {
		'/api/auth/login': 'Login',
		'/api/info/notes': 'Notes',
		'/api/network/switch': 'Switch'
	};

	let records = $derived.by(() => {
		if (!data) return [];

		return data.map((record) => {
			const path = record.action?.path || '';
			const method = record.action?.method || '';
			let resolvedAction = method;

			const matchedEntry = Object.entries(pathToActionMap).find(([prefix]) =>
				path.startsWith(prefix)
			);

			if (matchedEntry) {
				const label = matchedEntry[1];
				switch (method.toUpperCase()) {
					case 'GET':
						resolvedAction = `${label} - View`;
						break;
					case 'POST':
						resolvedAction = `${label} - Create`;
						break;
					case 'PUT':
					case 'PATCH':
						resolvedAction = `${label} - Update`;
						break;
					case 'DELETE':
						resolvedAction = `${label} - Delete`;
						record.action.body = {
							id: record.id
						};
						break;
					default:
						resolvedAction = label;
				}
			}

			if (resolvedAction === 'Login - Create') {
				resolvedAction = 'Login';
			}

			return {
				...record,
				resolvedAction
			};
		});
	});

	export function formatStatus(status: string): string {
		switch (status) {
			case 'success':
				return 'OK';
			case 'client_error':
				return 'Bad Request';
			case 'server_error':
				return 'Server Error';
			default:
				return status;
		}
	}
</script>

<Tabs.Root value="cluster" class="flex h-full w-full flex-col">
	<Tabs.Content value="cluster" class="flex h-full flex-col border-x border-b">
		<div class="flex h-full flex-col overflow-hidden">
			<Table.Root class="w-full table-fixed border-collapse">
				<Table.Header class="bg-background sticky top-0 z-[50] ">
					<Table.Row class="dark:hover:bg-background ">
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>Start Time</Table.Head
						>
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>End Time</Table.Head
						>
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>Node</Table.Head
						>
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>User</Table.Head
						>
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>Action</Table.Head
						>
						<Table.Head class="h-10 px-4 py-2 font-semibold text-black dark:text-white"
							>Status</Table.Head
						>
					</Table.Row>
				</Table.Header>

				<Table.Body class="flex-grow overflow-auto pb-32">
					{#each records as record, i (i)}
						<Table.Row>
							<Table.Cell class="h-10 px-4 py-2">{convertDbTime(record.started)}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{convertDbTime(record.ended)}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{record.node}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{`${record.user}@${record.authType}`}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2" title={JSON.stringify(record.action.body)}
								>{record.resolvedAction}</Table.Cell
							>
							<Table.Cell
								class="h-10 px-4 py-2"
								title={record.action?.response != null
									? typeof record.action.response === 'string'
										? record.action.response
										: JSON.stringify(record.action.response)
									: 'No response'}
							>
								{formatStatus(record.status)}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	</Tabs.Content>
</Tabs.Root>
