<script lang="ts">
	import { default as TreeView } from '$lib/components/custom/TreeView.svelte';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname } from '$lib/stores/basic';
	import { getTranslation } from '$lib/utils/i18n';

	let openCategories: { [key: string]: boolean } = $state({});
	let node = $hostname;

	const toggleCategory = (label: string) => {
		openCategories[label] = !openCategories[label];
	};

	const tree = [
		{
			label: getTranslation('common.datacenter', 'Data Center'),
			icon: 'fa-solid:server',
			children: [
				{
					label: node,
					icon: 'mdi:dns',
					href: `/${node}`,
					children: [
						// {
						// 	label: '100 (Firewall)',
						// 	icon: 'tabler:prison',
						// 	href: `/${node}/100_firewall`
						// },
					]
				}
			]
		}
	];
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
