import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { wuchale } from 'wuchale';

export default defineConfig({
	server: {
		allowedHosts: ['sylve.lan']
	},
	plugins: [wuchale(), sveltekit()],
	optimizeDeps: {
		esbuildOptions: {
			target: 'esnext'
		},
		exclude: ['xterm', 'Xterm.svelte', '@battlefieldduck/xterm-svelte']
	},
	build: {
		target: 'esnext'
	}
});
