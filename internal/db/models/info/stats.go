// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoModels

import "time"

type CPU struct {
	ID        uint      `gorm:"primarykey" json:"id,omitempty"`
	Usage     float64   `json:"usage"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
}

type IODelay struct {
	ID        uint      `gorm:"primarykey" json:"id,omitempty"`
	Delay     float64   `json:"delay"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt,omitempty"`
}
