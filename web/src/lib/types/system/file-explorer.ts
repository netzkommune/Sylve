import { z } from 'zod/v4';

export const FileNodeSchema = z.object({
	id: z.string(),
	date: z.string().transform((s) => new Date(s)),
	type: z.string(),
	lazy: z.boolean().optional(),
	size: z.number().min(0).optional()
});

export type FileNode = z.infer<typeof FileNodeSchema>;
