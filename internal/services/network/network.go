// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package network

import (
	libvirtServiceInterfaces "sylve/internal/interfaces/services/libvirt"
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	"sync"

	"gorm.io/gorm"
)

var _ networkServiceInterfaces.NetworkServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	syncMutex sync.Mutex

	LibVirt libvirtServiceInterfaces.LibvirtServiceInterface
}

func NewNetworkService(db *gorm.DB, libvirt libvirtServiceInterfaces.LibvirtServiceInterface) networkServiceInterfaces.NetworkServiceInterface {
	return &Service{
		DB:      db,
		LibVirt: libvirt,
	}
}
