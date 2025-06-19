import { getBasicInfo } from '$lib/api/info/basic';
import { getCPUInfo } from '$lib/api/info/cpu';
import { getRAMInfo, getSwapInfo } from '$lib/api/info/ram';
import { getIODelay } from '$lib/api/zfs/pool';
import { SEVEN_DAYS } from '$lib/utils';
import { cachedFetch } from '$lib/utils/http';
import { getTotalDiskUsage } from '$lib/utils/zfs';

export async function load() {
	const cacheDuration = SEVEN_DAYS;
	const [
		basicInfo,
		cpuInfo,
		cpuInfoHistorical,
		ramInfo,
		ramInfoHistorical,
		swapInfo,
		swapInfoHistorical,
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
		cachedFetch(
			'ramInfoHistorical',
			() => getRAMInfo({ queryKey: ['ramInfoHistorical'], meta: undefined }),
			cacheDuration
		),
		cachedFetch('swapInfo', getSwapInfo, cacheDuration),
		cachedFetch(
			'swapInfoHistorical',
			() => getSwapInfo({ queryKey: ['swapInfoHistorical'], meta: undefined }),
			cacheDuration
		),
		cachedFetch(
			'ioDelay',
			() => getIODelay({ queryKey: ['ioDelay'], meta: undefined }),
			cacheDuration
		),
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
		ramInfoHistorical,
		swapInfo,
		swapInfoHistorical,
		ioDelay,
		ioDelayHistorical,
		totalDiskUsage
	};
}
