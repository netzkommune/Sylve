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

export async function destroyDisk(disk: string): Promise<APIResponse> {
	return await apiRequest(`/disk/wipe`, APIResponseSchema, 'POST', {
		device: disk
	});
}

export async function destroyPartition(partition: string): Promise<APIResponse> {
	return await apiRequest(`/disk/delete-partition`, APIResponseSchema, 'POST', {
		device: partition
	});
}

export async function initializeGPT(disk: string): Promise<APIResponse> {
	return await apiRequest(`/disk/initialize-gpt`, APIResponseSchema, 'POST', {
		device: disk
	});
}

export async function createPartitions(disk: string, sizes: number[]): Promise<APIResponse> {
	return await apiRequest(`/disk/create-partitions`, APIResponseSchema, 'POST', {
		device: disk,
		sizes
	});
}
