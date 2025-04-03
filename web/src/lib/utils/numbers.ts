/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import humanFormat from 'human-format';

export function floatToNDecimals(value: number | undefined, n: number): number {
	if (!value) return 0.0;
	return parseFloat(value.toFixed(n));
}

export function bytesToHumanReadable(value: number | undefined): string {
	if (!value) return '0 B';
	return humanFormat(value, { unit: 'B' });
}

export function getDeterministicIdFromString(str: string): string {
	const hash = Array.from(str).reduce((acc, char) => acc + char.charCodeAt(0), 0);
	return `id-${hash}`;
}
