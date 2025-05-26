// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package network

import (
	networkServiceInterfaces "sylve/internal/interfaces/services/network"
	"sync"

	"gorm.io/gorm"
)

var _ networkServiceInterfaces.NetworkServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	syncMutex sync.Mutex
}

func NewNetworkService(db *gorm.DB) networkServiceInterfaces.NetworkServiceInterface {
	return &Service{
		DB: db,
	}
}
