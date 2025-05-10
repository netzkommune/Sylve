import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { SwitchListSchema, type SwitchList } from '$lib/types/network/switch';
import { apiRequest } from '$lib/utils/http';

export async function getSwitches(): Promise<SwitchList> {
	return await apiRequest('/network/switch', SwitchListSchema, 'GET');
}

export async function createSwitch(
	name: string,
	mtu: number,
	vlan: number,
	address: string,
	privateSw: boolean,
	ports: string[]
): Promise<APIResponse> {
	const body = {
		name,
		mtu,
		vlan,
		address,
		private: privateSw,
		ports
	};

	return await apiRequest('/network/switch/standard', APIResponseSchema, 'POST', body);
}

export async function deleteSwitch(id: number): Promise<APIResponse> {
	return await apiRequest(`/network/switch/standard/${id}`, APIResponseSchema, 'DELETE');
}
