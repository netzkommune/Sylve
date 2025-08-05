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
	GuestOk         bool           `json:"guestOk" gorm:"default:false"`
	ReadOnly        bool           `json:"readOnly" gorm:"default:false"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
}

type SambaAuditLog struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Share     string    `json:"share"`
	User      string    `json:"user"`
	IP        string    `json:"ip"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Path      string    `json:"path"`
	Target    string    `json:"target"`
	Folder    string    `json:"folder"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}
