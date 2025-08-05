import { z } from 'zod/v4';

export interface CreateData {
	name: string;
	id: number;
	description: string;
	storage: {
		dataset: string;
		base: string;
	};
	network: {
		switch: number;
		mac: number;
		ipv4: number;
		ipv4Gateway: number;
		ipv6: number;
		ipv6Gateway: number;
		dhcp: boolean;
		slaac: boolean;
	};
	hardware: {
		cpuCores: number;
		ram: number;
		startAtBoot: boolean;
		bootOrder: number;
	};
}

export const SimpleJailSchema = z.object({
	id: z.number().int(),
	name: z.string(),
	ctId: z.number().int()
});

export const NetworkSchema = z.object({
	switchId: z.number().int(),
	macId: z.number().int().optional(),
	ipv4Id: z.number().int().optional(),
	ipv4GwId: z.number().int().optional(),
	ipv6Id: z.number().int().optional(),
	ipv6GwId: z.number().int().optional(),
	ctId: z.number().int()
});

export const JailSchema = SimpleJailSchema.extend({
	description: z.string().nullable(),
	dataset: z.string(),
	base: z.string(),
	startAtBoot: z.boolean(),
	startOrder: z.number().int(),
	networks: z.array(NetworkSchema).optional().default([]),
	createdAt: z.string(),
	updatedAt: z.string()
});

export type SimpleJail = z.infer<typeof SimpleJailSchema>;
export type Jail = z.infer<typeof JailSchema>;
