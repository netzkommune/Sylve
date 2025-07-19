// @ts-check
import { adapter as svelte } from '@wuchale/svelte';
import { defineConfig } from 'wuchale';

export default defineConfig({
	locales: {
		mal: { name: 'Malayalam' },
		'cn-simplified': { name: '简体中文' },
		ar: { name: 'Arabic' },
		ru: { name: 'Russian' },
		tu: { name: 'Türkçe' }
	},
	adapters: {
		main: svelte({
			catalog: 'src/lib/locales/{locale}'
		})
	}
	// hmr: false
});
