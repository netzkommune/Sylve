package zfs

import (
	"fmt"
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

func (s *Service) DeleteSnapshot(guid string) error {
	datasets, err := zfs.Snapshots("")

	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for _, v := range properties {
			if v == guid {
				err := dataset.Destroy(zfs.DestroyDefault)

				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	return fmt.Errorf("snapshot with guid %s not found", guid)
}

func (s *Service) CreateSnapshot(guid string, name string, recursive bool) error {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for k, v := range properties {
			if k == "guid" {
				if v == guid {
					shot, err := dataset.Snapshot(name, recursive)
					if err != nil {
						return err
					}

					if shot.Name == dataset.Name+"@"+name {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("dataset with guid %s not found", guid)
}

func (s *Service) CreateFilesystem(name string, props map[string]string) error {
	// find parent from props
	// if parent is not found, return error
	// parent := ""

	// for k, v := range props {
	// 	if k == "parent" {
	// 		parent = v
	// 		break
	// 	}
	// }

	// if parent == "" {
	// 	return fmt.Errorf("parent_not_found")
	// }

	// dataset, err := zfs.Datasets(name)
	// if err != nil {
	// 	return err
	// }

	// if dataset.Name == name {
	// 	return nil
	// }

	// return fmt.Errorf("failed to create filesystem %s", name)

	return nil
}
