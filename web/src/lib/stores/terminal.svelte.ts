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
	id: string;
	isOpen: boolean;
	isMinimized: boolean;
	position: { x: number; y: number };
	title: string;
	tabs: Tab[];
	activeTabId: string;
	xterm: Xterm | null;
}

export let terminalStore: {
	value: Terminal;
} = $state({
	value: {
		id: '',
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
		isMinimized: false
	};

	const id = nanoid(6);
	const tabId = nanoid(6);
	const xOffset = 30;
	const yOffset = 30;

	const newTerminal: Terminal = {
		id,
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
