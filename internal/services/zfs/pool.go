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
	"sort"
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

	if err := s.Libvirt.CreateStoragePool(pool.Name); err != nil {
		return fmt.Errorf("libvirt_create_pool_failed: %v", err)
	}

	return nil
}

func (s *Service) DeletePool(poolName string) error {
	err := zfs.DestroyPool(poolName)

	if err != nil {
		return err
	}

	if err := s.Libvirt.DeleteStoragePool(poolName); err != nil {
		return err
	}

	return nil
}

func (s *Service) SyncToLibvirt() error {
	return nil
}

func (s *Service) GetZpoolHistoricalStats(intervalMinutes int, limit int) (map[string][]zfsServiceInterfaces.PoolStatPoint, error) {
	if intervalMinutes <= 0 {
		return nil, fmt.Errorf("invalid interval: must be > 0")
	}

	var records []infoModels.ZPoolHistorical
	err := s.DB.Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}

	intervalMs := int64(intervalMinutes) * 60 * 1000
	buckets := make(map[string]map[int64]zfsServiceInterfaces.PoolStatPoint)

	for _, rec := range records {
		t := rec.CreatedAt
		bucket := t - (t % intervalMs)

		p := zfs.Zpool(rec.Pools)
		poolName := p.Name

		if buckets[poolName] == nil {
			buckets[poolName] = make(map[int64]zfsServiceInterfaces.PoolStatPoint)
		}

		if _, exists := buckets[poolName][bucket]; !exists {
			buckets[poolName][bucket] = zfsServiceInterfaces.PoolStatPoint{
				Time:       bucket,
				Allocated:  p.Allocated,
				Free:       p.Free,
				Size:       p.Size,
				DedupRatio: p.DedupRatio,
			}
		}
	}

	result := make(map[string][]zfsServiceInterfaces.PoolStatPoint)
	for name, points := range buckets {
		var sorted []zfsServiceInterfaces.PoolStatPoint
		for _, point := range points {
			sorted = append(sorted, point)
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Time < sorted[j].Time
		})
		if limit > 0 && len(sorted) > limit {
			sorted = sorted[len(sorted)-limit:]
		}
		result[name] = sorted
	}

	return result, nil
}
