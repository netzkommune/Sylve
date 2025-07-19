<script lang="ts">
	import '@fontsource/noto-sans';
	import '@fontsource/noto-sans/700.css';

	import { goto, pushState, replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { isTokenValid, login } from '$lib/api/auth';
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
	import { setCatalog } from '@wuchale/svelte/runtime.svelte.js';
	import { ModeWatcher } from 'mode-watcher';
	import { onMount, tick } from 'svelte';

	import type { Locales } from '$lib/types/common';
	import { toast } from 'svelte-sonner';
	import '../app.css';

	async function setLocale(locale: Locales) {
		const mod = await import(`$lib/locales/${locale}.svelte.js`);
		if (mod) {
			setCatalog(mod);
		} else {
			console.error('Failed to load locale:', locale);
		}
	}

	$effect.pre(() => {
		setLocale($language as Locales);
	});

	const queryClient = new QueryClient();
	let { children } = $props();
	let isLoggedIn = $state(false);
	let isLoading = $state(true);

	$effect(() => {
		if (isLoggedIn && $hostname) {
			const path = window.location.pathname;
			if (path === '/' || !path.startsWith(`/${$hostname}`)) {
				goto(`/${$hostname}/summary`, { replaceState: true });
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

		await preloadIcons();
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
		let isError = false;

		try {
			setLocale(language as Locales);
			if (await login(username, password, type, remember, language)) {
				isLoading = true;
				isLoggedIn = true;
				const path = window.location.pathname;

				if (path === '/' || !path.startsWith(`/${$hostname}`)) {
					await goto(`/${$hostname}/summary`, { replaceState: true });
				}
			} else {
				isError = true;
				isLoggedIn = false;
			}
		} catch (error) {
			isError = true;
			isLoggedIn = false;
		} finally {
			if (!isError) {
				await sleep(2500);
				isLoading = false;
			}
		}
		return;
	}
</script>

<svelte:head>
	<!-- @wc-ignore -->
	<title>Sylve</title>
</svelte:head>

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
