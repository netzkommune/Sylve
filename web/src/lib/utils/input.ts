export function generateComboboxOptions(values: string[]) {
	return values.map((option) => ({
		label: option,
		value: option
	}));
}
