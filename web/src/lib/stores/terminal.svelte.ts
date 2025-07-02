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
import { localStore } from '@layerstack/svelte-stores';
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
    title: string;
    tabs: Tab[];
    activeTabId: string;
}

export const terminalStore = localStore<Terminal>('terminal', {
    isOpen: false,
    isMinimized: false,
    title: '',
    tabs: [],
    activeTabId: ''
});

export function getDefaultTitle() {
    return `${getUsername()}@${get(hostname)}:~`;
}

export function openTerminal() {
    terminalStore.set({
        ...get(terminalStore),
        isOpen: true,
        isMinimized: false
    });

    let store = get(terminalStore);
    if (store.tabs.length > 0) {
        return;
    }

    const tabId = nanoid(9);
    const newTerminal: Terminal = {
        isOpen: true,
        isMinimized: false,
        title: 'Terminal',
        tabs: [
            {
                id: tabId,
                title: getDefaultTitle()
            }
        ],

        activeTabId: tabId
    };

    terminalStore.set(newTerminal);
}
