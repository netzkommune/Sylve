import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import {
	DatasetSchema,
	PeriodicSnapshotSchema,
	type Dataset,
	type PeriodicSnapshot
} from '$lib/types/zfs/dataset';

import { apiRequest } from '$lib/utils/http';

export async function getDatasets(): Promise<Dataset[]> {
	return await apiRequest('/zfs/datasets', DatasetSchema.array(), 'GET');
}

export async function deleteSnapshot(
	snapshot: Dataset,
	recursive: boolean = false
): Promise<APIResponse> {
	const param = recursive ? '?recursive=true' : '';
	return await apiRequest(
		`/zfs/datasets/snapshot/${snapshot.properties.guid}${param}`,
		APIResponseSchema,
		'DELETE'
	);
}

export async function createSnapshot(
	dataset: Dataset,
	name: string,
	recursive: boolean
): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/snapshot', APIResponseSchema, 'POST', {
		name: name,
		recursive: recursive,
		guid: dataset.properties.guid
	});
}

export async function getPeriodicSnapshots(): Promise<PeriodicSnapshot[]> {
	return await apiRequest('/zfs/datasets/snapshot/periodic', PeriodicSnapshotSchema.array(), 'GET');
}

export async function createPeriodicSnapshot(
	dataset: Dataset,
	prefix: string,
	recursive: boolean,
	interval: number,
	cronExpr: string
): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/snapshot/periodic', APIResponseSchema, 'POST', {
		guid: dataset.properties.guid,
		prefix: prefix,
		recursive: recursive,
		interval: interval,
		cronExpr: cronExpr
	});
}

export async function deletePeriodicSnapshot(guid: string): Promise<APIResponse> {
	return await apiRequest(`/zfs/datasets/snapshot/periodic/${guid}`, APIResponseSchema, 'DELETE');
}

export async function createFileSystem(
	name: string,
	parent: string,
	properties: Record<string, string>
): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/filesystem', APIResponseSchema, 'POST', {
		name: name,
		parent: parent,
		properties: properties
	});
}

export async function editFileSystem(
	guid: string,
	properties: Record<string, string>
): Promise<APIResponse> {
	return await apiRequest(`/zfs/datasets/filesystem`, APIResponseSchema, 'PATCH', {
		guid: guid,
		properties: properties
	});
}

export async function deleteFileSystem(dataset: Dataset): Promise<APIResponse> {
	return await apiRequest(
		`/zfs/datasets/filesystem/${dataset.properties.guid}`,
		APIResponseSchema,
		'DELETE'
	);
}

export async function rollbackSnapshot(guid: string): Promise<APIResponse> {
	return await apiRequest(`/zfs/datasets/snapshot/rollback`, APIResponseSchema, 'POST', {
		guid: guid,
		destroyMoreRecent: true
	});
}

export async function createVolume(
	name: string,
	parent: string,
	props: Record<string, string>
): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/volume', APIResponseSchema, 'POST', {
		name: name,
		parent: parent,
		properties: props
	});
}

export async function editVolume(
	dataset: Dataset,
	properties: Record<string, string>
): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/volume', APIResponseSchema, 'PATCH', {
		name: dataset.name,
		properties: properties
	});
}

export async function deleteVolume(dataset: Dataset): Promise<APIResponse> {
	return await apiRequest(
		`/zfs/datasets/volume/${dataset.properties.guid}`,
		APIResponseSchema,
		'DELETE'
	);
}

export async function bulkDelete(datasets: Dataset[]): Promise<APIResponse> {
	const guids = datasets.map((dataset) => dataset.properties.guid);
	return await apiRequest('/zfs/datasets/bulk-delete', APIResponseSchema, 'POST', {
		guids: guids
	});
}

export async function flashVolume(guid: string, uuid: string): Promise<APIResponse> {
	return await apiRequest('/zfs/datasets/volume/flash', APIResponseSchema, 'POST', {
		guid: guid,
		uuid: uuid
	});
}
