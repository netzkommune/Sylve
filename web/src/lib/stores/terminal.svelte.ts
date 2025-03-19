import { localStore } from '@layerstack/svelte-stores';
import { addDays } from 'date-fns';
import { nanoid } from 'nanoid';

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
}

export let terminals: {
	value: Terminal[];
} = $state({
	value: []
});

//test
export function openTerminal() {
	const existingTerminal = terminals.value.find((terminal) => terminal.isOpen);

	if (existingTerminal) {
		terminals.value = terminals.value.map((terminal) =>
			terminal.id === existingTerminal.id ? { ...terminal, isMinimized: false } : terminal
		);
		return terminals.value.map((terminal) =>
			terminal.id === existingTerminal.id ? { ...terminal, isMinimized: false } : terminal
		);
	}

	const id = nanoid(6);
	const tabId = nanoid(6);
	const xOffset = terminals.value.length * 30;
	const yOffset = terminals.value.length * 30;

	const newTerminal: Terminal = {
		id,
		isOpen: true,
		isMinimized: false,
		position: {
			x: window.innerWidth / 2 - 400 + xOffset,
			y: window.innerHeight / 2 - 300 + yOffset
		},

		title: `Terminal ${terminals.value.length + 1}`,
		tabs: [
			{
				id: tabId,
				title: `Tab 1`
			}
		],

		activeTabId: tabId
	};

	terminals.value = [...terminals.value, newTerminal];
	return [...terminals.value, newTerminal];
}
