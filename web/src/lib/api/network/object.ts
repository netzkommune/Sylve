import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { NetworkObjectSchema, type NetworkObject } from '$lib/types/network/object';
import { apiRequest } from '$lib/utils/http';

export async function getNetworkObjects(): Promise<NetworkObject[]> {
	return await apiRequest('/network/object', NetworkObjectSchema.array(), 'GET');
}

export async function createNetworkObject(
	name: string,
	type: string,
	values: string[]
): Promise<APIResponse> {
	const body = {
		name,
		type,
		values
	};

	return await apiRequest('/network/object', APIResponseSchema, 'POST', body);
}

export async function deleteNetworkObject(id: number): Promise<APIResponse> {
	return await apiRequest(`/network/object/${id}`, APIResponseSchema, 'DELETE');
}
