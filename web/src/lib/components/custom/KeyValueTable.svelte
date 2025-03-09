<script lang="ts">
	import * as Table from '$lib/components/ui/table/index.js';

	let { data }: { data: { key: string; value: string }[] } = $props();
</script>

<Table.Root>
	<Table.Header></Table.Header>
	<Table.Body>
		{#each data as item, i (i)}
			<Table.Row>
				<Table.Cell class="uppercase">{item.key}</Table.Cell>
				<Table.Cell>{item.value}</Table.Cell>
			</Table.Row>
		{/each}
	</Table.Body>
</Table.Root>
