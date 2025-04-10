<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { getTranslation } from '$lib/utils/i18n';
	import Icon from '@iconify/svelte';
	import { _ } from 'svelte-i18n';
	import { slide } from 'svelte/transition';
	import SidebarElement from './TreeView.svelte';

	interface SidebarProps {
		label: string;
		icon: string;
		href?: string;
		children?: SidebarProps[];
	}

	interface Props {
		item: SidebarProps;
		onToggle: (label: string) => void;
	}

	let { item, onToggle }: Props = $props();

	let isOpen = $state(false);

	const toggle = (e: MouseEvent) => {
		e.preventDefault();

		if (item.children) {
			isOpen = !isOpen;
			onToggle(item.label);
		}

		if (item.href) {
			goto(item.href, { replaceState: false, noScroll: false });
		}
	};

	const sidebarActive = 'rounded-md bg-primary/10 dark:bg-muted font-inter font-medium';

	function isItemActive(menuItem: SidebarProps, currentUrl: string): boolean {
		if (menuItem.href && currentUrl.startsWith(menuItem.href)) {
			return true;
		}
		if (menuItem.children) {
			return menuItem.children.some((child) => isItemActive(child, currentUrl));
		}
		return false;
	}

	let activeUrl = $derived($page.url.pathname);
	let isActive = $derived(isItemActive(item, activeUrl));
	let lastActiveUrl = $derived.by(() => {
		const segments = activeUrl.split('/');
		return segments[segments.length - 1];
	});

	function isItemOpen(menuItem: SidebarProps, currentUrl: string): boolean {
		if (menuItem.href && currentUrl.startsWith(menuItem.href)) {
			return true;
		}
		if (menuItem.children) {
			return menuItem.children.some((child) => isItemOpen(child, currentUrl));
		}
		return false;
	}

	$effect(() => {
		isOpen = isItemOpen(item, activeUrl);
	});
</script>

<li class={`w-full`}>
	<a
		class={`my-0.5 flex w-full items-center justify-between px-1.5 py-0.5 ${isActive ? sidebarActive : 'rounded-md hover:bg-primary/10 dark:hover:bg-muted'}${lastActiveUrl === item.label ? '!text-primary' : ' '}`}
		href={item.href}
		onclick={toggle}
	>
		<div class="flex items-center space-x-1 text-sm">
			<Icon icon={item.icon} width="18" />
			<p class="font-inter cursor-pointer whitespace-nowrap">
				{getTranslation(`node.${item.label}`, item.label)}
			</p>
		</div>
		{#if item.children}
			<Icon
				icon={isOpen ? 'teenyicons:down-solid' : 'teenyicons:right-solid'}
				class="h-3.5 w-3.5"
			/>
		{/if}
	</a>
</li>

{#if isOpen && item.children}
	<ul class="pl-5" transition:slide={{ duration: 200, easing: (t) => t }} style="overflow: hidden;">
		{#each item.children as child}
			<SidebarElement item={child} {onToggle} />
		{/each}
	</ul>
{/if}
