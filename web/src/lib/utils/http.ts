/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import { api } from '$lib/api/common';
import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import type { QueryFunctionContext } from '@sveltestack/svelte-query';
import adze from 'adze';
import toast from 'svelte-french-toast';
import { z } from 'zod/v4';
import { getValidationError } from './i18n';

export async function apiRequest<T extends z.ZodType>(
	endpoint: string,
	schema: T,
	method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH',
	body?: unknown
): Promise<z.infer<T>> {
	try {
		const config = {
			method,
			url: endpoint,
			...(body ? { data: body } : {})
		};

		const response = await api.request({ ...config, validateStatus: () => true });
		const apiResponse = APIResponseSchema.safeParse(response.data);

		/* Couldn't parse response data into APIResponse so we'll just return the data? */
		if (!apiResponse.success) {
			return apiResponse.data;
		}

		/* Caller asked for a raw response */
		if (schema._def.description === 'APIResponseSchema') {
			return apiResponse.data as z.infer<T>;
		}

		if (apiResponse.data.data) {
			const parsedResult = schema.safeParse(apiResponse.data.data);
			if (parsedResult.success) {
				return parsedResult.data;
			} else {
				// console.log('Validation Error', parsedResult.error);
				adze.withEmoji.warn('Zod Validation Error', parsedResult.error);
				return getDefaultValue(schema, apiResponse.data);
			}
		}

		return getDefaultValue(schema, apiResponse.data);
	} catch (error) {
		return getDefaultValue(schema, { status: 'error' });
	}
}

function getDefaultValue<T extends z.ZodType>(schema: T, response: APIResponse): z.infer<T> {
	if (schema instanceof z.ZodArray) {
		return [] as z.infer<T>;
	}

	if (schema instanceof z.ZodObject) {
		return response as z.infer<T>;
	}

	return undefined as z.infer<T>;
}

type CacheEntry<T> = {
	timestamp: number;
	data: T;
};

export async function cachedFetch<T>(
	key: string,
	fetchFunction: (queryObj?: QueryFunctionContext) => Promise<T>,
	duration: number
): Promise<T> {
	const now = Date.now();
	const storedEntry = localStorage.getItem(key);

	if (storedEntry) {
		try {
			const entry: CacheEntry<T> = JSON.parse(storedEntry);
			if (now - entry.timestamp < duration) {
				return entry.data;
			}
		} catch (error) {
			console.error(`Failed to parse cached data for key "${key}"`, error);
		}
	}

	const data = await fetchFunction();
	const entry: CacheEntry<T> = { timestamp: now, data };
	localStorage.setItem(key, JSON.stringify(entry));
	return data;
}

export async function updateCache<T>(key: string, obj: T): Promise<void> {
	const now = Date.now();
	const storedEntry = localStorage.getItem(key);

	if (storedEntry) {
		try {
			const entry: CacheEntry<T> = JSON.parse(storedEntry);
			entry.data = obj;
			entry.timestamp = now;
			localStorage.setItem(key, JSON.stringify(entry));
		} catch (error) {
			console.error(`Failed to parse cached data for key "${key}"`, error);
		}
	}
}

export function isAPIResponse(obj: any): obj is APIResponse {
	return (
		obj &&
		typeof obj.status === 'string' &&
		(typeof obj.message === 'string' || typeof obj.error === 'string')
	);
}

export function handleValidationErrors(result: APIResponse, section: string): void {
	adze.withEmoji.error('Validation Error', result);
	if (result.error === 'validation_error') {
		if (Array.isArray(result.data)) {
			if (result.data.length > 0) {
				toast.error(getValidationError(result.data[0], section), {
					position: 'bottom-center'
				});
			}
		}
	}
}

export function handleAPIError(result: APIResponse): void {
	adze.withEmoji.error('API Error', result);
}
