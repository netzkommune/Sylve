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
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	utilitiesServiceInterfaces "sylve/internal/interfaces/services/utilities"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/logger"
	"sync"
	"time"

	"sylve/pkg/pkg"
	sysctl "sylve/pkg/utils/sysctl"

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
}

func NewStartupService(db *gorm.DB,
	info infoServiceInterfaces.InfoServiceInterface,
	zfs zfsServiceInterfaces.ZfsServiceInterface,
	network networkServiceInterfaces.NetworkServiceInterface,
	libvirt libvirtServiceInterfaces.LibvirtServiceInterface,
	utiliies utilitiesServiceInterfaces.UtilitiesServiceInterface,
	system systemServiceInterfaces.SystemServiceInterface,
) serviceInterfaces.StartupServiceInterface {
	return &Service{
		DB:        db,
		Info:      info,
		ZFS:       zfs,
		Network:   network,
		Libvirt:   libvirt,
		Utilities: utiliies,
		System:    system,
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

func (s *Service) SysctlSync() error {
	intVals := map[string]int32{
		"net.inet.ip.forwarding":      1,
		"net.link.bridge.inherit_mac": 1,
	}

	for k, v := range intVals {
		_, err := sysctl.GetInt64(k)
		if err != nil {
			logger.L.Error().Msgf("Error getting sysctl %s: %v, skipping!", k, err)
			continue
		}

		err = sysctl.SetInt32(k, v)
		if err != nil {
			logger.L.Error().Msgf("Error setting sysctl %s: %v", k, err)
		}
	}

	return nil
}

func (s *Service) InitFirewall() error {
	// if len(config.ParsedConfig.WANInterfaces) == 0 {
	// 	return fmt.Errorf("no WAN interfaces found in config")
	// }

	return nil
}

func (s *Service) CheckPackageDepdencies() error {
	requiredPackages := []string{
		"libvirt",
		"bhyve-firmware",
		"smartmontools",
		"tmux",
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(requiredPackages))

	for _, p := range requiredPackages {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !pkg.IsPackageInstalled(p) {
				errCh <- fmt.Errorf("Required package %s is not installed", p)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) Initialize(authService serviceInterfaces.AuthServiceInterface) error {
	if err := s.CheckPackageDepdencies(); err != nil {
		return err
	}

	if err := s.InitKeys(authService); err != nil {
		return err
	}

	if err := s.Libvirt.CheckVersion(); err != nil {
		return err
	}

	if err := s.ZFS.SyncLibvirtPools(); err != nil {
		return err
	}

	if err := s.InitFirewall(); err != nil {
		return err
	}

	go s.Info.Cron()
	go s.ZFS.Cron()
	go s.ZFS.StartSnapshotScheduler(context.Background())
	go s.Libvirt.StoreVMUsage()

	if err := s.SysctlSync(); err != nil {
		return err
	}

	err := s.Network.SyncStandardSwitches(nil, "sync")
	if err != nil {
		logger.L.Error().Msgf("Error syncing standard switches: %v", err)
	}

	if err := s.System.SyncPPTDevices(); err != nil {
		return fmt.Errorf("failed to sync passthrough devices: %w", err)
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

	return nil
}
