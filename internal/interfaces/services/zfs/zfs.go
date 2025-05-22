// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsServiceInterfaces

import (
	"context"
	infoModels "sylve/internal/db/models/info"
	zfsModels "sylve/internal/db/models/zfs"
)

type ZfsServiceInterface interface {
	GetTotalIODelayHisorical() ([]infoModels.IODelay, error)
	GetZpoolHistoricalStats(intervalMinutes int, limit int) (map[string][]PoolStatPoint, int, error)

	CreatePool(Zpool) error
	DeletePool(poolName string) error

	GetDatasets() ([]Dataset, error)
	BulkDeleteDataset(guids []string) error

	CreateSnapshot(guid string, name string, recursive bool) error
	RollbackSnapshot(guid string, destroyMoreRecent bool) error
	DeleteSnapshot(guid string, recursive bool) error

	GetPeriodicSnapshots() ([]zfsModels.PeriodicSnapshot, error)
	AddPeriodicSnapshot(guid string, prefix string, recursive bool, interval int) error
	DeletePeriodicSnapshot(guid string) error
	StartSnapshotScheduler(ctx context.Context)

	CreateFilesystem(name string, props map[string]string) error
	DeleteFilesystem(guid string) error

	SyncLibvirt() error

	StoreStats(interval int)
	Cron()
}
