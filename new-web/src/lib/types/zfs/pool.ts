import { z } from 'zod/v4';

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

export const ZpoolDeviceSchema: z.ZodType<any> = z.lazy(() =>
	z.object({
		name: z.string(),
		state: z.string(),
		read: z.number(),
		write: z.number(),
		cksum: z.number(),
		note: z.string(),
		children: z.array(ZpoolDeviceSchema).optional().default([])
	})
);

export const ZpoolStatusSchema = z.object({
	name: z.string(),
	state: z.string(),
	status: z.string(),
	action: z.string(),
	scan: z.string(),
	devices: z.array(ZpoolDeviceSchema).optional().default([]),
	errors: z.string()
});

export const ZpoolSpareSchema = z.object({
	name: z.string(),
	size: z.number(),
	health: z.string()
});

export const ZpoolPropertySchema = z.object({
	property: z.string(),
	value: z.string(),
	source: z.string()
});

export const ZpoolSchema = z.object({
	id: z.string(),
	name: z.string(),
	guid: z.string(),
	health: z.string(),
	allocated: z.number(),
	size: z.number(),
	free: z.number(),
	readOnly: z.boolean(),
	freeing: z.number(),
	leaked: z.number(),
	dedupRatio: z.number(),
	vdevs: z.array(VdevSchema),
	properties: z.array(ZpoolPropertySchema).optional().default([]),
	status: ZpoolStatusSchema,
	spares: z.array(ZpoolSpareSchema).optional().default([])
});

export const CreateVdevSchema = z.object({
	name: z.string(),
	devices: z.array(z.string())
});

export const ZpoolRaidTypeSchema = z.union([
	z.enum(['mirror', 'raidz', 'raidz2', 'raidz3', 'stripe']),
	z.undefined()
]);

export const CreateZpoolSchema = z.object({
	name: z
		.string()
		.min(1, 'Name must be at least 1 character long')
		.max(24, 'Name must be at most 24 characters long')
		.regex(/^[a-zA-Z0-9]+$/, 'Name must be alphanumeric'),
	raidType: ZpoolRaidTypeSchema,
	vdevs: z.array(CreateVdevSchema),
	properties: z.record(z.string(), z.string()).optional(),
	createForce: z.boolean().default(false),
	spares: z.array(z.string()).optional()
});

export const ReplaceDeviceSchema = z.object({
	name: z.string(),
	old: z.string(),
	new: z.string()
});

export const PoolStatPointSchema = z.object({
	allocated: z.number(),
	free: z.number(),
	size: z.number(),
	dedupRatio: z.number(),
	time: z.number()
});

export const PoolStatPointsSchema = z.record(
	z.string(),
	z
		.array(PoolStatPointSchema)
		.refine((obj) => Object.keys(obj).length > 0, { message: 'No Data Found' })
);

export const PoolStatPointsResponseSchema = z.object({
	poolStatPoint: PoolStatPointsSchema,
	intervalMap: z.array(
		z.object({ value: z.number().transform((v) => v.toString()), label: z.string() })
	)
});

export type IODelay = z.infer<typeof IODelaySchema>;
export type IODelayHistorical = z.infer<typeof IODelayHistoricalSchema>;
export type Zpool = z.infer<typeof ZpoolSchema>;
export type ReplaceDevice = z.infer<typeof ReplaceDeviceSchema>;
export type CreateZpool = z.infer<typeof CreateZpoolSchema>;
export type ZpoolRaidType = z.infer<typeof ZpoolRaidTypeSchema>;
export type PoolStatPointsResponse = z.infer<typeof PoolStatPointsResponseSchema>;
