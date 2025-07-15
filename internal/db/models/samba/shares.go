package sambaModels

import (
	"sylve/internal/db/models"
	"time"
)

type SambaShare struct {
	ID              int            `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"uniqueIndex"`
	Dataset         string         `json:"dataset" gorm:"uniqueIndex"`
	ReadOnlyGroups  []models.Group `json:"readOnlyGroups" gorm:"many2many:samba_share_read_only_groups;"`
	WriteableGroups []models.Group `json:"writeableGroups" gorm:"many2many:samba_share_writeable_groups;"`
	CreateMask      string         `json:"createMask" gorm:"default:'0664'"`
	DirectoryMask   string         `json:"directoryMask" gorm:"default:'2775'"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
}
