<script lang="ts">
	import { listDisks } from '$lib/api/disk/disk';
	import { createPool, deletePool, getPools, replaceDevice, scrubPool } from '$lib/api/zfs/pool';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import CustomCheckbox from '$lib/components/ui/custom-input/checkbox.svelte';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import type { Row } from '$lib/types/components/tree-table';
	import type { Disk, Partition } from '$lib/types/disk/disk';
	import type { Zpool, ZpoolRaidType } from '$lib/types/zfs/pool';
	import {
		getDiskSize,
		stripDev,
		zpoolUseableDisks,
		zpoolUseablePartitions
	} from '$lib/utils/disk';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { flip } from 'svelte/animate';
	import { slide } from 'svelte/transition';

	import { Textarea } from '$lib/components/ui/textarea';
	import { draggable, dropzone } from '$lib/utils/dnd';
	import { isValidPoolName } from '$lib/utils/zfs';
	import {
		generateTableData,
		getPoolByDevice,
		isPool,
		isReplaceableDevice,
		parsePoolActionError,
		raidTypeArr
	} from '$lib/utils/zfs/pool';
	import humanFormat from 'human-format';
	import { untrack } from 'svelte';
	import toast from 'svelte-french-toast';

	import AlertDialogModal from '$lib/components/custom/AlertDialog.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Badge } from '$lib/components/ui/badge';

	import TreeTable from '$lib/components/custom/TreeTable.svelte';
	import Search from '$lib/components/custom/TreeTable/Search.svelte';
	import { updateCache } from '$lib/utils/http';
	import { getTranslation } from '$lib/utils/i18n';
	import { capitalizeFirstLetter } from '$lib/utils/string';

	interface Data {
		disks: Disk[];
		pools: Zpool[];
	}

	interface VdevContainer {
		id: string;
		disks: Disk[];
		partitions: Partition[];
	}

	let { data }: { data: Data } = $props();

	const results = useQueries([
		{
			queryKey: ['diskList'],
			queryFn: async () => {
				return await listDisks();
			},
			refetchInterval: 1000,
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

	let activeRow: Row | null = $state(null);
	let replaceInProgress: boolean = $state(false);
	let scrubInProgress: boolean = $state(false);
	let raidTypes = $state(raidTypeArr);

	interface SelectSpares {
		value: string;
		label: string;
		disabled: boolean;
	}

	let modal = $state({
		open: false,
		name: '',
		vdevCount: 1,
		vdevContainers: [] as VdevContainer[],
		raidType: 'stripe' as ZpoolRaidType,
		mountPoint: '',
		advanced: false,
		forceCreate: false,
		properties: {
			comment: '',
			ashift: 12,
			autoexpand: 'off',
			autotrim: 'off',
			delegation: 'off',
			failmode: 'wait'
		},
		useable: 0,
		creating: false,
		spares: [] as SelectSpares[],
		close: () => {
			modal.name = '';
			modal.open = false;
			modal.vdevCount = 1;
			modal.vdevContainers = [];
			modal.advanced = false;
			modal.properties = {
				comment: '',
				ashift: 12,
				autoexpand: 'off',
				autotrim: 'off',
				delegation: 'off',
				failmode: 'wait'
			};
			modal.raidType = 'stripe';
			modal.mountPoint = '';
			modal.useable = 0;
			modal.forceCreate = false;
		}
	});

	let confirmModals = $state({
		active: '' as 'statusPool' | 'deletePool' | 'replaceDevice',
		statusPool: {
			open: false,
			data: {
				status: {} as Zpool['status']
			},
			title: getTranslation('zfs.pool.pool_status', 'Pool Status')
		},
		deletePool: {
			open: false,
			data: '',
			title: getTranslation('zfs.pool.delete_pool', 'Delete Pool')
		},
		replaceDevice: {
			open: false,
			data: {
				pool: '',
				old: '',
				new: ''
			},
			title: capitalizeFirstLetter(getTranslation('zfs.pool.replace_device', 'Replace Device'))
		}
	});

	let disks = $derived($results[0].data as Disk[]);
	let pools = $derived($results[1].data as Zpool[]);
	let useableDisks = $derived(zpoolUseableDisks(disks, pools));
	let useablePartitions = $derived(zpoolUseablePartitions(disks, pools));
	let tableData = $derived(generateTableData(pools, disks));
	let sPool = $derived(pools.find((p) => p.name === activeRow?.name)?.status as Zpool['status']);
	let sPoolSpares = $derived(
		pools.find((p) => p.name === activeRow?.name)?.spares as Zpool['spares']
	);

	$effect(() => {
		modal.vdevCount = Math.max(1, Math.min(128, modal.vdevCount));
		untrack(() => {
			setRedundancyAvailability();
			setUsableSpace();
		});
	});

	let previousVdevCount = $state(modal.vdevCount);

	$effect(() => {
		if (modal.vdevCount < previousVdevCount) {
			const removedContainers = modal.vdevContainers.slice(modal.vdevCount);
			removedContainers.forEach((container) => {
				container.disks.forEach((disk) => {
					if (!useableDisks.some((ud) => ud.uuid === disk.uuid)) {
						useableDisks = [...useableDisks, disk];
					}
				});
				container.partitions.forEach((partition) => {
					if (!useableDisks.some((ud) => ud.partitions.some((p) => p.name === partition.name))) {
						const parentDisk = disks.find((d) =>
							d.partitions.some((p) => p.name === partition.name)
						);
						if (parentDisk) {
							useableDisks = [...useableDisks, { ...parentDisk }];
						}
					}
				});
			});
			modal.vdevContainers = modal.vdevContainers.slice(0, modal.vdevCount);
		}
		previousVdevCount = modal.vdevCount;
	});

	function setRedundancyAvailability() {
		const vdevLengths = modal.vdevContainers.map(
			(vdev) => vdev.disks.length + vdev.partitions.length
		);

		raidTypes = raidTypes.map((type) => {
			switch (type.value) {
				case 'stripe':
					return { ...type, available: true };
				case 'mirror':
					const allMirrors = vdevLengths.every((length) => length === 2) && vdevLengths.length > 0;
					return { ...type, available: allMirrors };
				case 'raidz':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 3) && vdevLengths.length > 0
					};
				case 'raidz2':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 4) && vdevLengths.length > 0
					};
				case 'raidz3':
					return {
						...type,
						available: vdevLengths.every((length) => length >= 5) && vdevLengths.length > 0
					};
				default:
					return type;
			}
		});

		if (!raidTypes.find((rt) => rt.value === modal.raidType)?.available) {
			modal.raidType = (raidTypes.find((rt) => rt.available)?.value as ZpoolRaidType) || 'stripe';
		}

		setUsableSpace();
	}

	function setUsableSpace() {
		let totalUsable = 0;

		for (const vdev of modal.vdevContainers) {
			const sizes = [
				...(vdev.disks ?? []).map((d) => d.size),
				...(vdev.partitions ?? []).map((p) => p.size)
			].filter((size) => typeof size === 'number');

			if (sizes.length === 0) continue;

			sizes.sort((a, b) => a - b);

			const total = sizes.reduce((sum, s) => sum + s, 0);

			switch (modal.raidType) {
				case 'stripe':
					totalUsable += total;
					break;
				case 'mirror':
					totalUsable += sizes[0];
					break;
				case 'raidz':
					if (sizes.length > 1) {
						totalUsable += total - sizes[sizes.length - 1];
					}
					break;
				case 'raidz2':
					if (sizes.length > 2) {
						totalUsable += total - sizes.slice(-2).reduce((a, b) => a + b, 0);
					}
					break;
				case 'raidz3':
					if (sizes.length > 3) {
						totalUsable += total - sizes.slice(-3).reduce((a, b) => a + b, 0);
					}
					break;
				default:
					console.warn(`Unknown RAID type: ${modal.raidType}`);
			}
		}

		modal.useable = totalUsable;
	}

	function isDiskInVdev(diskId: string | undefined | string[]): boolean {
		if (!diskId) return false;

		if (typeof diskId === 'string') {
			return modal.vdevContainers.some((vdev) => {
				return vdev.disks.some((disk) => disk.uuid === diskId);
			});
		}

		if (Array.isArray(diskId)) {
			return modal.vdevContainers.some((vdev) => {
				return vdev.partitions.some((partition) => diskId.includes(partition.name));
			});
		}

		return false;
	}

	function handleDropToVdev(containerId: number, event: DragEvent) {
		const diskId = event.dataTransfer?.getData('application/disk');

		if (!modal.vdevContainers[containerId]) {
			modal.vdevContainers[containerId] = {
				id: `vdev-${containerId}`,
				disks: [],
				partitions: []
			};
		}

		const disk = disks.find((d) => d.uuid === diskId);

		if (disk) {
			const existingDisk = modal.vdevContainers[containerId].disks.find(
				(d) => d.uuid === disk.uuid
			);
			if (!existingDisk) {
				modal.vdevContainers[containerId].disks.push(disk);
				useableDisks = useableDisks.filter((ud) => ud.uuid !== disk.uuid);
			}
		}

		if (!disk) {
			const diskContainingPartition = disks.find((d) =>
				d.partitions.some((p) => p.name === diskId)
			);

			if (diskContainingPartition) {
				const partition = diskContainingPartition.partitions.find((p) => p.name === diskId);
				if (partition) {
					const existingPartition = modal.vdevContainers[containerId].partitions.find(
						(p) => p.name === partition.name
					);
					if (!existingPartition) {
						modal.vdevContainers[containerId].partitions.push(partition);
						useableDisks = useableDisks.filter(
							(ud) => !ud.partitions.some((p) => p.name === partition.name)
						);
					}
				}
			}
		}

		setRedundancyAvailability();
		setUsableSpace();
	}

	function vdevContains(id: number): boolean {
		const vdev = modal.vdevContainers[id];
		if (!vdev) return false;

		return vdev.disks.length > 0 || vdev.partitions.length > 0;
	}

	function removeFromVdev(id: number, diskId: string) {
		const vdev = modal.vdevContainers[id];
		if (!vdev) return;

		const diskIndex = vdev.disks.findIndex((d) => d.uuid === diskId);
		if (diskIndex !== -1) {
			const removedDisk = vdev.disks.splice(diskIndex, 1)[0];
			if (!useableDisks.some((ud) => ud.uuid === removedDisk.uuid)) {
				useableDisks = [...useableDisks, removedDisk];
			}
		}

		const partitionIndex = vdev.partitions.findIndex((p) => p.name === diskId);
		if (partitionIndex !== -1) {
			const removedPartition = vdev.partitions.splice(partitionIndex, 1)[0];
			const parentDisk = disks.find((d) =>
				d.partitions.some((p) => p.name === removedPartition.name)
			);
			if (
				parentDisk &&
				!useableDisks.some((ud) => ud.partitions.some((p) => p.name === removedPartition.name))
			) {
				useableDisks = [...useableDisks, { ...parentDisk }];
			}
		}

		setRedundancyAvailability();
		setUsableSpace();
	}

	function getVdevErrors(id: number): string {
		const vdev = modal.vdevContainers[id];
		const disks = vdev?.disks || [];
		const partitions = vdev?.partitions || [];
		const diskSizes = disks.map((disk) => disk.size);
		const partSizes = partitions.map((partition) => partition.size);
		const allSizes = [...diskSizes, ...partSizes];

		const diskTypes = disks.map((disk) => disk.type);
		for (let i = 0; i < diskTypes.length - 1; i++) {
			if (diskTypes[i] !== diskTypes[i + 1]) {
				return getTranslation(
					'zfs.pool.warnings.disks_within_vdev_should_be_same_type',
					'Disks within a VDEV should ideally be the same type'
				);
			}
		}

		const partitionTypes = partitions.map((partition) => {
			const disk = useableDisks.find((d) => d.partitions.some((p) => p.name === partition.name));
			return disk ? disk.type : null;
		});

		for (let i = 0; i < partitionTypes.length - 1; i++) {
			if (partitionTypes[i] !== partitionTypes[i + 1]) {
				return getTranslation(
					'disks_within_vdev_should_be_same_drive_type',
					'Disks within a VDEV should ideally be the same drive type'
				);
			}
		}

		for (let i = 0; i < allSizes.length - 1; i++) {
			if (allSizes[i] !== allSizes[i + 1]) {
				if (partSizes.length === 0) {
					return getTranslation(
						'zfs.pool.warnings.disks_within_vdev_should_be_same_size',
						'Disks within a VDEV should ideally be the same size'
					);
				} else if (diskSizes.length === 0) {
					return getTranslation(
						'zfs.pool.warnings.partitions_within_vdev_should_be_same_size',
						'Partitions within a VDEV should ideally be the same size'
					);
				} else {
					return getTranslation(
						'zfs.pool.warnings.disks_or_partitions_within_vdev_should_be_same_drive_type',
						'Disks/Partitions within a VDEV should ideally be the same drive type'
					);
				}
			}
		}

		return '';
	}

	async function makePool() {
		if (modal.creating) return;

		if (useableDisks.length === 0 && useablePartitions.length === 0) {
			toast.error(
				getTranslation(
					'zfs.pool.errors.pool_create_failed_no_disks',
					'No available disks or partitions'
				),
				{
					position: 'bottom-center'
				}
			);
			return;
		}

		if (!isValidPoolName(modal.name)) {
			toast.error(
				getTranslation('zfs.pool.errors.pool_create_failed_invalid_name', 'Invalid pool name'),
				{
					position: 'bottom-center'
				}
			);
			return;
		}

		if (modal.vdevContainers.length === 0) {
			toast.error(
				getTranslation(
					'zfs.pool.errors.pool_create_failed_need_atleast_one',
					'Please add at least one disk'
				),
				{
					position: 'bottom-center'
				}
			);
			return;
		}

		if (
			modal.vdevContainers.some((vdev) => {
				return vdev.disks.length === 0 && vdev.partitions.length === 0;
			})
		) {
			modal.vdevContainers = modal.vdevContainers.filter((vdev) => {
				return vdev.disks.length > 0 || vdev.partitions.length > 0;
			});
			return;
		}

		let raidType: ZpoolRaidType = modal.raidType;

		if (modal.raidType === 'stripe') {
			raidType = undefined;
		}

		modal.creating = true;
		let biggestSize = 0;

		for (const vdev of modal.vdevContainers) {
			const sizes = [
				...(vdev.disks ?? []).map((d) => d.size),
				...(vdev.partitions ?? []).map((p) => p.size)
			].filter((size) => typeof size === 'number');

			if (sizes.length === 0) continue;
			sizes.sort((a, b) => a - b);
			biggestSize = Math.max(biggestSize, ...sizes);
		}

		if (modal.spares.length !== 0) {
			const spareSizes = modal.spares.map((spare) => {
				const disk = useableDisks.find((d) => d.device === spare.value);
				if (disk) {
					return disk.size;
				}
				const partition = useablePartitions.find((p) => p.name === spare.value);
				if (partition) {
					return partition.size;
				}
				return 0;
			});

			const minSpareSize = Math.min(...spareSizes);
			if (minSpareSize < biggestSize) {
				toast.error(
					getTranslation(
						'zfs.pool.errors.pool_create_failed_spare_smaller',
						'Spares must be larger than the largest disk in the pool'
					),
					{
						position: 'bottom-center'
					}
				);
				modal.creating = false;
				return;
			}
		}

		const response = await createPool({
			name: modal.name,
			raidType: raidType,
			vdevs: modal.vdevContainers.map((vdev) => ({
				name: vdev.id,
				devices: [
					...vdev.disks.map((disk) => disk.device),
					...vdev.partitions.map((partition) => partition.name)
				]
			})),
			properties: {
				comment: modal.properties.comment,
				ashift: modal.properties.ashift.toString(),
				autoexpand: modal.properties.autoexpand,
				autotrim: modal.properties.autotrim,
				delegation: modal.properties.delegation,
				failmode: modal.properties.failmode
			},
			spares: modal.spares.map((spare) => spare.value),
			createForce: modal.forceCreate
		});

		modal.creating = false;

		if (response.status === 'error') {
			toast.error(parsePoolActionError(response), {
				position: 'bottom-center'
			});
		} else {
			toast.success(getTranslation(`zfs.pool.${response.message}`, 'Pool created'), {
				position: 'bottom-center'
			});

			modal.close();
		}
	}

	async function confirmAction() {
		if (confirmModals.active === 'deletePool') {
			const response = await deletePool(confirmModals.deletePool.data);
			if (response.status === 'error') {
				toast.error(parsePoolActionError(response), {
					position: 'bottom-center'
				});
			} else {
				toast.success(getTranslation(`zfs.pool.${response.message}`, 'Pool deleted'), {
					position: 'bottom-center'
				});
			}
		}

		if (confirmModals.active === 'replaceDevice') {
			const { old: oldName, new: newName, pool: poolName } = confirmModals.replaceDevice.data;
			const pool = pools.find((p) => p.name === poolName);
			const vdev = pool?.vdevs.find((v) => v.devices.some((d) => d.name === oldName));
			const oldDevice = vdev?.devices.find((d) => d.name === oldName);
			const newDevice = useableDisks.find((d) => d.device === newName);
			const disks = {
				old: oldDevice,
				new: newDevice
			};

			if (disks.old && disks.new) {
				if (parseInt(getDiskSize(disks.new)) < disks.old.size) {
					toast.error(
						getTranslation(
							'zfs.pool.errors.pool_create_failed_new_device_smaller',
							'New device is smaller than old device'
						),
						{
							position: 'bottom-center'
						}
					);

					return;
				}
			} else if (!disks.new && disks.old) {
				const partition = useablePartitions.find((p) => p.name === newName);
				if (partition) {
					if (partition.size < disks.old.size) {
						toast.error(
							getTranslation(
								'zfs.pool.errors.pool_create_failed_new_partition_smaller',
								'New partition is smaller than old device'
							),
							{
								position: 'bottom-center'
							}
						);

						return;
					}
				}
			}

			replaceInProgress = true;

			try {
				await toast.promise(
					replaceDevice({
						name: poolName,
						old: oldName,
						new: newName
					}),
					{
						loading: `${getTranslation('zfs.pool.replacing', 'Replacing')} ${stripDev(oldName)} ${getTranslation('common.with', 'with')} ${stripDev(newName)}`,
						success: (response) => {
							return getTranslation(`zfs.pool.${response.message}`, 'Device replacement started');
						},
						error: (error) => {
							return getTranslation(`zfs.pool.${error.message}`, 'Error replacing device');
						}
					},
					{
						position: 'bottom-center'
					}
				);
			} finally {
				replaceInProgress = false;
			}
		}

		confirmModals[confirmModals.active].open = false;
	}

	$effect(() => {
		if (
			JSON.stringify(tableData).toLowerCase().includes('replaced') &&
			JSON.stringify(tableData).toLowerCase().includes('replacing')
		) {
			replaceInProgress = true;
		} else {
			replaceInProgress = false;
		}

		if (JSON.stringify(pools).toLowerCase().includes('scrub in progress since')) {
			scrubInProgress = true;
		} else {
			scrubInProgress = false;
		}
	});

	let possibleSpares: string[] = $derived.by(() => {
		const uD: string[] = useableDisks
			.filter((disk) => {
				return !modal.vdevContainers.some((vdev) => {
					return vdev.disks.some((d) => d.uuid === disk.uuid);
				});
			})
			.map((disk) => disk.device);

		const uP: string[] = useablePartitions
			.filter((partition) => {
				return !modal.vdevContainers.some((vdev) => {
					return vdev.partitions.some((p) => p.name === partition.name);
				});
			})
			.map((partition) => partition.name);

		return [...uD, ...uP].filter((device) => {
			return device !== 'da0' && device !== 'cd0';
		});
	});

	let query: string = $state('');
</script>

{#snippet button(type: string)}
	{#if activeRow !== null}
		{#if type === 'pool-status'}
			{#if isPool(pools, activeRow.name)}
				<Button
					on:click={() => {
						confirmModals.active = 'statusPool';
						confirmModals.statusPool.open = true;
						confirmModals.statusPool.title = activeRow?.name;
						confirmModals.statusPool.data.status = pools.find((p) => p.name === activeRow?.name)
							?.status as Zpool['status'];
					}}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
				>
					<Icon icon="mdi:eye" class="mr-1 h-4 w-4" />
					{getTranslation('common.status', 'Status')}
				</Button>
			{/if}
		{/if}

		{#if type === 'pool-scrub'}
			{#if isPool(pools, activeRow.name)}
				<Button
					on:click={async () => {
						const response = await scrubPool(activeRow?.name);
						if (response.status === 'error') {
							toast.error(parsePoolActionError(response), {
								position: 'bottom-center'
							});
						} else {
							toast.success(getTranslation(`zfs.pool.${response.message}`, 'Scrub started'), {
								position: 'bottom-center'
							});
						}
					}}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
					disabled={scrubInProgress}
					title={scrubInProgress
						? getTranslation('zfs.pool.scrub_in_progress', 'A scrub is already in progress')
						: ''}
				>
					<Icon icon="cil:scrubber" class="mr-1 h-4 w-4" />
					{getTranslation('zfs.pool.scrub', 'Scrub')}
				</Button>
			{/if}
		{/if}

		{#if type === 'pool-delete'}
			{#if isPool(pools, activeRow.name)}
				<Button
					on:click={() => {
						confirmModals.active = 'deletePool';
						confirmModals.deletePool.open = true;
						confirmModals.deletePool.title = activeRow?.name;
						confirmModals.deletePool.data = activeRow?.name;
					}}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
					disabled={replaceInProgress}
					title={replaceInProgress
						? getTranslation(
								'zfs.pool.warnings.cannot_delete_pool_while_replacing_device',
								'Cannot delete pool while replacing device in any pool'
							)
						: ''}
				>
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					{getTranslation('zfs.pool.delete', 'Delete')}
				</Button>
			{/if}
		{/if}

		{#if type === 'replace-device'}
			{#if isReplaceableDevice(pools, activeRow.name)}
				<Button
					on:click={() => {
						confirmModals.active = 'replaceDevice';
						confirmModals.replaceDevice.open = true;
						confirmModals.replaceDevice.title = activeRow?.name;
						confirmModals.replaceDevice.data = {
							pool: getPoolByDevice(pools, activeRow?.name),
							old: activeRow?.name,
							new: ''
						};
					}}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
					disabled={replaceInProgress}
					title={replaceInProgress
						? 'Cannot replace device while replacing device in any pool'
						: ''}
				>
					<Icon icon="mdi:swap-horizontal" class="mr-1 h-4 w-4" />
					{capitalizeFirstLetter(getTranslation('zfs.pool.replace_device', 'Replace Device'))}
				</Button>
			{/if}
		{/if}
	{/if}
{/snippet}

{#snippet diskContainer(type: string)}
	<div id="{type.toLowerCase()}-container">
		<Label>{type}</Label>
		<div class="bg-primary/10 dark:bg-background mt-1 rounded-lg p-4">
			<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
				<div class="flex min-h-[80px] items-center justify-center gap-4">
					{#each useableDisks.filter((disk) => disk.type === type && disk.partitions.length === 0 && !isDiskInVdev(disk.uuid)) as disk (disk.uuid)}
						<div class="text-center" animate:flip={{ duration: 300 }}>
							<div class="cursor-move" use:draggable={disk.uuid ?? ''}>
								{#if type === 'HDD'}
									<Icon icon="mdi:harddisk" class="h-11 w-11 text-green-500" />
								{:else if type === 'SSD'}
									<Icon icon="icon-park-outline:ssd" class="h-11 w-11 text-blue-500" />
								{:else if type === 'NVMe'}
									<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
								{/if}
							</div>
							<div class="max-w-[64px] truncate text-xs">
								{disk.device.replaceAll('/dev/', '')}
							</div>
							<div class="text-muted-foreground text-xs">
								{humanFormat(disk.size)}
							</div>
						</div>
					{/each}

					{#if useableDisks.filter((disk) => disk.type === type).length === 0 || useableDisks.filter((disk) => disk.type === type && disk.partitions.length === 0 && !isDiskInVdev(disk.uuid)).length === 0}
						<div class="text-muted-foreground/80 flex h-16 w-full items-center justify-center">
							{getTranslation('zfs.pool.no_available_disks', 'No available disks')}
						</div>
					{/if}
				</div>
			</ScrollArea>
		</div>
	</div>
{/snippet}

{#snippet partitionsContainer()}
	<div id="partitions-container">
		<Label>{getTranslation('zfs.pool.partitions', 'Partitions')}</Label>
		<div class="bg-primary/10 dark:bg-background mt-1 rounded-lg p-4">
			<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
				<div class="flex min-h-[80px] items-center justify-center gap-4">
					{#each useablePartitions.filter((partition) => !modal.vdevContainers
								.flatMap((vdev) => vdev.partitions)
								.some((p) => p.name === partition.name)) as partition (partition.name)}
						<div class="text-center" animate:flip={{ duration: 100 }}>
							<div class="cursor-move" use:draggable={partition.name}>
								<Icon
									icon="ant-design:partition-outlined"
									class="h-11 w-11 rotate-90 text-blue-500"
								/>
							</div>
							<div class="max-w-[64px] truncate text-xs">
								{partition.name}
							</div>
							<div class="text-muted-foreground text-xs">
								{humanFormat(partition.size)}
							</div>
						</div>
					{/each}

					{#if useablePartitions.length === 0 || useablePartitions.filter((partition) => !modal.vdevContainers
									.flatMap((vdev) => vdev.partitions)
									.some((p) => p.name === partition.name)).length === 0}
						<div class="flex h-16 w-full items-center justify-center text-neutral-400">
							{getTranslation('zfs.pool.no_available_partitions', 'No available partitions')}
						</div>
					{/if}
				</div>
			</ScrollArea>
		</div>
	</div>
{/snippet}

{#snippet vdevContainer(id: number)}
	{#each modal.vdevContainers[id]?.disks || [] as disk (disk.uuid)}
		<div animate:flip={{ duration: 300 }} class="relative">
			{#if disk.type === 'HDD'}
				<Icon icon="mdi:harddisk" class="h-11 w-11 text-green-500" />
			{:else if disk.type === 'SSD'}
				<Icon icon="icon-park-outline:ssd" class="h-11 w-11 text-blue-500" />
			{:else if disk.type === 'NVMe'}
				<Icon icon="bi:nvme" class="h-11 w-11 rotate-90 text-blue-500" />
			{/if}

			<div class="max-w-[48px] truncate text-center text-xs">
				{disk.device.split('/').pop()}
			</div>

			<button
				class="absolute -right-1 -top-1 rounded-full bg-red-500 p-0.5 text-white hover:bg-red-600"
				onclick={() => removeFromVdev(id, disk.uuid as string)}
			>
				<Icon icon="mdi:close" class="h-3 w-3" />
			</button>
		</div>
	{/each}

	{#each modal.vdevContainers[id]?.partitions || [] as partition (partition.name)}
		<div animate:flip={{ duration: 300 }} class="relative">
			<Icon icon="ant-design:partition-outlined" class="h-11 w-11 rotate-90 text-blue-500" />

			<div class="max-w-[48px] truncate text-center text-xs">
				{partition.name.split('/').pop()}
			</div>

			<button
				class="absolute -right-1 -top-1 rounded-full bg-red-500 p-0.5 text-white hover:bg-red-600"
				onclick={() => removeFromVdev(id, partition.name)}
			>
				<Icon icon="mdi:close" class="h-3 w-3" />
			</button>
		</div>
	{/each}
{/snippet}

{#snippet vdevErrors(id: number)}
	{#if getVdevErrors(id) !== ''}
		<div class="absolute right-1 top-1 z-50 cursor-pointer text-yellow-700 hover:text-yellow-600">
			<Tooltip.Root>
				<Tooltip.Trigger><Icon icon="carbon:warning-filled" class="h-5 w-5" /></Tooltip.Trigger>
				<Tooltip.Content>
					<p>
						{@html getVdevErrors(id)}
					</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</div>
	{/if}
{/snippet}

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2">
		<Search bind:query />
		<Button on:click={() => (modal.open = !modal.open)} size="sm" class="h-6">
			<Icon icon="gg:add" class="mr-1 h-4 w-4" />
			{capitalizeFirstLetter(getTranslation('common.new', 'New'))}
		</Button>

		{@render button('pool-status')}
		{@render button('pool-scrub')}
		{@render button('pool-delete')}
		{@render button('replace-device')}
	</div>

	<TreeTable data={tableData} name="tt-zfsPool" bind:parentActiveRow={activeRow} bind:query />
</div>

<Dialog.Root bind:open={modal.open} closeOnOutsideClick={false}>
	<Dialog.Content
		class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
	>
		<div class="flex items-center justify-between px-4 py-3">
			<Dialog.Header class="p-0">
				<Dialog.Title>{getTranslation('zfs.pool.create_zfs_pool', 'Create ZFS Pool')}</Dialog.Title>
			</Dialog.Header>

			<Dialog.Close
				class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
				onclick={() => modal.close()}
			>
				<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
			</Dialog.Close>
		</div>
		<Tabs.Root value="tab-devices" class="w-full overflow-hidden">
			<Tabs.List class="grid w-full grid-cols-2 p-0 px-4">
				<Tabs.Trigger value="tab-devices" class="border-b">
					{capitalizeFirstLetter(getTranslation('common.devices', 'Devices'))}
				</Tabs.Trigger>
				<Tabs.Trigger value="tab-options" class="border-b"
					>{capitalizeFirstLetter(getTranslation('common.options', 'Options'))}</Tabs.Trigger
				>
			</Tabs.List>
			<Tabs.Content class="mt-0" value="tab-devices">
				<Card.Root class="border-none pb-4">
					<Card.Content class="flex gap-4 p-4 !pb-0">
						<CustomValueInput
							label={capitalizeFirstLetter(getTranslation('common.name', 'Name'))}
							placeholder="tank"
							bind:value={modal.name}
							classes="flex-1 space-y-1"
						/>

						<CustomValueInput
							label="{capitalizeFirstLetter(
								getTranslation('common.virtual', 'Virtual')
							)}{capitalizeFirstLetter(getTranslation('common.devices', 'Devices'))}"
							placeholder="1"
							bind:value={modal.vdevCount}
							classes="flex-1 space-y-1"
							type="number"
						></CustomValueInput>

						<div class="flex-1 space-y-1">
							<Label class="w-24 whitespace-nowrap text-sm" for="raid"
								>{capitalizeFirstLetter(
									getTranslation('zfs.pool.redundancy.redundancy', 'Redundancy')
								)}
								<span class="font-semibold text-green-500 {modal.useable ? '' : 'hidden'}"
									>({humanFormat(modal.useable)})</span
								></Label
							>
							<Select.Root
								selected={{
									label: raidTypes.find((rt) => rt.value === modal.raidType)?.label,
									value: raidTypes.find((rt) => rt.value === modal.raidType)?.value
								}}
								onSelectedChange={(value) => {
									modal.raidType = value?.value as ZpoolRaidType;
									setRedundancyAvailability();
								}}
							>
								<Select.Trigger class="w-full">
									<Select.Value placeholder="Select Redundancy" />
								</Select.Trigger>
								<Select.Content class="max-h-36 overflow-y-auto">
									<Select.Group>
										{#each raidTypes as raidType}
											{#if raidType.available}
												<Select.Item value={raidType.value} label={raidType.label}
													>{raidType.label}</Select.Item
												>
											{/if}
										{/each}
									</Select.Group>
								</Select.Content>
							</Select.Root>
						</div>
					</Card.Content>

					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div id="vdev-containers">
							<Label>VDEVs</Label>
							<ScrollArea class="w-full whitespace-nowrap rounded-md" orientation="horizontal">
								<div
									class="bg-muted mt-1 flex w-full items-center justify-center gap-7 overflow-hidden rounded-lg border-y border-none p-4 pr-4"
								>
									{#each Array(modal.vdevCount) as _, i}
										<div class="relative flex flex-col">
											{@render vdevErrors(i)}

											<div
												class={`bg-primary/10 dark:bg-background relative h-28 w-48 flex-shrink-0 overflow-auto rounded-lg p-2 ${getVdevErrors(i) ? 'border border-yellow-700 ' : ''}`}
												use:dropzone={{
													on_dropzone: (_: unknown, event: DragEvent) => handleDropToVdev(i, event),
													dragover_class: 'droppable'
												}}
											>
												{#if !vdevContains(i)}
													<div
														class="text-muted-foreground/60 flex h-full flex-col items-center justify-center gap-1"
													>
														<span class="text-muted-foreground/60">{i + 1}</span>
														<span>Drop disks here</span>
													</div>
												{:else}
													<div class="flex h-full flex-wrap items-center justify-center gap-2">
														{@render vdevContainer(i)}
													</div>
												{/if}
											</div>
										</div>
									{/each}
								</div></ScrollArea
							>
						</div>
					</Card.Content>

					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div id="disk-containers">
							<Label>Disks</Label>
							<div
								class="bg-muted mt-1 grid grid-cols-4 gap-6 overflow-hidden border-y border-none p-4"
							>
								{@render diskContainer('HDD')}
								{@render diskContainer('SSD')}
								{@render diskContainer('NVMe')}
								{@render partitionsContainer()}
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tabs.Content>

			<Tabs.Content class="mt-0" value="tab-options">
				<Card.Root class="min-h-[20vh] border-none pb-6">
					<Card.Content class="flex flex-col gap-4 p-4 !pb-0">
						<div transition:slide class="grid grid-cols-1 gap-4">
							<div class="flex-1 space-y-1">
								<Label for="comment">Comment</Label>
								<Textarea
									id="comment"
									placeholder="Comments about the pool"
									bind:value={modal.properties.comment}
								/>
							</div>

							<div transition:slide class="grid grid-cols-1 items-center gap-4 md:grid-cols-3">
								<CustomValueInput
									type="text"
									label={capitalizeFirstLetter(
										getTranslation('zfs.pool.mount_point', 'Mount Point')
									)}
									placeholder="/tank"
									bind:value={modal.mountPoint}
									classes="flex-1 space-y-1"
								></CustomValueInput>

								<div class="col-span-2 flex items-center gap-6 md:mt-4">
									<CustomCheckbox
										label="Force Create"
										bind:checked={modal.forceCreate}
										classes="flex items-center gap-2"
									></CustomCheckbox>

									<CustomCheckbox
										label="Advanced"
										bind:checked={modal.advanced}
										classes="flex items-center gap-2"
									></CustomCheckbox>
								</div>
							</div>
						</div>

						{#if modal.advanced}
							<div transition:slide class="grid grid-cols-1 gap-4 md:grid-cols-3">
								<!-- Ashift -->
								<div class="h-full space-y-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="ashift">Ashift</Label>
									<Select.Root
										selected={{
											label: [
												{ value: 0, label: '0 (auto)' },
												...Array.from({ length: 8 }, (_, i) => {
													const val = i + 9;
													return { value: val, label: `${val}` };
												})
											].find((opt) => opt.value === modal.properties.ashift)?.label,
											value: [
												{ value: 0, label: '0 (auto)' },
												...Array.from({ length: 8 }, (_, i) => {
													const val = i + 9;
													return { value: val, label: `${val}` };
												})
											].find((opt) => opt.value === modal.properties.ashift)?.value
										}}
										onSelectedChange={(value) => {
											modal.properties.ashift = value?.value || 0;
										}}
									>
										<Select.Trigger class="w-full">
											<Select.Value placeholder="Select Ashift" />
										</Select.Trigger>
										<Select.Content class="max-h-36 overflow-y-auto">
											<Select.Group>
												<Select.Item value={0} label="0 (auto)">0 (auto)</Select.Item>
												{#each Array.from({ length: 8 }, (_, i) => i + 9) as val}
													<Select.Item value={val} label={`${val}`}>{val}</Select.Item>
												{/each}
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>

								<!-- Auto Expand -->
								<div class="h-full space-y-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="autoexpand">Auto Expand</Label>
									<Select.Root
										selected={{
											label:
												modal.properties.autoexpand === 'on'
													? 'Yes'
													: modal.properties.autoexpand === 'off'
														? 'No'
														: undefined,
											value: modal.properties.autoexpand
										}}
										onSelectedChange={(value) => {
											modal.properties.autoexpand = value?.value || 'off';
										}}
									>
										<Select.Trigger class="w-full">
											<Select.Value placeholder="Select Autoexpand" />
										</Select.Trigger>
										<Select.Content class="max-h-36 overflow-y-auto">
											<Select.Group>
												<Select.Item value="on" label="Yes">Yes</Select.Item>
												<Select.Item value="off" label="No">No</Select.Item>
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>

								<!-- Auto Trim -->
								<div class="h-full space-y-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="autotrim">Auto Trim</Label>
									<Select.Root
										selected={{
											label:
												modal.properties.autotrim === 'on'
													? 'Yes'
													: modal.properties.autotrim === 'off'
														? 'No'
														: undefined,
											value: modal.properties.autotrim
										}}
										onSelectedChange={(value) => {
											modal.properties.autotrim = value?.value || 'off';
										}}
									>
										<Select.Trigger class="w-full">
											<Select.Value placeholder="Select Auto Trim" />
										</Select.Trigger>
										<Select.Content class="max-h-36 overflow-y-auto">
											<Select.Group>
												<Select.Item value="on" label="Yes">Yes</Select.Item>
												<Select.Item value="off" label="No">No</Select.Item>
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>

								<!-- Delegation -->
								<div class="h-full space-y-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="delegation">Delegation</Label>
									<Select.Root
										selected={{
											label:
												modal.properties.delegation === 'on'
													? 'Yes'
													: modal.properties.delegation === 'off'
														? 'No'
														: undefined,
											value: modal.properties.delegation
										}}
										onSelectedChange={(value) => {
											modal.properties.delegation = value?.value || 'off';
										}}
									>
										<Select.Trigger class="w-full">
											<Select.Value placeholder="Select Delegation" />
										</Select.Trigger>
										<Select.Content class="max-h-36 overflow-y-auto">
											<Select.Group>
												<Select.Item value="on" label="Yes">Yes</Select.Item>
												<Select.Item value="off" label="No">No</Select.Item>
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>

								<!-- Fail Mode -->
								<div class="h-full space-y-1">
									<Label class="w-24 whitespace-nowrap text-sm" for="failmode">Fail Mode</Label>
									<Select.Root
										selected={{
											label:
												modal.properties.failmode === 'wait'
													? 'Wait'
													: modal.properties.failmode === 'continue'
														? 'Continue'
														: modal.properties.failmode === 'panic'
															? 'Panic'
															: undefined,
											value: modal.properties.failmode
										}}
										onSelectedChange={(value) => {
											modal.properties.failmode = value?.value || 'wait';
										}}
									>
										<Select.Trigger class="w-full">
											<Select.Value placeholder="Select Delegation" />
										</Select.Trigger>
										<Select.Content class="max-h-36 overflow-y-auto">
											<Select.Group>
												<Select.Item value="wait" label="Wait">Wait</Select.Item>
												<Select.Item value="continue" label="Continue">Continue</Select.Item>
												<Select.Item value="panic" label="Panic">Panic</Select.Item>
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>

								{#if possibleSpares && possibleSpares.length > 0 && modal.raidType !== 'stripe'}
									<div class="h-full space-y-1">
										<Label class="w-24 whitespace-nowrap text-sm" for="spares">Spares</Label>
										<Select.Root
											multiple={true}
											name="spares"
											selected={modal.spares}
											onSelectedChange={(v) => {
												modal.spares = v as SelectSpares[];
											}}
										>
											<Select.Trigger>
												{#if possibleSpares.length > 0}
													<span>
														{modal.spares.length > 0
															? modal.spares.map((s) => s.label).join(', ')
															: 'Select spares'}
													</span>
												{:else}
													<span>Select spares</span>
												{/if}
											</Select.Trigger>
											<Select.Content>
												<Select.Group>
													{#each possibleSpares as spare}
														<Select.Item value={spare} label={spare}>
															{spare}
														</Select.Item>
													{/each}
												</Select.Group>
											</Select.Content>
										</Select.Root>
									</div>
								{/if}
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</Tabs.Content>
		</Tabs.Root>

		<Dialog.Footer class="flex justify-between gap-2 border-t px-4 py-3">
			<div class="flex gap-2">
				<Button
					variant="default"
					class="h-8 bg-blue-600 text-white hover:bg-blue-700"
					onclick={() => makePool()}
				>
					{#if modal.creating}
						<Icon icon="mdi:loading" class="mr-1 h-4 w-4 animate-spin" />
					{:else}
						{capitalizeFirstLetter(getTranslation('common.create', 'Create'))}
					{/if}
				</Button>
			</div>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

{#if confirmModals.active == 'deletePool'}
	<AlertDialogModal
		open={confirmModals.active && confirmModals[confirmModals.active].open}
		names={{
			parent: 'pool',
			element: confirmModals.active ? confirmModals[confirmModals.active].title || '' : ''
		}}
		actions={{
			onConfirm: () => {
				if (confirmModals.active) {
					confirmAction();
				}
			},
			onCancel: () => {
				if (confirmModals.active) {
					confirmModals[confirmModals.active].open = false;
				}
			}
		}}
	></AlertDialogModal>
{/if}

{#snippet dtEl(device: Zpool['status']['devices'][0], showNote: boolean)}
	<div
		class="mr-3 h-2.5 w-2.5 rounded-full
        {device.state === 'ONLINE'
			? 'bg-green-500'
			: device.state === 'DEGRADED'
				? 'bg-yellow-500'
				: device.state === 'FAULTED'
					? 'bg-red-500'
					: 'bg-gray-500'}"
		title={device.state}
	></div>

	<div class="flex items-center gap-2 font-medium">
		<div>{device.name}</div>
		{#if device.note && showNote}
			<div
				class="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-100"
			>
				{#if device.note === '(resilvering)'}
					<span>Resilvering</span>
				{:else}
					<span>{device.note}</span>
				{/if}
			</div>
		{/if}
	</div>

	{#if device.read > 0 || device.write > 0 || device.cksum > 0}
		<div class="ml-auto flex gap-2 text-xs">
			{#if device.read > 0}
				<span class="rounded bg-red-100 px-2 py-0.5 text-red-800 dark:bg-red-900 dark:text-red-100"
					>READ: {device.read}</span
				>
			{/if}
			{#if device.write > 0}
				<span class="rounded bg-red-100 px-2 py-0.5 text-red-800 dark:bg-red-900 dark:text-red-100"
					>WRITE: {device.write}</span
				>
			{/if}
			{#if device.cksum > 0}
				<span class="rounded bg-red-100 px-2 py-0.5 text-red-800 dark:bg-red-900 dark:text-red-100"
					>CKSUM: {device.cksum}</span
				>
			{/if}
		</div>
	{/if}
{/snippet}

{#snippet deviceTreeNode(device: Zpool['status']['devices'][0], showNote: boolean)}
	<div class="device-tree relative">
		<div class="bg-background relative flex items-center rounded-md border p-1.5">
			{#if showNote && !device.__isLast}
				<div
					class="bg-secondary absolute -left-6 bottom-0 top-0 w-0.5"
					style="height: calc(100% + 0.7rem);"
				></div>
			{:else}
				<div class="bg-secondary absolute -left-6 bottom-0 top-0 w-0.5" style="height: 18px;"></div>
			{/if}
			{@render dtEl(device, showNote)}
		</div>

		{#if device.children && device.children.length > 0 && !device.name.startsWith('replacing')}
			<div class=" ml-5 mt-2 space-y-2 pl-4">
				{#each device.children as child, index (child.name)}
					<div class="relative">
						<div
							class="bg-secondary h-0.5 w-6"
							style="position: absolute;left: -23px;top:18px"
						></div>
						{@render deviceTreeNode(
							{ ...child, __isLast: index === device.children.length - 1 },
							true
						)}
					</div>
				{/each}
			</div>
		{/if}

		{#if device.name.startsWith('replacing') && device.children && device.children.length > 0}
			<div class="border-border ml-5 mt-2 space-y-2 border-l-2 pl-4">
				{#each device.children as replaceDisk}
					{@render deviceTreeNode(replaceDisk, true)}
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

{#if confirmModals.active == 'statusPool'}
	<Dialog.Root
		bind:open={confirmModals.statusPool.open}
		onOutsideClick={() => {
			confirmModals.statusPool.open = false;
		}}
		closeOnOutsideClick={true}
		closeOnEscape={false}
	>
		<Dialog.Content
			class="fixed left-1/2 top-1/2 max-h-[90vh] w-[80%] -translate-x-1/2 -translate-y-1/2 transform gap-0 overflow-visible overflow-y-auto p-0 transition-all duration-300 ease-in-out lg:max-w-[70%]"
		>
			<div class="flex items-center justify-between px-4 py-3">
				<Dialog.Header>
					<Dialog.Title class="flex items-center">
						<span class="text-primary font-semibold">Pool Status</span>
						<span class="text-muted-foreground mx-2">•</span>
						<span class="text-xl font-medium">{confirmModals.statusPool.data.status.name}</span>
						<Badge
							variant={sPool.state === 'ONLINE'
								? 'success'
								: sPool.state === 'DEGRADED'
									? 'warning'
									: sPool.state === 'FAULTED'
										? 'destructive'
										: 'secondary'}
							class="ml-3">{sPool.state}</Badge
						>
					</Dialog.Title>
				</Dialog.Header>

				<Dialog.Close
					class="flex h-5 w-5 items-center justify-center rounded-sm opacity-70 transition-opacity hover:opacity-100"
					onclick={() => {
						confirmModals.statusPool.open = false;
					}}
				>
					<Icon icon="material-symbols:close-rounded" class="h-5 w-5" />
				</Dialog.Close>
			</div>

			<div class="space-y-4 px-4 py-3">
				{#if sPool}
					{#if sPool.status && sPool.status.length > 0}
						<div
							class="rounded-md border border-yellow-200 bg-yellow-50 p-4 text-yellow-800 dark:border-yellow-800 dark:bg-yellow-950 dark:text-yellow-200"
						>
							<div class="flex gap-3">
								<Icon icon="mdi:alert-circle" class="mt-0.5 h-5 w-5 flex-shrink-0" />
								<div>
									<p class="font-medium">{sPool.status}</p>
									{#if sPool.action && sPool.action.length > 0}
										<p class="mt-2 text-sm">{sPool.action}</p>
									{/if}
								</div>
							</div>
						</div>
					{/if}

					<div class="space-y-4 overflow-hidden rounded-md">
						<div class="border">
							<div class="bg-muted flex items-center gap-2 px-4 py-2">
								<Icon icon="mdi:magnify" class="text-primary h-5 w-5" />
								<span class="font-semibold">Scan Activity</span>
							</div>
							<div class="p-4">
								{#if sPool.scan && sPool.scan.length > 0}
									{#if sPool.scan.includes('in progress') || sPool.scan.includes('resilver in progress')}
										{@const progressMatch = sPool.scan.match(/(\d+\.\d+)%/)}
										{@const progress = progressMatch ? parseFloat(progressMatch[1]) : 0}
										{@const isResilver = sPool.scan.includes('resilver')}

										<div class="text-muted-foreground text-sm">
											{capitalizeFirstLetter(sPool.scan)}
										</div>
										<div class="bg-secondary mt-3 h-2.5 w-full overflow-hidden rounded-full">
											<div
												class="h-full rounded-full {isResilver ? 'bg-blue-500' : 'bg-primary'}"
												style="width: {progress}%"
											></div>
										</div>
									{:else}
										<div class="text-muted-foreground text-sm">
											{capitalizeFirstLetter(sPool.scan)}
										</div>
									{/if}
								{:else}
									<div class="text-muted-foreground flex items-center gap-2 py-1">
										<Icon icon="material-symbols:info" class="h-4 w-4" />
										<span>No recent scan activity</span>
									</div>
								{/if}
							</div>
						</div>

						<div class="border">
							<div class="bg-muted flex items-center gap-2 px-4 py-2">
								<Icon icon="tabler:topology-bus" class="text-primary h-5 w-5" />
								<span class="font-semibold">Device Topology</span>
							</div>
							<div class="h-full max-h-28 overflow-auto p-4 md:max-h-44 xl:max-h-72">
								{#if sPool.devices && sPool.devices.length > 0}
									<div class="space-y-3">
										{#each sPool.devices as device, index (device.name)}
											{@render deviceTreeNode(
												{ ...device, __isLast: index === device.children.length - 1 },
												false
											)}
										{/each}
									</div>

									{#if sPoolSpares && sPoolSpares.length > 0}
										<div class="device-tree relative mt-2">
											<div
												class="bg-background relative flex items-center gap-2 rounded-md border p-1.5"
											>
												<div
													class="bg-secondary absolute -left-6 bottom-0 top-0 w-0.5"
													style="height: calc(100% + 0.7rem);"
												></div>
												<Icon icon="tabler:replace" />
												<span class="font-medium">Spares</span>
											</div>

											<div class="ml-5 mt-2 space-y-2 pl-4">
												{#each sPoolSpares as spare, index (spare.name)}
													<div class="relative">
														<div
															class="bg-secondary h-0.5 w-6"
															style="position: absolute; left: -23px; top: 18px"
														></div>
														<div class="device-tree relative">
															<div
																class="bg-background relative flex items-center gap-2 rounded-md border p-1.5"
															>
																<div
																	class="bg-secondary absolute -left-6 bottom-0 top-0 w-0.5"
																	style="height: 18px;"
																></div>
																<div
																	class="h-2 w-2 rounded-full"
																	class:bg-green-500={spare.health === 'AVAIL'}
																	class:bg-yellow-400={spare.health !== 'AVAIL'}
																></div>

																<span>{spare.name}</span>
															</div>
														</div>
													</div>
												{/each}
											</div>
										</div>
									{/if}
								{:else}
									<div class="text-muted-foreground flex items-center gap-2 py-2">
										<Icon icon="material-symbols:info" class="h-4 w-4" />
										<span>No devices found</span>
									</div>
								{/if}
							</div>
						</div>

						<div class="border">
							<div class="bg-muted flex items-center gap-2 px-4 py-2">
								<Icon icon="mdi:alert" class="text-primary h-5 w-5" />
								<span class="font-semibold">Error Status</span>
							</div>
							<div class="p-4">
								<div
									class="flex items-center gap-2 rounded-md border p-2 {sPool.errors.includes(
										'No known data errors'
									)
										? 'border-green-200 bg-green-50 text-green-800 dark:border-green-800 dark:bg-green-950 dark:text-green-200'
										: 'border-red-200 bg-red-50 text-red-800 dark:border-red-800 dark:bg-red-950 dark:text-red-200'}"
								>
									<Icon
										icon={sPool.errors.includes('No known data errors')
											? 'mdi:check-circle'
											: 'mdi:alert-circle'}
										class="h-5 w-5"
									/>
									<span>{sPool.errors}</span>
								</div>
							</div>
						</div>
					</div>
				{/if}
			</div>

			<Dialog.Footer class="px-4 py-3">
				<Button
					variant="default"
					class="h-8 bg-blue-700 text-white hover:bg-blue-600"
					onclick={() => {
						confirmModals.statusPool.open = false;
					}}>Close</Button
				>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/if}

{#if confirmModals.active == 'replaceDevice'}
	<AlertDialog.Root
		bind:open={confirmModals.replaceDevice.open}
		closeOnOutsideClick={false}
		closeOnEscape={false}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title
					>Replace {confirmModals.replaceDevice.data.old} in {confirmModals.replaceDevice.data
						.pool}</AlertDialog.Title
				>
			</AlertDialog.Header>

			<div class="space-y-1 py-1">
				<div>
					<Select.Root
						selected={{
							label:
								useableDisks.find((d) => d.device === confirmModals.replaceDevice.data?.new)
									?.device ||
								useablePartitions.find((d) => d.name === confirmModals.replaceDevice.data?.new)
									?.name,
							value: confirmModals.replaceDevice.data?.new
						}}
						onSelectedChange={(value) => {
							confirmModals.replaceDevice.data = {
								...confirmModals.replaceDevice.data,
								new: value?.value as string
							};
						}}
					>
						<Select.Trigger class="w-full">
							<Select.Value placeholder="Select replacement device" />
						</Select.Trigger>
						<Select.Content class="max-h-36 overflow-y-auto">
							<Select.Group>
								{#each useableDisks as disk}
									<Select.Item value={disk.device} label={disk.device}>
										{disk.device}
									</Select.Item>
								{/each}

								{#each useablePartitions as partition}
									<Select.Item value={partition.name} label={partition.name}>
										{partition.name}
									</Select.Item>
								{/each}
							</Select.Group>
						</Select.Content>
					</Select.Root>
				</div>
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel
					onclick={() => {
						confirmModals.replaceDevice.open = false;
					}}>Cancel</AlertDialog.Cancel
				>
				<AlertDialog.Action
					disabled={!confirmModals.replaceDevice.data?.new}
					onclick={() => {
						confirmAction();
					}}
				>
					Replace
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
