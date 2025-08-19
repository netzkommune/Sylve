package utilitiesModels

import "time"

func (WoL) TableName() string {
	return "wols"
}

type WoL struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Mac       string    `json:"mac"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
