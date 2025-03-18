import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import {
	DiskActionSchema,
	DiskInfoSchema,
	type Disk,
	type DiskInfo,
	type Partition
} from '$lib/types/disk/disk';
import { apiRequest } from '$lib/utils/http';

export async function listDisks(): Promise<DiskInfo> {
	return await apiRequest('/disk/list', DiskInfoSchema, 'GET');
}

export async function destroyDiskOrPartition(disk: string): Promise<APIResponse> {
	return await apiRequest(`/disk/wipe`, APIResponseSchema, 'POST', {
		device: disk
	});
}

export async function initializeGPT(disk: string): Promise<APIResponse> {
	return await apiRequest(`/disk/initialize-gpt`, APIResponseSchema, 'POST', {
		device: disk
	});
}
