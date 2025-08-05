<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { deleteJail, getJails } from '$lib/api/jail/jail';
	import AlertDialog from '$lib/components/custom/Dialog/Alert.svelte';
	import LoadingDialog from '$lib/components/custom/Dialog/Loading.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { hostname } from '$lib/stores/basic';
	import type { Jail } from '$lib/types/jail/jail';
	import { sleep } from '$lib/utils';
	import { updateCache } from '$lib/utils/http';
	import Icon from '@iconify/svelte';
	import { useQueries } from '@sveltestack/svelte-query';
	import { toast } from 'svelte-sonner';

	interface Data {
		jails: Jail[];
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
		}
	]);

	let jail: Jail = $derived(
		($results[0].data as Jail[]).find((jail: Jail) => jail.ctId === parseInt(ctId)) || ({} as Jail)
	);

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
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-4">
		<Button
			onclick={() => {
				modalState.isDeleteOpen = true;
				modalState.title = `${jail.name} (${jail.ctId})`;
			}}
			size="sm"
			class="bg-muted-foreground/40 dark:bg-muted h-6 text-black disabled:!pointer-events-auto disabled:hover:bg-neutral-600 dark:text-white"
		>
			<Icon icon="mdi:delete" class="mr-1 h-4 w-4" />
			{'Delete'}
		</Button>
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
/>
