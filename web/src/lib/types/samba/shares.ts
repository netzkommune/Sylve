import { z } from 'zod/v4';
import { GroupSchema } from '../auth';

export const SambaShareSchema = z.object({
	id: z.number(),
	name: z.string(),
	dataset: z.string(),
	readOnlyGroups: z.preprocess((val) => (val == null ? [] : val), z.array(GroupSchema)),
	writeableGroups: z.preprocess((val) => (val == null ? [] : val), z.array(GroupSchema)),
	createMask: z.string(),
	directoryMask: z.string(),
	guestOk: z.boolean(),
	readOnly: z.boolean(),
	createdAt: z.string(),
	updatedAt: z.string()
});

export type SambaShare = z.infer<typeof SambaShareSchema>;
