<script lang="ts">
	import { createS3Storage } from '$lib/api/cluster/storage';
	import { Button } from '$lib/components/ui/button/index.js';
	import CustomComboBox from '$lib/components/ui/custom-input/combobox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import type { ClusterStorages } from '$lib/types/cluster/storage';
	import { handleAPIError } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		reload: boolean;
		storages: ClusterStorages;
	}

	let { open = $bindable(), reload = $bindable(), storages }: Props = $props();
	let options = {
		s3: {
			name: '',
			endpoint: '',
			region: '',
			bucket: '',
			accessKey: '',
			secretKey: ''
		}
	};

	let properties = $state(options);
	let loading = $state(false);

	let type = $state({
		combobox: {
			open: false,
			value: '' as '' | 's3'
		}
	});

	async function create() {
		if (type.combobox.value === 's3') {
			const data = properties.s3;
			if (
				!data.name ||
				!data.endpoint ||
				!data.region ||
				!data.bucket ||
				!data.accessKey ||
				!data.secretKey
			) {
				toast.error('Missing required fields', {
					position: 'bottom-center'
				});
				return;
			}

			loading = true;

			const response = await createS3Storage(
				data.name,
				data.endpoint,
				data.region,
				data.bucket,
				data.accessKey,
				data.secretKey
			);

			loading = false;
			reload = true;

			if (response.error) {
				handleAPIError(response);
				toast.error('Failed to create S3 storage', {
					position: 'bottom-center'
				});
				return;
			}

			toast.success('S3 storage created', {
				position: 'bottom-center'
			});

			open = false;
		}
	}
</script>

{#snippet s3Input(name: string, label: string)}
	<CustomValueInput
		bind:value={properties.s3[name as keyof typeof properties.s3]}
		placeholder={label}
		classes="flex-1 space-y-1.5"
	/>
{/snippet}

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header class="p-0">
			<Dialog.Title class="flex justify-between gap-1 text-left">
				<div class="flex items-center gap-2">
					<Icon icon="mdi:storage" class="h-6 w-6" />
					<span>Create Storage</span>
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
							properties = options;
						}}
					>
						<Icon icon="material-symbols:close-rounded" class="pointer-events-none h-4 w-4" />
						<span class="sr-only">Close</span>
					</Button>
				</div>
			</Dialog.Title>
		</Dialog.Header>

		<CustomComboBox
			bind:open={type.combobox.open}
			label="Type"
			bind:value={type.combobox.value}
			data={[{ value: 's3', label: 'S3' }]}
			classes="flex-1 space-y-1"
			placeholder="Select Type"
			triggerWidth="w-full"
			width="w-full lg:w-[75%]"
		></CustomComboBox>

		{#if type.combobox.value === 's3'}
			<div class="mt-0 grid grid-cols-2 gap-4">
				{@render s3Input('name', 'Name')}
				{@render s3Input('endpoint', 'Endpoint')}
				{@render s3Input('region', 'Region')}
				{@render s3Input('bucket', 'Bucket')}
				{@render s3Input('accessKey', 'Access Key')}
				{@render s3Input('secretKey', 'Secret Key')}
			</div>
		{/if}

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
