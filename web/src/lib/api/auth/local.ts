import { UserSchema, type User } from '$lib/types/auth';
import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';
import { z } from 'zod/v4';

export async function listUsers(): Promise<User[]> {
	return await apiRequest('/auth/users', z.array(UserSchema), 'GET');
}
