/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import type { Disk, DiskInfo, SmartAttribute, SmartCtl, SmartNVME } from '$lib/types/disk/disk';
import type { Zpool } from '$lib/types/zfs/pool';

export async function simplifyDisks(disks: DiskInfo): Promise<Disk[]> {
	const transformed: Disk[] = [];
	for (const disk of disks) {
		if (disk.size > 0) {
			const o = {
				Device: `/dev/${disk.device}`,
				Type: disk.type,
				Usage: disk.usage,
				Size: disk.size,
				GPT: disk.gpt,
				Model: disk.model,
				Serial: disk.serial,
				Wearout:
					typeof disk.wearOut === 'number' && !isNaN(disk.wearOut) ? `-` : `${disk.wearOut} %`,
				'S.M.A.R.T.': 'Passed',
				Partitions: disk.partitions,
				SmartData: disk.smartData ?? null
			};

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

export function parseSMART(disk: Disk): SmartAttribute | SmartAttribute[] {
	if (disk.Type === 'NVMe') {
		return {
			'Available Spare': (disk.SmartData as SmartNVME).availableSpare,
			'Available Spare Threshold': (disk.SmartData as SmartNVME).availableSpareThreshold,
			'Controller Busy Time': (disk.SmartData as SmartNVME).controllerBusyTime,
			'Critical Warning': (disk.SmartData as SmartNVME).criticalWarning,
			'Critical Warning State': {
				'Available Spare': (disk.SmartData as SmartNVME).criticalWarningState.availableSpare,
				'Device Reliability': (disk.SmartData as SmartNVME).criticalWarningState.deviceReliability,
				'Read Only': (disk.SmartData as SmartNVME).criticalWarningState.readOnly,
				Temperature: (disk.SmartData as SmartNVME).criticalWarningState.temperature,
				'Volatile Memory Backup': (disk.SmartData as SmartNVME).criticalWarningState
					.volatileMemoryBackup
			},
			'Data Units Read': (disk.SmartData as SmartNVME).dataUnitsRead,
			'Data Units Written': (disk.SmartData as SmartNVME).dataUnitsWritten,
			'Error Info Log Entries': (disk.SmartData as SmartNVME).errorInfoLogEntries,
			'Host Read Commands': (disk.SmartData as SmartNVME).hostReadCommands,
			'Host Write Commands': (disk.SmartData as SmartNVME).hostWriteCommands,
			'Media Errors': (disk.SmartData as SmartNVME).mediaErrors,
			'Percentage Used': (disk.SmartData as SmartNVME).percentageUsed,
			'Power Cycles': (disk.SmartData as SmartNVME).powerCycles,
			'Power On Hours': (disk.SmartData as SmartNVME).powerOnHours,
			Temperature: (disk.SmartData as SmartNVME).temperature,
			'Temperature 1 Transition Count': (disk.SmartData as SmartNVME).temperature1TransitionCnt,
			'Temperature 2 Transition Count': (disk.SmartData as SmartNVME).temperature2TransitionCnt,
			'Total Time For Temperature 1': (disk.SmartData as SmartNVME).totalTimeForTemperature1,
			'Total Time For Temperature 2': (disk.SmartData as SmartNVME).totalTimeForTemperature2,
			'Unsafe Shutdowns': (disk.SmartData as SmartNVME).unsafeShutdowns,
			'Warning Composite Temp Time': (disk.SmartData as SmartNVME).warningCompositeTempTime
		};
	} else if (disk.Type === 'HDD' || disk.Type === 'SSD') {
		const data = disk.SmartData as SmartCtl;
		const attributes: SmartAttribute[] = [];

		if (data?.ata_smart_attributes?.table?.length) {
			for (const element of data.ata_smart_attributes.table) {
				attributes.push({
					ID: element.id,
					Name: element.name,
					Value: element.value,
					Worst: element.worst,
					Threshold: element.thresh,
					Flags: element.flags.string,
					Failing: element.when_failed || '-'
				});
			}
		}

		if (attributes.length > 0) {
			return attributes;
		}
	}

	return {};
}

export function diskSpaceAvailable(disk: Disk, required: number): boolean {
	if (disk.Usage === 'Partitions') {
		const total = disk.Size;
		const used = disk.Partitions.reduce((acc, cur) => acc + cur.size, 0);
		return total - used >= required;
	}

	return disk.Size >= required;
}

export function getGPTLabel(disk: Disk, pools: Zpool[]): string {
	if (disk.GPT) {
		return 'Yes';
	}

	if (disk.GPT === false) {
		if (pools.length > 0) {
			if (pools.some((pool) => pool.vdevs.some((vdev) => vdev.name === disk.Device))) {
				return '-';
			}
		}
	}

	return 'No';
}
