import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import wuchale from 'wuchale';

export default defineConfig({
	plugins: [
		wuchale({
			localesDir: 'src/lib/locales',
			otherLocales: ['mal']
		}),
		tailwindcss(),
		sveltekit()
	],
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
