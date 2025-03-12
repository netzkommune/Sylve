import { DiskInfoSchema, type Disk, type DiskInfo, type Partition } from '$lib/types/disk/disk';
import { apiRequest } from '$lib/utils/http';

export async function listDisks(): Promise<DiskInfo> {
	return await apiRequest('/disk/list', DiskInfoSchema, 'GET');
}
