import { createNote, deleteNote, getNotes } from '$lib/api/info/notes';
import { toast } from 'svelte-french-toast';

export async function load() {
	const notes = await getNotes();

	if (!Array.isArray(notes) && notes.status === 'error') {
		toast.error(notes.message);
	}

	return {
		notes
	};
}
