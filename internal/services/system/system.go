// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package system

import (
	"sync"

	systemServiceInterfaces "github.com/alchemillahq/sylve/internal/interfaces/services/system"

	"gorm.io/gorm"
)

var _ systemServiceInterfaces.SystemServiceInterface = (*Service)(nil)

type Service struct {
	DB        *gorm.DB
	syncMutex sync.Mutex
	achMutex  sync.Mutex
}

func NewSystemService(db *gorm.DB) systemServiceInterfaces.SystemServiceInterface {
	return &Service{
		DB: db,
	}
}
