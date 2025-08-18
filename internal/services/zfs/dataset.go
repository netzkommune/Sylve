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

func (s *Service) GetDatasets(t string) ([]*zfsServiceInterfaces.Dataset, error) {
	var datasets []*zfs.Dataset
	var err error

	if t == "" || t == "all" {
		datasets, err = zfs.Datasets("")
	} else if t == "filesystem" {
		datasets, err = zfs.Filesystems("")
	} else if t == "snapshot" {
		datasets, err = zfs.Snapshots("")
	} else if t == "volume" {
		datasets, err = zfs.Volumes("")
	}

	if err != nil {
		return nil, err
	}

	var results []*zfsServiceInterfaces.Dataset

	for _, dataset := range datasets {
		results = append(results, &zfsServiceInterfaces.Dataset{
			Name:          dataset.Name,
			Origin:        dataset.Origin,
			GUID:          dataset.GUID,
			Used:          dataset.Used,
			Avail:         dataset.Avail,
			Mountpoint:    dataset.Mountpoint,
			Compression:   dataset.Compression,
			Type:          dataset.Type,
			Written:       dataset.Written,
			Volsize:       dataset.Volsize,
			VolBlockSize:  dataset.VolBlockSize,
			Logicalused:   dataset.Logicalused,
			Usedbydataset: dataset.Usedbydataset,
			Quota:         dataset.Quota,
			Referenced:    dataset.Referenced,
			Mounted:       dataset.Mounted,
			Checksum:      dataset.Checksum,
			Dedup:         dataset.Dedup,
			ACLInherit:    dataset.ACLInherit,
			ACLMode:       dataset.ACLMode,
			PrimaryCache:  dataset.PrimaryCache,
			VolMode:       dataset.VolMode,
		})
	}

	return results, nil
}

func (s *Service) GetDatasetByGUID(guid string) (*zfs.Dataset, error) {
	datasets, err := zfs.Datasets("")
	if err != nil {
		return nil, err
	}

	for _, dataset := range datasets {
		gguid, err := dataset.GetProperty("guid")
		if err != nil {
			return nil, err
		}

		if gguid == guid {
			return dataset, nil
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
