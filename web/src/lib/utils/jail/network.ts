import type { NetworkObject } from '$lib/types/network/object';

export function ipGatewayFormatter(
	networkObjects: NetworkObject[],
	ipId: number,
	ipGwId: number
): string {
	const found = networkObjects.find((obj) => obj.entries?.some((entry) => entry.objectId === ipId));
	const entry = found?.entries?.[0];

	const foundGw = networkObjects.find((obj) =>
		obj.entries?.some((entry) => entry.objectId === ipGwId)
	);
	const gwEntry = foundGw?.entries?.[0];

	return `${found?.name || '0'} (${entry?.value || '0'}) <br> ${foundGw?.name || '0'} (${gwEntry?.value || '0'})`;
}
