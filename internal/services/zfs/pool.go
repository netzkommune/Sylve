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
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/pkg/zfs"
)

func (s *Service) GetTotalIODelayHisorical() ([]infoModels.IODelay, error) {
	historicalData, err := db.GetHistorical[infoModels.IODelay](s.DB, 128)

	if err != nil {
		return nil, err
	}

	return historicalData, nil
}

func (s *Service) CreatePool(pool zfsServiceInterfaces.Zpool) error {
	if !zfs.IsValidPoolName(pool.Name) {
		return fmt.Errorf("invalid_pool_name")
	}

	_, err := zfs.GetZpool(pool.Name)
	if err == nil {
		return fmt.Errorf("pool_name_taken")
	}

	if pool.RaidType != "" {
		validRaidTypes := map[string]int{
			"mirror": 2,
			"raidz":  3,
			"raidz2": 4,
			"raidz3": 5,
		}
		minDevices, ok := validRaidTypes[pool.RaidType]
		if !ok {
			return fmt.Errorf("invalid_raidz_type")
		}

		for _, vdev := range pool.Vdevs {
			if len(vdev.VdevDevices) < minDevices {
				return fmt.Errorf("vdev %s has insufficient devices for %s (minimum %d)", vdev.Name, pool.RaidType, minDevices)
			}
		}
	} else {
		for _, vdev := range pool.Vdevs {
			if len(vdev.VdevDevices) == 0 {
				return fmt.Errorf("vdev %s has no devices", vdev.Name)
			}
		}
	}

	var args []string
	if pool.CreateForce {
		args = append(args, "-f")
	}

	var vdevArgs []string
	for _, vdev := range pool.Vdevs {
		if pool.RaidType != "" {
			vdevArgs = append(vdevArgs, pool.RaidType)
		}
		vdevArgs = append(vdevArgs, vdev.VdevDevices...)
	}
	args = append(args, vdevArgs...)

	if len(pool.Spares) > 0 {
		args = append(args, "spare")
		args = append(args, pool.Spares...)
	}

	_, err = zfs.CreateZpool(pool.Name, pool.Properties, args...)
	if err != nil {
		return fmt.Errorf("zpool_create_failed: %v", err)
	}

	return nil
}
