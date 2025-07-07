<script lang="ts">
	import { attachNetwork } from '$lib/api/vm/network';
	import SimpleSelect from '$lib/components/custom/SimpleSelect.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { SwitchList } from '$lib/types/network/switch';
	import type { VM } from '$lib/types/vm/vm';
	import { handleAPIError } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		switches: SwitchList;
		vm: VM | null;
	}

	let { open = $bindable(), switches, vm }: Props = $props();
	let usable = $derived.by(() => {
		return switches.standard?.filter((s) => {
			return !vm?.networks.some((n) => n.switchId === s.id);
		});
	});

	let options = {
		emulation: '',
		mac: '',
		switchId: ''
	};

	let properties = $state(options);

	async function addNetwork() {
		let error = '';

		if (!properties.switchId) {
			error = 'Switch is required';
		} else if (!properties.emulation) {
			error = 'Emulation is required';
		}

		if (error) {
			toast.error(error, {
				position: 'bottom-center'
			});
			return;
		}

		const response = await attachNetwork(
			vm?.vmId ?? 0,
			Number(properties.switchId),
			properties.emulation,
			properties.mac
		);
		if (response.error) {
			handleAPIError(response);
			toast.error('Error attaching VM to switch', {
				position: 'bottom-center'
			});
			return;
		} else {
			toast.success('VM attached to switch', {
				position: 'bottom-center'
			});
			open = false;
			properties = options;
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="w-md overflow-hidden p-5 lg:max-w-2xl">
		<Dialog.Header class="">
			<Dialog.Title class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Icon icon="mdi:network" class="h-5 w-5" />
					<span>New Network</span>
				</div>

				<div class="flex items-center gap-0.5">
					<Button
						size="sm"
						variant="link"
						title={'Reset'}
						class="h-4"
						onclick={() => {
							properties = options;
						}}
					>
						<Icon icon="radix-icons:reset" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">{'Reset'}</span>
					</Button>
					<Button
						size="sm"
						variant="link"
						class="h-4"
						title={'Close'}
						onclick={() => {
							properties = options;
							open = false;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">{'Close'}</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		<SimpleSelect
			label="Switch"
			placeholder="Select Switch"
			options={usable?.map((s) => ({
				value: s.id.toString(),
				label: s.name
			})) || []}
			bind:value={properties.switchId}
			onChange={(value) => (properties.switchId = value)}
		/>

		<div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
			<SimpleSelect
				label="Emulation"
				placeholder="Select Emulation"
				options={[
					{ value: 'virtio', label: 'VirtIO' },
					{ value: 'e1000', label: 'E1000' }
				]}
				bind:value={properties.emulation}
				onChange={(value) => (properties.emulation = value)}
			/>

			<CustomValueInput
				label="MAC Address"
				placeholder="58:9c:ff:ff:ff:a8"
				bind:value={properties.mac}
				classes="flex-1 space-y-1"
			/>
		</div>
		<Dialog.Footer class="flex justify-end">
			<div class="flex w-full items-center justify-end gap-2">
				<Button onclick={addNetwork} type="submit" size="sm">{'Save'}</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
