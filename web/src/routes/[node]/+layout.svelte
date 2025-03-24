<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import TreeView from '$lib/components/custom/TreeView.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Resizable from '$lib/components/ui/resizable';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname, paneSizes } from '$lib/stores/basic';
	import { triggers } from '$lib/utils/keyboard-shortcuts';
	import { shortcut, type ShortcutTrigger } from '@svelte-put/shortcut';
	import CircleHelp from 'lucide-svelte/icons/circle-help';
	import { onMount } from 'svelte';

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
			icon: 'arcticons:notes',
			href: `/${node}/notes`
		},
		{
			label: 'Storage',
			icon: 'mdi:storage',
			children: [
				{
					label: 'Disks',
					icon: 'mdi:harddisk',
					href: `/${node}/storage/disks`
				},
				{
					label: 'ZFS',
					icon: 'file-icons:openzfs',
					href: `/${node}/storage/zfs`
				}
			]
		},
		{
			label: 'notifications',
			icon: 'mdi:notifications'
		}
	]);

	const hundredItems = $state([
		{
			label: 'Search',
			icon: 'mdi:magnify'
		},
		{
			label: 'Summary',
			icon: 'basil:document-outline'
		},
		{
			label: 'Resources',
			icon: 'arcticons:notes'
		},
		{
			label: 'Network',
			icon: 'fa-solid:server'
		},
		{
			label: 'DNS',
			icon: 'tabler:fingerprint'
		},
		{
			label: 'Options',
			icon: 'material-symbols-light:settings-outline-rounded'
		},
		{
			label: 'Task History',
			icon: 'mdi:database'
		},
		{
			label: 'Backup',
			icon: 'tdesign:system-storage'
		},
		{
			label: 'Replication',
			icon: 'icomoon-free:loop'
		},
		{
			label: 'Snapshots',
			icon: 'icomoon-free:loop'
		},
		{
			label: 'Permissions',
			icon: 'bi:unlock-fill'
		},
		{
			label: 'Firewall',
			icon: 'mdi:home',
			children: [
				{
					label: 'Options',
					icon: 'mdi:phone',
					href: ''
				},
				{
					label: 'Alias',
					icon: 'mdi:phone',
					href: ''
				},
				{
					label: 'IPSet',
					icon: 'mdi:phone',
					href: ''
				},
				{
					label: 'Log',
					icon: 'mdi:phone',
					href: ''
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
		return shouldShowHundredItems ? hundredItems : nodeItems;
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
	<div class="flex h-10 w-full items-center justify-between border p-2">
		<p>Datacenter</p>
		<Button size="sm" class="h-6 bg-neutral-500 text-white hover:bg-neutral-400">
			<CircleHelp class="mr-2 h-3 w-3" />
			Help
		</Button>
	</div>

	<Resizable.PaneGroup direction="horizontal" class="h-full w-full">
		<Resizable.Pane
			defaultSize={$paneSizes.middle}
			minSize={5}
			onResize={(e) => ($paneSizes.middle = e)}
		>
			<div class="h-full px-1">
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
		<Resizable.Pane defaultSize={$paneSizes.right} onResize={(e) => ($paneSizes.right = e)}>
			<div class="h-full">
				{@render children?.()}
			</div>
		</Resizable.Pane>
	</Resizable.PaneGroup>
</div>
