// @ts-check
import { defineConfig } from 'wuchale';
import { adapter as svelte } from '@wuchale/svelte';

export default defineConfig({
	locales: {
		mal: { name: 'Malayalam' },
		'cn-simplified': { name: '简体中文' }
	},
    adapters: {
        main: svelte({
            catalog: 'src/lib/locales/{locale}',
        })
    }
	// hmr: false
});
