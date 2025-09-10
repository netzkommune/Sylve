// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsServiceInterfaces

import zfsModels "github.com/alchemillahq/sylve/internal/db/models/zfs"

type ZFSS3Backup struct {
	Job  zfsModels.ZFSS3Backup
	Runs []zfsModels.ZFSS3BackupRun
}
