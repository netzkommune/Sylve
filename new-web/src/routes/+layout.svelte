<script lang="ts">
	import '@fontsource/noto-sans';
	import '@fontsource/noto-sans/700.css';

	import { goto, replaceState } from '$app/navigation';
	import { page } from '$app/state';
	import { isTokenValid, login } from '$lib/api/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import '../app.css';

	let { children } = $props();
</script>

<Toaster />

{@render children()}
