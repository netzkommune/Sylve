import {
	IODelayHistoricalSchema,
	IODelaySchema,
	ZpoolSchema,
	type IODelay,
	type IODelayHistorical,
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
				'/zfs/pool/io-delay?historical=1',
				IODelayHistoricalSchema,
				'GET'
			);
			return IODelayHistoricalSchema.parse(data);
		}
	}

	const data = await apiRequest('/zfs/pool/io-delay', IODelaySchema, 'GET');
	return IODelaySchema.parse(data);
}

export async function getPools(): Promise<Zpool[]> {
	const data = await apiRequest('/zfs/pool/list', ZpoolSchema.array(), 'GET');
	return ZpoolSchema.array().parse(data);
}
