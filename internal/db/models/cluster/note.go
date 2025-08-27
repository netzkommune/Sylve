package clusterModels

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClusterNote struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func upsertNote(db *gorm.DB, n *ClusterNote) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "content", "created_at", "updated_at"}),
	}).Create(n).Error
}
