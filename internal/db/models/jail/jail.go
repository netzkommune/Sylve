package jailModels

import (
	networkModels "sylve/internal/db/models/network"
	"time"
)

func (Network) TableName() string {
	return "jail_networks"
}

type Network struct {
	ID uint `gorm:"primaryKey" json:"id"`

	SwitchID uint                         `json:"switchId" gorm:"not null;index"`
	Switch   networkModels.StandardSwitch `gorm:"foreignKey:SwitchID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	MacID         *uint                 `json:"macId" gorm:"column:mac_id"`
	MacAddressObj *networkModels.Object `json:"macObj" gorm:"foreignKey:MacID"`

	IPv4ID    *uint                 `json:"ipv4Id" gorm:"column:ipv4_id"`
	IPv4Obj   *networkModels.Object `json:"ipv4Obj" gorm:"foreignKey:IPv4ID"`
	IPv4GwID  *uint                 `json:"ipv4GwId" gorm:"column:ipv4_gw_id"`
	IPv4GwObj *networkModels.Object `json:"ipv4GwObj" gorm:"foreignKey:IPv4GwID"`

	IPv6ID    *uint                 `json:"ipv6Id" gorm:"column:ipv6_id"`
	IPv6Obj   *networkModels.Object `json:"ipv6Obj" gorm:"foreignKey:IPv6ID"`
	IPv6GwID  *uint                 `json:"ipv6GwId" gorm:"column:ipv6_gw_id"`
	IPv6GwObj *networkModels.Object `json:"ipv6GwObj" gorm:"foreignKey:IPv6GwID"`

	CTID uint `json:"ctId" gorm:"index"`
}

type Jail struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	CTID        int    `json:"ctId" gorm:"unique;not null;uniqueIndex"`
	Name        string `json:"name" gorm:"not null;unique"`
	Description string `json:"description"`
	Dataset     string `json:"dataset"`
	Base        string `json:"base"`
	StartAtBoot bool   `json:"startAtBoot"`
	StartOrder  int    `json:"startOrder"`

	Networks []Network `json:"networks" gorm:"foreignKey:CTID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
