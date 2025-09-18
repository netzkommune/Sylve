<script lang="ts">
	import '@fontsource/noto-sans';
	import '@fontsource/noto-sans/700.css';

	import { goto } from '$app/navigation';
	import { isClusterTokenValid, isTokenValid, login } from '$lib/api/auth';
	import Login from '$lib/components/custom/Login.svelte';
	import Throbber from '$lib/components/custom/Throbber.svelte';
	import Shell from '$lib/components/skeleton/Shell.svelte';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import { store as token } from '$lib/stores/auth';
	import { hostname, language } from '$lib/stores/basic';
	import '$lib/utils/i18n';
	import { preloadIcons } from '$lib/utils/icons';
	import { addTabulatorFilters } from '$lib/utils/table';
	import { QueryClient, QueryClientProvider } from '@sveltestack/svelte-query';
	import { ModeWatcher } from 'mode-watcher';
	import { onMount, tick } from 'svelte';
	import { loadLocale } from 'wuchale/run-client';

	import type { Locales } from '$lib/types/common';
	import { sleep } from '$lib/utils';
	import '../app.css';

	$effect.pre(() => {
		loadLocale($language as Locales);
	});

	const queryClient = new QueryClient();
	let { children } = $props();
	let isLoggedIn = $state(false);
	let loading = $state({
		throbber: true,
		login: false
	});

	$effect(() => {
		if (isLoggedIn && $hostname) {
			const path = window.location.pathname;
			if (path === '/') {
				goto('/datacenter/summary', { replaceState: true });
			}
		}
	});

	onMount(async () => {
		addTabulatorFilters();
		const faviconEl = document.getElementById('favicon');
		if (faviconEl) {
			const darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
			if (darkMode) {
				faviconEl.setAttribute('href', '/logo/white.svg');
			} else {
				faviconEl.setAttribute('href', '/logo/black.svg');
			}
		}

		if ($token) {
			try {
				if ((await isTokenValid()) && (await isClusterTokenValid())) {
					isLoggedIn = true;
				} else {
					$token = '';
				}
			} catch (error) {
				console.error('Token validation error:', error);
				$token = '';
			}
		}

		await preloadIcons();
		await sleep(1000);
		loading.throbber = false;
		await tick();
	});

	async function handleLogin(
		username: string,
		password: string,
		type: string,
		language: string,
		remember: boolean
	) {
		let isError = false;
		loading.login = true;

		await sleep(500);

		try {
			loadLocale(language as Locales);
			if (await login(username, password, type, remember, language)) {
				isLoggedIn = true;
				loading.login = false;
				const path = window.location.pathname;

				if (path === '/') {
					await goto('/datacenter/summary', { replaceState: true });
				}
			} else {
				isError = true;
				isLoggedIn = false;
				loading.login = false;
			}
		} catch (error) {
			isError = true;
			isLoggedIn = false;
			loading.login = false;
		} finally {
			if (!isError) {
				await sleep(2500);
			}
		}

		loading.login = false;
		loading.throbber = false;
		return;
	}
</script>

<svelte:head>
	<!-- @wc-ignore -->
	<title>Sylve</title>
</svelte:head>

<Toaster />
<ModeWatcher />

{#if loading.throbber}
	<Throbber />
{:else if isLoggedIn && $hostname}
	<QueryClientProvider client={queryClient}>
		<Shell>
			{@render children()}
		</Shell>
	</QueryClientProvider>
{:else}
	<Login onLogin={handleLogin} loading={loading.login} />
{/if}
