import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

export async function storageDetach(vmId: number, storageId: number): Promise<APIResponse> {
	return await apiRequest(`/vm/storage/detach`, APIResponseSchema, 'POST', {
		vmId,
		storageId
	});
}

export async function storageAttach(
	vmId: number,
	storageType: string,
	dataset: string,
	emulation: string,
	size: number
): Promise<APIResponse> {
	return await apiRequest(`/vm/storage/attach`, APIResponseSchema, 'POST', {
		vmId,
		storageType,
		dataset,
		emulation,
		size
	});
}
