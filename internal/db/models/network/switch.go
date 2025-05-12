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
	Private    bool   `json:"private" gorm:"default:false"`

	Ports []NetworkPort `gorm:"foreignKey:SwitchID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

type NetworkPort struct {
	ID       int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string         `json:"name" gorm:"unique;not null"`
	SwitchID int            `json:"switchId" gorm:"not null"`
	Switch   StandardSwitch `gorm:"foreignKey:SwitchID"`
}
