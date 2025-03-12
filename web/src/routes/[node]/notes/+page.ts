import { createNote, deleteNote, getNotes } from '$lib/api/info/notes';
import type { Notes } from '$lib/types/info/notes';
import { getTranslation } from '$lib/utils/i18n';
import { toast } from 'svelte-french-toast';

export async function load() {
	let notes = await getNotes();

	if (!Array.isArray(notes) && notes.status === 'error') {
		const d = getTranslation(notes.message || 'Error loading notes', 'Error loading notes');
		toast.error(d, {
			position: 'bottom-center'
		});

		notes = [];
	}

	return {
		notes: notes as Notes
	};
}
