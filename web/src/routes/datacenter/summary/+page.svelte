<script lang="ts">
	import { getNodes } from '$lib/api/cluster/cluster';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import type { ClusterNode } from '$lib/types/cluster/cluster';
	import { updateCache } from '$lib/utils/http';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { dateToAgo } from '$lib/utils/time';
	import Icon from '@iconify/svelte';
	import { useQueries, useQueryClient } from '@sveltestack/svelte-query';

	interface Data {
		nodes: ClusterNode[];
	}

	let { data }: { data: Data } = $props();

	const queryClient = useQueryClient();
	let results = useQueries([
		{
			queryKey: 'cluster-nodes',
			queryFn: async () => {
				return (await getNodes()) as ClusterNode[];
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.nodes,
			refetchOnMount: 'always',
			onSuccess: (data: ClusterNode[]) => {
				updateCache('cluster-nodes', data);
			}
		}
	]);

	let nodes = $derived($results[0].data ?? []);
</script>

<div class="flex h-full w-full flex-col">
	<div class="min-h-0 flex-1">
		<div class="p-4">
			<Card.Root class="gap-2">
				<Card.Header>
					<Card.Title>
						<div class="flex items-center gap-2">
							<Icon icon="fa7-solid:hexagon-nodes" />
							<span>Nodes</span>
						</div>
					</Card.Title>
				</Card.Header>
				<Card.Content>
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Status</Table.Head>
								<Table.Head>Hostname</Table.Head>
								<Table.Head>ID</Table.Head>
								<Table.Head>Last Ping</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each nodes as node (node.id)}
								<Table.Row>
									<Table.Cell>
										<Badge variant="outline" class="text-muted-foreground px-1.5">
											{#if node.status === 'online'}
												<Icon icon="mdi:check-circle" class="text-green-500" />
											{:else}
												<Icon icon="mdi:close-circle" class="text-red-500" />
											{/if}
											{capitalizeFirstLetter(node.status)}
										</Badge>
									</Table.Cell>
									<Table.Cell>{node.hostname}</Table.Cell>
									<Table.Cell>{node.nodeUUID}</Table.Cell>
									<Table.Cell>{dateToAgo(node.updatedAt)}</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
				<Card.Footer></Card.Footer>
			</Card.Root>
		</div>
	</div>
</div>
