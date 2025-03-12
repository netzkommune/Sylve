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

export const ZpoolSchema = z.object({
	name: z.string(),
	health: z.string(),
	allocated: z.number(),
	size: z.number(),
	free: z.number(),
	readOnly: z.boolean(),
	freeing: z.number(),
	leaked: z.number(),
	dedupRatio: z.number()
});

export type IODelay = z.infer<typeof IODelaySchema>;
export type IODelayHistorical = z.infer<typeof IODelayHistoricalSchema>;
export type Zpool = z.infer<typeof ZpoolSchema>;
