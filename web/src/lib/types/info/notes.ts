import { z } from 'zod';

export const NoteSchema = z.object({
	id: z.number().default(0),
	title: z.string().default(''),
	content: z.string().default(''),
	createdAt: z.string().default('')
});

export const NotesSchema = z.array(NoteSchema);

export type Notes = z.infer<typeof NotesSchema>;
export type Note = z.infer<typeof NoteSchema>;
