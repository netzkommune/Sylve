import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { NotesSchema, type Notes } from '$lib/types/info/notes';
import { apiRequest } from '$lib/utils/http';

async function notesRequest(
	endpoint: string,
	method: 'GET' | 'POST' | 'PUT' | 'DELETE',
	body?: object
): Promise<Notes | APIResponse> {
	const schema = method === 'GET' ? NotesSchema : APIResponseSchema;
	const data = await apiRequest(endpoint, schema, method, body);
	return schema.parse(data);
}

export const getNotes = () => notesRequest('/info/notes', 'GET');
export const createNote = (title: string, content: string) =>
	notesRequest('/info/notes', 'POST', { title, content });
export const deleteNote = (id: number) => notesRequest(`/info/notes/${id}`, 'DELETE');
export const updateNote = (id: number, title: string, content: string) =>
	notesRequest(`/info/notes/${id}`, 'PUT', { title, content });
