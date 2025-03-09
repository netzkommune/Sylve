import { RAMInfoSchema, type RAMInfo } from '$lib/types/info/ram';
import { apiRequest } from '$lib/utils/http';

export async function getRAMInfo(): Promise<RAMInfo> {
	const data = await apiRequest('/info/ram', RAMInfoSchema, 'GET');
	return RAMInfoSchema.parse(data);
}

export async function getSwapInfo(): Promise<RAMInfo> {
	const data = await apiRequest('/info/swap', RAMInfoSchema, 'GET');
	return RAMInfoSchema.parse(data);
}
