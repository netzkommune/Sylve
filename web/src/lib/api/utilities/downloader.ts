import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { DownloadSchema, type Download } from '$lib/types/utilities/downloader';
import { apiRequest } from '$lib/utils/http';

export async function getDownloads(): Promise<Download[]> {
	return await apiRequest('/utilities/downloads', DownloadSchema.array(), 'GET');
}

export async function startDownload(url: string): Promise<APIResponse> {
	return await apiRequest('/utilities/download', APIResponseSchema, 'POST', {
		url
	});
}
