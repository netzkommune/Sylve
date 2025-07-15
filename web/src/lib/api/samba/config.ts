import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { SambaConfigSchema, type SambaConfig } from '$lib/types/samba/config';
import { apiRequest } from '$lib/utils/http';

export async function getSambaConfig(): Promise<SambaConfig> {
	return await apiRequest('/samba/config', SambaConfigSchema, 'GET');
}

export async function updateSambaConfig(config: Partial<SambaConfig>): Promise<APIResponse> {
	return await apiRequest('/samba/config', APIResponseSchema, 'POST', config);
}
