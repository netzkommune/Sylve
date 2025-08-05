// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package services

import (
	serviceInterfaces "sylve/internal/interfaces/services"
	diskServiceInterfaces "sylve/internal/interfaces/services/disk"
	infoServiceInterfaces "sylve/internal/interfaces/services/info"
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	sambaServiceInterfaces "sylve/internal/interfaces/services/samba"
	systemServiceInterfaces "sylve/internal/interfaces/services/system"
	utilitiesServiceInterfaces "sylve/internal/interfaces/services/utilities"
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/auth"
	"sylve/internal/services/disk"
	"sylve/internal/services/info"
	"sylve/internal/services/libvirt"
	"sylve/internal/services/network"
	"sylve/internal/services/samba"
	"sylve/internal/services/startup"
	"sylve/internal/services/system"
	"sylve/internal/services/utilities"
	"sylve/internal/services/zfs"

	"gorm.io/gorm"
)

type ServiceRegistry struct {
	AuthService      serviceInterfaces.AuthServiceInterface
	StartupService   serviceInterfaces.StartupServiceInterface
	InfoService      infoServiceInterfaces.InfoServiceInterface
	ZfsService       zfsServiceInterfaces.ZfsServiceInterface
	DiskService      diskServiceInterfaces.DiskServiceInterface
	NetworkService   networkServiceInterfaces.NetworkServiceInterface
	LibvirtService   libvirtServiceInterfaces.LibvirtServiceInterface
	UtilitiesService utilitiesServiceInterfaces.UtilitiesServiceInterface
	SystemService    systemServiceInterfaces.SystemServiceInterface
	SambaService     sambaServiceInterfaces.SambaServiceInterface
}

func NewService[T any](db *gorm.DB, dependencies ...interface{}) interface{} {
	switch any(new(T)).(type) {
	case *auth.Service:
		return auth.NewAuthService(db)
	case *system.Service:
		return system.NewSystemService(db)
	case *startup.Service:
		infoService := dependencies[0].(infoServiceInterfaces.InfoServiceInterface)
		zfsService := dependencies[1].(zfsServiceInterfaces.ZfsServiceInterface)
		networkService := dependencies[2].(networkServiceInterfaces.NetworkServiceInterface)
		libvirtService := dependencies[3].(libvirtServiceInterfaces.LibvirtServiceInterface)
		utilitiesService := dependencies[4].(utilitiesServiceInterfaces.UtilitiesServiceInterface)
		systemService := dependencies[5].(systemServiceInterfaces.SystemServiceInterface)
		sambaService := dependencies[6].(sambaServiceInterfaces.SambaServiceInterface)

		return startup.NewStartupService(db, infoService, zfsService, networkService, libvirtService, utilitiesService, systemService, sambaService)
	case *info.Service:
		return info.NewInfoService(db)
	case *zfs.Service:
		return zfs.NewZfsService(db, dependencies[0].(libvirtServiceInterfaces.LibvirtServiceInterface))
	case *disk.Service:
		return disk.NewDiskService(db, dependencies[0].(zfsServiceInterfaces.ZfsServiceInterface))
	case *network.Service:
		return network.NewNetworkService(db, dependencies[0].(libvirtServiceInterfaces.LibvirtServiceInterface))
	case *libvirt.Service:
		return libvirt.NewLibvirtService(db)
	case *utilities.Service:
		return utilities.NewUtilitiesService(db)
	case *samba.Service:
		zfsService := dependencies[0].(zfsServiceInterfaces.ZfsServiceInterface)

		return samba.NewSambaService(db, zfsService)
	default:
		return nil
	}
}

func NewServiceRegistry(db *gorm.DB) *ServiceRegistry {
	authService := NewService[auth.Service](db)
	infoService := NewService[info.Service](db)
	libvirtService := NewService[libvirt.Service](db)
	zfsService := NewService[zfs.Service](db, libvirtService)
	utilitiesService := NewService[utilities.Service](db)
	systemService := NewService[system.Service](db)
	sambaService := NewService[samba.Service](db, zfsService)
	networkService := NewService[network.Service](db, libvirtService)

	return &ServiceRegistry{
		AuthService:      authService.(serviceInterfaces.AuthServiceInterface),
		StartupService:   NewService[startup.Service](db, infoService, zfsService, networkService, libvirtService, utilitiesService, systemService, sambaService).(*startup.Service),
		InfoService:      infoService.(infoServiceInterfaces.InfoServiceInterface),
		ZfsService:       zfsService.(*zfs.Service),
		DiskService:      NewService[disk.Service](db, zfsService).(*disk.Service),
		NetworkService:   NewService[network.Service](db, libvirtService).(*network.Service),
		LibvirtService:   libvirtService.(libvirtServiceInterfaces.LibvirtServiceInterface),
		UtilitiesService: utilitiesService.(utilitiesServiceInterfaces.UtilitiesServiceInterface),
		SystemService:    systemService.(systemServiceInterfaces.SystemServiceInterface),
		SambaService:     sambaService.(sambaServiceInterfaces.SambaServiceInterface),
	}
}
