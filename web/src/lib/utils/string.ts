export function capitalizeFirstLetter(str: string): string {
	return str.length > 0 ? str[0].toLocaleUpperCase() + str.slice(1) : str;
}
