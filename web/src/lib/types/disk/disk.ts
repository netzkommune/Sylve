import { z } from 'zod';

export const SmartNVMESchema = z.object({
	device: z.string(),
	criticalWarning: z.string(),
	criticalWarningState: z.object({
		availableSpare: z.union([z.number(), z.string()]),
		temperature: z.union([z.number(), z.string()]),
		deviceReliability: z.union([z.number(), z.string()]),
		readOnly: z.union([z.number(), z.string()]),
		volatileMemoryBackup: z.union([z.number(), z.string()])
	}),
	temperature: z.number(),
	availableSpare: z.number(),
	availableSpareThreshold: z.number(),
	percentageUsed: z.number(),
	dataUnitsRead: z.number(),
	dataUnitsWritten: z.number(),
	hostReadCommands: z.number(),
	hostWriteCommands: z.number(),
	controllerBusyTime: z.number(),
	powerCycles: z.number(),
	powerOnHours: z.number(),
	unsafeShutdowns: z.number(),
	mediaErrors: z.number(),
	errorInfoLogEntries: z.number(),
	warningCompositeTempTime: z.number(),
	temperature1TransitionCnt: z.number(),
	temperature2TransitionCnt: z.number(),
	totalTimeForTemperature1: z.number(),
	totalTimeForTemperature2: z.number()
});

export const SmartCtlSchema = z.object({
	json_format_version: z.array(z.number()),
	smartctl: z.object({
		version: z.array(z.number()),
		pre_release: z.boolean(),
		svn_revision: z.string(),
		platform_info: z.string(),
		build_info: z.string().optional(),
		argv: z.array(z.string()),
		drive_database_version: z
			.object({
				string: z.string()
			})
			.optional(),
		exit_status: z.number()
	}),
	local_time: z.object({
		time_t: z.number(),
		asctime: z.string()
	}),
	device: z.object({
		name: z.string(),
		info_name: z.string(),
		type: z.string(),
		protocol: z.string()
	}),
	smart_status: z.object({
		passed: z.boolean()
	}),
	power_on_time: z.object({
		hours: z.number()
	}),
	power_cycle_count: z.number(),
	temperature: z.object({
		current: z.number()
	}),
	ata_smart_attributes: z.object({
		revision: z.number(),
		table: z.array(
			z.object({
				id: z.number(),
				name: z.string(),
				value: z.number(),
				worst: z.number(),
				thresh: z.number(),
				when_failed: z.string(),
				flags: z.object({
					value: z.number(),
					string: z.string(),
					prefailure: z.boolean(),
					updated_online: z.boolean(),
					performance: z.boolean(),
					error_rate: z.boolean(),
					event_count: z.boolean(),
					auto_keep: z.boolean()
				}),
				raw: z.object({
					value: z.number(),
					string: z.string()
				})
			})
		)
	})
});

export const PartitionSchema = z.object({
	name: z.string(),
	usage: z.string(),
	size: z.number()
});

export const DiskInfoSchema = z.array(
	z.object({
		device: z.string(),
		type: z.string(),
		usage: z.string(),
		size: z.number(),
		gpt: z.boolean(),
		model: z.string(),
		serial: z.string(),
		smartData: z.union([SmartNVMESchema, SmartCtlSchema, z.null()]).optional(),
		wearOut: z.union([z.number(), z.string()]).optional(),
		partitions: z.array(PartitionSchema).optional().default([])
	})
);

export type SmartNVME = z.infer<typeof SmartNVMESchema>;
export type DiskInfo = z.infer<typeof DiskInfoSchema>;
export type Partition = z.infer<typeof PartitionSchema>;

export interface Disk {
	Device: string;
	Type: string;
	Usage: string;
	Size: number;
	GPT: boolean;
	Model: string;
	Serial: string;
	'S.M.A.R.T.': string;
	Wearout: string | number | undefined;
	Partitions: Partition[];
}
