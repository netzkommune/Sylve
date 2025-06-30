// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfs

import (
	"fmt"
	vmModels "sylve/internal/db/models/vm"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/pkg/utils"
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

func (s *Service) GetDatasetByGUID(guid string) (*zfsServiceInterfaces.Dataset, error) {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return nil, err
		}

		for _, v := range properties {
			if v == guid {
				propMap := make(map[string]string, len(properties))
				for k, v := range properties {
					propMap[k] = v
				}

				return &zfsServiceInterfaces.Dataset{
					Dataset:    *dataset,
					Properties: propMap,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("dataset with guid %s not found", guid)
}

func (s *Service) BulkDeleteDataset(guids []string) error {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	defer s.Libvirt.RescanStoragePools()

	var count int64
	if err := s.DB.Model(&vmModels.Storage{}).Where("dataset IN ?", guids).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if datasets are in use: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("datasets_in_use_by_vm")
	}

	guidsMap := make(map[string]struct{})
	for _, guid := range guids {
		guidsMap[guid] = struct{}{}
	}

	datasets, err := zfs.Datasets("")
	if err != nil {
		return err
	}

	matched := make(map[string]*zfs.Dataset)

	for _, dataset := range datasets {
		properties, err := dataset.GetAllProperties()
		if err != nil {
			return err
		}

		for _, v := range properties {
			if _, ok := guidsMap[v]; ok {
				matched[v] = dataset
				delete(guidsMap, v)
				break
			}
		}
	}

	if len(guidsMap) > 0 {
		return fmt.Errorf("datasets with guids %v not found", utils.MapKeys(guidsMap))
	}

	for guid, dataset := range matched {
		if err := dataset.Destroy(zfs.DestroyDefault); err != nil {
			return fmt.Errorf("failed to delete dataset with guid %s: %w", guid, err)
		}
	}

	return nil
}

func (s *Service) IsDatasetInUse(guid string, failEarly bool) bool {
	var count int64
	if err := s.DB.Model(&vmModels.Storage{}).Where("dataset = ?", guid).
		Count(&count).Error; err != nil {
		return false
	}

	if count > 0 {
		if failEarly {
			return true
		}

		var storage vmModels.Storage

		if err := s.DB.Model(&vmModels.Storage{}).Where("dataset = ?", guid).
			First(&storage).Error; err != nil {
			return false
		}

		if storage.VMID > 0 {
			var vm vmModels.VM
			if err := s.DB.Model(&vmModels.VM{}).Where("id = ?", storage.VMID).
				First(&vm).Error; err != nil {
				return false
			}

			domain, err := s.Libvirt.GetLvDomain(vm.VmID)
			if err != nil {
				return false
			}

			if domain != nil {
				if domain.Status == "Running" || domain.Status == "Paused" {
					return true
				} else {
					return false
				}
			}
		}
	}

	return false
}
