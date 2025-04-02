export function createEmptyArrayOfArrays(length: number): Array<Array<any>> {
	return Array.from({ length }, () => []);
}
