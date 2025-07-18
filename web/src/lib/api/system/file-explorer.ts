import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { FileNodeSchema, type FileNode } from '$lib/types/system/file-explorer';
import { apiRequest } from '$lib/utils/http';

export async function getFiles(id?: string): Promise<FileNode[]> {
	let url = '/system/file-explorer';

	if (id) {
		url += `?id=${encodeURIComponent(id)}`;
	}

	return await apiRequest(url, FileNodeSchema.array(), 'GET');
}

export async function addFileOrFolder(
	path: string,
	name: string,
	isFolder: boolean
): Promise<APIResponse> {
	const body = {
		path,
		name,
		isFolder
	};

	return await apiRequest('/system/file-explorer', APIResponseSchema, 'POST', body);
}

export async function deleteFileOrFolder(path: string): Promise<APIResponse> {
	return await apiRequest(
		'/system/file-explorer?id=' + encodeURIComponent(path),
		APIResponseSchema,
		'DELETE'
	);
}

export async function renameFileOrFolder(id: string, newName: string): Promise<APIResponse> {
	const body = {
		id,
		newName
	};

	return await apiRequest('/system/file-explorer/rename', APIResponseSchema, 'POST', body);
}

export async function copyOrMoveFileOrFolder(
	id: string,
	newPath: string,
	cut: boolean
): Promise<APIResponse> {
	const body = {
		id,
		newPath,
		cut
	};

	return await apiRequest('/system/file-explorer/copy-or-move', APIResponseSchema, 'POST', body);
}

export async function deleteFilesOrFolders(paths: string[]): Promise<APIResponse> {
	const body = {
		paths
	};

	return await apiRequest('/system/file-explorer/delete', APIResponseSchema, 'POST', body);
}
