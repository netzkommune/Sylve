<script lang="ts">
	import { page } from '$app/state';
	import { getVMs } from '$lib/api/vm/vm';
	import type { VM, VMDomain } from '$lib/types/vm/vm';
	import { updateCache } from '$lib/utils/http';
	import { useQueries } from '@sveltestack/svelte-query';

	interface Data {
		vms: VM[];
		domain: VMDomain;
	}

	let { data }: { data: Data } = $props();
	const vmId = page.url.pathname.split('/')[2];

	const results = useQueries([
		{
			queryKey: ['vm-list'],
			queryFn: async () => {
				return await getVMs();
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.vms,
			onSuccess: (data: VM[]) => {
				updateCache('vm-list', data);
			}
		},
		{
			queryKey: [`vm-domain-${vmId}`],
			queryFn: async () => {
				return data.domain;
			},
			refetchInterval: 1000,
			keepPreviousData: true,
			initialData: data.domain,
			onSuccess: (data: VMDomain) => {
				updateCache(`vm-domain-${vmId}`, data);
			}
		}
	]);

	$effect(() => {
		console.log('Data received:', data);
	});
</script>

<div class="flex h-full w-full flex-col">
	<div class="flex h-10 w-full items-center gap-2 border p-2"></div>
</div>
