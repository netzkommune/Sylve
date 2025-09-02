/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { parseJwt } from '$lib/utils/string';
import { localStore } from '@layerstack/svelte-stores';
import { addDays } from 'date-fns';
import { get } from 'svelte/store';

export const store = localStore('token', '', {
	expiry: addDays(new Date(), 1)
});

export const oldStore = localStore('oldToken', '', {
	expiry: addDays(new Date(), 1)
});

export const clusterStore = localStore('clusterToken', '', {
	expiry: addDays(new Date(), 1)
});

export const currentHostname = localStore('currentHostname', '', {
	expiry: addDays(new Date(), 1280)
});

export function getUsername(): string {
	try {
		const token = get(store);
		if (!token) return 'unknown';

		const decoded = parseJwt(token);
		return decoded.custom_claims.username;
	} catch (e) {
		return 'unknown';
	}
}
