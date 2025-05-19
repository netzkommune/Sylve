import { z } from 'zod/v4';

export const AuditLogSchema = z.array(
	z.object({
		id: z.number(),
		userId: z.number(),
		user: z.string(),
		authType: z.string(),
		node: z.string(),
		started: z.string(),
		ended: z.string(),
		action: z.string(),
		status: z.string(),
		createdAt: z.string(),
		updatedAt: z.string()
	})
);

export type AuditLog = z.infer<typeof AuditLogSchema>;
