<script lang="ts">
	import { createCluster } from '$lib/api/cluster/cluster';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { handleAPIError } from '$lib/utils/http';
	import { isValidIPv4, isValidIPv6, isValidPortNumber } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { onMount } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		reload: boolean;
	}

	let { open = $bindable(), reload = $bindable() }: Props = $props();
	let options = {
		ip: '',
		port: 8182
	};

	let properties = $state(options);
	let loading = $state(false);

	$effect(() => {
		if (open) {
			if (window && window.location.hostname) {
				if (isValidIPv4(window.location.hostname) || isValidIPv6(window.location.hostname)) {
					properties.ip = window.location.hostname;
				}
			}
		}
	});

	async function create() {
		let error = '';

		if (!isValidIPv4(properties.ip) && !isValidIPv6(properties.ip)) {
			error = 'Invalid IP address';
		} else if (!isValidPortNumber(properties.port)) {
			error = 'Invalid port number';
		}

		if (error) {
			toast.error(error, {
				position: 'bottom-center'
			});

			return;
		}

		loading = true;

		const response = await createCluster(properties.ip, properties.port);
		reload = true;
		if (response.error) {
			handleAPIError(response);
			toast.error('Failed to create cluster', {
				position: 'bottom-center'
			});
		} else {
			toast.success('Cluster created', {
				position: 'bottom-center'
			});
			open = false;
			properties = options;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex  justify-between gap-1 text-left">
				<div class="flex items-center gap-2">
					<Icon icon="oui:ml-create-population-job" class="h-6 w-6" />
					<span>Create Cluster</span>
				</div>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Reset'}
						onclick={() => {
							properties = options;
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Reset</span>
					</Button>

					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Close'}
						onclick={() => {
							open = false;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Close</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		<CustomValueInput
			bind:value={properties.ip}
			placeholder="Node IP"
			classes="flex-1 space-y-1.5"
		/>

		<CustomValueInput
			bind:value={properties.port}
			placeholder="Node Port"
			classes="flex-1 space-y-1.5"
			type="number"
		/>

		<Dialog.Footer class="flex justify-end">
			<div class="flex w-full items-center justify-end gap-2">
				<Button onclick={create} type="submit" size="sm" disabled={loading}>
					{#if loading}
						<Icon icon="mdi:loading" class="h-4 w-4 animate-spin" />
					{:else}
						Create
					{/if}
				</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
