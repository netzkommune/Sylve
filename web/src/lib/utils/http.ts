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
import { APIResponseSchema } from '$lib/types/common';
import type { QueryFunctionContext } from '@sveltestack/svelte-query';
import adze from 'adze';
import { z } from 'zod';

export async function apiRequest<T extends z.ZodType>(
	endpoint: string,
	schema: T,
	method: 'GET' | 'POST' | 'PUT' | 'DELETE',
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

		if (!apiResponse.success) {
			return getDefaultValue(schema);
		}

		if (apiResponse.data.status !== 'success') {
			return apiResponse.data;
		}

		const responseData = apiResponse.data.data;
		const parsedResult = schema.safeParse(responseData);

		if (parsedResult.success) {
			return parsedResult.data;
		}

		if (responseData && typeof responseData === 'object') {
			const nestedResult = schema.safeParse((responseData as any).data);
			if (nestedResult.success) {
				return nestedResult.data;
			}
		}

		return getDefaultValue(schema);
	} catch (error) {
		return getDefaultValue(schema);
	}
}

function getDefaultValue<T extends z.ZodType>(schema: T): z.infer<T> {
	if (schema instanceof z.ZodArray) {
		return [] as z.infer<T>;
	}

	if (schema instanceof z.ZodObject) {
		return {} as z.infer<T>;
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
