package networkModels

import (
	iface "sylve/pkg/network/iface"
)

type IPv6 struct {
	ID           int     `json:"id" gorm:"primaryKey"`
	InterfaceID  int     `json:"interfaceId" gorm:"index:idx_ipv6_unique,unique"`
	Protocol     string  `json:"protocol" gorm:"not null;index:idx_ipv6_unique,unique"`
	Address      *string `json:"address" gorm:"index:idx_ipv6_unique,unique"`
	Options      *string `json:"options"`
	PrefixLength *int    `json:"prefixLength"`
	IsAlias      bool    `json:"isAlias" gorm:"default:false"`
}

type IPv4 struct {
	ID          int     `json:"id" gorm:"primaryKey"`
	InterfaceID int     `json:"interfaceId" gorm:"index:idx_ipv4_unique,unique"`
	Protocol    string  `json:"protocol" gorm:"not null;index:idx_ipv4_unique,unique"`
	Address     *string `json:"address" gorm:"index:idx_ipv4_unique,unique"`
	Netmask     *string `json:"netmask"`
	Options     *string `json:"options"`
	IsAlias     bool    `json:"isAlias" gorm:"default:false"`
}

type NetworkInterface struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"unique;not null"`
	MAC    string `json:"mac" gorm:"unique;not null"`
	MTU    *int   `json:"mtu" gorm:"default:1500"`
	Metric *int   `json:"metric" gorm:"default:0"`
	IPv4s  []IPv4 `json:"ipv4s" gorm:"foreignKey:InterfaceID;references:ID;constraint:OnDelete:CASCADE"`
	IPv6s  []IPv6 `json:"ipv6s" gorm:"foreignKey:InterfaceID;references:ID;constraint:OnDelete:CASCADE"`

	Interface *iface.Interface `json:"interface" gorm:"-"`
}
