import { z } from 'zod';

export const NotesSchema = z.array(
	z.object({
		id: z.number().default(0),
		title: z.string().default(''),
		content: z.string().default(''),
		createdAt: z.string().default('')
	})
);

export type Notes = z.infer<typeof NotesSchema>;
export type Note = z.infer<typeof NotesSchema>['0'];
