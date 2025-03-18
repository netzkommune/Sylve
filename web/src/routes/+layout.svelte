<script lang="ts">
	import '@fontsource/noto-sans';
	import '@fontsource/noto-sans/700.css';

	import { goto } from '$app/navigation';
	import { navigating } from '$app/state';
	import { isTokenValid, login } from '$lib/api/auth';
	import Login from '$lib/components/custom/Login.svelte';
	import Throbber from '$lib/components/custom/Throbber.svelte';
	import Shell from '$lib/components/Skeleton/Shell.svelte';
	import { store as token } from '$lib/stores/auth';
	import { hostname } from '$lib/stores/basic';
	import '$lib/utils/i18n';
	import { QueryClient, QueryClientProvider } from '@sveltestack/svelte-query';
	import { ModeWatcher } from 'mode-watcher';
	import { onMount, tick } from 'svelte';
	import { Toaster } from 'svelte-french-toast';
	import '../app.css';

	const queryClient = new QueryClient();
	let { children } = $props();
	let isLoggedIn = $state(false);
	let isLoading = $state(true);

	$effect(() => {
		if (isLoggedIn && $hostname && !navigating) {
			const path = window.location.pathname;
			if (path === '/' || !path.startsWith(`/${$hostname}`)) {
				goto(`/${$hostname}/summary`, { replaceState: true });
			}
		}
	});

	onMount(async () => {
		if ($token) {
			await sleep(4000);
			try {
				if (await isTokenValid()) {
					isLoggedIn = true;
				} else {
					$token = '';
				}
			} catch (error) {
				console.error('Token validation error:', error);
				$token = '';
			}
		}

		isLoading = false;
		await tick();
	});

	function sleep(ms: number) {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	async function handleLogin(
		username: string,
		password: string,
		type: string,
		language: string,
		remember: boolean
	) {
		isLoading = true;
		try {
			if (await login(username, password, type, remember, language)) {
				isLoggedIn = true;
				const path = window.location.pathname;

				if (path === '/' || !path.startsWith(`/${$hostname}`)) {
					await goto(`/${$hostname}/summary`, { replaceState: true });
				}
			} else {
				alert('Login failed');
			}
		} catch (error) {
			console.error('Login error:', error);
			alert('Login failed: An error occurred');
		} finally {
			isLoading = false;
		}
		return;
	}
</script>

<Toaster />
<ModeWatcher />

{#if isLoading}
	<Throbber />
{:else if isLoggedIn && $hostname}
	<QueryClientProvider client={queryClient}>
		<Shell>
			{@render children()}
		</Shell>
	</QueryClientProvider>
{:else}
	<Login onLogin={handleLogin} />
{/if}
