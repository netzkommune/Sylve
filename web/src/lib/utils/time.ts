/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { format, formatDistance, formatRelative, fromUnixTime, parseISO, subDays } from 'date-fns';
import { formatInTimeZone } from 'date-fns-tz';

/**
 * Converts a Unix timestamp (in seconds) to a human-readable "X time ago" string
 *
 * @param seconds - Unix timestamp in seconds since epoch
 * @returns Formatted time ago string (e.g., "9 hours ago")
 */
export function secondsToHoursAgo(seconds: number | undefined): string {
	if (!seconds) return 'Unknown';

	// If the value is small (like 32893), it's likely seconds of uptime rather than a Unix timestamp
	if (seconds < 86400 * 365) {
		// Less than a year in seconds
		// Convert uptime seconds to a date by subtracting from now
		const uptimeDate = new Date(Date.now() - seconds * 1000);
		return formatDistance(uptimeDate, new Date(), { addSuffix: true });
	}

	// Otherwise treat as a Unix timestamp
	return formatDistance(fromUnixTime(seconds), new Date(), { addSuffix: true });
}

export function convertDbTime(time: string): string {
	try {
		const userTimeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
		const date = new Date(time);
		const zonedDate = formatInTimeZone(date, userTimeZone, 'yyyy-MM-dd hh:mm a');

		return zonedDate;
	} catch (e) {
		return time;
	}
}
