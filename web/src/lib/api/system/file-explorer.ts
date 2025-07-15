import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { FileNodeSchema, type FileNode } from '$lib/types/system/file-explorer';
import { apiRequest } from '$lib/utils/http';

export async function getFiles(id?: string): Promise<FileNode[]> {
	let url = '/system/file-explorer/files';

	if (id) {
		url += `?id=${encodeURIComponent(id)}`;
	}

	return await apiRequest(url, FileNodeSchema.array(), 'GET');
}
