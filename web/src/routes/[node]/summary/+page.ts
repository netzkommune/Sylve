import { getBasicInfo } from '$lib/api/info/basic';
import { getCPUInfo } from '$lib/api/info/cpu';
import { getRAMInfo, getSwapInfo } from '$lib/api/info/ram';
import { getIODelay } from '$lib/api/zfs/pool';
import { cachedFetch } from '$lib/utils/http';
import { getTotalDiskUsage } from '$lib/utils/zfs';

export async function load() {
	const cacheDuration = 360 * 1000;
	const [
		basicInfo,
		cpuInfo,
		cpuInfoHistorical,
		ramInfo,
		swapInfo,
		ioDelay,
		ioDelayHistorical,
		totalDiskUsage
	] = await Promise.all([
		cachedFetch('basicInfo', getBasicInfo, cacheDuration),
		cachedFetch('cpuInfo', getCPUInfo, cacheDuration),
		cachedFetch(
			'cpuInfoHistorical',
			() =>
				getCPUInfo({
					queryKey: ['cpuInfoHistorical'],
					meta: undefined
				}),
			cacheDuration
		),
		cachedFetch('ramInfo', getRAMInfo, cacheDuration),
		cachedFetch('swapInfo', getSwapInfo, cacheDuration),
		cachedFetch('ioDelay', getIODelay, cacheDuration),
		cachedFetch(
			'ioDelayHistorical',
			() => getIODelay({ queryKey: ['ioDelayHistorical'], meta: undefined }),
			cacheDuration
		),
		cachedFetch('totalDiskUsage', getTotalDiskUsage, cacheDuration)
	]);

	return {
		basicInfo,
		cpuInfo,
		cpuInfoHistorical,
		ramInfo,
		swapInfo,
		ioDelay,
		ioDelayHistorical,
		totalDiskUsage
	};
}
