/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { getIcon, loadIcon } from '@iconify/svelte';
import { Mnemonic } from './vendor/mnemonic';

export function capitalizeFirstLetter(str: string, firstOnly: boolean = false): string {
	if (firstOnly) {
		return str.charAt(0).toLocaleUpperCase() + str.slice(1);
	}

	return str
		.split(' ')
		.map((word) => word.charAt(0).toLocaleUpperCase() + word.slice(1))
		.join(' ');
}

export function parseJwt(token: string) {
	let base64Url = token.split('.')[1];
	let base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
	let jsonPayload = decodeURIComponent(
		window
			.atob(base64)
			.split('')
			.map(function (c) {
				return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
			})
			.join('')
	);

	return JSON.parse(jsonPayload);
}

export function shortenString(str: string, maxLength: number): string {
	if (str.length <= maxLength) return str;
	return str.slice(0, maxLength) + '...';
}

export function generatePassword(): string {
	return new Mnemonic().toWords().slice(0, 6).join('-');
}

export async function iconToSVG(icon: string): Promise<string> {
	await loadIcon(icon);
	const i = getIcon(icon);
	if (i) {
		const { body, width, height, left, top } = i;
		const viewBox = `${left} ${top} ${width} ${height}`;
		const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="${viewBox}">${body}</svg>`;
		return svg;
	}

	return ''; // Ensure the function always returns a string
}
