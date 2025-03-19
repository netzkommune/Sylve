<script lang="ts">
	import { store } from '$lib/stores/auth';
	import { terminals } from '$lib/stores/terminal.svelte';
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

	let ws: WebSocket | null = $state(null);
	let xtremTerminal: Terminal | undefined = $state();
	let options: ITerminalOptions & ITerminalInitOnlyOptions = {
		cursorBlink: true
	};

	async function onLoad(terminal: Terminal) {
		const fitAddon = new (await XtermAddon.FitAddon()).FitAddon();
		terminal.loadAddon(fitAddon);
		fitAddon.fit();

		try {
			ws = new WebSocket(`/api/info/terminal`, ['Bearer', $store]);
			ws.binaryType = 'arraybuffer';
		} catch (error) {
			adze.error('Failed to connect to terminal WebSocket', { error });
		}

		if (!ws) {
			adze.error('WebSocket connection is not available');
			return;
		}

		ws.onopen = () => {
			adze.info('Terminal WebSocket connected');
		};

		ws.onmessage = (event) => {
			if (event.data instanceof ArrayBuffer) {
				terminal.write(new Uint8Array(event.data));
			} else {
				console.warn('Unexpected message:', event.data);
			}
		};

		ws.onclose = () => {
			adze.info('Terminal WebSocket disconnected');
			terminal.write('\x1b[31mDisconnected from server.\x1b[0m\r\n');
		};
	}

	function onData(data: string) {
		ws?.send(new TextEncoder().encode('\x00' + data));
	}

	function addTab(terminalId: string) {
		terminals.value = terminals.value.map((t) => {
			if (t.id === terminalId) {
				const newTabId = nanoid(6);
				const newTabNumber = t.tabs.length + 1;
				return {
					...t,
					tabs: [
						...t.tabs,
						{
							id: newTabId,
							title: `Tab ${newTabNumber}`
						}
					],
					activeTabId: newTabId
				};
			}
			return t;
		});
	}

	function removeTab(terminalId: string, tabId: string) {
		terminals.value = terminals.value.map((t) => {
			if (t.id === terminalId) {
				if (t.tabs.length <= 1) return t;

				const newTabs = t.tabs.filter((tab) => tab.id !== tabId);
				let newActiveTabId = t.activeTabId;

				if (t.activeTabId === tabId) {
					newActiveTabId = newTabs[0].id;
				}

				return {
					...t,
					tabs: newTabs,
					activeTabId: newActiveTabId
				};
			}
			return t;
		});
	}

	function setActiveTab(terminalId: string, tabId: string) {
		terminals.value = terminals.value.map((t) => {
			if (t.id === terminalId) {
				return { ...t, activeTabId: tabId };
			}
			return t;
		});
	}

	function closeDialog(id: string) {
		terminals.value = terminals.value.filter((t) => t.id !== id);
	}

	function minimizeDialog(id: string) {
		terminals.value = terminals.value.map((t) => {
			if (t.id === id) {
				return { ...t, isMinimized: true };
			}
			return t;
		});
	}

	function restoreDialog(id: string) {
		terminals.value = terminals.value.map((t) => {
			if (t.id === id) {
				return { ...t, isMinimized: false };
			}
			return t;
		});
	}

	function getMinimizedPosition(index: number) {
		const width = 160;
		const gap = 10;
		return (width + gap) * index + 20;
	}

	$effect(() => {
		console.log('Terminals:', terminals.value);
	});
</script>

{#if terminals.value.some((t) => t.isOpen && !t.isMinimized)}
	<div
		class="fixed inset-0 z-[9998] bg-black/30 backdrop-blur-sm transition-all duration-300"
	></div>
{/if}

{#each terminals.value as terminal, i (terminal.id)}
	{#if terminal.isOpen}
		{#if terminal.isMinimized}
			<div
				class="bg-muted fixed bottom-0 z-[9999] flex h-10 w-40 items-center justify-between rounded-t-lg px-3 text-white shadow-lg transition-all duration-300"
				style="left: {getMinimizedPosition(i)}px"
				ondblclick={() => restoreDialog(terminal.id)}
				in:fly={{ y: 50, duration: 300 }}
				out:fly={{ y: 50, duration: 300 }}
			>
				<span class="truncate text-sm">{terminal.title}</span>
				<div class="flex gap-2">
					<button class="text-white hover:text-gray-300" onclick={() => restoreDialog(terminal.id)}>
						<Icon icon="mdi:window-restore" class="h-4 w-4" />
					</button>
					<button
						class="text-white hover:text-red-300"
						onclick={() => closeDialog(terminal.id)}
						title="Close"
					>
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
								onclick={() => addTab(terminal.id)}
								title="Add new tab"
							>
								<Icon icon="mdi:plus" class="h-4 w-4" />
							</button>
							<span>{terminal.title}</span>
						</div>
						<div class="flex space-x-2">
							<button
								class="rounded-full bg-yellow-400 p-1"
								onclick={() => minimizeDialog(terminal.id)}
								title="Minimize"
							>
								<Icon icon="mdi:window-minimize" class="h-3 w-3 text-gray-800" />
							</button>
							<button
								class="rounded-full bg-red-500 p-1"
								onclick={() => closeDialog(terminal.id)}
								title="Close"
							>
								<Icon icon="mdi:close" class="h-3 w-3 text-gray-800" />
							</button>
						</div>
					</div>

					<!-- Tab Bar -->
					<div class="border-muted-foreground bg-muted/30 flex overflow-x-auto border-b">
						{#each terminal.tabs as tab}
							<div
								class="border-muted-foreground flex cursor-pointer items-center border-r px-3 py-1 {tab.id ===
								terminal.activeTabId
									? 'bg-muted-foreground/40 dark:bg-muted-foreground/40'
									: 'hover:bg-muted/50'}"
								onclick={() => setActiveTab(terminal.id, tab.id)}
							>
								<span class="mr-2 whitespace-nowrap text-sm">{tab.title}</span>
								{#if terminal.tabs.length > 1}
									<button class="hover:text-red-400" onclick={() => removeTab(terminal.id, tab.id)}>
										<Icon icon="mdi:close" class="h-3 w-3" />
									</button>
								{/if}
							</div>
						{/each}
					</div>

					<div class="bg-muted relative h-full w-full flex-grow overflow-hidden">
						{#each terminal.tabs as tab}
							{#if tab.id === terminal.activeTabId}
								<div class="absolute inset-0">
									<Xterm
										terminal={xtremTerminal}
										{options}
										{onLoad}
										{onData}
										class="h-full w-full"
									/>
								</div>
							{/if}
						{/each}
					</div>
				</div>
			</div>
		{/if}
	{/if}
{/each}
