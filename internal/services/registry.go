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
	zfsServiceInterfaces "sylve/internal/interfaces/services/zfs"
	"sylve/internal/services/auth"
	"sylve/internal/services/disk"
	"sylve/internal/services/info"
	"sylve/internal/services/libvirt"
	"sylve/internal/services/network"
	"sylve/internal/services/startup"
	"sylve/internal/services/zfs"

	"gorm.io/gorm"
)

type ServiceRegistry struct {
	AuthService    serviceInterfaces.AuthServiceInterface
	StartupService serviceInterfaces.StartupServiceInterface
	InfoService    infoServiceInterfaces.InfoServiceInterface
	ZfsService     zfsServiceInterfaces.ZfsServiceInterface
	DiskService    diskServiceInterfaces.DiskServiceInterface
	NetworkService networkServiceInterfaces.NetworkServiceInterface
	LibvirtService libvirtServiceInterfaces.LibvirtServiceInterface
}

func NewService[T any](db *gorm.DB, dependencies ...interface{}) interface{} {
	switch any(new(T)).(type) {
	case *auth.Service:
		return auth.NewAuthService(db)
	case *startup.Service:
		infoService := dependencies[0].(infoServiceInterfaces.InfoServiceInterface)
		zfsService := dependencies[1].(zfsServiceInterfaces.ZfsServiceInterface)
		networkService := dependencies[2].(networkServiceInterfaces.NetworkServiceInterface)
		libvirtService := dependencies[3].(libvirtServiceInterfaces.LibvirtServiceInterface)
		return startup.NewStartupService(db, infoService, zfsService, networkService, libvirtService)
	case *info.Service:
		return info.NewInfoService(db)
	case *zfs.Service:
		return zfs.NewZfsService(db, dependencies[0].(libvirtServiceInterfaces.LibvirtServiceInterface))
	case *disk.Service:
		return disk.NewDiskService(db, dependencies[0].(zfsServiceInterfaces.ZfsServiceInterface))
	case *network.Service:
		return network.NewNetworkService(db)
	case *libvirt.Service:
		return libvirt.NewLibvirtService(db)
	default:
		return nil
	}
}

func NewServiceRegistry(db *gorm.DB) *ServiceRegistry {
	authService := NewService[auth.Service](db)
	infoService := NewService[info.Service](db)
	networkService := NewService[network.Service](db)
	libvirtService := NewService[libvirt.Service](db)
	zfsService := NewService[zfs.Service](db, libvirtService)

	return &ServiceRegistry{
		AuthService:    authService.(serviceInterfaces.AuthServiceInterface),
		StartupService: NewService[startup.Service](db, infoService, zfsService, networkService, libvirtService).(*startup.Service),
		InfoService:    infoService.(infoServiceInterfaces.InfoServiceInterface),
		ZfsService:     zfsService.(*zfs.Service),
		DiskService:    NewService[disk.Service](db, zfsService).(*disk.Service),
		NetworkService: networkService.(networkServiceInterfaces.NetworkServiceInterface),
		LibvirtService: libvirtService.(libvirtServiceInterfaces.LibvirtServiceInterface),
	}
}
