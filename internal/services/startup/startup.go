// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package startup

import (
	"context"
	"os"
	serviceInterfaces "sylve/internal/interfaces/services"
	infoServiceInterfaces "sylve/internal/interfaces/services/info"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"

	"gorm.io/gorm"
)

var _ serviceInterfaces.StartupServiceInterface = (*Service)(nil)

type Service struct {
	DB   *gorm.DB
	Info infoServiceInterfaces.InfoServiceInterface
	ZFS  zfsServiceInterfaces.ZfsServiceInterface
}

func NewStartupService(db *gorm.DB,
	info infoServiceInterfaces.InfoServiceInterface,
	zfs zfsServiceInterfaces.ZfsServiceInterface) serviceInterfaces.StartupServiceInterface {
	return &Service{
		DB:   db,
		Info: info,
		ZFS:  zfs,
	}
}

func (s *Service) InitKeys(authService serviceInterfaces.AuthServiceInterface) error {
	if err := authService.InitSecret("JWTSecret", 6); err != nil {
		return err
	}

	if err := os.MkdirAll("/etc/zfs/keys", os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (s *Service) Initialize(authService serviceInterfaces.AuthServiceInterface) error {
	if err := s.InitKeys(authService); err != nil {
		return err
	}

	go s.Info.Cron()
	go s.ZFS.Cron()
	go s.ZFS.StartSnapshotScheduler(context.Background())

	return nil
}
