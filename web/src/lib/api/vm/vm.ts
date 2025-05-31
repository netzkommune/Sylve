import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { VMSchema, type CreateData, type VM } from '$lib/types/vm/vm';
import { apiRequest } from '$lib/utils/http';
import { z } from 'zod/v4';

export async function getVMs(): Promise<VM[]> {
	return await apiRequest('/vm/list', z.array(VMSchema), 'GET');
}

export async function newVM(data: CreateData): Promise<APIResponse> {
	return await apiRequest('/vm/create', APIResponseSchema, 'POST', {
		name: data.name,
		vmId: data.id,
		iso: data.storage.iso,
		storageType: data.storage.type,
		storageDataset: data.storage.guid,
		storageSize: data.storage.size,
		storageEmulationType: data.storage.emulation,
		switchId: data.network.switch,
		switchEmulationType: data.network.emulation,
		macAddress: data.network.mac,
		cpuSockets: data.hardware.sockets,
		cpuCores: data.hardware.cores,
		cpuThreads: data.hardware.threads,
		ram: data.hardware.memory,
		vncPort: data.advanced.vncPort,
		vncPassword: data.advanced.vncPassword,
		vncWait: data.advanced.vncWait,
		vncResolution: data.advanced.vncResolution,
		startAtBoot: data.advanced.startAtBoot,
		bootOrder: data.advanced.bootOrder,
		pciDevices: data.hardware.passthroughIds
	});
}

export async function deleteVM(id: number): Promise<APIResponse> {
	return await apiRequest(`/vm/remove/${id}`, APIResponseSchema, 'DELETE');
}
