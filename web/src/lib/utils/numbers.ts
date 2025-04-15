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

export function generateNumberFromString(str: string): number {
	let hash = 0x811c9dc5;
	for (let i = 0; i < str.length; i++) {
		hash ^= str.charCodeAt(i);
		hash = Math.imul(hash, 0x01000193);
	}
	return hash >>> 0;
}

export function isValidSize(value: string): boolean {
	try {
		const parsed = humanFormat.parse(value);
		return parsed !== null && !isNaN(parsed);
	} catch (e) {
		return false;
	}
}
