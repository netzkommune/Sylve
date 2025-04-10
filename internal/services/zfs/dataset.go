package zfs

import (
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/pkg/zfs"
)

func (s *Service) GetDatasets() ([]zfsServiceInterfaces.Dataset, error) {
	var results []zfsServiceInterfaces.Dataset

	datasets, err := zfs.Datasets("")
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		props, err := dataset.GetAllProperties()
		if err != nil {
			return nil, err
		}

		propMap := make(map[string]string, len(props))
		for k, v := range props {
			propMap[k] = v
		}

		results = append(results, zfsServiceInterfaces.Dataset{
			Dataset:    *dataset,
			Properties: propMap,
		})
	}

	return results, nil
}
