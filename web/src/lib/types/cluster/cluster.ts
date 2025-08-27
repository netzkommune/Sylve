import { z } from 'zod/v4';

export const ClusterSchema = z.object({
	id: z.number(),
	enabled: z.boolean(),
	key: z.string(),
	raftBootstrap: z.boolean().nullable(),
	raftIP: z.string(),
	raftPort: z.number().min(0).max(65535).optional()
});

export const RaftNodeSchema = z.object({
	id: z.string(),
	address: z.string(),
	suffrage: z.string(),
	isLeader: z.boolean()
});

export const ClusterDetailsSchema = z.object({
	cluster: ClusterSchema,
	nodeId: z.string(),
	nodes: z.array(RaftNodeSchema).default([]),
	leaderId: z.string().optional(),
	leaderAddress: z.string().optional(),
	partial: z.boolean()
});

export type Cluster = z.infer<typeof ClusterSchema>;
export type RaftNode = z.infer<typeof RaftNodeSchema>;
export type ClusterDetails = z.infer<typeof ClusterDetailsSchema>;
