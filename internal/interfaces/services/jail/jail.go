// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package jailServiceInterfaces

type CreateJailRequest struct {
	Name        string `json:"name" binding:"required"`
	CTID        *int   `json:"ctId" binding:"required"`
	Description string `json:"description"`
	Dataset     string `json:"dataset"`
	Base        string `json:"base"`

	SwitchId *int `json:"switchId"`

	DHCP  *bool `json:"dhcp"`
	SLAAC *bool `json:"slaac"`

	IPv4   *int `json:"ipv4"`
	IPv4Gw *int `json:"ipv4Gw"`

	IPv6   *int `json:"ipv6"`
	IPv6Gw *int `json:"ipv6Gw"`

	MAC *int `json:"mac"`

	StartAtBoot *bool `json:"startAtBoot" binding:"required"`
	StartOrder  int   `json:"startOrder"`
}

type SimpleList struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	CTID int    `json:"ctId"`
}

type JailServiceInterface interface {
	// GetJails() ([]Jail, error)
}
