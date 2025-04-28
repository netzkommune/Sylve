// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package zfsModels

import "time"

type PeriodicSnapshot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GUID      string    `gorm:"uniqueIndex:uniq_dataset_guid" json:"guid"`
	Recursive bool      `json:"recursive"`
	Interval  int       `json:"interval"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
	LastRunAt time.Time `json:"lastRunAt,omitempty"`
}
