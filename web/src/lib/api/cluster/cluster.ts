import { ClusterDetailsSchema, type ClusterDetails } from '$lib/types/cluster/cluster';
import { APIResponseSchema, type APIResponse } from '$lib/types/common';
import { apiRequest } from '$lib/utils/http';

export async function getDetails(): Promise<ClusterDetails> {
	return await apiRequest('/cluster', ClusterDetailsSchema, 'GET');
}

export async function createCluster(ip: string, port: number): Promise<APIResponse> {
	return await apiRequest('/cluster', APIResponseSchema, 'POST', {
		ip: ip,
		port: port
	});
}

export async function joinCluster(
	nodeId: string,
	nodeIp: string,
	nodePort: number,
	leaderApi: string,
	clusterKey: string
): Promise<APIResponse> {
	return await apiRequest('/cluster/join', APIResponseSchema, 'POST', {
		nodeId: nodeId,
		nodeIp: nodeIp,
		nodePort: nodePort,
		leaderApi: leaderApi,
		clusterKey: clusterKey
	});
}
