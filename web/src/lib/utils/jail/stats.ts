import type { Jail, JailStat } from '$lib/types/jail/jail';

export function cleanStats(
	stats: JailStat[],
	jail: Jail
): { cpu: { value: number; date: string }[]; memory: { value: number; date: string }[] } {
	const result = stats.map((stat) => {
		const cpu = parseFloat(Math.min((stat.cpuUsage * 100) / (100 * jail.cores), 100).toFixed(2));
		const memory = stat.memoryUsage;

		return { cpu, memory, createdAt: stat.createdAt };
	});

	return {
		cpu: result.map((r) => ({ value: r.cpu, date: r.createdAt })),
		memory: result.map((r) => ({ value: r.memory, date: r.createdAt }))
	};
}
