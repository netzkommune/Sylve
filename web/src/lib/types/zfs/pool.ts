import { z } from 'zod';

export const IODelaySchema = z.object({
	delay: z.number().default(0)
});

export const IODelayHistoricalSchema = z.array(
	z.object({
		id: z.number().default(0),
		delay: z.number().default(0),
		createdAt: z.string().default('')
	})
);

export const RWSchema = z.object({
	read: z.number(),
	write: z.number()
});

export const VdevDeviceSchema = z.object({
	name: z.string(),
	size: z.number(),
	health: z.string()
});

export const ReplacingVdevDeviceSchema = z.object({
	name: z.string(),
	health: z.string(),
	oldDrive: VdevDeviceSchema,
	newDrive: VdevDeviceSchema
});

export const VdevSchema = z.object({
	name: z.string(),
	alloc: z.number(),
	free: z.number(),
	size: z.number(),
	health: z.string(),
	operations: RWSchema,
	bandwidth: RWSchema,
	devices: z.array(VdevDeviceSchema),
	replacingDevices: z.array(ReplacingVdevDeviceSchema).optional()
});

export const ZpoolSchema = z.object({
	name: z.string(),
	health: z.string(),
	allocated: z.number(),
	size: z.number(),
	free: z.number(),
	readOnly: z.boolean(),
	freeing: z.number(),
	leaked: z.number(),
	dedupRatio: z.number(),
	vdevs: z.array(VdevSchema)
});

export const CreateVdevSchema = z.object({
	name: z.string(),
	devices: z.array(z.string())
});

export const CreateZpoolSchema = z.object({
	name: z
		.string()
		.min(1, 'Name must be at least 1 character long')
		.max(24, 'Name must be at most 24 characters long')
		.regex(/^[a-zA-Z0-9]+$/, 'Name must be alphanumeric'),
	raidType: z.enum(['mirror', 'raidz', 'raidz2', 'raidz3']).optional(),
	vdevs: z.array(CreateVdevSchema),
	properties: z.record(z.string()).optional(),
	createForce: z.boolean().default(false)
});

export const ReplaceDeviceSchema = z.object({
	name: z.string(),
	old: z.string(),
	new: z.string()
});

export type IODelay = z.infer<typeof IODelaySchema>;
export type IODelayHistorical = z.infer<typeof IODelayHistoricalSchema>;
export type Zpool = z.infer<typeof ZpoolSchema>;
export type ReplaceDevice = z.infer<typeof ReplaceDeviceSchema>;
export type CreateZpool = z.infer<typeof CreateZpoolSchema>;
