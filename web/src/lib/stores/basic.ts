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
