<script lang="ts">
	import { store } from '$lib/stores/auth';
	import { getDefaultTitle, terminalStore } from '$lib/stores/terminal.svelte';
	import type {
		ITerminalInitOnlyOptions,
		ITerminalOptions,
		Terminal
	} from '@battlefieldduck/xterm-svelte';
	import { Xterm, XtermAddon } from '@battlefieldduck/xterm-svelte';
	import Icon from '@iconify/svelte';
	import adze from 'adze';
	import { nanoid } from 'nanoid';
	import { fly, scale } from 'svelte/transition';

	let wsConnections: Record<string, WebSocket> = $state({});
	let terminalInstances: Record<string, Terminal> = $state({});
	let currentActiveTabId: string = $state('');
	let terminalHistories: Record<string, string> = $state({});

	let options: ITerminalOptions & ITerminalInitOnlyOptions = {
		cursorBlink: true
	};

	function saveTerminalHistory(tabId: string, serializeAddon: any) {
		if (terminalInstances[tabId]) {
			const history = serializeAddon.serialize();
			terminalHistories[tabId] = history;
		}
	}

	async function onLoad(terminal: Terminal) {
		const tabId = terminal.element?.parentElement?.getAttribute('data-id');
		if (!tabId) return;

		terminalInstances[tabId] = terminal;
		currentActiveTabId = tabId;

		const fitAddon = new (await XtermAddon.FitAddon()).FitAddon();
		const serializeAddon = new (await XtermAddon.SerializeAddon()).SerializeAddon();
		terminal.loadAddon(fitAddon);
		terminal.loadAddon(serializeAddon);
		fitAddon.fit();

		if (terminalHistories[tabId]) {
			terminal.write(terminalHistories[tabId]);
		}

		if (!wsConnections[tabId]) {
			try {
				const newWs = new WebSocket(`/api/info/terminal`, ['Bearer', $store]);
				newWs.binaryType = 'arraybuffer';
				wsConnections[tabId] = newWs;

				newWs.onopen = () => {
					adze.info(`Terminal WebSocket connected for tab ${tabId}`);
					const dimensions = fitAddon.proposeDimensions();
					newWs.send(
						new TextEncoder().encode(
							'\x01' + JSON.stringify({ rows: dimensions?.rows, cols: dimensions?.cols })
						)
					);
				};

				newWs.onmessage = (event) => {
					if (event.data instanceof ArrayBuffer) {
						const terminal = terminalInstances[tabId];
						if (terminal) {
							terminal.write(new Uint8Array(event.data));
							saveTerminalHistory(tabId, serializeAddon);
						}
					}
				};

				newWs.onclose = () => {
					adze.info(`Terminal WebSocket disconnected for tab ${tabId}`);
					const terminal = terminalInstances[tabId];
					if (terminal) {
						terminal.write('\x1b[31mDisconnected from server.\x1b[0m\r\n');
					}
					delete wsConnections[tabId];
				};
			} catch (error) {
				adze.error('Failed to connect to terminal WebSocket', { error });
			}
		}
	}

	function onData(data: string) {
		if (currentActiveTabId && wsConnections[currentActiveTabId]) {
			wsConnections[currentActiveTabId].send(new TextEncoder().encode('\x00' + data));
		}
	}

	function addTab() {
		terminalStore.value.tabs = [
			...terminalStore.value.tabs,
			{
				id: nanoid(6),
				title: getDefaultTitle()
			}
		];
	}

	function removeTab(terminalId: string, tabId: string) {
		console.log(terminalId, tabId);
		if (wsConnections[tabId]) {
			wsConnections[tabId].close();
			delete wsConnections[tabId];
			delete terminalInstances[tabId];
		}

		terminalStore.value.tabs = terminalStore.value.tabs.filter((tab) => tab.id !== tabId);

		if (terminalStore.value.activeTabId === tabId) {
			if (terminalStore.value.tabs.length > 0) {
				const newActiveTab = terminalStore.value.tabs[terminalStore.value.tabs.length - 1];
				terminalStore.value.activeTabId = newActiveTab.id;
			} else {
				terminalStore.value.isOpen = false;
			}
		}
	}

	async function setActiveTab(tabId: string) {
		currentActiveTabId = tabId;
		terminalStore.value.activeTabId = tabId;

		setTimeout(async () => {
			const termInstance = terminalInstances[tabId];
			if (termInstance) {
				const fitAddon = new (await XtermAddon.FitAddon()).FitAddon();
				termInstance.loadAddon(fitAddon);
				fitAddon.fit();
				termInstance.focus();
			}
		}, 50);
	}

	function closeDialog() {
		terminalStore.value.tabs.forEach((tab) => {
			if (wsConnections[tab.id]) {
				wsConnections[tab.id].close();
				delete wsConnections[tab.id];
				delete terminalInstances[tab.id];
			}
		});

		terminalStore.value.isOpen = false;
	}

	function minimizeDialog() {
		terminalStore.value.isMinimized = true;
	}

	function restoreDialog() {
		terminalStore.value.isMinimized = false;
		currentActiveTabId = terminalStore.value.activeTabId;

		const activeTerminal = terminalInstances[terminalStore.value.activeTabId];
		if (activeTerminal) {
			activeTerminal.focus();
		}
	}
</script>

{#if terminalStore.value.isOpen && !terminalStore.value.isMinimized}
	<div
		class="fixed inset-0 z-[9998] bg-black/30 backdrop-blur-sm transition-all duration-300"
	></div>
{/if}

{#if terminalStore.value.isOpen}
	{#if terminalStore.value.isMinimized}
		<div
			class="bg-muted fixed bottom-0 z-[9999] flex h-10 w-40 items-center justify-between rounded-t-lg px-3 text-white shadow-lg transition-all duration-300"
			ondblclick={() => restoreDialog()}
			in:fly={{ y: 50, duration: 300 }}
			out:fly={{ y: 50, duration: 300 }}
		>
			<span class="truncate text-sm">{terminalStore.value.title}</span>
			<div class="flex gap-2">
				<button class="text-white hover:text-gray-300" onclick={() => restoreDialog()}>
					<Icon icon="mdi:window-restore" class="h-4 w-4" />
				</button>
				<button class="text-white hover:text-red-300" onclick={() => closeDialog()} title="Close">
					<Icon icon="mdi:close" class="h-4 w-4" />
				</button>
			</div>
		</div>
	{:else}
		<div
			class="fixed inset-0 z-[9999] flex items-center justify-center transition-all duration-300"
			in:scale={{ start: 0.8, duration: 300 }}
			out:scale={{ start: 0.8, duration: 300 }}
		>
			<div
				class="border-muted-foreground bg-muted-foreground/10 relative flex h-[70%] w-[60%] flex-col rounded-lg border shadow-lg"
			>
				<div
					class="border-muted-foreground bg-primary-foreground flex items-center justify-between rounded-t-lg border-b p-1"
				>
					<div class="flex items-center gap-2">
						<button
							class="dark:hover-bg-muted hover:bg-muted-foreground/40 flex h-6 w-6 items-center justify-center rounded"
							onclick={() => addTab()}
							title="Add new tab"
						>
							<Icon icon="mdi:plus" class="h-4 w-4" />
						</button>
						<span>{terminalStore.value.title}</span>
					</div>
					<div class="flex space-x-2">
						<button
							class="rounded-full bg-yellow-400 p-1"
							onclick={() => minimizeDialog()}
							title="Minimize"
						>
							<Icon icon="mdi:window-minimize" class="h-3 w-3 text-gray-800" />
						</button>
						<button class="rounded-full bg-red-500 p-1" onclick={() => closeDialog()} title="Close">
							<Icon icon="mdi:close" class="h-3 w-3 text-gray-800" />
						</button>
					</div>
				</div>

				<div class="border-muted-foreground bg-muted/30 flex overflow-x-auto border-b">
					{#each terminalStore.value.tabs as tab}
						<div
							class="border-muted-foreground flex cursor-pointer items-center border-r px-3 py-1 {tab.id ===
							terminalStore.value.activeTabId
								? 'bg-muted-foreground/40 dark:bg-muted-foreground/40'
								: 'hover:bg-muted/50'}"
							onclick={() => setActiveTab(tab.id)}
						>
							<span class="mr-2 whitespace-nowrap text-sm">{tab.title}</span>
							{#if terminalStore.value.tabs.length > 1}
								<button
									class="hover:text-red-400"
									onclick={() => removeTab(terminalStore.value.id, tab.id)}
								>
									<Icon icon="mdi:close" class="h-3 w-3" />
								</button>
							{/if}
						</div>
					{/each}
				</div>

				<div class="bg-muted relative h-full w-full flex-grow overflow-hidden">
					{#each terminalStore.value.tabs as tab}
						<Xterm
							terminal={terminalInstances[tab.id]}
							{options}
							{onLoad}
							{onData}
							class="h-full w-full"
							hidden={tab.id !== terminalStore.value.activeTabId}
							data-id={tab.id}
							onclick={() => {
								currentActiveTabId = tab.id;
								setTimeout(() => terminalInstances[tab.id]?.focus(), 0);
							}}
						/>
					{/each}
				</div>
			</div>
		</div>
	{/if}
{/if}
