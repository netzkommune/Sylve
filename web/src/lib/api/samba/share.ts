import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { SambaShareSchema, type SambaShare } from '$lib/types/samba/shares';
import { apiRequest } from '$lib/utils/http';
import { z } from 'zod/v4';

export async function getSambaShares(): Promise<SambaShare[]> {
	return await apiRequest('/samba/shares', z.array(SambaShareSchema), 'GET');
}

export async function createSambaShare(
	name: string,
	dataset: string,
	readOnlyGroups: string[] = [],
	writeableGroups: string[] = [],
	createMask: string = '',
	directoryMask: string = ''
): Promise<APIResponse> {
	return await apiRequest('/samba/shares', APIResponseSchema, 'POST', {
		name,
		dataset,
		readOnlyGroups,
		writeableGroups,
		createMask,
		directoryMask
	});
}

export async function deleteSambaShare(id: number): Promise<APIResponse> {
	return await apiRequest(`/samba/shares/${id}`, APIResponseSchema, 'DELETE');
}
