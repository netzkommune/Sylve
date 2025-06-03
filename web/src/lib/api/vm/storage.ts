import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

export async function storageDetach(vmId: number, storageId: number): Promise<APIResponse> {
	return await apiRequest(`/vm/storage/detach`, APIResponseSchema, 'POST', {
		vmId,
		storageId
	});
}
