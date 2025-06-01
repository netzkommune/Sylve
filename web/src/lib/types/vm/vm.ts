import { z } from 'zod/v4';

export interface CreateData {
	name: string;
	id: number;
	description: string;
	storage: {
		type: string;
		guid: string;
		size: number;
		emulation: string;
		iso: string;
	};
	network: {
		switch: number;
		mac: string;
		emulation: string;
	};
	hardware: {
		sockets: number;
		cores: number;
		threads: number;
		memory: number;
		passthroughIds: number[];
	};
	advanced: {
		vncPort: number;
		vncPassword: string;
		vncWait: boolean;
		vncResolution: string;
		startAtBoot: boolean;
		bootOrder: number;
	};
}

export const VMStorageSchema = z.object({
	id: z.number().int(),
	type: z.string(),
	dataset: z.string(),
	size: z.number().int(),
	emulation: z.string(),
	vmId: z.number().int().optional()
});

export const VMNetworkSchema = z.object({
	id: z.number().int(),
	mac: z.string(),
	switchId: z.number().int(),
	emulation: z.string(),
	vmId: z.number().int().optional()
});

export const VMSchema = z.object({
	id: z.number().int(),
	name: z.string(),
	description: z.string(),
	vmId: z.number().int(),
	cpuSockets: z.number().int(),
	cpuCores: z.number().int(),
	cpuThreads: z.number().int(),
	ram: z.number().int(),
	vncPort: z.number().int(),
	vncPassword: z.string(),
	vncResolution: z.string(),
	vncWait: z.boolean(),
	startAtBoot: z.boolean(),
	startOrder: z.number().int(),

	storages: z.array(VMStorageSchema),
	networks: z.array(VMNetworkSchema),
	pciDevices: z.array(z.number().int()),

	createdAt: z.string(),
	updatedAt: z.string()
});

export const VMDomainSchema = z.object({
	id: z.number().int(),
	uuid: z.string(),
	name: z.string(),
	status: z.string()
});

export type VM = z.infer<typeof VMSchema>;
export type VMStorage = z.infer<typeof VMStorageSchema>;
export type VMNetwork = z.infer<typeof VMNetworkSchema>;
export type VMDomain = z.infer<typeof VMDomainSchema>;
