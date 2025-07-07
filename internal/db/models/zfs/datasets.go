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
	GUID      string    `gorm:"uniqueIndex:uniq_guid_interval_prefix,priority:1" json:"guid"`
	Interval  int       `gorm:"uniqueIndex:uniq_guid_interval_prefix,priority:2" json:"interval"`
	Prefix    string    `gorm:"uniqueIndex:uniq_guid_interval_prefix,priority:3" json:"prefix"`
	Recursive bool      `json:"recursive"`
	CronExpr  string    `json:"cronExpr"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
	LastRunAt time.Time `json:"lastRunAt,omitempty"`
}
