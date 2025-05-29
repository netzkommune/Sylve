package vmModels

import (
	"sylve/internal/db/models"
	"time"
)

type Storage struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Dataset       string `json:"dataset"`
	Size          int64  `json:"size"`
	EmulationType string `json:"emulationType"`

	VMID uint `json:"vmId" gorm:"index"`
}

type Network struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	MAC    string `json:"mac" gorm:"unique"`
	Switch string `json:"switch" gorm:"unique"`

	VMID uint `json:"vmId" gorm:"index"`
}

type VM struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Name          string `json:"name"`
	VmID          string `json:"vmId"`
	CPUSockets    int    `json:"cpuSockets"`
	CPUCores      int    `json:"cpuCores"`
	CPUsThreads   int    `json:"cpuThreads"`
	RAM           int    `json:"ram"`
	VNCPort       int    `json:"vncPort"`
	VNCPassword   string `json:"vncPassword"`
	VNCResolution string `json:"vncResolution"`
	StartAtBoot   bool   `json:"startAtBoot"`
	StartOrder    int    `json:"startOrder"`

	Storages   []Storage                 `json:"storages" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Networks   []Network                 `json:"networks" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PCIDevices []models.PassedThroughIDs `json:"pciDevices" gorm:"foreignKey:VMID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
