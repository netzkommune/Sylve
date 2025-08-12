<script lang="ts">
	import { page } from '$app/state';
	import { disinheritHostNetwork, getJails, inheritHostNetwork } from '$lib/api/jail/jail';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { Jail, JailState } from '$lib/types/jail/jail';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';

	interface Data {
		jails: Jail[];
		jailStates: JailState[];
		jail: Jail;
	}

	let { data }: { data: Data } = $props();
	const ctId = page.url.pathname.split('/')[3];
	const results = useQueries([
		{
			queryKey: ['jail-list'],
			queryFn: async () => {
				return await getJails();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.jails,
			onSuccess: (data: Jail[]) => {
				updateCache('jail-list', data);
			}
		}
	]);

	let jails = $derived($results[0].data || []);
	let jail = $derived.by(() => {
		if (jails.length > 0) {
			const found = jails.find((j) => j.ctId === parseInt(ctId));
			return found || data.jail;
		}

		return data.jail;
	});

	let inherited = $derived.by(() => {
		if (jail) {
			return jail.inheritIPv4 || jail.inheritIPv6;
		}

		return false;
	});

	let modals = $state({
		inherit: {
			open: false,
			ipv4: false,
			ipv6: false
		}
	});

	async function inherit() {
		if (!jail) return;
		if (!modals.inherit.ipv4 && !modals.inherit.ipv6) {
			toast.error('You must select at least one protocol to inherit', {
				position: 'bottom-center'
			});

			return;
		}

		const response = await inheritHostNetwork(jail.ctId, modals.inherit.ipv4, modals.inherit.ipv6);
		if (response.error) {
			handleAPIError(response);
			toast.error('Failed to inherit network', {
				position: 'bottom-center'
			});
		} else {
			toast.success('Host network inherited', {
				position: 'bottom-center'
			});
		}

		modals.inherit.open = false;
	}

	async function disinherit() {
		if (!jail) return;
		const response = await disinheritHostNetwork(jail.ctId);
		if (response.error) {
			handleAPIError(response);
			toast.error('Failed to disinherit network', {
				position: 'bottom-center'
			});
		} else {
			toast.success('Host network disinherited', {
				position: 'bottom-center'
			});
		}

		modals.inherit.open = false;
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Button
			onclick={() => {
				modals.inherit.open = true;
			}}
			size="sm"
			variant="outline"
			class="h-6.5"
		>
			<div class="flex items-center">
				{#if inherited}
					<Icon icon="mdi:close-network" class="mr-1 h-4 w-4" />
					<span>Disinherit Network</span>
				{:else}
					<Icon icon="mdi:plus-network" class="mr-1 h-4 w-4" />
					<span>Inherit Network</span>
				{/if}
			</div>
		</Button>
	</div>
</div>

<Dialog.Root bind:open={modals.inherit.open}>
	<Dialog.Content>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex items-center justify-between text-left">
				<div class="flex items-center">
					<Icon icon="mdi:network" class="mr-2 h-5 w-5" />
					{#if inherited}
						Disinherit Network
					{:else}
						Inherit Network
					{/if}
				</div>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Close'}
						onclick={() => {
							modals.inherit.open = false;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Close</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		{#if inherited}
			<span class="text-muted-foreground text-justify text-sm">
				This option will <b>disinherit the network configuration</b> from the host. Choose this if
				you want to <b>attach a custom network switch</b> to this jail or <b>disable networking</b>
				entirely.
				<b>Changes will take effect after restarting the jail.</b>
			</span>
		{:else}
			<span class="text-muted-foreground text-justify text-sm">
				This option will inherit the <b>network configuration</b> from the host. Choose this if you
				want the jail to <b>share the host's networking</b>. You can select which <b>protocols</b>
				to inherit below.
				<b>Changes will take effect after restarting the jail.</b>
			</span>

			<div>
				<div class="flex flex-row gap-2">
					<CustomCheckbox
						label="IPv4"
						bind:checked={modals.inherit.ipv4}
						classes="flex items-center gap-2"
					></CustomCheckbox>
					<CustomCheckbox
						label="IPv6"
						bind:checked={modals.inherit.ipv6}
						classes="flex items-center gap-2"
					></CustomCheckbox>
				</div>
			</div>
		{/if}

		<Dialog.Footer class="flex justify-end">
			<div class="flex w-full items-center justify-end gap-2">
				{#if !inherited}
					<Button onclick={inherit} type="submit" size="sm">Save</Button>
				{:else}
					<Button onclick={disinherit} type="submit" size="sm">Disinherit</Button>
				{/if}
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
