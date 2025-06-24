import { z } from 'zod/v4';

export const NetworkInterfaceInfoSchema = z.object({
	id: z.number().default(0),
	name: z.string().default(''),
	flags: z.string().default(''),
	network: z.string().default(''),
	address: z.string().default(''),
	receivedPackets: z.number().int().default(0),
	receivedErrors: z.number().int().default(0),
	droppedPackets: z.number().int().default(0),
	receivedBytes: z.number().int().default(0),
	sentPackets: z.number().int().default(0),
	sendErrors: z.number().int().default(0),
	sentBytes: z.number().int().default(0),
	collisions: z.number().int().default(0),
	createdAt: z.string(),
	updatedAt: z.string()
});

export const HistoricalNetworkInterfaceSchema = z.object({
	sentBytes: z.number().int().default(0),
	receivedBytes: z.number().int().default(0),
	createdAt: z.string()
});

export type HistoricalNetworkInterface = z.infer<typeof HistoricalNetworkInterfaceSchema>;
export type NetworkInterfaceInfo = z.infer<typeof NetworkInterfaceInfoSchema>;
