export function generateComboboxOptions(values: string[], additional?: string[]) {
	const combined = [...values, ...(additional ?? [])];
	return combined.map((option) => ({
		label: option,
		value: option
	}));
}
