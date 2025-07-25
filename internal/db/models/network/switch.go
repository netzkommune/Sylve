// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package networkModels

import (
	"time"
)

type StandardSwitch struct {
	ID         int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string `json:"name" gorm:"unique;not null"`
	BridgeName string `gorm:"unique;not null"`
	MTU        int    `json:"mtu" gorm:"default:1500"`
	VLAN       int    `json:"vlan" gorm:"default:0"`
	Address    string `json:"address"`
	Address6   string `json:"address6"`

	AddressID  *uint   `json:"addressId" gorm:"column:address_object_id"`
	AddressObj *Object `json:"addressObj" gorm:"foreignKey:AddressID"`

	Address6ID  *uint   `json:"address6Id" gorm:"column:address6_object_id"`
	Address6Obj *Object `json:"address6Obj" gorm:"foreignKey:Address6ID"`

	DisableIPv6 bool `json:"disableIPv6" gorm:"default:false"`
	Private     bool `json:"private" gorm:"default:false"`

	Ports []NetworkPort `json:"ports" gorm:"foreignKey:SwitchID;constraint:OnDelete:CASCADE"`

	DHCP  bool `json:"dhcp" gorm:"default:false"`
	SLAAC bool `json:"slaac" gorm:"default:false"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type NetworkPort struct {
	ID       int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string         `json:"name" gorm:"unique;not null"`
	SwitchID int            `json:"switchId" gorm:"not null"`
	Switch   StandardSwitch `gorm:"foreignKey:SwitchID"`
}

func (sw *StandardSwitch) IPv4() string {
	if sw.AddressObj != nil && len(sw.AddressObj.Entries) > 0 {
		return sw.AddressObj.Entries[0].Value
	}
	return ""
}

func (sw *StandardSwitch) IPv6() string {
	if sw.Address6Obj != nil && len(sw.Address6Obj.Entries) > 0 {
		return sw.Address6Obj.Entries[0].Value
	}
	return ""
}
