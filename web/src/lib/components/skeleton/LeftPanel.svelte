<script lang="ts">
	import { getVMs } from '$lib/api/vm/vm';
	import { default as TreeView } from '$lib/components/custom/TreeView.svelte';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname } from '$lib/stores/basic';
	import type { VM } from '$lib/types/vm/vm';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { useQueries } from '@sveltestack/svelte-query';

	let openCategories: { [key: string]: boolean } = $state({});
	let node = $hostname;

	const toggleCategory = (label: string) => {
		openCategories[label] = !openCategories[label];
	};

	const results = useQueries([
		{
			queryKey: ['vms-list'],
			queryFn: async () => {
				return await getVMs();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: [] as VM[],
			refetchOnMount: 'always'
		}
	]);

	const vms = $derived($results[0].data || []);
	const tree = $derived([
		{
			label: capitalizeFirstLetter(getTranslation('common.datacenter', 'Data Center')),
			icon: 'fa-solid:server',
			children: [
				{
					label: node,
					icon: 'mdi:dns',
					href: `/${node}`,
					children: vms.map((vm) => ({
						label: `${vm.name} (${vm.vmId})`,
						icon: 'material-symbols:monitor-outline',
						href: `/${node}/${vm.name}`
					}))
				}
			]
		}
	]);
</script>

<div class="h-full overflow-y-auto px-1.5 pt-1">
	<nav aria-label="Difuse-sidebar" class="menu thin-scrollbar w-full">
		<ul>
			<ScrollArea orientation="both" class="h-full w-full">
				{#each tree as item}
					<TreeView {item} onToggle={toggleCategory} bind:this={openCategories} />
				{/each}
			</ScrollArea>
		</ul>
	</nav>
</div>
