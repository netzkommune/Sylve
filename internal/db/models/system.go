// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package models

type DefaultRoutes struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

type System struct {
	ID            int           `json:"id" gorm:"primaryKey"`
	Initialized   bool          `json:"initialized"`
	Hostname      string        `json:"hostname"`
	DefaultRoutes DefaultRoutes `json:"defaultRoutes" gorm:"embedded"`
	ISODir        string        `json:"isoDir"`
}
