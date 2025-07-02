import { AuditLogSchema, type AuditLog } from '$lib/types/info/audit';
import { apiRequest } from '$lib/utils/http';

export async function getAuditLogs(): Promise<AuditLog> {
	return await apiRequest('/info/audit-logs', AuditLogSchema, 'GET');
}

export function formatAction(action: string): string {
	return action;
}

export function formatStatus(status: string): string {
	switch (status) {
		case 'started':
			return 'Started';
		case 'success':
			return 'OK';
		case 'failure':
			return 'Failed';
		case 'failed':
			return 'Failed';
		case 'progress':
			return 'In Progress';
		default:
			return status;
	}
}
