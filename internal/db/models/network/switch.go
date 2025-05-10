package networkModels

import (
	"time"
)

type StandardSwitch struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"unique;not null"`
	MTU       int       `json:"mtu" gorm:"default:1500"`
	VLAN      int       `json:"vlan" gorm:"default:0"`
	Address   string    `json:"address"`
	Private   bool      `json:"private" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	Ports []NetworkPort `json:"ports" gorm:"foreignKey:SwitchID;constraint:OnDelete:CASCADE"`
}

type NetworkPort struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"unique;not null"`
	SwitchID int    `json:"switchId"`
}
