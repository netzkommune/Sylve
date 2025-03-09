/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { logOut } from '$lib/api/auth';

export interface KeyboardTrigger {
	key: string;
	modifier: string[][];
	callback: () => void;
}

export const triggers: KeyboardTrigger[] = [
	{
		key: 'Q',
		modifier: [
			['ctrl', 'shift'],
			['meta', 'shift']
		],
		callback: logOut
	}
];
