import { NotesSchema, type Notes } from '$lib/types/info/notes';
import { apiRequest } from '$lib/utils/http';

async function notesRequest(
	endpoint: string,
	method: 'GET' | 'POST' | 'PUT' | 'DELETE',
	body?: object
): Promise<Notes> {
	const data = await apiRequest(endpoint, NotesSchema, method, body);
	return NotesSchema.parse(data);
}

export const getNotes = () => notesRequest('/info/notes', 'GET');
export const createNote = (title: string, content: string) =>
	notesRequest('/info/notes', 'POST', { title, content });
export const deleteNote = (id: number) => notesRequest(`/info/notes/${id}`, 'DELETE');
export const updateNote = (id: number, title: string, content: string) =>
	notesRequest(`/info/notes/${id}`, 'PUT', { title, content });
