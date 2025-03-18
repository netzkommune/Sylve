import { AuditLogSchema, type AuditLog } from '$lib/types/info/audit';
import { apiRequest } from '$lib/utils/http';
import { getTranslation } from '$lib/utils/i18n';
import { capitalizeFirstLetter } from '$lib/utils/string';

export async function getAuditLogs(): Promise<AuditLog> {
	return await apiRequest('/info/audit-logs', AuditLogSchema, 'GET');
}

export function formatAction(action: string): string {
	if (action.includes('|-|')) {
		const split = action.split('|-|').join(' ');
		const parts = split.split(' ');

		if (parts.length >= 2) {
			const command = parts[0];
			const argument = parts[1];

			const parent = command.split('_').pop();
			const child = command.split('_').shift();

			return (
				capitalizeFirstLetter(getTranslation(`${parent}.${child}`, `${parent}.${child}`)) +
				' ' +
				argument
			);
		}
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
		case 'progress':
			return 'In Progress';
		default:
			return status;
	}
}
