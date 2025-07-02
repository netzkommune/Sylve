/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

export function sampleByInterval(
	data: Array<{ date: Date; value: number }>,
	intervalMin: number
): Array<{ date: Date; value: number }> {
	if (!data.length || intervalMin <= 0) return [];

	const ms = intervalMin * 60 * 1000;
	const sorted = data.slice().sort((a, b) => a.date.getTime() - b.date.getTime());
	const buckets = new Map();

	for (const pt of sorted) {
		const bucketKey = Math.floor(pt.date.getTime() / ms) * ms;
		if (!buckets.has(bucketKey)) {
			buckets.set(bucketKey, pt);
		}
	}

	return Array.from(buckets.values()).sort((a, b) => a.date - b.date);
}
