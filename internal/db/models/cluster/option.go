package clusterModels

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClusterOption struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	KeyboardLayout string    `json:"keyboardLayout"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func upsertOption(db *gorm.DB, o *ClusterOption) error {
	if o.ID == 0 {
		o.ID = 1
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"keyboard_layout", "created_at", "updated_at"}),
	}).Create(o).Error
}
