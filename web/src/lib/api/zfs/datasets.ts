import { APIResponseSchema } from '$lib/types/common';
import { DatasetSchema, type Dataset } from '$lib/types/zfs/dataset';

import { apiRequest } from '$lib/utils/http';

export async function getDatasets(): Promise<Dataset[]> {
	return await apiRequest('/zfs/datasets', DatasetSchema.array(), 'GET');
}
