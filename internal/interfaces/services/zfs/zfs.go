// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsServiceInterfaces

import infoModels "sylve/internal/db/models/info"

type ZfsServiceInterface interface {
	GetPoolNames() ([]string, error)
	GetPool(name string) (Zpool, error)
	GetPools() ([]Zpool, error)
	GetPoolIODelay(poolName string) float64
	GetTotalIODelay() float64
	GetTotalIODelayHisorical() ([]infoModels.IODelay, error)

	Cron()
}
