<script lang="ts">
	import { formatAction, formatStatus, getAuditLogs } from '$lib/api/info/audit';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import type { AuditLog } from '$lib/types/info/audit';
	import { convertDbTime } from '$lib/utils/time';
	import { useQueries } from '@sveltestack/svelte-query';

	const results = useQueries([
		{
			queryKey: ['auditLog'],
			queryFn: async () => {
				return await getAuditLogs();
			},
			refetchInterval: 1000,
			keepPreviousData: true
		}
	]);

	let logs = $derived($results[0].data as AuditLog);
</script>

<Tabs.Root value="cluster" class="flex h-full w-full flex-col">
	<Tabs.Content value="cluster" class="flex h-full flex-col border">
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
					{#each logs as log, i (i)}
						<Table.Row>
							<Table.Cell class="h-10 px-4 py-2">{convertDbTime(log.started)}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{convertDbTime(log.ended)}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{log.node}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{`${log.user}@${log.authType}`}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{formatAction(log.action)}</Table.Cell>
							<Table.Cell class="h-10 px-4 py-2">{formatStatus(log.status)}</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	</Tabs.Content>
</Tabs.Root>
