/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { z } from 'zod';

export const APIResponseSchema = z.object({
	status: z.string(),
	message: z.string().optional(),
	error: z.string().optional(),
	data: z.unknown().optional()
});

export interface HistoricalBase {
	id?: number | string;
	createdAt?: string | Date;
	[key: string]: number | string | Date | undefined;
}

export interface HistoricalData {
	date: Date;
	[key: string]: number | string | Date;
}
