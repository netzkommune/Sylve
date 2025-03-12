import { RAMInfoSchema, type RAMInfo } from '$lib/types/info/ram';
import { apiRequest } from '$lib/utils/http';

export async function getRAMInfo(): Promise<RAMInfo> {
	return await apiRequest('/info/ram', RAMInfoSchema, 'GET');
}

export async function getSwapInfo(): Promise<RAMInfo> {
	return await apiRequest('/info/swap', RAMInfoSchema, 'GET');
}
