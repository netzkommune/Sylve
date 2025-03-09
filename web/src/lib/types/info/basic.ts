import { z } from 'zod';

export const BasicInfoSchema = z.object({
	hostname: z.string().default('Unknown'),
	os: z.string().default('Unknown'),
	uptime: z.number().default(0),
	loadAverage: z.string().default('Unknown'),
	bootMode: z.string().default('Unknown'),
	sylveVersion: z.string().default('Unknown')
});

export type BasicInfo = z.infer<typeof BasicInfoSchema>;
