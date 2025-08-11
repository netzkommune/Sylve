import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

export async function modifyRAM(ctId: number, bytes: number): Promise<APIResponse> {
	return await apiRequest('/jail/memory', APIResponseSchema, 'PUT', {
		ctId: ctId,
		memory: bytes
	});
}

export async function modifyCPU(ctId: number, cores: number): Promise<APIResponse> {
	return await apiRequest('/jail/cpu', APIResponseSchema, 'PUT', {
		ctId: ctId,
		cores: parseInt(cores.toString(), 10)
	});
}
