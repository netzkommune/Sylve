package infoModels

import "time"

type NetworkInterface struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"index" json:"name"`
	Flags string `json:"flags"`

	Network string `json:"network"`
	Address string `json:"address"`

	ReceivedPackets int64 `gorm:"default:0" json:"receivedPackets"`
	ReceivedErrors  int64 `gorm:"default:0" json:"receivedErrors"`
	DroppedPackets  int64 `gorm:"default:0" json:"droppedPackets"`
	ReceivedBytes   int64 `gorm:"default:0" json:"receivedBytes"`

	SentPackets int64 `gorm:"default:0" json:"sentPackets"`
	SendErrors  int64 `gorm:"default:0" json:"sendErrors"`
	SentBytes   int64 `gorm:"default:0" json:"sentBytes"`

	Collisions int64 `gorm:"default:0" json:"collisions"`

	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
