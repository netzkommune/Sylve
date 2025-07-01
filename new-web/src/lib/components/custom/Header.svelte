<script>
	import { logOut } from '$lib/api/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { openTerminal, terminalStore } from '$lib/stores/terminal.svelte';
	import Icon from '@iconify/svelte';
	import { mode, toggleMode } from 'mode-watcher';
	import CreateVM from './CreateVM/CreateVM.svelte';

	let menuData = $state({
		createVM: {
			open: false
		},
		menuItems: [
			{ icon: 'ic:baseline-settings', label: 'My Settings', shortcut: '⌘S' },
			{ icon: 'solar:key-bold', label: 'Password', shortcut: '⇧⌘P' },
			{ icon: 'ic:round-lock', label: 'TFA', shortcut: '⌘K' },
			{ icon: 'mdi:palette', label: 'Color Theme', shortcut: '⌘⇧T' },
			{ icon: 'meteor-icons:language', label: 'Language', shortcut: '⌘K' }
		]
	});
</script>

<header class="sticky top-0 flex h-[5vh] items-center gap-4 border-x border-b px-2 md:h-[4vh]">
	<nav
		class="hidden flex-col gap-2 text-lg font-medium md:items-center md:gap-2 md:text-sm lg:flex lg:flex-row lg:gap-4"
	>
		<div class="flex items-center space-x-2">
			{#if mode.current === 'dark'}
				<img src="/logo/white.svg" alt="Sylve Logo" class="h-6 w-auto max-w-[100px]" />
			{:else}
				<img src="/logo/black.svg" alt="Sylve Logo" class="h-6 w-auto max-w-[100px]" />
			{/if}
			<p class="font-normal tracking-[.45em]">SYLVE</p>
		</div>
	</nav>
	<Sheet.Root>
		<Sheet.Trigger>
			<Button variant="outline" size="icon" class="h-7 shrink-0 lg:hidden">
				<Icon icon="material-symbols:menu-rounded" class="h-5 w-5" />
				<span class="sr-only">Toggle navigation menu</span>
			</Button>
		</Sheet.Trigger>
		<Sheet.Content side="left">
			<!-- mobile view -->
			<nav class="flex flex-col text-lg font-medium">
				<div class="mt-4 flex items-center space-x-2">
					<img src="/logo/white.svg" alt="Sylve Logo" class="h-6 w-auto max-w-[100px]" />
					<p class="font-normal tracking-[.45em]">SYLVE</p>
				</div>
				<p class="mt-4 whitespace-nowrap">Virtual Environment 0.0.1</p>
			</nav>
		</Sheet.Content>
	</Sheet.Root>
	<div class="flex w-full items-center justify-end gap-2 md:ml-auto">
		<!-- desktop view -->
		<div class="mr-2 hidden items-center gap-4 lg:inline-flex">
			<Button
				size="icon"
				variant="link"
				class="relative z-[9999] flex  w-auto items-center justify-center "
				onclick={() => openTerminal()}
			>
				<Icon icon="garden:terminal-cli-stroke-16" class="h-6 w-6" />
				{#if $terminalStore.tabs.length > 0}
					<span
						class="absolute -right-1 -top-1 flex h-4 min-w-[8px] items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white"
					>
						{$terminalStore.tabs.length}
					</span>
				{/if}
			</Button>

			<Button
				class="h-6"
				size="sm"
				onclick={() => (menuData.createVM.open = !menuData.createVM.open)}
			>
				<Icon icon="material-symbols:monitor-outline-rounded" class="mr-1.5 h-5 w-5" />
				Create VM
			</Button>

			{#if menuData.createVM.open}
				<CreateVM bind:open={menuData.createVM.open} />
			{/if}
		</div>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				<Button variant="outline" size="sm" class="h-6.5"
					><Icon icon="mdi:user" class="h-4 w-4" /> Root <Icon
						icon="famicons:chevron-down"
						class="h-4 w-4"
					/>
					<span class="sr-only">Toggle user menu</span></Button
				>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="w-56">
				<DropdownMenu.Group>
					{#each menuData.menuItems as { icon, label, shortcut }}
						<DropdownMenu.Item
							class="cursor-pointer"
							onclick={() => label === 'Color Theme' && toggleMode()}
						>
							<Icon {icon} class="mr-2 h-4 w-4" />
							<span>{label}</span>
							{#if shortcut}
								<DropdownMenu.Shortcut>{shortcut}</DropdownMenu.Shortcut>
							{/if}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Group>

				<DropdownMenu.Separator />
				<DropdownMenu.Item class="cursor-pointer" onclick={() => logOut()}>
					<Icon icon="ic:twotone-logout" class="mr-2 h-4 w-4" />
					<span>Log out</span>
					<DropdownMenu.Shortcut>⌘⇧Q</DropdownMenu.Shortcut>
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>
</header>
