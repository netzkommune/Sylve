<script lang="ts">
	import { store } from '$lib/stores/auth';
	import { getDefaultTitle, terminalStore } from '$lib/stores/terminal.svelte';
	import {
		Xterm,
		XtermAddon,
		type FitAddon,
		type ITerminalInitOnlyOptions,
		type ITerminalOptions,
		type Terminal
	} from '@battlefieldduck/xterm-svelte';
	import Icon from '@iconify/svelte';
	import adze from 'adze';
	import { nanoid } from 'nanoid';
	import { untrack } from 'svelte';
	import { fade, scale } from 'svelte/transition';

	let terminal = $state<Terminal>();
	let ws = $state<WebSocket>();
	let fitAddonGlobal = $state<FitAddon>();
	let options: ITerminalOptions & ITerminalInitOnlyOptions = {
		cursorBlink: true
	};

	let tabsCount = $derived.by(() => {
		return terminalStore.value.tabs.length;
	});

	let currentTab = $derived.by(() => {
		return terminalStore.value.tabs.find((tab) => tab.id === terminalStore.value.activeTabId);
	});

	async function killSession(sessionId: string): Promise<boolean> {
		return new Promise((resolve) => {
			if (!ws || ws.readyState !== WebSocket.OPEN) {
				resolve(false);
				return;
			}

			const onMessage = (event: MessageEvent) => {
				if (event.data.includes(`Session killed: ${sessionId}`)) {
					ws?.removeEventListener('message', onMessage);
					resolve(true);
				}
			};

			ws.addEventListener('message', onMessage);
			ws.send(new TextEncoder().encode('\x02' + JSON.stringify({ kill: sessionId })));

			// Timeout after 2 seconds
			setTimeout(() => {
				ws?.removeEventListener('message', onMessage);
				resolve(false);
			}, 2000);
		});
	}

	async function onLoad() {
		try {
			if (!currentTab) return;

			ws?.close();
			terminal?.clear();
			terminal?.reset();

			const fitAddon = new (await XtermAddon.FitAddon()).FitAddon();
			terminal?.loadAddon(fitAddon);
			fitAddon.fit();

			ws = new WebSocket(`/api/info/terminal?id=${currentTab?.id}`, ['Bearer', $store]);
			ws.binaryType = 'arraybuffer';
			ws.onopen = () => {
				adze.info(`Terminal WebSocket connected for tab ${currentTab?.id}`);
				if (terminal) {
					const dimensions = fitAddon.proposeDimensions();
					(ws as WebSocket).send(
						new TextEncoder().encode(
							'\x01' + JSON.stringify({ rows: dimensions?.rows, cols: dimensions?.cols })
						)
					);

					fitAddonGlobal = fitAddon;
				}
			};

			ws.onmessage = (event) => {
				if (event.data instanceof ArrayBuffer) {
					if (terminal) {
						terminal.write(new Uint8Array(event.data));
					}
				}
			};

			ws.onclose = () => {
				adze.info(`Terminal WebSocket disconnected for tab ${currentTab?.id}`);
				if (terminal) {
					terminal.write('\x1b[31mDisconnected from server.\x1b[0m\r\n');
				}
			};
		} catch (e) {
			adze.error('Failed to connect to terminal WebSocket', { error: e });
		}
	}

	function onData(data: string) {
		ws?.send(new TextEncoder().encode('\x00' + data));
	}

	async function visiblityAction(t: string, e?: MouseEvent | string) {
		if (t === 'window-minimize') {
			terminalStore.value.isMinimized = true;
			return;
		}

		if (t === 'window-close') {
			const tabsToKill = [...terminalStore.value.tabs];
			for (const tab of tabsToKill) {
				await killSession(tab.id);
			}

			terminalStore.value.tabs = [];
			terminalStore.value.isOpen = false;
			ws?.close();
		}

		if (t === 'tab-close') {
			const event = e as MouseEvent;
			if (event) {
				const target = event.target as HTMLElement;
				const parent = target.closest('button');
				if (parent) {
					const tabId = parent.getAttribute('data-id');
					if (tabId) {
						await killSession(tabId);
						terminalStore.value.tabs = terminalStore.value.tabs.filter((tab) => tab.id !== tabId);
						if (terminalStore.value.tabs.length > 0) {
							terminalStore.value.activeTabId = terminalStore.value.tabs[0].id;
						}
					}
				}
			}
		}

		if (t === 'tab-select') {
			const tabId = e as string;
			terminalStore.value.activeTabId = tabId;
		}
	}

	function addTab() {
		const newTab = {
			id: nanoid(5),
			title: getDefaultTitle()
		};

		terminalStore.value.tabs = [...terminalStore.value.tabs, newTab];
		terminalStore.value.activeTabId = newTab.id;
	}

	let innerWidth = $state(0);

	$effect(() => {
		if (innerWidth) {
			untrack(() => {
				fitAddonGlobal?.fit();
				const dimensions = fitAddonGlobal?.proposeDimensions();
				ws?.send(
					new TextEncoder().encode(
						'\x01' + JSON.stringify({ rows: dimensions?.rows, cols: dimensions?.cols })
					)
				);
			});
		}
	});
</script>

<svelte:window bind:innerWidth />

{#if terminalStore.value.isOpen && !terminalStore.value.isMinimized}
	<div
		class="fixed inset-0 z-[9998] bg-black/30 backdrop-blur-sm transition-all duration-150"
	></div>
	<div
		class="fixed inset-0 z-[9999] flex items-center justify-center transition-all duration-150"
		in:scale={{ start: 0.9, duration: 150 }}
		out:scale={{ start: 0.9, duration: 150 }}
	>
		<div
			class="relative flex w-[60%] flex-col rounded-lg border-4 border-muted bg-muted-foreground/10"
		>
			<div class="flex items-center justify-between bg-primary-foreground p-2">
				<!-- Add Tab Button -->
				<div class="flex items-center gap-2">
					<span>{terminalStore.value.title}</span>
				</div>
				<!-- Minimize / Close -->
				<div class="flex space-x-3">
					<button
						class="rounded-full transition-colors duration-300 ease-in-out hover:bg-yellow-600 hover:text-white"
						onclick={() => visiblityAction('window-minimize')}
						title="Minimize"
					>
						<Icon icon="mdi:window-minimize" class="h-5 w-5" />
					</button>
					<button
						class="rounded-full transition-colors duration-300 ease-in-out hover:bg-red-500 hover:text-white"
						onclick={() => visiblityAction('window-close')}
						title="Close"
					>
						<Icon icon="mdi:close" class="h-5 w-5" />
					</button>
				</div>
			</div>

			<!-- Available Tabs -->
			<div class="flex overflow-x-auto bg-white dark:bg-muted/30">
				{#each terminalStore.value.tabs as tab}
					<div
						class="flex cursor-pointer items-center border-muted-foreground/40 px-3.5 py-2 {tab.id ===
						terminalStore.value.activeTabId
							? 'bg-muted-foreground/40 dark:bg-muted-foreground/25 '
							: 'border-x border-t border-muted-foreground/25 hover:bg-muted-foreground/25'}"
						onclick={() => visiblityAction('tab-select', tab.id)}
					>
						<span class="mr-2 whitespace-nowrap text-sm">{tab.title}</span>
						{#if tabsCount > 1}
							<button
								class="rounded-full transition-colors duration-300 ease-in-out hover:bg-red-500 hover:text-white"
								data-id={tab.id}
								onclick={(e) => {
									e.stopPropagation();
									visiblityAction('tab-close', e);
								}}
							>
								<Icon icon="mdi:close" class="h-4 w-4" />
							</button>
						{/if}
					</div>
				{/each}
				<div
					class="flex items-center justify-center border px-1 hover:border-muted-foreground/30 hover:bg-muted-foreground/30"
				>
					<button
						class="dark:hover-bg-muted flex h-6 w-6 items-center justify-center rounded"
						onclick={() => addTab()}
						title="Add new tab"
					>
						<Icon icon="ic:sharp-plus" class="h-5 w-5" />
					</button>
				</div>
			</div>

			<!-- Terminal Body -->
			<div
				id="terminal-container"
				class="relative min-h-[456px] w-full flex-grow overflow-hidden bg-black"
			>
				{#each terminalStore.value.tabs as tab}
					{#if tab.id === terminalStore.value.activeTabId}
						<div in:fade={{ duration: 150 }}>
							<Xterm bind:terminal {options} {onLoad} {onData} />
						</div>
					{/if}
				{/each}
			</div>
		</div>
	</div>
{/if}
