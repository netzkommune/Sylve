import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

export async function modifyHardware(
	vmId: number,
	cpuSockets: number,
	cpuCores: number,
	cpuThreads: number,
	ram: number,
	cpuPinning: number[],
	vncPort: number,
	vncResolution: string,
	vncPassword: string,
	vncWait: boolean,
	pciDevices: number[]
): Promise<APIResponse> {
	return await apiRequest(`/vm/hardware/${vmId}`, APIResponseSchema, 'PUT', {
		cpuSockets,
		cpuCores,
		cpuThreads,
		ram,
		cpuPinning,
		vncPort,
		vncResolution,
		vncPassword,
		vncWait,
		pciDevices
	});
}

export async function modifyCPU(
	vmId: number,
	cpuSockets: number,
	cpuCores: number,
	cpuThreads: number,
	cpuPinning: number[]
): Promise<APIResponse> {
	return await apiRequest(`/vm/hardware/cpu/${vmId}`, APIResponseSchema, 'PUT', {
		cpuSockets,
		cpuCores,
		cpuThreads,
		cpuPinning
	});
}

export async function modifyRAM(vmId: number, ram: number): Promise<APIResponse> {
	return await apiRequest(`/vm/hardware/ram/${vmId}`, APIResponseSchema, 'PUT', {
		ram
	});
}

export async function modifyVNC(
	vmId: number,
	vncPort: number,
	vncResolution: string,
	vncPassword: string,
	vncWait: boolean
): Promise<APIResponse> {
	return await apiRequest(`/vm/hardware/vnc/${vmId}`, APIResponseSchema, 'PUT', {
		vncPort,
		vncResolution,
		vncPassword,
		vncWait
	});
}

export async function modifyPPT(vmId: number, pciDevices: number[]): Promise<APIResponse> {
	return await apiRequest(`/vm/hardware/ppt/${vmId}`, APIResponseSchema, 'PUT', {
		pciDevices
	});
}
