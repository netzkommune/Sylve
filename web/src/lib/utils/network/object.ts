import type { NetworkObject } from '$lib/types/network/object';
import { isValidIPv4, isValidIPv6 } from '../string';

export function generateIPOptions(
	networkObjects: NetworkObject[] | undefined,
	type: string
): { label: string; value: string }[] {
	if (!networkObjects || networkObjects.length === 0) {
		return [];
	}

	const options = [] as { label: string; value: string }[];
	const objects = networkObjects?.filter((obj) => obj.type === 'Host');
	if (!objects || objects.length === 0) {
		return [];
	}

	for (const object of objects) {
		if (object.entries && object.entries.length === 1) {
			for (const entry of object.entries) {
				const validator = type == 'IPv4' ? isValidIPv4 : isValidIPv6;
				if (validator(entry.value)) {
					options.push({
						label: `${object.name} (${entry.value})`,
						value: object.id.toString()
					});
				}
			}
		}
	}

	return options;
}

export function generateNetworkOptions(
	networkObjects: NetworkObject[] | undefined,
	type: string
): { label: string; value: string }[] {
	if (!networkObjects || networkObjects.length === 0) {
		return [];
	}

	const options = [] as { label: string; value: string }[];
	const objects = networkObjects?.filter((obj) => obj.type === 'Network');
	if (!objects || objects.length === 0) {
		return [];
	}

	for (const object of objects) {
		if (object.entries && object.entries.length > 0) {
			for (const entry of object.entries) {
				if (type === 'IPv4' && isValidIPv4(entry.value, true)) {
					options.push({
						label: `${object.name} (${entry.value})`,
						value: object.id.toString()
					});
				} else if (type === 'IPv6' && isValidIPv6(entry.value, true)) {
					options.push({
						label: `${object.name} (${entry.value})`,
						value: object.id.toString()
					});
				}
			}
		}
	}

	return options;
}

export function generateMACOptions(
	networkObjects: NetworkObject[] | undefined
): { label: string; value: string }[] {
	if (!networkObjects || networkObjects.length === 0) {
		return [];
	}

	const options = [] as { label: string; value: string }[];
	const objects = networkObjects?.filter((obj) => obj.type === 'Mac');
	if (!objects || objects.length === 0) {
		return [];
	}

	for (const object of objects) {
		if (object.entries && object.entries.length > 0) {
			for (const entry of object.entries) {
				options.push({
					label: `${object.name} (${entry.value})`,
					value: object.id.toString()
				});
			}
		}
	}

	return options;
}
