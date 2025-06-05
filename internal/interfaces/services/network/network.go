// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package networkServiceInterfaces

import networkModels "sylve/internal/db/models/network"

type NetworkServiceInterface interface {
	SyncStandardSwitches(previous *networkModels.StandardSwitch, action string) error
	GetStandardSwitches() ([]networkModels.StandardSwitch, error)
	NewStandardSwitch(name string, mtu int, vlan int, address string, address6 string, ports []string, private bool, dhcp bool) error
	EditStandardSwitch(id int, mtu int, vlan int, address string, address6 string, ports []string, private bool) error
	DeleteStandardSwitch(id int) error
}
