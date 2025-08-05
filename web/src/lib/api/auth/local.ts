import { UserSchema, type User } from '$lib/types/auth';
import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';
import { z } from 'zod/v4';

export async function listUsers(): Promise<User[]> {
	return await apiRequest('/auth/users', z.array(UserSchema), 'GET');
}

export async function createUser(
	username: string,
	email: string,
	password: string,
	admin: boolean
): Promise<APIResponse> {
	const body = {
		username,
		password,
		email,
		admin
	};

	return await apiRequest('/auth/users', APIResponseSchema, 'POST', body);
}

export async function deleteUser(id: string): Promise<APIResponse> {
	return await apiRequest(`/auth/users/${id}`, APIResponseSchema, 'DELETE');
}

export async function editUser(
	id: number,
	username: string,
	email: string,
	password: string,
	admin: boolean
): Promise<APIResponse> {
	const body = {
		id,
		username,
		email,
		password,
		admin
	};

	return await apiRequest('/auth/users', APIResponseSchema, 'PUT', body);
}
