import { AuditLogSchema, type AuditLog } from '$lib/types/info/audit';
import { apiRequest } from '$lib/utils/http';
import { getTranslation } from '$lib/utils/i18n';
import { capitalizeFirstLetter } from '$lib/utils/string';

export async function getAuditLogs(): Promise<AuditLog> {
	return await apiRequest('/info/audit-logs', AuditLogSchema, 'GET');
}

export function formatAction(action: string): string {
	if (action.includes('|-|')) {
		const parts = action.split('|-|');
		return capitalizeFirstLetter(getTranslation(parts[0], parts[0]), true) + ' ' + parts[1];
	}

	switch (action) {
		case 'login':
			return getTranslation('auth.login', 'Login');
		case 'revoke_token':
			return getTranslation('auth.logout', 'Logout');
		default:
			return action;
	}
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
