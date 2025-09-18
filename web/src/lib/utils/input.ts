/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

export function generateComboboxOptions(values: string[], additional?: string[]) {
	const combined = [...values, ...(additional ?? [])];
	return combined
		.map((option) => ({
			label: option,
			value: option
		}))
		.filter((option, index, self) => self.findIndex((o) => o.value === option.value) === index);
}
