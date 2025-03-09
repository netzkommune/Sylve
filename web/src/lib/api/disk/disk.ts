import { DiskInfoSchema, type Disk, type DiskInfo, type Partition } from '$lib/types/disk/disk';
import { apiRequest } from '$lib/utils/http';

export async function listDisks(): Promise<DiskInfo> {
	const data = await apiRequest('/disk/list', DiskInfoSchema, 'GET');
	return DiskInfoSchema.parse(data);
}

export async function simplifyDisks(disks: DiskInfo): Promise<Disk[]> {
	const transformed: Disk[] = [];
	for (const disk of disks) {
		if (disk.size > 0) {
			const o = {
				Device: `/dev/${disk.device}`,
				Type: disk.type,
				Usage: disk.usage,
				Size: disk.size,
				GPT: disk.gpt ? 'Yes' : 'No',
				Model: disk.model,
				Serial: disk.serial,
				Wearout:
					typeof disk.wearOut === 'number' && !isNaN(disk.wearOut) ? `-` : `${disk.wearOut} %`,
				'S.M.A.R.T.': 'Passed',
				Partitions: disk.partitions
			};

			// sort the partitions by size asc
			o.Partitions.sort((a, b) => a.size - b.size);

			transformed.push(o);
		}
	}

	transformed.sort((a, b) => {
		if (a.Usage === 'Partitions' && b.Usage !== 'Partitions') {
			return -1;
		}
		if (a.Usage !== 'Partitions' && b.Usage === 'Partitions') {
			return 1;
		}
		return 0;
	});

	return transformed;
}
