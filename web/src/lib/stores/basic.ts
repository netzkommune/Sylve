/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { localStore } from '@layerstack/svelte-stores';
import { addDays } from 'date-fns';

export const hostname = localStore('hostname', '', {
	expiry: addDays(new Date(), 1)
});

export const language = localStore('language', 'en', {
	expiry: addDays(new Date(), 720)
});

export const paneSizes = localStore(
	'paneSizes',
	{
		main: 70,
		left: 13,
		middle: 87,
		bottom: 30,
		right: 87
	},
	{
		expiry: addDays(new Date(), 720)
	}
);
