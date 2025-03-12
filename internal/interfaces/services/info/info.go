// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoServiceInterfaces

import infoModels "sylve/internal/db/models/info"

type InfoServiceInterface interface {
	GetBasicInfo() (basicInfo BasicInfo, err error)
	GetCPUInfo(usageOnly bool) (cpuInfo CPUInfo, err error)
	GetRAMInfo() (ramInfo RAMInfo, err error)
	GetSwapInfo() (swapInfo SwapInfo, err error)

	GetNotes() ([]infoModels.Note, error)
	AddNote(title, note string) (infoModels.Note, error)
	DeleteNoteByID(id int) error
	UpdateNoteByID(id int, title, note string) error

	GetAuditLogs(limit int) ([]infoModels.AuditLog, error)

	StoreStats()
	Cron()
}
