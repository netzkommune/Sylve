/**
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Copyright (c) 2025 The FreeBSD Foundation.
 *
 * This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
 * of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
 * under sponsorship from the FreeBSD Foundation.
 */

import humanFormat, { type ScaleLike } from 'human-format';
import {
	TabulatorFull,
	type CellComponent,
	type RowComponent,
	type Tabulator
} from 'tabulator-tables';
import { iconCache } from './icons';

export function deselectAllRows(table: Tabulator | null) {
	if (table) {
		table.getRows().forEach((row) => {
			console.log(row);
			row.deselect();
		});
	}
}

export function selectOneRow(table: Tabulator | null, row: RowComponent) {
	if (table) {
		deselectAllRows(table);
		row.select();
	}
}

export function getTable(id: string): Tabulator | null {
	const table = TabulatorFull.findTable(`#${id}`);
	if (table) {
		return table[0];
	}

	return null;
}

export function deleteRowByFieldValue(tableId: string, field: string, value: string | number) {
	try {
		const table = getTable(tableId);
		if (table) {
			const rows = table.getRows();
			for (const row of rows) {
				if (row.getData()[field] === value) {
					row.delete();
				}

				if (row.getData()['children']) {
					if (Array.isArray(row.getData()['children'])) {
						const children = row.getData()['children'];
						for (const child of children) {
							if (child[field] === value) {
								row.delete();
							}
						}
					}
				}
			}
		}
	} catch (error) {
		console.error('Error deleting row:', error);
	}
}

export const renderWithIcon = (iconKey: string, suffix: string, extraClass?: string) => {
	const icon = iconCache[iconKey] || '';
	return `
                        <span class="inline-flex items-center gap-1 whitespace-nowrap text-ellipsis overflow-hidden">
                            <span class="shrink-0 ${extraClass || ''}">${icon}</span>
                            <span>${suffix}</span>
                        </span>
                    `.trim();
};

export function sizeFormatter(cell: CellComponent) {
	try {
		const sizeOptions = {
			scale: 'binary' as ScaleLike,
			unit: 'B',
			maxDecimals: 1
		};

		return humanFormat(cell.getValue(), sizeOptions);
	} catch (e) {
		return cell.getValue();
	}
}
