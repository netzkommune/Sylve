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
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/pkg/zfs"
	"time"

	"gorm.io/gorm"
)

var _ zfsServiceInterfaces.ZfsServiceInterface = (*Service)(nil)

type Service struct {
	DB *gorm.DB
}

func NewZfsService(db *gorm.DB) zfsServiceInterfaces.ZfsServiceInterface {
	return &Service{
		DB: db,
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
