<script lang="ts">
	import { destroyDisk, destroyPartition, initializeGPT, listDisks } from '$lib/api/disk/disk';
	import { getPools } from '$lib/api/zfs/pool';
	import AlertDialog from '$lib/components/custom/AlertDialog.svelte';
	import KvTableModal from '$lib/components/custom/KVTableModal.svelte';
	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import CreatePartition from '$lib/components/disk/CreatePartition.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { Row } from '$lib/types/components/tree-table';
	import { type Disk, type Partition } from '$lib/types/disk/disk';
	import type { Zpool } from '$lib/types/zfs/pool';
	import { diskSpaceAvailable, generateTableData, parseSMART } from '$lib/utils/disk';
	import { handleAPIError, updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';

	interface Data {
		disks: Disk[];
		pools: Zpool[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['diskList'],
			queryFn: async () => {
				return await listDisks();
			},
			refetchInterval: 2000,
			keepPreviousData: true,
			initialData: data.disks,
			onSuccess: (data: Disk[]) => {
				updateCache('disks', data);
			}
		},
		{
			queryKey: ['poolList'],
			queryFn: async () => {
				return await getPools();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.pools,
			onSuccess: (data: Zpool[]) => {
				updateCache('pools', data);
			}
		}
	]);

	let disks = $derived($results[0].data as Disk[]);
	let activeRows: Row[] | null = $state(null);
	let activeRow: Row | null = $derived(activeRows ? (activeRows[0] as Row) : ({} as Row));
	let { rows, columns } = $derived(generateTableData(disks));

	let wipeModal = $state({
		open: false,
		title: ''
	});

	let partitionModal: {
		open: boolean;
		disk: Disk | null;
	} = $state({
		open: false,
		disk: null
	});

	let smartModal = $state({
		open: false,
		title: '',
		KV: {},
		type: ''
	});

	let activeDisk: Disk | null = $derived.by(() => {
		if (activeRow !== null) {
			return disks.find((disk) => disk.device === activeRow?.device) || null;
		}
		return null;
	});

	let activePartition: Partition | null = $derived.by(() => {
		if (activeRow !== null) {
			const partition = disks.filter((disk) => {
				return disk.partitions.some((part) => part.name === activeRow?.device);
			});

			if (partition.length > 0) {
				return partition[0].partitions.find((part) => part.name === activeRow?.device) || null;
			} else {
				return null;
			}
		}
		return null;
	});

	async function diskAction(action: string) {
		if (action === 'smart') {
			if (activeDisk) {
				smartModal.open = false;
				smartModal.title = `${getTranslation('disk.smart', 'S.M.A.R.T')} Values (${activeDisk.device})`;
				if (activeDisk.type === 'NVMe') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					smartModal.open = true;
					smartModal.type = 'kv';
				} else if (activeDisk.type === 'HDD' || activeDisk.type === 'SSD') {
					smartModal.KV = parseSMART($state.snapshot(activeDisk));
					smartModal.open = true;
					smartModal.type = 'array';
				}
			}
		}

		if (action === 'wipe') {
			wipeModal.open = true;
			if (activePartition !== null) {
				wipeModal.title = `${getTranslation('common.this_action_cannot_be_undone', 'This action cannot be undone')}. ${getTranslation(
					'common.this_will_permanently',
					'This will permanently'
				)} <b>${getTranslation('common.delete', 'delete')}</b> ${getTranslation('disk.partition', 'disk')} <b>${activePartition.name}</b>.`;
			} else if (activeDisk !== null) {
				wipeModal.title = `${getTranslation('common.this_action_cannot_be_undone', 'This action cannot be undone')}. ${getTranslation(
					'common.this_will_permanently',
					'This will permanently'
				)} <b>${getTranslation('disk.wipe', 'wipe')}</b> ${getTranslation('disk.disk', 'disk')} <b>${activeDisk.device}</b>.`;
			}
		}

		if (action === 'gpt') {
			if (activeDisk) {
				const response = await initializeGPT(activeDisk.device);
				if (response.status === 'success') {
					toast.success(
						`${capitalizeFirstLetter(getTranslation('disk.disk', 'Disk'))} ${activeDisk.device} ${getTranslation(
							'disk.gpt_initialized',
							'initialized with GPT'
						)}`,
						{ position: 'bottom-center' }
					);
				} else {
					handleAPIError(response);
				}
			}
		}

		if (action === 'partition') {
			partitionModal.open = true;
			partitionModal.disk = activeDisk;
		}
	}

	let buttonAbilities = $state({
		smart: {
			ability: false
		},
		gpt: {
			ability: false
		},
		wipe: {
			ability: false
		},
		createPartition: {
			ability: false
		}
	});

	$effect(() => {
		if (activeDisk) {
			untrack(() => {
				buttonAbilities.smart.ability = activeDisk.smartData !== null;
				buttonAbilities.gpt.ability = !activeDisk.gpt;

				if (activeDisk.usage === 'ZFS') {
					buttonAbilities.gpt.ability = false;
					buttonAbilities.wipe.ability = false;
				}

				if (activeDisk.usage === 'Unused' || activeDisk.usage === 'Partitions') {
					if (activeDisk.gpt) {
						buttonAbilities.wipe.ability = true;
					} else {
						buttonAbilities.wipe.ability = false;
					}
				}

				buttonAbilities.createPartition.ability =
					activeDisk.gpt &&
					diskSpaceAvailable(activeDisk, 128 * 1024 * 1024) &&
					activeDisk.usage !== 'ZFS';
			});
		} else if (activePartition) {
			untrack(() => {
				buttonAbilities.gpt.ability = false;
				buttonAbilities.wipe.ability = true;
				buttonAbilities.createPartition.ability = false;
				buttonAbilities.smart.ability = false;
			});
		} else {
			untrack(() => {
				buttonAbilities.gpt.ability = false;
				buttonAbilities.wipe.ability = false;
				buttonAbilities.createPartition.ability = false;
				buttonAbilities.smart.ability = false;
			});
		}
	});

	let query = $state('');
</script>

{#snippet button(type: string)}
	{#if type == 'smart' && buttonAbilities.smart.ability}
		<Button
			onclick={() => diskAction('smart')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="icon-park-outline:hdd" class="mr-1 h-4 w-4" />
			{getTranslation('disk.smart_values', 'S.M.A.R.T Values')}
		</Button>
	{/if}

	{#if type == 'gpt' && buttonAbilities.gpt.ability}
		<Button
			onclick={() => diskAction('gpt')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="carbon:logical-partition" class="mr-1 h-4 w-4" />
			{getTranslation('disk.initialize_gpt', 'Initialize GPT')}
		</Button>
	{/if}

	{#if type == 'wipe-disk' && buttonAbilities.wipe.ability && activeDisk !== null}
		<Button
			onclick={() => diskAction('wipe')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('disk.wipe_disk', 'Wipe Disk'))}
		</Button>
	{/if}

	{#if type == 'wipe-partition' && buttonAbilities.wipe.ability && activePartition !== null}
		<Button
			onclick={() => diskAction('wipe')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('disk.delete_partition', 'Delete Partition'))}
		</Button>
	{/if}

	{#if type == 'partition' && buttonAbilities.createPartition.ability}
		<Button
			onclick={() => diskAction('partition')}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="ant-design:partition-outlined" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('disk.create_partition', 'Create Partition'))}
		</Button>
	{/if}
{/snippet}

<div class="flex h-full flex-col overflow-hidden">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />

		{@render button('smart')}
		{@render button('gpt')}
		{@render button('partition')}
		{@render button('wipe-disk')}
		{@render button('wipe-partition')}
	</div>

	<KvTableModal
		titles={{
			main: smartModal.title,
			key: getTranslation('disk.attribute', 'Attribute'),
			value: getTranslation('disk.value', 'Value')
		}}
		open={smartModal.open}
		KV={smartModal.KV}
		type={smartModal.type}
		actions={{
			close: () => {
				smartModal.open = false;
			}
		}}
	></KvTableModal>

	<TreeTable
		data={{
			rows: rows,
			columns: columns
		}}
		name={'tt-disks'}
		bind:parentActiveRow={activeRows}
		multipleSelect={false}
		bind:query
	/>
</div>

<AlertDialog
	open={wipeModal.open}
	names={{ parent: 'disks', element: wipeModal.title || '' }}
	actions={{
		onConfirm: async () => {
			if (activeDisk || activePartition) {
				const message = activeDisk
					? getTranslation('disk.full_wipe_success', 'Disk wiped successfully')
					: getTranslation('disk.partition_wipe_success', 'Disk wiped successfully');

				const result = activeDisk
					? await destroyDisk(`/dev/${activeDisk.device}`)
					: await destroyPartition(`/dev/${activePartition?.name}`);

				if (result.status === 'success') {
					toast.success(message, { position: 'bottom-center' });
					activeRow = null;
				} else {
					handleAPIError(result);
				}
			}
			wipeModal.title = '';
			wipeModal.open = false;
		},
		onCancel: () => {
			wipeModal.title = '';
			wipeModal.open = false;
		}
	}}
	customTitle={wipeModal.title}
></AlertDialog>

<CreatePartition
	open={partitionModal.open}
	disk={partitionModal.disk}
	onCancel={() => {
		partitionModal.open = false;
		partitionModal.disk = null;
	}}
/>
