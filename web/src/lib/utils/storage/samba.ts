function pathFormatter(cell: CellComponent) {
	const row = cell.getRow();
	const share = data.shares.find((s) => s.name === row.getData().share);
	if (share) {
		const dataset = data.datasets.find((d) => d.properties.guid === share.dataset);
		if (dataset?.mountpoint) {
			const path = cell.getValue().replace(dataset.mountpoint, '');
			return path;
		}
	}
	return cell.getValue() || '-';
}
