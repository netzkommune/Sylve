import { AuditLogSchema, type AuditLog } from '$lib/types/info/audit';
import { apiRequest } from '$lib/utils/http';

export async function getAuditLogs(): Promise<AuditLog> {
	const data = await apiRequest('/info/audit-logs', AuditLogSchema, 'GET');
	return AuditLogSchema.parse(data);
}

export function formatAction(action: string): string {
	switch (action) {
		case 'login':
			return 'Login';
		case 'revoke_token':
			return 'Logout';
		default:
			return action;
	}
}

export function formatStatus(status: string): string {
	switch (status) {
		case 'success':
			return 'OK';
		case 'failure':
			return 'Failed';
		case 'progress':
			return 'In Progress';
		default:
			return status;
	}
}
