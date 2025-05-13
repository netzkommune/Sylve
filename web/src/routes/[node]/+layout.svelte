<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import TreeView from '$lib/components/custom/TreeView.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Resizable from '$lib/components/ui/resizable';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname } from '$lib/stores/basic';
	import { getTranslation } from '$lib/utils/i18n';
	import { triggers } from '$lib/utils/keyboard-shortcuts';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import { shortcut, type ShortcutTrigger } from '@svelte-put/shortcut';
	import CircleHelp from 'lucide-svelte/icons/circle-help';

	let openCategories: { [key: string]: boolean } = $state({});

	const toggleCategory = (label: string) => {
		openCategories[label] = !openCategories[label];
	};

	let node = $hostname;

	const nodeItems = $state([
		{
			label: 'summary',
			icon: 'basil:document-outline',
			href: `/${node}/summary`
		},
		{
			label: 'notes',
			icon: 'mdi:notes',
			href: `/${node}/notes`
		},
		{
			label: 'network',
			icon: 'mdi:network',
			children: [
				{
					label: 'interfaces',
					icon: 'carbon:network-interface',
					href: `/${node}/network/interfaces`
				},
				{
					label: 'switches',
					icon: 'clarity:network-switch-line',
					href: `/${node}/network/switches`
				}
			]
		},
		{
			label: 'storage',
			icon: 'mdi:storage',
			children: [
				{
					label: 'disks',
					icon: 'mdi:harddisk',
					href: `/${node}/storage/disks`
				},
				{
					label: 'zfs',
					icon: 'file-icons:openzfs',
					children: [
						{
							label: 'Dashboard',
							icon: 'material-symbols:dashboard',
							href: `/${node}/storage/zfs/dashboard`
						},
						{
							label: 'pools',
							icon: 'bi:hdd-stack-fill',
							href: `/${node}/storage/zfs/pools`
						},
						{
							label: 'datasets',
							icon: 'material-symbols:dataset',
							children: [
								{
									label: 'file_systems',
									icon: 'eos-icons:file-system',
									href: `/${node}/storage/zfs/datasets/fs`
								},
								{
									label: 'volumes',
									icon: 'carbon:volume-block-storage',
									href: `/${node}/storage/zfs/datasets/volumes`
								},
								{
									label: 'snapshots',
									icon: 'carbon:ibm-cloud-vpc-block-storage-snapshots',
									href: `/${node}/storage/zfs/datasets/snapshots`
								}
							]
						}
					]
				}
			]
		},
		{
			label: 'utilities',
			icon: 'mdi:tools',
			children: [
				{
					label: 'downloader',
					icon: 'material-symbols:download',
					href: `/${node}/utilities/downloader`
				}
			]
		}
	]);

	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();

	let mainTree = $derived.by(() => {
		const pathname = page.url.pathname;
		const shouldShowHundredItems = /1|local/.test(pathname);
		return shouldShowHundredItems ? [] : nodeItems;
	});

	$effect(() => {
		if (page.url.pathname === `/${$hostname}`) {
			goto(`/${node}/summary`);
		}
	});
</script>

<svelte:window
	use:shortcut={{
		trigger: triggers as ShortcutTrigger[]
	}}
/>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center justify-between border-b p-2">
		<p>{capitalizeFirstLetter(getTranslation('common.datacenter', 'Datacenter'))}</p>
		<Button size="sm" class="h-6 ">
			<CircleHelp class="mr-2 h-3 w-3" />
			Help
		</Button>
	</div>

	<Resizable.PaneGroup
		direction="horizontal"
		class="h-full w-full"
		id="main-pane-auto"
		autoSaveId="main-pane-auto-save"
	>
		<Resizable.Pane defaultSize={15}>
			<div class="h-full px-1.5">
				<div class="h-full overflow-y-auto">
					<nav aria-label="Difuse-sidebar" class="menu thin-scrollbar w-full">
						<ul>
							<ScrollArea orientation="both" class="h-full w-full">
								{#each mainTree as item}
									<TreeView {item} onToggle={toggleCategory} bind:this={openCategories} />
								{/each}
							</ScrollArea>
						</ul>
					</nav>
				</div>
			</div>
		</Resizable.Pane>
		<Resizable.Handle withHandle />
		<Resizable.Pane>
			<div class="h-full overflow-auto">
				{@render children?.()}
			</div>
		</Resizable.Pane>
	</Resizable.PaneGroup>
</div>
