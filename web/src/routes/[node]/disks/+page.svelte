<script lang="ts">
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { type Disk } from '$lib/types/disk/disk';

	interface Data {
		disks: Disk[];
	}

	let { data }: { data: Data } = $props();

	console.log();
</script>

<div class="flex h-full flex-col overflow-hidden">
	<div class="inline-flex w-full gap-2 border-b px-3 py-2">
		<Button size="sm" class="h-8  bg-neutral-600 text-white hover:bg-neutral-700">Reload</Button>
		<Button size="sm" class="h-8  bg-neutral-600 text-white hover:bg-neutral-700">
			Show S.M.A.R.T values
		</Button>
		<Button size="sm" class="h-8  bg-neutral-600 text-white hover:bg-neutral-700">
			Initialize Disk with GPT
		</Button>
		<Button size="sm" class="h-8  bg-neutral-600 text-white hover:bg-neutral-700">Wipe Disk</Button>
	</div>

	<div class="flex flex-1 flex-col overflow-hidden">
		<div class="flex-1 overflow-hidden">
			<TreeTable data={data.disks} />
		</div>
	</div>
</div>
