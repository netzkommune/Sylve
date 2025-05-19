/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { getIcon, loadIcon } from '@iconify/svelte';

export const iconCache: Record<string, string> = {};
const icons = [
	'carbon:ibm-cloud-vpc-block-storage-snapshots',
	'material-symbols:files',
	'carbon:volume-block-storage',
	'bi:hdd-stack-fill',
	'mdi:harddisk',
	'bi:nvme',
	'icon-park-outline:ssd',
	'ant-design:partition-outlined',
	'mdi:magnet',
	'mdi:download',
	'mdi:file',
	'mingcute:file-fill',
	'line-md:downloading-loop',
	'lets-icons:check-fill',
	'material-symbols-light:private-connectivity-outline',
	'material-symbols:public',
	'clarity:network-switch-line',
	'ic:baseline-loop',
	'mdi:ethernet',
	'eos-icons:three-dots-loading'
];

export async function preloadIcons() {
	for (const icon of icons) {
		await loadIcon(icon);
		const i = getIcon(icon);
		if (i) {
			const { body, width, height, left, top } = i;
			const viewBox = `${left} ${top} ${width} ${height}`;
			iconCache[icon] = `
				<svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 inline" viewBox="${viewBox}" width="${width}" height="${height}">
					${body}
				</svg>
			`.trim();
		}
	}
}
