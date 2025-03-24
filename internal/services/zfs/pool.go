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
)

func (s *Service) GetTotalIODelayHisorical() ([]infoModels.IODelay, error) {
	historicalData, err := db.GetHistorical[infoModels.IODelay](s.DB, 128)

	if err != nil {
		return nil, err
	}

	return historicalData, nil
}

func (s *Service) CreatePool(zfsServiceInterfaces.Zpool) error {
	return nil
}
