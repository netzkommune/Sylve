import { z } from 'zod';

export const RAMInfoSchema = z.object({
	total: z.number().default(0),
	free: z.number().default(0),
	usedPercent: z.number().default(0)
});

export const RAMInfoResponseSchema = z.object({
	info: RAMInfoSchema
});

export type RAMInfo = z.infer<typeof RAMInfoSchema>;
