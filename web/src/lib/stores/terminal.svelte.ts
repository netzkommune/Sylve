/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { hostname } from '$lib/stores/basic';
import type { Terminal as Xterm } from '@battlefieldduck/xterm-svelte';
import { nanoid } from 'nanoid';
import { get } from 'svelte/store';
import { getUsername } from './auth';
interface Tab {
	id: string;
	title: string;
}
interface Terminal {
	isOpen: boolean;
	isMinimized: boolean;
	position: { x: number; y: number };
	title: string;
	tabs: Tab[];
	activeTabId: string;
	xterm: Xterm | null;
}

export const terminalStore: {
	value: Terminal;
} = $state({
	value: {
		isOpen: false,
		isMinimized: false,
		position: { x: 0, y: 0 },
		title: '',
		tabs: [],
		activeTabId: '',
		xterm: null
	}
});

export function getDefaultTitle() {
	return `${getUsername()}@${get(hostname)}:~`;
}

export function openTerminal() {
	terminalStore.value = {
		...terminalStore.value,
		isMinimized: false,
		isOpen: true
	};

	if (terminalStore.value.tabs.length > 0) {
		return;
	}

	const tabId = nanoid(9);
	const xOffset = 30;
	const yOffset = 30;

	const newTerminal: Terminal = {
		isOpen: true,
		isMinimized: false,
		position: {
			x: window.innerWidth / 2 - 400 + xOffset,
			y: window.innerHeight / 2 - 300 + yOffset
		},
		title: 'Terminal',
		tabs: [
			{
				id: tabId,
				title: getDefaultTitle()
			}
		],

		activeTabId: tabId,
		xterm: null
	};

	terminalStore.value = newTerminal;
}
