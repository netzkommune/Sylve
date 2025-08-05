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
	"fmt"
	"os"
	serviceInterfaces "sylve/internal/interfaces/services"
	infoServiceInterfaces "sylve/internal/interfaces/services/info"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	sambaServiceInterfaces "sylve/internal/interfaces/services/samba"
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	utilitiesServiceInterfaces "sylve/internal/interfaces/services/utilities"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/logger"
	"time"

	"gorm.io/gorm"
)

var _ serviceInterfaces.StartupServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	Info      infoServiceInterfaces.InfoServiceInterface
	ZFS       zfsServiceInterfaces.ZfsServiceInterface
	Network   networkServiceInterfaces.NetworkServiceInterface
	Libvirt   libvirtServiceInterfaces.LibvirtServiceInterface
	Utilities utilitiesServiceInterfaces.UtilitiesServiceInterface
	System    systemServiceInterfaces.SystemServiceInterface
	Samba     sambaServiceInterfaces.SambaServiceInterface
}

func NewStartupService(db *gorm.DB,
	info infoServiceInterfaces.InfoServiceInterface,
	zfs zfsServiceInterfaces.ZfsServiceInterface,
	network networkServiceInterfaces.NetworkServiceInterface,
	libvirt libvirtServiceInterfaces.LibvirtServiceInterface,
	utiliies utilitiesServiceInterfaces.UtilitiesServiceInterface,
	system systemServiceInterfaces.SystemServiceInterface,
	samba sambaServiceInterfaces.SambaServiceInterface,
) serviceInterfaces.StartupServiceInterface {
	return &Service{
		DB:        db,
		Info:      info,
		ZFS:       zfs,
		Network:   network,
		Libvirt:   libvirt,
		Utilities: utiliies,
		System:    system,
		Samba:     samba,
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

func (s *Service) PreFlightChecklist() error {
	if err := s.FreeBSDCheck(); err != nil {
		return err
	}

	if err := s.CheckPackageDependencies(); err != nil {
		return err
	}

	if err := s.CheckServiceDependencies(); err != nil {
		return err
	}

	if err := s.CheckKernelModules(); err != nil {
		return err
	}

	if err := s.CheckSyslogConfig(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Initialize(authService serviceInterfaces.AuthServiceInterface) error {
	if err := s.PreFlightChecklist(); err != nil {
		return fmt.Errorf("Pre-flight check failed: %w", err)
	}

	s.SysctlSync()

	if err := s.InitKeys(authService); err != nil {
		return err
	}

	if err := s.Libvirt.CheckVersion(); err != nil {
		return err
	}

	if err := s.ZFS.SyncLibvirtPools(); err != nil {
		return err
	}

	if err := s.Libvirt.StartTPM(); err != nil {
		return err
	}

	go s.Info.Cron()
	go s.ZFS.Cron()
	go s.ZFS.StartSnapshotScheduler(context.Background())
	go s.Libvirt.StoreVMUsage()

	err := s.Network.SyncStandardSwitches(nil, "sync")
	if err != nil {
		logger.L.Error().Msgf("Error syncing standard switches: %v", err)
	}

	if err := s.System.SyncPPTDevices(); err != nil {
		return fmt.Errorf("failed to sync passthrough devices: %w", err)
	}

	if err := s.InitSamba(); err != nil {
		return fmt.Errorf("failed to initialize Samba: %w", err)
	}

	if err := s.InitSambaAdmins(); err != nil {
		return fmt.Errorf("failed to initialize Samba admins: %w", err)
	}

	go func() {
		for {
			err := s.Utilities.SyncDownloadProgress()
			if err != nil {
				logger.L.Fatal().Msgf("Failed to sync progress for downloads: %v", err)
			}

			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			if err := s.Libvirt.StoreVMUsage(); err != nil {
				logger.L.Error().Msgf("Failed to store VM usage: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			if err := s.Samba.ParseAuditLogs(); err != nil {
				logger.L.Error().Msgf("Failed to parse Samba audit logs: %v", err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}
