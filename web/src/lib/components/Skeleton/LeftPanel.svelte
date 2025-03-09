<script lang="ts">
	import { page } from '$app/state';
	import { default as TreeView } from '$lib/components/custom/TreeView.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { hostname } from '$lib/stores/basic';
	import Settings from 'lucide-svelte/icons/settings';

	let openCategories: { [key: string]: boolean } = $state({});

	const toggleCategory = (label: string) => {
		openCategories[label] = !openCategories[label];
	};

	let node = $hostname;
	const options = [
		{ value: 'server', label: 'Server view' },
		{ value: 'folder', label: 'Folder View' },
		{ value: 'pool', label: 'Pool View' }
	];

	const tree = [
		{
			label: 'datacenter',
			icon: 'fa-solid:server',
			children: [
				{
					label: node,
					icon: 'mdi:dns',
					href: `/${node}/summary`,
					children: [
						{
							label: '100 (Firewall)',
							icon: 'tabler:prison',
							href: `/${node}/100_firewall`
						},
						{
							label: '101 (Windows)',
							icon: 'mi:computer',
							href: `/${node}/100_firewall`
						},
						{
							label: '102 (test-store)',
							icon: 'mdi:database',
							href: `/${node}/106_tg_wallet`
						}
					]
				}
			]
		}
	];
</script>

<div class="h-full overflow-y-auto">
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
