import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import {
	JailSchema,
	SimpleJailSchema,
	type CreateData,
	type Jail,
	type SimpleJail
} from '$lib/types/jail/jail';
import { apiRequest } from '$lib/utils/http';
import { z } from 'zod/v4';

export async function newJail(data: CreateData): Promise<APIResponse> {
	console.log(data);
	return await apiRequest('/jail', APIResponseSchema, 'POST', {
		name: data.name,
		ctId: parseInt(data.id.toString(), 10),
		description: data.description,
		dataset: data.storage.dataset,
		base: data.storage.base,
		switchId: data.network.switch,
		dhcp: data.network.dhcp,
		slaac: data.network.slaac,
		ipv4: data.network.ipv4,
		ipv4Gw: data.network.ipv4Gateway,
		ipv6: data.network.ipv6,
		ipv6Gw: data.network.ipv6Gateway,
		mac: data.network.mac,
		startAtBoot: data.hardware.startAtBoot,
		startOrder: data.hardware.bootOrder
	});
}

export async function getSimpleJails(): Promise<SimpleJail[]> {
	return await apiRequest('/jail/simple', z.array(SimpleJailSchema), 'GET');
}

export async function getJails(): Promise<Jail[]> {
	return await apiRequest('/jail', z.array(JailSchema), 'GET');
}

export async function deleteJail(ctId: number): Promise<APIResponse> {
	return await apiRequest(`/jail/${ctId}`, APIResponseSchema, 'DELETE');
}
