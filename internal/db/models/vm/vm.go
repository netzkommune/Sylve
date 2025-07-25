// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package vmModels

import (
	networkModels "sylve/internal/db/models/network"
	"time"
)

type Storage struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `json:"name" gorm:"default:''"`
	Type      string `json:"type"`
	Dataset   string `json:"dataset"`
	Size      int64  `json:"size"`
	Emulation string `json:"emulation"`

	VMID uint `json:"vmId" gorm:"index"`
}

type Network struct {
	ID  uint   `gorm:"primaryKey" json:"id"`
	MAC string `json:"mac" gorm:"unique"`

	SwitchID uint                         `json:"switchId" gorm:"not null;index"`
	Switch   networkModels.StandardSwitch `gorm:"foreignKey:SwitchID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Emulation string `json:"emulation"`

	VMID uint `json:"vmId" gorm:"index"`
}

type VMStats struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	VMID        uint    `json:"vmId" gorm:"index"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	MemoryUsed  float64 `json:"memoryUsed"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

type VM struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	VmID          int    `json:"vmId"`
	CPUSockets    int    `json:"cpuSockets"`
	CPUCores      int    `json:"cpuCores"`
	CPUsThreads   int    `json:"cpuThreads"`
	RAM           int    `json:"ram"`
	VNCPort       int    `json:"vncPort"`
	VNCPassword   string `json:"vncPassword"`
	VNCResolution string `json:"vncResolution"`
	VNCWait       bool   `json:"vncWait"`
	StartAtBoot   bool   `json:"startAtBoot"`
	TPMEmulation  bool   `json:"tpmEmulation"`
	StartOrder    int    `json:"startOrder"`

	ISO        string    `json:"iso"`
	Storages   []Storage `json:"storages" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Networks   []Network `json:"networks" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PCIDevices []int     `json:"pciDevices" gorm:"serializer:json;type:json"`
	CPUPinning []int     `json:"cpuPinning" gorm:"serializer:json;type:json"`

	Stats []VMStats `json:"-" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	StartedAt *time.Time `json:"startedAt" gorm:"default:null"`
	StoppedAt *time.Time `json:"stoppedAt" gorm:"default:null"`
}
