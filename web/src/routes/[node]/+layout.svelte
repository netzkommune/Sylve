<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import TreeView from '$lib/components/custom/TreeView.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Resizable from '$lib/components/ui/resizable';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname } from '$lib/stores/basic';
	import { triggers } from '$lib/utils/keyboard-shortcuts';
	import { shortcut, type ShortcutTrigger } from '@svelte-put/shortcut';
	import CircleHelp from 'lucide-svelte/icons/circle-help';

	let openCategories: { [key: string]: boolean } = $state({});

	const toggleCategory = (label: string) => {
		openCategories[label] = !openCategories[label];
	};

	let node = $hostname;

	interface NodeItem {
		label: string;
		icon: string;
		href?: string;
		children?: NodeItem[];
	}

	let nodeItems: NodeItem[] = $derived.by(() => {
		if (page.url.pathname.startsWith(`/${$hostname}/vm`)) {
			const vmName = page.url.pathname.split('/')[3];
			return [
				{
					label: 'Summary',
					icon: 'basil:document-outline',
					href: `/${node}/vm/${vmName}/summary`
				},
				{
					label: 'Console',
					icon: 'mdi:monitor',
					href: `/${node}/vm/${vmName}/console`
				},
				{
					label: 'Storage',
					icon: 'mdi:storage',
					href: `/${node}/vm/${vmName}/storage`
				},
				{
					label: 'Hardware',
					icon: 'ix:hardware-cabinet',
					href: `/${node}/vm/${vmName}/hardware`
				},
				{
					label: 'Network',
					icon: 'mdi:network',
					href: `/${node}/vm/${vmName}/network`
				}
			];
		} else {
			return [
				{
					label: 'Summary',
					icon: 'basil:document-outline',
					href: `/${node}/summary`
				},
				{
					label: 'Notes',
					icon: 'mdi:notes',
					href: `/${node}/notes`
				},
				{
					label: 'Network',
					icon: 'mdi:network',
					children: [
						{
							label: 'Interfaces',
							icon: 'carbon:network-interface',
							href: `/${node}/network/interfaces`
						},
						{
							label: 'Switches',
							icon: 'clarity:network-switch-line',
							href: `/${node}/network/switches`
						}
					]
				},
				{
					label: 'Storage',
					icon: 'mdi:storage',
					children: [
						{
							label: 'Explorer',
							icon: 'bxs:folder-open',
							href: `/${node}/storage/explorer`
						},
						{
							label: 'Disks',
							icon: 'mdi:harddisk',
							href: `/${node}/storage/disks`
						},
						{
							label: 'ZFS',
							icon: 'file-icons:openzfs',
							children: [
								{
									label: 'Dashboard',
									icon: 'mdi:monitor-dashboard',
									href: `/${node}/storage/zfs/dashboard`
								},
								{
									label: 'Pools',
									icon: 'bi:hdd-stack-fill',
									href: `/${node}/storage/zfs/pools`
								},
								{
									label: 'Datasets',
									icon: 'material-symbols:dataset',
									children: [
										{
											label: 'File Systems',
											icon: 'eos-icons:file-system',
											href: `/${node}/storage/zfs/datasets/fs`
										},
										{
											label: 'Volumes',
											icon: 'carbon:volume-block-storage',
											href: `/${node}/storage/zfs/datasets/volumes`
										},
										{
											label: 'Snapshots',
											icon: 'carbon:ibm-cloud-vpc-block-storage-snapshots',
											href: `/${node}/storage/zfs/datasets/snapshots`
										}
									]
								}
							]
						},
						{
							label: 'Samba',
							icon: 'material-symbols:smb-share',
							children: [
								{
									label: 'Shares',
									icon: 'mdi:folder-network',
									href: `/${node}/storage/samba/shares`
								},
								{
									label: 'Settings',
									icon: 'mdi:folder-settings-variant',
									href: `/${node}/storage/samba/settings`
								}
							]
						}
					]
				},
				{
					label: 'Utilities',
					icon: 'mdi:tools',
					children: [
						{
							label: 'Downloader',
							icon: 'material-symbols:download',
							href: `/${node}/utilities/downloader`
						}
					]
				},
				{
					label: 'Settings',
					icon: 'material-symbols:settings',
					children: [
						{
							label: 'PCI Passthrough',
							icon: 'eos-icons:hardware-circuit',
							href: `/${node}/settings/device-passthrough`
						},
						{
							label: 'Authentication',
							icon: 'mdi:shield-key',
							children: [
								{
									label: 'Users',
									icon: 'mdi:account',
									href: `/${node}/settings/authentication/users`
								},
								{
									label: 'Groups',
									icon: 'mdi:account-group',
									href: `/${node}/settings/authentication/groups`
								}
							]
						}
					]
				}
			];
		}
	});

	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();

	$effect(() => {
		if (page.url.pathname === `/${$hostname}`) {
			goto(`/${node}/summary`);
		} else if (page.url.pathname.startsWith(`/${$hostname}/vm`)) {
			const vmId = page.url.pathname.split('/')[3];
			if (page.url.pathname === `/${node}/vm/${vmId}`) {
				goto(`/${node}/vm/${vmId}/summary`, { replaceState: true });
			}
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
		<span>Data Center</span>
		<Button
			size="sm"
			class="h-6"
			onclick={() => (window.location.href = 'https://github.com/AlchemillaHQ/Sylve')}
		>
			<div class="flex items-center">
				<CircleHelp class="mr-2 h-5 w-5" />
				<span>Help</span>
			</div>
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
								{#each nodeItems as item}
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
