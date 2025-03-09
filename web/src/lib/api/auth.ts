/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { oldStore, store } from '$lib/stores/auth';
import { hostname } from '$lib/stores/basic';
import adze from 'adze';
import axios, { AxiosError } from 'axios';
import { get } from 'svelte/store';

interface Token {
	value: string;
	expiry: number;
}

export async function login(
	username: string,
	password: string,
	authType: string,
	remember: boolean
) {
	try {
		const response = await axios.post('/api/auth/login', {
			username,
			password,
			authType,
			remember
		});
		store.set(response.data.token);
		hostname.set(response.data.hostname);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			const axiosError = error as AxiosError;
			if (axiosError.response) {
				if (axiosError.response.status === 401) {
					// showToast({
					// 	text: 'Invalid credentials',
					// 	type: 'error',
					// 	timeout: 5000
					// });
					console.log('Invalid credentials');
				} else {
					// showToast({
					// 	text: 'An error occurred during login',
					// 	type: 'error',
					// 	timeout: 5000
					// });
					console.log('An error occurred during login');
				}
			} else if (axiosError.request) {
				// showToast({
				// 	text: 'No response from server',
				// 	type: 'error',
				// 	timeout: 5000
				// });
				console.log('No response from server');
			} else {
				// showToast({ text: 'An error occurred', type: 'error', timeout: 5000 });
				console.log('An error occurred');
			}
		} else {
			// showToast({
			// 	text: 'An unexpected error occurred',
			// 	type: 'error',
			// 	timeout: 5000
			// });
			console.log('An unexpected error occurred');
		}
		return false;
	}
}

export function getToken(): string | null {
	if (browser) {
		try {
			const parsed = JSON.parse(localStorage.getItem('token') || '');
			return parsed.value;
		} catch (_e: unknown) {
			return null;
		}
	}

	return null;
}

export async function isTokenValid(): Promise<boolean> {
	try {
		const response = await axios.get('/api/health/basic', {
			headers: {
				Authorization: `Bearer ${getToken()}`
			}
		});

		if (response.status < 400) {
			if (response.data?.hostname) {
				hostname.set(response.data.hostname);
			}
			return true;
		}
	} catch (_e: unknown) {
		return false;
	}

	return false;
}

export async function logOut() {
	const token = getToken();
	if (token) {
		oldStore.set(token);
	}

	store.set('');
	hostname.set('');
	if (browser) {
		localStorage.removeItem('token');
		localStorage.removeItem('hostname');
	}

	goto('/?loggedOut=true');
}

export async function revokeJWT() {
	try {
		const oldToken = get(oldStore);
		if (oldToken) {
			await axios.get('/api/auth/logout', {
				headers: {
					Authorization: `Bearer ${oldToken}`
				}
			});

			oldStore.set('');
		}
	} catch (_e: unknown) {
		adze.error('Failed to revoke JWT');
	}
}
