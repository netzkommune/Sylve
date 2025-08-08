<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		deleteJail,
		getJailLogs,
		getJails,
		getJailStates,
		jailAction,
		updateDescription
	} from '$lib/api/jail/jail';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import LoadingDialog from '$lib/components/custom/Dialog/Loading.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import CustomValueInput from '$lib/components/ui/custom-input/value.svelte';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { hostname } from '$lib/stores/basic';
	import type { Jail, JailState } from '$lib/types/jail/jail';
	import { sleep } from '$lib/utils';
	import { updateCache } from '$lib/utils/http';
	import { dateToAgo, secondsToHoursAgo } from '$lib/utils/time';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import humanFormat from 'human-format';
	import { toast } from 'svelte-sonner';

	interface Data {
		jails: Jail[];
		jailStates: JailState[];
		jail: Jail;
	}

	let { data }: { data: Data } = $props();
	const ctId = page.url.pathname.split('/')[3];

	let modalState = $state({
		isDeleteOpen: false,
		title: '',
		loading: {
			open: false,
			title: '',
			description: '',
			iconColor: ''
		}
	});

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
		},
		{
			queryKey: ['jail-states'],
			queryFn: async () => {
				return await getJailStates();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.jailStates,
			onSuccess: (data: JailState[]) => {
				updateCache('jail-states', data);
			}
		},
		{
			queryKey: [`jail-${ctId}-start-logs`],
			queryFn: async () => {
				return await getJailLogs(data.jail.id, true);
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: { logs: '' },
			onSuccess: (data: any) => {
				updateCache(`jail-${ctId}-start-logs`, data);
			}
		},
		{
			queryKey: [`jail-${ctId}-stop-logs`],
			queryFn: async () => {
				return await getJailLogs(data.jail.id, false);
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: { logs: '' },
			onSuccess: (data: any) => {
				updateCache(`jail-${ctId}-stop-logs`, data);
			}
		}
	]);

	let jail: Jail = $derived(
		($results[0].data as Jail[]).find((jail: Jail) => jail.ctId === parseInt(ctId)) || ({} as Jail)
	);

	let jState: JailState = $derived(
		($results[1].data as JailState[]).find((state: JailState) => state.ctId === parseInt(ctId)) ||
			({} as JailState)
	);

	let startLogs = $derived($results[2].data);
	let stopLogs = $derived($results[3].data);
	let jailDesc = $state(jail.description || '');

	$inspect(startLogs, 'startLogs');

	$effect(() => {
		if (jailDesc) {
			updateDescription(jail.id, jailDesc);
		} else {
			updateDescription(jail.id, '');
		}
	});

	let udTime = $derived.by(() => {
		if (jState.state === 'ACTIVE') {
			if (jail.startedAt) {
				return `Started ${dateToAgo(jail.startedAt)}`;
			}
		} else if (jState.state === 'INACTIVE' || jState.state === 'UNKNOWN') {
			if (jail.stoppedAt) {
				return `Stopped ${dateToAgo(jail.stoppedAt)}`;
			}
		}
		return '';
	});

	async function handleDelete() {
		modalState.isDeleteOpen = false;
		modalState.loading.open = true;
		modalState.loading.title = 'Deleting Jail';
		modalState.loading.description = `Please wait while Jail <b>${jail.name} (${jail.ctId})</b> is being deleted`;

		await sleep(1000);
		modalState.loading.open = false;

		await sleep(1000);
		const result = await deleteJail(jail.ctId);
		modalState.loading.open = false;

		console.log(result, 'result');

		if (result.status === 'error') {
			toast.error('Error deleting jail', {
				duration: 5000,
				position: 'bottom-center'
			});
		} else if (result.status === 'success') {
			goto(`/${$hostname}/summary`);
			toast.success('Jail deleted', {
				duration: 5000,
				position: 'bottom-center'
			});
		}
	}

	async function handleStop() {
		modalState.loading.open = true;
		modalState.loading.title = 'Stopping Jail';
		modalState.loading.description = `Please wait while Jail <b>${jail.name} (${jail.ctId})</b> is being stopped`;
		modalState.loading.iconColor = 'text-red-500';

		await sleep(1000);
		await jailAction(jail.ctId, 'stop');

		modalState.loading.open = false;
	}

	async function handleStart() {
		modalState.loading.open = true;
		modalState.loading.title = 'Starting Jail';
		modalState.loading.description = `Please wait while Jail <b>${jail.name} (${jail.ctId})</b> is being started`;
		modalState.loading.iconColor = 'text-green-500';

		await sleep(1000);
		await jailAction(jail.ctId, 'start');

		modalState.loading.open = false;
	}
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-4">
		{#if jState?.state === 'ACTIVE'}
			<Button
				onclick={handleStop}
				size="sm"
				class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
			>
				<Icon icon="mdi:stop" class="mr-1 h-4 w-4" />
				{'Stop'}
			</Button>
		{:else}
			<div class="flex items-center gap-2">
				<Button
					onclick={handleStart}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
				>
					<Icon icon="mdi:play" class="mr-1 h-4 w-4" />
					{'Start'}
				</Button>

				<Button
					onclick={handleDelete}
					size="sm"
					class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
				>
					<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
					{'Delete'}
				</Button>
			</div>
		{/if}
	</div>

	<div class="min-h-0 flex-1">
		<ScrollArea orientation="both" class="h-full">
			<div class="grid grid-cols-1 gap-4 p-4 lg:grid-cols-2">
				<Card.Root class="w-full gap-0 p-4">
					<Card.Header class="p-0">
						<Card.Description class="text-md  font-normal text-blue-600 dark:text-blue-500">
							{`${jail?.name} ${udTime ? `(${udTime})` : ''}`}
						</Card.Description>
					</Card.Header>
					<Card.Content class="mt-3 p-0">
						<div class="flex items-start">
							<div class="flex items-center">
								<Icon icon="fluent:status-12-filled" class="mr-1 h-5 w-5" />
								{'Status'}
							</div>
							<div class="ml-auto">
								{jState.state === 'ACTIVE'
									? 'Running'
									: jState.state === 'INACTIVE'
										? 'Stopped'
										: jState.state}
							</div>
						</div>

						<div class="mt-2">
							<div class="flex w-full justify-between pb-1">
								<p class="inline-flex items-center">
									<Icon icon="solar:cpu-bold" class="mr-1 h-5 w-5" />
									{'CPU Usage'}
								</p>
								<p class="ml-auto">
									{#if jState.state === 'ACTIVE'}
										{`${Math.min((jState.pcpu * 100) / (100 * jail.cores), 100).toFixed(2)}% of ${jail.cores} Core(s)`}
									{:else}
										{`0% of ${jail.cores} Core(s)`}
									{/if}
								</p>
							</div>

							{#if jState.state === 'ACTIVE'}
								<Progress
									value={(jState.pcpu / (100 * jail.cores)) * 100}
									max={100}
									class="ml-auto h-2"
								/>
							{:else}
								<Progress value={0} max={100} class="ml-auto h-2" />
							{/if}
						</div>

						<div class="mt-2">
							<div class="flex w-full justify-between pb-1">
								<p class="inline-flex items-center">
									<Icon icon="ph:memory" class="mr-1 h-5 w-5" />
									{'RAM Usage'}
								</p>
								<p class="ml-auto">
									{#if jState.state === 'ACTIVE'}
										{`${((jState.memory / (jail.memory || 1)) * 100).toFixed(2)}% of ${humanFormat(jail.memory || 0)}`}
									{:else}
										{`0% of ${humanFormat(jail.memory || 0)}`}
									{/if}
								</p>
							</div>

							{#if jState.state === 'ACTIVE'}
								<Progress
									value={(jState.memory / (jail.memory || 1)) * 100}
									max={100}
									class="ml-auto h-2"
								/>
							{:else}
								<Progress value={0} max={100} class="ml-auto h-2" />
							{/if}
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="w-full gap-0 p-4">
					<Card.Header class="p-0">
						<Card.Description class="text-md font-normal text-blue-600 dark:text-blue-500">
							Description
						</Card.Description>
					</Card.Header>
					<Card.Content class="mt-3 p-0">
						<CustomValueInput
							label={''}
							placeholder="Notes"
							bind:value={jailDesc}
							classes=""
							textAreaClasses="!h-32"
							type="textarea"
						/>
					</Card.Content>
				</Card.Root>
			</div>
		</ScrollArea>
	</div>
</div>

<AlertDialog
	open={modalState.isDeleteOpen}
	names={{ parent: 'Jail', element: modalState?.title || '' }}
	actions={{
		onConfirm: async () => {
			handleDelete();
		},
		onCancel: () => {
			modalState.isDeleteOpen = false;
		}
	}}
></AlertDialog>

<LoadingDialog
	bind:open={modalState.loading.open}
	title={modalState.loading.title}
	description={modalState.loading.description}
	iconColor={modalState.loading.iconColor}
	logs={modalState.loading.title.toLowerCase().includes('deleting')
		? ''
		: startLogs?.logs || stopLogs?.logs || ''}
/>
