import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	server: {
		allowedHosts: ['sylve.lan']
	},
	plugins: [sveltekit()],
	optimizeDeps: {
		exclude: ['xterm', 'Xterm.svelte', '@battlefieldduck/xterm-svelte']
	}
});
