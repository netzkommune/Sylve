import { z } from 'zod/v4';

export const JWTClaimsSchema = z.object({
	exp: z.number(),
	jti: z.string(),
	custom_claims: z.object({
		userId: z.number(),
		username: z.string(),
		authType: z.string()
	})
});

export const UserSchema = z.object({
	id: z.number().int(),
	username: z.string(),
	email: z.string(),
	notes: z.string(),
	totp: z.string(),
	admin: z.boolean(),
	createdAt: z.string(),
	updatedAt: z.string(),
	lastLoginTime: z.string()
});

export type JWTClaims = z.infer<typeof JWTClaimsSchema>;
export type User = z.infer<typeof UserSchema>;
