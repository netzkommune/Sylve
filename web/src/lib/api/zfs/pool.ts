import { APIResponseSchema } from '$lib/types/common';
import {
	IODelayHistoricalSchema,
	IODelaySchema,
	PoolStatPointsResponseSchema,
	ZpoolSchema,
	type CreateZpool,
	type IODelay,
	type IODelayHistorical,
	type PoolStatPointsResponse,
	type ReplaceDevice,
	type Zpool
} from '$lib/types/zfs/pool';
import { apiRequest } from '$lib/utils/http';
import type { QueryFunctionContext } from '@sveltestack/svelte-query';

export async function getIODelay(
	queryObj: QueryFunctionContext | undefined
): Promise<IODelay | IODelayHistorical> {
	if (queryObj) {
		if (queryObj.queryKey.includes('ioDelayHistorical')) {
			const data = await apiRequest(
				'/zfs/pool/io-delay/historical',
				IODelayHistoricalSchema,
				'GET'
			);
			return IODelayHistoricalSchema.parse(data);
		}
	}

	return await apiRequest('/zfs/pool/io-delay', IODelaySchema, 'GET');
}

export async function getPools(): Promise<Zpool[]> {
	return await apiRequest('/zfs/pools', ZpoolSchema.array(), 'GET');
}

export async function createPool(data: CreateZpool) {
	return await apiRequest('/zfs/pools', APIResponseSchema, 'POST', {
		...data
	});
}

export async function replaceDevice(data: ReplaceDevice) {
	return await apiRequest(`/zfs/pools/${data.name}/replace-device`, APIResponseSchema, 'POST', {
		...data
	});
}

export async function deletePool(name: string) {
	return await apiRequest(`/zfs/pools/${name}`, APIResponseSchema, 'DELETE');
}

export async function scrubPool(name: string) {
	return await apiRequest(`/zfs/pools/${name}/scrub`, APIResponseSchema, 'POST');
}

export async function getPoolStats(
	interval: number,
	limit: number
): Promise<PoolStatPointsResponse> {
	return await apiRequest(
		`/zfs/pool/stats/${interval}/${limit}`,
		PoolStatPointsResponseSchema,
		'GET'
	);
}
