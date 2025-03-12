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
import { hostname, language as langStore } from '$lib/stores/basic';
import adze from 'adze';
import axios, { AxiosError } from 'axios';
import { toast } from 'svelte-french-toast';
import { get } from 'svelte/store';

export async function login(
	username: string,
	password: string,
	authType: string,
	remember: boolean,
	language: string
) {
	try {
		const response = await axios.post('/api/auth/login', {
			username,
			password,
			authType,
			remember
		});

		if (response.status === 200 && response.data) {
			if (response.data.data?.hostname && response.data.data?.token) {
				langStore.set(language);
				hostname.set(response.data.data.hostname);
				store.set(response.data.data.token);
				return true;
			} else {
				toast.error('Error logging in', {
					position: 'bottom-center'
				});
			}
		} else {
			toast.error('Error logging in', {
				position: 'bottom-center'
			});
			return false;
		}

		return false;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			const axiosError = error as AxiosError;
			if (axiosError.response) {
				if (axiosError.response.status === 401) {
					toast.error('Invalid credentials', {
						position: 'bottom-center'
					});
				} else {
					toast.error('Error logging in', {
						position: 'bottom-center'
					});
				}
			} else if (axiosError.request) {
				toast.error('Error logging in', {
					position: 'bottom-center'
				});
			} else {
				toast.error('Error logging in', {
					position: 'bottom-center'
				});
			}
		} else {
			toast.error('Error logging in', {
				position: 'bottom-center'
			});
		}

		adze.withEmoji.error('Login failed', error);
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
