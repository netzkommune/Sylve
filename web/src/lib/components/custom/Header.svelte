<script>
	import { logOut } from '$lib/api/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import Icon from '@iconify/svelte';
	import { mode, toggleMode } from 'mode-watcher';
	import CreateDialog from './CreateDialog.svelte';

	const vmTabs = [
		{ value: 'vm_general', label: 'General' },
		{ value: 'vm_os', label: 'OS' },
		{ value: 'vm_system', label: 'System' },
		{ value: 'vm_disks', label: 'Disks' },
		{ value: 'vm_cpu', label: 'CPU' },
		{ value: 'vm_memory', label: 'Memory' },
		{ value: 'vm_network', label: 'Network' },
		{ value: 'vm_confirm', label: 'Confirm' }
	];

	const ctTabs = [
		{ value: 'ct_general', label: 'General' },
		{ value: 'ct_template', label: 'Template' },
		{ value: 'ct_disks', label: 'Disks' },
		{ value: 'ct_cpu', label: 'CPU' },
		{ value: 'ct_memory', label: 'Memory' },
		{ value: 'ct_network', label: 'Network' },
		{ value: 'ct_dns', label: 'DNS' },
		{ value: 'ct_confirm', label: 'Confirm' }
	];

	const menuItems = [
		{ icon: 'ic:baseline-settings', label: 'My Settings', shortcut: '⌘S' },
		{ icon: 'solar:key-bold', label: 'Password', shortcut: '⇧⌘P' },
		{ icon: 'ic:round-lock', label: 'TFA', shortcut: '⌘K' },
		{ icon: 'mdi:palette', label: 'Color Theme', shortcut: '⌘⇧T' },
		{ icon: 'meteor-icons:language', label: 'Language', shortcut: '⌘K' }
	];

	$effect(() => {
		console.log('mode', $mode);
	});
</script>

<header class="bg-background sticky top-0 flex h-[5vh] items-center gap-4 border-b px-2 md:h-[4vh]">
	<nav
		class="hidden flex-col gap-2 text-lg font-medium md:items-center md:gap-2 md:text-sm lg:flex lg:flex-row lg:gap-4"
	>
		<div class="flex items-center space-x-2">
			{#if $mode === 'dark'}
				<img src="/logo/white.svg" alt="Sylve Logo" class="h-6 w-auto max-w-[100px]" />
			{:else}
				<img src="/logo/black.svg" alt="Sylve Logo" class="h-6 w-auto max-w-[100px]" />
			{/if}
			<p class="font-normal tracking-[.45em]">SYLVE</p>
		</div>
		<form class="ml-auto flex-1 sm:flex-initial">
			<div class="relative">
				<Icon
					icon="ic:sharp-search"
					class="text-muted-foreground absolute left-2.5 top-1.5 h-4 w-4"
				/>
				<Input
					type="search"
					placeholder="Search products..."
					class="h-7 pl-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
				/>
			</div>
		</form>
	</nav>
	<Sheet.Root>
		<Sheet.Trigger asChild let:builder>
			<Button variant="outline" size="icon" class="h-7 shrink-0 lg:hidden" builders={[builder]}>
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
				<p class="mt-4 whitespace-nowrap">Virtual Environment 8.2.7</p>
				<Button size="sm" class="mt-4 h-8  bg-neutral-600 text-white hover:bg-neutral-700">
					<Icon icon="material-symbols-light:mail-outline-sharp" class="mr-2 h-4 w-4" />
					Documentation
				</Button>
				<CreateDialog
					title="Create: Virtual Machine"
					tabs={vmTabs}
					icon="material-symbols:monitor-outline-rounded"
					buttonText="Create VM"
					buttonClass="h-8 mt-4"
				/>
				<CreateDialog
					title="Create: LXC Container"
					tabs={ctTabs}
					icon="ph:cube-fill"
					buttonText="Create Jail"
					buttonClass="h-8 mt-4"
				/>
			</nav>
		</Sheet.Content>
	</Sheet.Root>
	<div class="flex w-full items-center justify-end gap-2 md:ml-auto">
		<div class="relative lg:hidden">
			<Icon
				icon="ic:sharp-search"
				class="text-muted-foreground absolute left-2.5 top-1.5 h-4 w-4"
			/>
			<Input
				type="search"
				placeholder="Search products..."
				class="h-7 pl-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
			/>
		</div>

		<!-- desktop view -->
		<div class="hidden items-center gap-2 lg:inline-flex">
			<Button size="sm" class="h-6 bg-neutral-600 text-white hover:bg-neutral-700 ">
				<Icon icon="material-symbols-light:mail-outline-sharp" class="mr-2 h-5 w-5" />
				Documentation
			</Button>

			<CreateDialog
				title="Create: Virtual Machine"
				tabs={vmTabs}
				icon="material-symbols:monitor-outline-rounded"
				buttonText="Create VM"
				buttonClass="h-6"
			/>

			<CreateDialog
				title="Create: LXC Container"
				tabs={ctTabs}
				icon="tabler:prison"
				buttonText="Create Jail"
				buttonClass="h-6"
			/>
		</div>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger asChild let:builder>
				<Button
					builders={[builder]}
					variant="outline"
					class="flex h-6 items-center gap-1 rounded-md border border-neutral-500"
					><Icon icon="mdi:user" class="h-4 w-4" /> Root <Icon
						icon="famicons:chevron-down"
						class="h-4 w-4"
					/>
					<span class="sr-only">Toggle user menu</span></Button
				>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="w-56">
				<DropdownMenu.Group>
					{#each menuItems as { icon, label, shortcut }}
						<DropdownMenu.Item
							class="cursor-pointer"
							on:click={() => label === 'Color Theme' && toggleMode()}
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
