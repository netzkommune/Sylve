<script lang="ts">
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';

	interface Props {
		name: string;
		id: number;
		description: string;
	}

	let { name = $bindable(), id = $bindable(), description = $bindable() }: Props = $props();
</script>

<div class="flex flex-col gap-4 p-4">
	<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
		<CustomValueInput
			label="VM Name"
			placeholder="my-virtual-machine"
			bind:value={name}
			classes="flex-1 space-y-1"
		/>

		<CustomValueInput
			label="VM ID"
			placeholder="100"
			type="number"
			bind:value={id}
			classes="flex-1 space-y-1"
		/>
	</div>

	<CustomValueInput
		label="Description"
		placeholder="Optional description for this virtual machine"
		type="textarea"
		textAreaCLasses="min-h-40"
		bind:value={description}
		classes="flex-1 space-y-1"
	/>
</div>
