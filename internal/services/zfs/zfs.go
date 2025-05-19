// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfs

import (
	"sylve/internal/db"
	infoModels "sylve/internal/db/models/info"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/logger"
	"sylve/pkg/zfs"
	"sync"
	"time"

	"gorm.io/gorm"
)

var _ zfsServiceInterfaces.ZfsServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	Libvirt   libvirtServiceInterfaces.LibvirtServiceInterface
	syncMutex *sync.Mutex
}

func NewZfsService(db *gorm.DB, libvirt libvirtServiceInterfaces.LibvirtServiceInterface) zfsServiceInterfaces.ZfsServiceInterface {
	return &Service{
		DB:        db,
		Libvirt:   libvirt,
		syncMutex: &sync.Mutex{},
	}
}

func (s *Service) StoreStats() {
	d := zfs.GetTotalIODelay()
	db.StoreAndTrimRecords(s.DB, &infoModels.IODelay{Delay: d}, 128)
}

func (s *Service) Cron() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	s.StoreStats()

	for range ticker.C {
		s.StoreStats()
	}
}

func (s *Service) SyncLibvirt() error {
	zfsPools, err := zfs.ListZpools()

	if err != nil {
		return err
	}

	lvPools, err := s.Libvirt.ListStoragePools()

	if err != nil {
		return err
	}

	for _, pool := range zfsPools {
		exists := false
		for _, lvPool := range lvPools {
			if pool.Name == lvPool.Source {
				exists = true
				break
			}
		}

		if !exists {
			err := s.Libvirt.CreateStoragePool(pool.Name)
			if err != nil {
				logger.L.Error().Err(err).Msgf("Failed to create storage pool %s in libvirt", pool.Name)
				return err
			}
		}
	}

	return nil
}
