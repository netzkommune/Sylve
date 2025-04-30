import { z } from 'zod';

export const FlagsSchema = z.object({
	raw: z.number(),
	desc: z.array(z.string()).default([])
});

export const IPv4Schema = z.object({
	ip: z.string(),
	netmask: z.string(),
	broadcast: z.string()
});

export const IPv6Schema = z.object({
	ip: z.string(),
	prefixLength: z.number(),
	scopeId: z.number(),
	autoConf: z.boolean(),
	detached: z.boolean(),
	deprecated: z.boolean(),
	lifeTimes: z.object({
		preferred: z.number(),
		valid: z.number()
	})
});

export const MediaSchema = z.object({
	type: z.string(),
	subtype: z.string(),
	options: z.array(z.string()).default([]),
	mode: z.string(),
	rawCurrent: z.number(),
	rawActive: z.number(),
	status: z.string()
});

export const ND6Schema = FlagsSchema;

export const IfaceSchema = z.object({
	name: z.string(),
	ether: z.string().default(''),
	flags: FlagsSchema,
	mtu: z.number().default(0),
	metric: z.number().default(0),
	capabilities: z.object({
		enabled: FlagsSchema,
		supported: FlagsSchema
	}),
	driver: z.string().default(''),
	ipv4: z.array(IPv4Schema).default([]),
	ipv6: z.array(IPv6Schema).default([]),
	media: MediaSchema.nullable().optional(),
	nd6: ND6Schema.nullable().optional()
});

export type Iface = z.infer<typeof IfaceSchema>;
